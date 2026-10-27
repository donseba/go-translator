package translator

import (
	"bufio"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/donseba/go-translator/v2/extract"

	"github.com/leonelquinteros/gotext"
)

var (
	// ErrorLanguageNotFound reports a language that is not loaded, or a
	// catalogue file that cannot be read.
	ErrorLanguageNotFound = fmt.Errorf("language not found")
	// ErrorLanguageAlreadyExists reports a NewLanguage for a loaded language.
	ErrorLanguageAlreadyExists = fmt.Errorf("language already exists")
	// ErrorInvalidCatalogue reports a catalogue that does not fit the loaded
	// language, see MergeDefaults and LoadLanguages.
	ErrorInvalidCatalogue = fmt.Errorf("invalid translation catalogue")
)

var (
	// TemplateExtension is the extension of the template files that ScanFiles
	// and CheckMissingTranslations scan.
	TemplateExtension = ".gohtml"
	// DefaultPotFile is the name of the POT file in the translations directory.
	DefaultPotFile = "translations.pot"
	// DefaultPoExtension is the extension of the catalogue files.
	DefaultPoExtension = ".po"
	// DefaultNoTranslationTL is the marker of a missing translation for Tl.
	DefaultNoTranslationTL = "*%s*"
	// DefaultNoTranslationTN is the marker of a missing translation for Tn.
	DefaultNoTranslationTN = "*%s/%s*"
	// DefaultNoTranslationCTL is the marker of a missing translation for Ctl.
	DefaultNoTranslationCTL = "*%s/%s*"
	// DefaultNoTranslationCTN is the marker of a missing translation for Ctn.
	DefaultNoTranslationCTN = "*%s/%s/%s*"
	// DefaultSeparator is the prefix separator of a new translator, see
	// SetPrefixSeparator.
	DefaultSeparator = "__."
)

type (
	// Localizer interface contains the methods that are needed for the translator
	Localizer interface {
		// GetLocale returns the locale of the localizer, ie. "en_US"
		GetLocale() string
	}

	// Translator looks up the translations of texts in the catalogues of its
	// languages and records the keys it has not seen in a POT file. Make one
	// with New, NewTranslator or Untranslated.
	Translator struct {
		catalogueMu     sync.RWMutex
		languages       map[string]*gotext.Po
		translationsDir string
		templateDir     string
		prefixSeparator string
		uniqueKeys      map[string]uniqueKey
		uniqueKeysCtx   map[string]map[string]uniqueKey
		potFile         string
		pot             *gotext.Po // nil until loadPot; guarded by catalogueMu
		recordMissing   atomic.Bool
		untranslated    func(Miss) string
		fallback        bool
	}

	uniqueKey struct{ singular, plural string }
)

// NewTranslator returns a translator for the catalogues in translationsDir and
// the templates below templateDir. Unlike New, it records every key a lookup
// has not seen in the POT file, which it reads the first time a key is
// recorded or extracted (see SetRecordMissing).
func NewTranslator(translationsDir, templateDir string) *Translator {
	tr := &Translator{
		translationsDir: translationsDir,
		templateDir:     templateDir,
		potFile:         DefaultPotFile,
		prefixSeparator: DefaultSeparator,
		languages:       make(map[string]*gotext.Po),
		uniqueKeys:      make(map[string]uniqueKey),
		uniqueKeysCtx:   make(map[string]map[string]uniqueKey),
	}

	tr.recordMissing.Store(true)

	return tr
}

// SetPrefixSeparator sets the separator that ends a key prefix,
// DefaultSeparator by default. A translated text loses everything up to and
// including the last separator.
func (t *Translator) SetPrefixSeparator(prefix string) {
	t.prefixSeparator = prefix
}

// PrefixSeparator returns the separator that ends a key prefix.
func (t *Translator) PrefixSeparator() string {
	return t.prefixSeparator
}

// SetPotFile sets the name of the POT file in the translations directory,
// DefaultPotFile by default. The file is read the next time a key is
// recorded or extracted.
func (t *Translator) SetPotFile(potFile string) {
	t.catalogueMu.Lock()
	defer t.catalogueMu.Unlock()

	t.potFile = potFile
	t.pot = nil
}

// PotFile returns the name of the POT file in the translations directory.
func (t *Translator) PotFile() string {
	return t.potFile
}

