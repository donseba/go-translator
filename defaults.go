package translator

import (
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/leonelquinteros/gotext"
)

// MergeDefaults adds missing messages from a PO catalogue to a loaded language.
// Existing entries, including blank translations, take precedence. Contexts and
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
	known := current.GetDomain().GetTranslations()
	for key, entry := range defaultEntries {
		if _, exists := known[key]; key != "" && !exists {
			appendEntry("", entry)
		}
	}
	contexts := current.GetDomain().GetCtxTranslations()
	for context, entries := range defaultContexts {
		for key, entry := range entries {
			if _, exists := contexts[context][key]; !exists {
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
