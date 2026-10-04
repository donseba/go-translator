package translator

import (
	"bufio"
	"fmt"
	"html/template"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"text/template/parse"

	"github.com/leonelquinteros/gotext"
)

var (
	ErrorLanguageNotFound      = fmt.Errorf("language not found")
	ErrorLanguageAlreadyExists = fmt.Errorf("language already exists")
	ErrorInvalidCatalogue      = fmt.Errorf("invalid translation catalogue")
)

var (
	TemplateExtension       = ".gohtml"
	DefaultPotFile          = "translations.pot"
	DefaultPoExtension      = ".po"
	DefaultNoTranslationTL  = "*%s*"
	DefaultNoTranslationTN  = "*%s/%s*"
	DefaultNoTranslationCTL = "*%s/%s*"
	DefaultNoTranslationCTN = "*%s/%s/%s*"
	DefaultSeparator        = "__."
)

type (
	// Localizer interface contains the methods that are needed for the translator
	Localizer interface {
		// GetLocale returns the locale of the localizer, ie. "en_US"
		GetLocale() string
	}

	Translator struct {
		catalogueMu     sync.RWMutex
		languages       map[string]*gotext.Po
		translationsDir string
		templateDir     string
		prefixSeparator string
		uniqueKeys      map[string]uniqueKey
		uniqueKeysCtx   map[string]map[string]uniqueKey
		potFile         string
		pot             *gotext.Po
	}

	uniqueKey struct{ singular, plural string }
)

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

	// load pot file if it exists
	tr.pot = gotext.NewPo()
	tr.pot.ParseFile(filepath.Join(tr.translationsDir, tr.potFile))

	return tr
}

func (t *Translator) SetPrefixSeparator(prefix string) {
	t.prefixSeparator = prefix
}

func (t *Translator) PrefixSeparator() string {
	return t.prefixSeparator
}

func (t *Translator) SetPotFile(potFile string) {
	t.potFile = potFile
}

func (t *Translator) PotFile() string {
	return t.potFile
}

func (t *Translator) SetTranslationsDir(translationsDir string) {
	t.translationsDir = translationsDir
}

func (t *Translator) TranslationsDir() string {
	return t.translationsDir
}

func (t *Translator) SetTemplateDir(templateDir string) {
	t.templateDir = templateDir
}

func (t *Translator) TemplateDir() string {
	return t.templateDir
}

// SetLanguage adds a new language to the translator by loading its .po file.
func (t *Translator) SetLanguage(lang string) error {
	po := gotext.NewPo()

	po.ParseFile(filepath.Join(t.translationsDir, lang+DefaultPoExtension))

	if po.Language == "" {
		return ErrorLanguageNotFound
	}

	t.languages[lang] = po
	return nil
}

// AddLanguage adds a new language to the translator by loading its .po file.
// Deprecated: use SetLanguage instead
func (t *Translator) AddLanguage(lang string) error {
	return t.SetLanguage(lang)
}

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