// SetRecordMissing turns runtime key recording on or off. Recording is on by
// default for NewTranslator and off for New and Untranslated: every lookup of
// a key that is not yet in the POT file appends it. An empty key is never
// recorded.
// Turn it off in production so serving requests never writes to disk;
// CheckMissingTranslations and ScanFiles keep working because they are
// explicit extraction steps.
func (t *Translator) SetRecordMissing(record bool) {
	t.recordMissing.Store(record)
}

// RecordMissing reports whether runtime lookups record unseen keys in the POT file.
func (t *Translator) RecordMissing() bool {
	return t.recordMissing.Load()
}

// SetTranslationsDir sets the directory of the catalogues and the POT file.
// The POT file is read the next time a key is recorded or extracted.
func (t *Translator) SetTranslationsDir(translationsDir string) {
	t.catalogueMu.Lock()
	defer t.catalogueMu.Unlock()

	t.translationsDir = translationsDir
	t.pot = nil
}

// TranslationsDir returns the directory of the catalogues and the POT file.
func (t *Translator) TranslationsDir() string {
	return t.translationsDir
}

// SetTemplateDir sets the directory that CheckMissingTranslations scans.
func (t *Translator) SetTemplateDir(templateDir string) {
	t.templateDir = templateDir
}

// TemplateDir returns the directory that CheckMissingTranslations scans. New
// leaves it empty.
func (t *Translator) TemplateDir() string {
	return t.templateDir
}

// SetLanguage loads the catalogue of lang from <translations dir>/<lang>.po.
// A catalogue already loaded under lang is replaced.
//
// When the file cannot be read, the error matches both ErrorLanguageNotFound
// and the cause, such as fs.ErrNotExist, with errors.Is. A catalogue without
// a Language header is loaded as lang. See SetLanguageFS and LoadLanguages
// to load catalogues from an fs.FS.
func (t *Translator) SetLanguage(lang string) error {
	data, err := os.ReadFile(filepath.Join(t.translationsDir, lang+DefaultPoExtension))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrorLanguageNotFound, err)
	}

	t.setCatalogue(lang, data)

	return nil
}

// AddLanguage adds a new language to the translator by loading its .po file.
//
// Deprecated: use SetLanguage instead.
func (t *Translator) AddLanguage(lang string) error {
	return t.SetLanguage(lang)
}

// EnsureLanguage creates <lang>.po in the translations directory, with the
// header of GetHeaderForLanguage, when it does not exist, and then loads it
// with SetLanguage. The translations directory must exist.
func (t *Translator) EnsureLanguage(lang string) error {
	poPath := filepath.Join(t.translationsDir, lang+DefaultPoExtension)
	if _, err := os.Stat(poPath); os.IsNotExist(err) {
		h := GetHeaderForLanguage(lang)
		f, err := os.Create(poPath)
		if err != nil {
			return fmt.Errorf("failed to create new language file: %w", err)
		}
		defer func() { _ = f.Close() }()
		_, err = f.WriteString(h.HeaderString())
		if err != nil {
			return fmt.Errorf("failed to write header to new language file: %w", err)
		}
	}

	return t.SetLanguage(lang)
}

// CheckMissingTranslations scans the template files below the template
// directory (see ScanFiles) and appends the keys that are not in the POT file
// to it, in the order they first appear in the templates, followed by keys
// recorded at runtime in sorted order. It returns the errors of every key it
// could not add, joined. To write a complete POT file with references
// instead, use the extract package.
func (t *Translator) CheckMissingTranslations() error {
	t.catalogueMu.Lock()
	defer t.catalogueMu.Unlock()

	catalog, err := t.scan(t.templateDir)
	if err != nil {
		return err
	}

	var (
		pot     = t.loadPot()
		written = map[potKey]bool{}
		errs    []error
	)

	for _, entry := range pot.GetDomain().GetTranslations() {
		written[potKey{id: entry.ID}] = true
	}

	for ctx, entries := range pot.GetDomain().GetCtxTranslations() {
		for _, entry := range entries {
			written[potKey{ctx: ctx, id: entry.ID}] = true
		}
	}

	add := func(ctx string, entry uniqueKey) {
		id := potKey{ctx: ctx, id: entry.singular}
		if written[id] {
			return
		}

		written[id] = true

		if err := t.addToPotFile(ctx, entry); err != nil {
			errs = append(errs, err)
		}
	}

	for _, message := range catalog.Messages {
		add(message.Context, t.recordedKey(message.Context, message.ID))
	}

	for _, key := range sortedKeys(t.uniqueKeys) {
		add("", t.uniqueKeys[key])
	}

	for _, ctx := range sortedKeys(t.uniqueKeysCtx) {
		for _, key := range sortedKeys(t.uniqueKeysCtx[ctx]) {
			add(ctx, t.uniqueKeysCtx[ctx][key])
		}
	}

	return errors.Join(errs...)
}

