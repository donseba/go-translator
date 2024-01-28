package translator

import (
	"bufio"
	"fmt"
	"github.com/donseba/gotext"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	re            = regexp.MustCompile(`{{\s*tl\s+(?:"[^"]+"|[^"]+)\s+"([^"]+)"`)
	rePlural      = regexp.MustCompile(`{{\s*tn\s+"[^"]+"\s+"([^"]+)"\s+"([^"]+)"\s+\b[0-9a-zA-Z_]+\b\s*}}`)
	rePotSingular = regexp.MustCompile(`msgid\s+"([^"]+)"\nmsgstr\s+"[^"]*"`)
	rePotPlural   = regexp.MustCompile(`msgid\s+"([^"]+)"\nmsgid_plural\s+"([^"]+)"`)
)

var (
	ErrorLanguageNotFound = fmt.Errorf("language not found")
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
		uniqueKeys      map[translationKey]uniqueKey
		potFile         string
	}

	uniqueKey struct{ singular, plural string }

	translationKey struct {
		value  string
		plural bool // true if it's a plural key
	}
)

func NewTranslator(translationsDir, templateDir string) *Translator {
	return &Translator{
		translationsDir: translationsDir,
		templateDir:     templateDir,
		languages:       make(map[string]*gotext.Po),
		uniqueKeys:      make(map[translationKey]uniqueKey),
		potFile:         "translations.pot",
	}
}

func (t *Translator) SetPrefixSeparator(prefix string) {
	t.prefixSeparator = prefix
}

func (t *Translator) PrefixSeparator() string {
	if t.prefixSeparator == "" {
		return "__."
	}

	return t.prefixSeparator
}

func (t *Translator) AddLanguage(lang string) error {
	po := gotext.NewPo()

	po.ParseFile(filepath.Join(t.translationsDir, lang+".po"))

	if po.Language == "" {
		return ErrorLanguageNotFound
	}

	t.languages[lang] = po
	return nil
}