// CheckMissingTranslations scans template files for missing translations and logs them.
func (t *Translator) CheckMissingTranslations() error {
	err := t.ScanFiles(t.templateDir)
	if err != nil {
		return err
	}
	t.catalogueMu.Lock()
	defer t.catalogueMu.Unlock()

	var (
		tr  = t.pot.GetDomain().GetTranslations()
		ctr = t.pot.GetDomain().GetCtxTranslations()
	)

	for key, entry := range t.uniqueKeys {
		found := false
		for _, potKey := range tr {
			if key == potKey.ID {
				found = true
				break
			}
		}

		if !found {
			err = t.addToPotFile("", entry)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

	for ctx, uniqueKeys := range t.uniqueKeysCtx {
		for key, entry := range uniqueKeys {
			found := false
			if _, ok := ctr[ctx]; ok {
				for _, potKey := range ctr[ctx] {
					if key == potKey.ID {
						found = true
						break
					}
				}
			}

			if !found {
				err = t.addToPotFile(ctx, entry)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}

	return nil
}

func (t *Translator) ScanFiles(root string) error {
	t.catalogueMu.Lock()
	defer t.catalogueMu.Unlock()

	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, TemplateExtension) {
			err = t.scanFile(path)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// recordKey serializes discovery and POT updates for concurrent lookups.
func (t *Translator) recordKey(ctx string, entry uniqueKey) {
	t.catalogueMu.RLock()
	if t.hasRecordedKey(ctx, entry.singular) {
		t.catalogueMu.RUnlock()
		return
	}
	t.catalogueMu.RUnlock()

	t.catalogueMu.Lock()
	defer t.catalogueMu.Unlock()
	if t.hasRecordedKey(ctx, entry.singular) {
		return
	}

	if err := t.addToPotFileIfNotExists(ctx, entry); err != nil {
		fmt.Println(err)
		return
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

func (t *Translator) scanFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	trees := make(map[string]*parse.Tree)
	tree := parse.New(filename)
	// Extraction does not execute templates or require application helpers.
	tree.Mode = parse.SkipFuncCheck
	if _, err := tree.Parse(string(data), "{{", "}}", trees); err != nil {
		return fmt.Errorf("scan translations in %s: %w", filename, err)
	}
	for _, tree := range trees {
		t.scanTemplateNode(tree.Root)
	}
	return nil
}

func (t *Translator) scanTemplateNode(node parse.Node) {
	switch node := node.(type) {
	case *parse.ListNode:
		if node != nil {
			for _, child := range node.Nodes {
				t.scanTemplateNode(child)
			}
		}
	case *parse.ActionNode:
		t.scanTemplateNode(node.Pipe)
	case *parse.IfNode:
		t.scanTemplateNode(node.Pipe)
		t.scanTemplateNode(node.List)
		t.scanTemplateNode(node.ElseList)
	case *parse.RangeNode:
		t.scanTemplateNode(node.Pipe)
		t.scanTemplateNode(node.List)
		t.scanTemplateNode(node.ElseList)
	case *parse.WithNode:
		t.scanTemplateNode(node.Pipe)
		t.scanTemplateNode(node.List)
		t.scanTemplateNode(node.ElseList)
	case *parse.TemplateNode:
		t.scanTemplateNode(node.Pipe)
	case *parse.PipeNode:
		if node != nil {
			for _, cmd := range node.Cmds {
				t.scanTemplateNode(cmd)
			}
		}
	case *parse.CommandNode:
		t.scanTranslationCommand(node)
		for _, arg := range node.Args {
			t.scanTemplateNode(arg)
		}
	}
}

func (t *Translator) scanTranslationCommand(cmd *parse.CommandNode) {
	if len(cmd.Args) == 0 {
		return
	}
	fn, ok := cmd.Args[0].(*parse.IdentifierNode)
	if !ok {
		return
	}
	count := map[string]int{"tl": 1, "tn": 2, "ctl": 2, "ctn": 3}[fn.Ident]
	if count == 0 || len(cmd.Args) < count+2 {
		return
	}
	args := make([]string, count)
	for i := range args {
		literal, ok := cmd.Args[i+2].(*parse.StringNode)
		if !ok {
			return // Dynamic keys are discovered by runtime translation calls.
		}
		args[i] = literal.Text
	}
	var ctx string
	entry := uniqueKey{}
	switch fn.Ident {
	case "tl":
		entry.singular = args[0]
	case "tn":
		entry.singular, entry.plural = args[0], args[1]
	case "ctl":
		ctx, entry.singular = args[0], args[1]
	case "ctn":
		ctx, entry.singular, entry.plural = args[0], args[1], args[2]
	}
	if ctx == "" {
		if entry.plural == "" {
			entry.plural = t.uniqueKeys[entry.singular].plural
		}
		t.uniqueKeys[entry.singular] = entry
		return
	}
	if t.uniqueKeysCtx[ctx] == nil {
		t.uniqueKeysCtx[ctx] = make(map[string]uniqueKey)
	}
	if entry.plural == "" {
		entry.plural = t.uniqueKeysCtx[ctx][entry.singular].plural
	}
	t.uniqueKeysCtx[ctx][entry.singular] = entry
}

func (t *Translator) addToPotFile(ctx string, entry uniqueKey) error {
	file, err := os.OpenFile(path.Join(t.translationsDir, t.potFile), os.O_APPEND|os.O_WRONLY, 0o644)
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

	// Reload pot file contents
	t.pot.ParseFile(filepath.Join(t.translationsDir, t.potFile))

	return nil
}

// addToPotFileIfNotExists must be called while catalogueMu is held.
func (t *Translator) addToPotFileIfNotExists(ctx string, entry uniqueKey) error {
	if ctx == "" {
		for _, existing := range t.pot.GetDomain().GetTranslations() {
			if existing.ID == entry.singular {
				return nil
			}
		}
	} else {
		for _, existing := range t.pot.GetDomain().GetCtxTranslations()[ctx] {
			if existing.ID == entry.singular {
				return nil
			}
		}
	}
	return t.addToPotFile(ctx, entry)
}

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
	idx := strings.LastIndex(s, t.PrefixSeparator())
	if idx != -1 {
		// Remove everything up to and including the prefix separator
		return s[idx+len(t.PrefixSeparator()):]
	}
	return s
}