// potKey identifies an entry of the POT file: its context and key.
type potKey struct {
	ctx string
	id  string
}

// sortedKeys returns the keys of m in sorted order, an empty slice for an
// empty map.
func sortedKeys[V any](m map[string]V) []string {
	keys := slices.AppendSeq(make([]string, 0, len(m)), maps.Keys(m))
	slices.Sort(keys)

	return keys
}

// ScanFiles extracts the keys of the TemplateExtension files below root with
// the extract package and remembers them, so CheckMissingTranslations can add
// them to the POT file. A template that does not parse is an error naming the
// file.
func (t *Translator) ScanFiles(root string) error {
	t.catalogueMu.Lock()
	defer t.catalogueMu.Unlock()

	_, err := t.scan(root)

	return err
}

// scan extracts the keys of the template files below root and remembers
// them. It must be called while catalogueMu is held for writing.
func (t *Translator) scan(root string) (*extract.Catalog, error) {
	var sources []extract.Source

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, TemplateExtension) {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		sources = append(sources, extract.Source{Name: path, Text: string(data)})

		return nil
	})
	if err != nil {
		return nil, err
	}

	catalog, err := extract.Extract(sources...)
	if err != nil {
		return nil, err
	}

	for _, message := range catalog.Messages {
		t.remember(message.Context, uniqueKey{singular: message.ID, plural: message.Plural})
	}

	return catalog, nil
}

// recordedKey returns the recorded key, with its plural; a key without
// plural when it is not recorded. It must be called while catalogueMu is
// held.
func (t *Translator) recordedKey(ctx, key string) uniqueKey {
	recorded := t.uniqueKeys
	if ctx != "" {
		recorded = t.uniqueKeysCtx[ctx]
	}

	entry, exists := recorded[key]
	if !exists {
		return uniqueKey{singular: key}
	}

	return entry
}

// recordKey serializes discovery and POT updates for concurrent lookups.
// An empty text is not recorded: msgid "" is the header of a POT file.
func (t *Translator) recordKey(ctx string, entry uniqueKey) {
	if !t.recordMissing.Load() || entry.singular == "" {
		return
	}

	t.catalogueMu.RLock()
	if t.hasRecordedKey(ctx, entry.singular) {
		t.catalogueMu.RUnlock()
		return
	}
	t.catalogueMu.RUnlock()

	// A lookup has no error to return, so a recording error is logged.
	if err := t.storeKey(ctx, entry); err != nil {
		err = fmt.Errorf("go-translator: record %q: %w", entry.singular, err)
		slog.Error("record missing translation", "error", err)
	}
}

// storeKey appends entry to the POT file, unless it is there already, and
// remembers it, under catalogueMu.
func (t *Translator) storeKey(ctx string, entry uniqueKey) error {
	t.catalogueMu.Lock()
	defer t.catalogueMu.Unlock()

	if t.hasRecordedKey(ctx, entry.singular) {
		return nil
	}

	if err := t.addToPotFileIfNotExists(ctx, entry); err != nil {
		return err
	}

	t.remember(ctx, entry)

	return nil
}

// remember adds a key to the recorded keys. A key without plural keeps the
// plural recorded earlier. It must be called while catalogueMu is held for
// writing.
func (t *Translator) remember(ctx string, entry uniqueKey) {
	if entry.plural == "" {
		entry.plural = t.recordedKey(ctx, entry.singular).plural
	}

	if ctx == "" {
		t.uniqueKeys[entry.singular] = entry

		return
	}

	if t.uniqueKeysCtx[ctx] == nil {
		t.uniqueKeysCtx[ctx] = make(map[string]uniqueKey)
	}

	t.uniqueKeysCtx[ctx][entry.singular] = entry
}

// hasRecordedKey requires catalogueMu to be held for reading or writing.
func (t *Translator) hasRecordedKey(ctx, key string) bool {
	if ctx == "" {
		_, exists := t.uniqueKeys[key]
		return exists
	}
	_, exists := t.uniqueKeysCtx[ctx][key]
	return exists
}