// CheckMissingTranslations scans template files for missing translations and logs them.
func (t *Translator) CheckMissingTranslations(potFile string) error {
	err := t.ScanFiles(t.templateDir)
	if err != nil {
		return err
	}

	t.potFile = potFile

	potKeys, err := t.loadKeysFromPotFile(t.translationsDir + "/" + potFile)
	if err != nil {
		return err
	}

	for key, entry := range t.uniqueKeys {
		found := false
		for _, potKey := range potKeys {
			if key.value == potKey.value && key.plural == potKey.plural {
				found = true
				break
			}
		}
		if !found {
			err = t.addToPotFile(t.translationsDir+"/"+potFile, entry)
			if err != nil {
				return err
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
		if !info.IsDir() && (strings.HasSuffix(path, ".gohtml")) {
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

	// Existing pattern for tl
	matches := re.FindAllStringSubmatch(string(data), -1)
	for _, match := range matches {
		if len(match) > 1 {
			t.uniqueKeys[translationKey{match[1], false}] = uniqueKey{singular: match[1]}
		}
	}

	// New pattern for tn (plural)
	matchesPlural := rePlural.FindAllStringSubmatch(string(data), -1)
	for _, match := range matchesPlural {
		if len(match) > 2 {
			t.uniqueKeys[translationKey{match[1], true}] = uniqueKey{singular: match[1], plural: match[2]}
		}
	}

	return nil
}

func (t *Translator) addToPotFile(filename string, entry uniqueKey) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	buf := bufio.NewWriter(file)

	// Write the new translation entry
	var content string
	if entry.plural == "" {
		content = fmt.Sprintf("\nmsgid \"%s\"\nmsgstr \"\"\n", entry.singular)
	} else {
		content = fmt.Sprintf("\nmsgid \"%s\"\nmsgid_plural \"%s\"\nmsgstr[0] \"\"\nmsgstr[1] \"\"\n", entry.singular, entry.plural)
	}

	if _, err := buf.WriteString(content); err != nil {
		return err
	}

	if err := buf.Flush(); err != nil {
		return err
	}

	return nil
}

func (t *Translator) loadKeysFromPotFile(filename string) ([]translationKey, error) {
	// Check if the file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// Create the file with a header if it doesn't exist
		header := `msgid ""
msgstr ""
"Project-Id-Version: dashblox\n"
"Content-Type: text/plain; charset=UTF-8\n"
`
		if err := os.WriteFile(filename, []byte(header), 0644); err != nil {
			return nil, err
		}
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var keys []translationKey
	matchesSingular := rePotSingular.FindAllStringSubmatch(string(data), -1)
	for _, match := range matchesSingular {
		keys = append(keys, translationKey{value: match[1], plural: false})
	}

	matchesPlural := rePotPlural.FindAllStringSubmatch(string(data), -1)
	for _, match := range matchesPlural {
		keys = append(keys, translationKey{value: match[1], plural: true})
	}

	return keys, nil
}

func (t *Translator) addToPotFileIfNotExists(key translationKey) error {
	// Load existing keys from the .pot file
	existingKeys, err := t.loadKeysFromPotFile(t.translationsDir + "/" + t.potFile)
	if err != nil {
		return fmt.Errorf("error loading keys from POT file: %w", err)
	}

	// Check if the key exists
	for _, existingKey := range existingKeys {
		if key == existingKey {
			return nil // Key already exists, no need to add
		}
	}

	// Get the corresponding uniqueKey entry
	entry, exists := t.uniqueKeys[key]
	if !exists {
		return fmt.Errorf("key not found in uniqueKeys")
	}

	// Add the key to the .pot file
	return t.addToPotFile(t.translationsDir+"/"+t.potFile, entry)
}

func (t *Translator) FuncMap() template.FuncMap {
	return template.FuncMap{
		"tl": t.tl,
		"tn": t.tn,
	}
}

// Tl translates a string based on the given language tag and key.
func (t *Translator) tl(loc Localizer, key string, args ...interface{}) string {
	translator, exists := t.languages[loc.GetLocale()]
	if !exists {
		return "*" + key + "*" // return key if language is not available
	}

	if !translator.IsTranslated(key) {
		// check if translation is in the pot file, if not, add it
		if _, ok := t.uniqueKeys[translationKey{key, false}]; !ok {
			t.uniqueKeys[translationKey{key, false}] = uniqueKey{singular: key}

			if t.potFile == "" {
				t.potFile = "translations.pot"
			}

			err := t.addToPotFileIfNotExists(translationKey{key, false})
			if err != nil {
				fmt.Println(err)
			}
		}
		return "*" + key + "*"
	}

	translated := translator.Get(key, args...)
	return t.removePrefix(translated)
}

// tn method for handling plurals
func (t *Translator) tn(loc Localizer, singular, plural string, n int, args ...any) string {
	translator, exists := t.languages[loc.GetLocale()]
	if !exists {
		return "*" + singular + "/" + plural + "*" // return key if language is not available
	}

	if !translator.IsTranslatedN(singular, n) {
		// check if translation is in the pot file, if not, add it
		if _, ok := t.uniqueKeys[translationKey{singular, true}]; !ok {
			t.uniqueKeys[translationKey{singular, true}] = uniqueKey{singular: singular, plural: plural}

			if t.potFile == "" {
				t.potFile = "translations.pot"
			}

			err := t.addToPotFileIfNotExists(translationKey{singular, true})
			if err != nil {
				fmt.Println(err)
			}
		}
		return "*" + singular + "/" + plural + "*"
	}

	translated := translator.GetN(singular, plural, n, args...)
	return t.removePrefix(translated)
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

// Tl translates a string based on the given language tag and key.
func (t *Translator) Tl(loc Localizer, key string, args ...interface{}) string {
	return t.tl(loc, key, args...)
}

func (t *Translator) Tn(loc Localizer, singular, plural string, n int) string {
	return t.tn(loc, singular, plural, n)
}
