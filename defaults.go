package translator

import (
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/leonelquinteros/gotext"
)

// MergeDefaults adds missing messages from a PO catalogue to a loaded language.
// Existing translated entries take precedence. Blank entries (an empty msgstr,
// or a plural entry whose msgstr[n] forms are all empty) count as missing, so a
// catalogue refreshed from a POT file does not hide the defaults. Contexts and
// every plural form are retained. This never writes the application's PO file.
// Configure catalogues before serving concurrent requests, as with SetLanguage.
func (t *Translator) MergeDefaults(language string, files fs.FS, name string) error {
	current, exists := t.languages[language]
	if !exists {
		return ErrorLanguageNotFound
	}
	data, err := fs.ReadFile(files, name)
	if err != nil {
		return err
	}
	defaults := gotext.NewPo()
	defaults.Parse(data)
	if defaults.Language == "" || defaults.Language != current.Language {
		return fmt.Errorf("%w: %s", ErrorInvalidCatalogue, name)
	}
	defaultEntries := defaults.GetDomain().GetTranslations()
	defaultContexts := defaults.GetDomain().GetCtxTranslations()
	plural := false
	for _, entry := range defaultEntries {
		plural = plural || entry.PluralID != ""
	}
	for _, entries := range defaultContexts {
		for _, entry := range entries {
			plural = plural || entry.PluralID != ""
		}
	}
	if plural && strings.Join(strings.Fields(defaults.PluralForms), "") != strings.Join(strings.Fields(current.PluralForms), "") {
		return fmt.Errorf("%w: plural rules in %s", ErrorInvalidCatalogue, name)
	}

	var additions strings.Builder
	appendEntry := func(context string, entry *gotext.Translation) {
		additions.WriteByte('\n')
		if context != "" {
			fmt.Fprintf(&additions, "msgctxt %q\n", context)
		}
		fmt.Fprintf(&additions, "msgid %q\n", entry.ID)
		if entry.PluralID == "" {
			fmt.Fprintf(&additions, "msgstr %q\n", entry.Trs[0])
			return
		}
		fmt.Fprintf(&additions, "msgid_plural %q\n", entry.PluralID)
		indices := make([]int, 0, len(entry.Trs))
		for index := range entry.Trs {
			indices = append(indices, index)
		}
		sort.Ints(indices)
		for _, index := range indices {
			fmt.Fprintf(&additions, "msgstr[%d] %q\n", index, entry.Trs[index])
		}
	}
	// Appended entries are parsed after the existing ones, so a default that
	// fills a blank entry replaces it in the merged catalogue.
	known := current.GetDomain().GetTranslations()
	for key, entry := range defaultEntries {
		if key != "" && needsDefault(known[key], entry) {
			appendEntry("", entry)
		}
	}
	contexts := current.GetDomain().GetCtxTranslations()
	for context, entries := range defaultContexts {
		for key, entry := range entries {
			if needsDefault(contexts[context][key], entry) {
				appendEntry(context, entry)
			}
		}
	}
	if additions.Len() == 0 {
		return nil
	}
	data, err = current.MarshalText()
	if err != nil {
		return err
	}
	merged := gotext.NewPo()
	merged.Parse(append(data, additions.String()...))
	t.languages[language] = merged
	return nil
}

// needsDefault reports whether the default entry should be used: the existing
// entry is missing, or it is blank while the default has a translation.
func needsDefault(existing, def *gotext.Translation) bool {
	if existing == nil {
		return true
	}

	return isBlank(existing) && !isBlank(def)
}

// isBlank reports whether an entry has no translated text in any form.
func isBlank(entry *gotext.Translation) bool {
	for _, translation := range entry.Trs {
		if translation != "" {
			return false
		}
	}

	return true
}