// ensurePotFile creates the POT file with a template header when it does not
// exist yet, so key discovery works on a fresh project. Like EnsureLanguage, it
// expects the translations directory to exist. It must be called while
// catalogueMu is held.
func (t *Translator) ensurePotFile() error {
	potPath := filepath.Join(t.translationsDir, t.potFile)

	_, err := os.Stat(potPath)
	if err == nil {
		return nil
	}

	if !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	if err := WritePOTFile(potPath); err != nil {
		return fmt.Errorf("create POT file: %w", err)
	}

	return nil
}

func (t *Translator) addToPotFile(ctx string, entry uniqueKey) error {
	if err := t.ensurePotFile(); err != nil {
		return err
	}

	file, err := os.OpenFile(filepath.Join(t.translationsDir, t.potFile), os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	buf := bufio.NewWriter(file)

	// Write the new translation entry
	var content string
	if ctx == "" {
		if entry.plural == "" {
			content = fmt.Sprintf("\nmsgid \"%s\"\nmsgstr \"\"\n", gotext.EscapeSpecialCharacters(entry.singular))
		} else {
			content = fmt.Sprintf("\nmsgid \"%s\"\nmsgid_plural \"%s\"\nmsgstr[0] \"\"\nmsgstr[1] \"\"\n", gotext.EscapeSpecialCharacters(entry.singular), gotext.EscapeSpecialCharacters(entry.plural))
		}
	} else {
		if entry.plural == "" {
			content = fmt.Sprintf("\nmsgctxt \"%s\"\nmsgid \"%s\"\nmsgstr \"\"\n", gotext.EscapeSpecialCharacters(ctx), gotext.EscapeSpecialCharacters(entry.singular))
		} else {
			content = fmt.Sprintf("\nmsgctxt \"%s\"\nmsgid \"%s\"\nmsgid_plural \"%s\"\nmsgstr[0] \"\"\nmsgstr[1] \"\"\n", gotext.EscapeSpecialCharacters(ctx), gotext.EscapeSpecialCharacters(entry.singular), gotext.EscapeSpecialCharacters(entry.plural))
		}
	}

	if _, err = buf.WriteString(content); err != nil {
		return err
	}

	if err = buf.Flush(); err != nil {
		return err
	}

	// The next key to record reads the POT file again, with this entry.
	t.pot = nil

	return nil
}

// loadPot returns the keys of the POT file, reading it the first time they
// are needed: to record a key or to extract keys from templates. A
// translator that does neither, such as one with recording off, never reads
// the file. A POT file that does not exist or cannot be read counts as
// empty. It must be called while catalogueMu is held for writing.
func (t *Translator) loadPot() *gotext.Po {
	if t.pot != nil {
		return t.pot
	}

	t.pot = gotext.NewPo()
	t.pot.ParseFile(filepath.Join(t.translationsDir, t.potFile))

	return t.pot
}

// addToPotFileIfNotExists must be called while catalogueMu is held.
func (t *Translator) addToPotFileIfNotExists(ctx string, entry uniqueKey) error {
	pot := t.loadPot()

	if ctx == "" {
		for _, existing := range pot.GetDomain().GetTranslations() {
			if existing.ID == entry.singular {
				return nil
			}
		}
	} else {
		for _, existing := range pot.GetDomain().GetCtxTranslations()[ctx] {
			if existing.ID == entry.singular {
				return nil
			}
		}
	}
	return t.addToPotFile(ctx, entry)
}

// FuncMap returns the template functions tl, tn, ctl and ctn. Each takes the
// localizer first, like Tl, Tn, Ctl and Ctn.
func (t *Translator) FuncMap() template.FuncMap {
	return template.FuncMap{
		"tl":  t.tl,
		"tn":  t.tn,
		"ctl": t.ctl,
		"ctn": t.ctn,
	}
}

// removePrefix removes any prefix ending with prefix separator from the translated string.
func (t *Translator) removePrefix(s string) string {
	return removeSeparator(s, t.PrefixSeparator())
}

// removeSeparator removes everything up to and including the last separator
// in s.
func removeSeparator(s, separator string) string {
	idx := strings.LastIndex(s, separator)
	if idx != -1 {
		return s[idx+len(separator):]
	}

	return s
}
