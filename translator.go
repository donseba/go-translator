package translator

import (
	"bufio"
	"fmt"
	"github.com/donseba/gotext"
	"html/template"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	re          = regexp.MustCompile(`{{\s*tl\s+(?:"[^"]+"|[^"]+)\s+"([^"]+)"`)
	rePlural    = regexp.MustCompile(`{{\s*tn\s+[.\$][a-zA-Z_]\w*\s+"([^"]+)"\s+"([^"]+)"(.*?)\s*}}`)
	reCtx       = regexp.MustCompile(`{{\s*ctl\s+[.\$][a-zA-Z_]\w*\s+"([^"]+)"\s+"([^"]+)"\s*}}`)
	reCtxPlural = regexp.MustCompile(`{{\s*ctn\s+[.\$][a-zA-Z_]\w*\s+"([^"]+)"\s+"([^"]+)"\s+"([^"]+)"`)
)

var (
	ErrorLanguageNotFound      = fmt.Errorf("language not found")
	ErrorLanguageAlreadyExists = fmt.Errorf("language already exists")
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
	//Localizer interface contains the methods that are needed for the translator
	Localizer interface {
		// GetLocale returns the locale of the localizer, ie. "en_US"
		GetLocale() string
	}

	Translator struct {
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

	translationKey struct {
		ctx    string
		value  string
		plural bool // true if it's a plural key
	}
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

func (t *Translator) AddLanguage(lang string) error {
	po := gotext.NewPo()

	po.ParseFile(filepath.Join(t.translationsDir, lang+DefaultPoExtension))

	if po.Language == "" {
		return ErrorLanguageNotFound
	}

	t.languages[lang] = po
	return nil
}

// CheckMissingTranslations scans template files for missing translations and logs them.
func (t *Translator) CheckMissingTranslations() error {
	err := t.ScanFiles(t.templateDir)
	if err != nil {
		return err
	}

	var (
		tr  = t.pot.GetTranslations()
		ctr = t.pot.GetCtxTranslations()
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
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(path, TemplateExtension)) {
			err = t.scanFile(path)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (t *Translator) scanFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	matches := re.FindAllStringSubmatch(string(data), -1)
	for _, match := range matches {
		if len(match) > 1 {
			t.uniqueKeys[match[1]] = uniqueKey{singular: match[1]}
		}
	}

	matchesPlural := rePlural.FindAllStringSubmatch(string(data), -1)
	for _, match := range matchesPlural {
		if len(match) > 2 {
			t.uniqueKeys[match[1]] = uniqueKey{singular: match[1], plural: match[2]}
		}
	}

	matchesCtx := reCtx.FindAllStringSubmatch(string(data), -1)
	for _, match := range matchesCtx {
		if len(match) > 1 {
			if t.uniqueKeysCtx[match[1]] == nil {
				t.uniqueKeysCtx[match[1]] = make(map[string]uniqueKey)
			}
			t.uniqueKeysCtx[match[1]][match[2]] = uniqueKey{singular: match[2]}
		}
	}

	matchesCtxPlural := reCtxPlural.FindAllStringSubmatch(string(data), -1)
	for _, match := range matchesCtxPlural {
		if len(match) > 2 {
			if t.uniqueKeysCtx[match[1]] == nil {
				t.uniqueKeysCtx[match[1]] = make(map[string]uniqueKey)
			}
			t.uniqueKeysCtx[match[1]][match[2]] = uniqueKey{singular: match[2], plural: match[3]}
		}
	}

	return nil
}

func (t *Translator) addToPotFile(ctx string, entry uniqueKey) error {
	file, err := os.OpenFile(path.Join(t.translationsDir, t.potFile), os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

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

func (t *Translator) addToPotFileIfNotExists(key translationKey) error {
	tr := t.pot.GetDomain().GetTranslations()

	if key.ctx == "" {
		for _, potKey := range tr {
			if key.value == potKey.ID {
				return nil
			}
		}

		return t.addToPotFile("", uniqueKey{singular: key.value})
	}

	ctr := t.pot.GetDomain().GetCtxTranslations()
	if ctr[key.ctx] == nil {
		return t.addToPotFile(key.ctx, uniqueKey{singular: key.value})
	}

	for _, potKey := range ctr[key.ctx] {
		if key.value == potKey.ID {
			return nil
		}
	}

	return t.addToPotFile(key.ctx, uniqueKey{singular: key.value})
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
