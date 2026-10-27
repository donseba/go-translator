package translator

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/leonelquinteros/gotext"
	"github.com/leonelquinteros/gotext/plurals"
)

// MergeDefaults adds missing messages from a PO catalogue to a loaded language.
// Existing translated entries take precedence. Blank entries (an empty msgstr,
// or a plural entry whose msgstr[n] forms are all empty) count as missing, so a
// catalogue refreshed from a POT file does not hide the defaults. Contexts and
// every plural form are retained. This never writes the application's PO file.
// Configure catalogues before serving concurrent requests, as with SetLanguage.
//
// The catalogue must be in the same base language as the loaded one: nl_NL
// defaults merge into nl, nl into nl_NL. When it has plural messages, its
// plural rules must choose the same forms (they are compared by evaluating
// them, not as text). It needs a Language header. Otherwise the error matches
// ErrorInvalidCatalogue.
func (t *Translator) MergeDefaults(language string, files fs.FS, name string) error {
	current, exists := t.languages[language]
	if !exists {
		return ErrorLanguageNotFound
	}

	data, err := fs.ReadFile(files, name)
	if err != nil {
		return err
	}

	merge := newMerger(language, current)

	if err := merge.add(data, name, ""); err != nil {
		return err
	}

	return t.applyMerge(merge)
}

// mergeFiles merges the file name of each of sources that has it into the
// loaded catalogue of language, earlier sources first, and parses the merged
// catalogue once. A file without a Language header is in language. When a
// file does not fit, nothing is merged.
func (t *Translator) mergeFiles(language string, sources []fs.FS, name string) error {
	merge := newMerger(language, t.languages[language])

	for _, source := range sources {
		data, err := fs.ReadFile(source, name)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}

		if err != nil {
			return err
		}

		if err := merge.add(data, name, language); err != nil {
			return err
		}
	}

	return t.applyMerge(merge)
}

// applyMerge replaces the catalogue of the merger's language with the merged
// one, when there is anything to add.
func (t *Translator) applyMerge(merge *merger) error {
	if merge.additions.Len() == 0 {
		return nil
	}

	data, err := merge.current.MarshalText()
	if err != nil {
		return err
	}

	merged := gotext.NewPo()
	merged.Parse(append(data, merge.additions.String()...))

	if merged.Language == "" {
		merged.Language = merge.current.Language
	}

	t.languages[merge.language] = merged

	return nil
}

// merger collects the default entries a loaded catalogue lacks, from one or
// more default catalogues, as PO text to append to it.
type merger struct {
	language  string
	current   *gotext.Po
	entries   map[string]*gotext.Translation
	contexts  map[string]map[string]*gotext.Translation
	additions strings.Builder
}

// newMerger starts a merge into the catalogue current, loaded as language.
func newMerger(language string, current *gotext.Po) *merger {
	return &merger{
		language: language,
		current:  current,
		entries:  current.GetDomain().GetTranslations(),
		contexts: current.GetDomain().GetCtxTranslations(),
	}
}

// add checks the default catalogue data and collects the entries it adds.
// An added entry counts as existing for the defaults that follow. A
// catalogue without a Language header is in language; with language empty,
// it does not fit.
func (m *merger) add(data []byte, name, language string) error {
	defaults := gotext.NewPo()
	defaults.Parse(data)

	if defaults.Language == "" {
		defaults.Language = language
	}

	if !m.sameLanguage(defaults.Language) {
		return fmt.Errorf("%w: %s", ErrorInvalidCatalogue, name)
	}

	entries := defaults.GetDomain().GetTranslations()
	contexts := defaults.GetDomain().GetCtxTranslations()

	if hasPlurals(entries, contexts) && !samePluralRules(defaults.PluralForms, m.current.PluralForms) {
		return fmt.Errorf("%w: plural rules in %s", ErrorInvalidCatalogue, name)
	}

	m.addEntries(entries)
	m.addContexts(contexts)

	return nil
}

// sameLanguage reports whether language, the language of a default
// catalogue, has the base language of the loaded catalogue: its Language
// header, or the language it is loaded as without one. An empty language
// does not fit.
func (m *merger) sameLanguage(language string) bool {
	current := m.current.Language
	if current == "" {
		current = m.language
	}

	return language != "" && sameBaseLanguage(language, current)
}

// hasPlurals reports whether one of the entries, with or without a context,
// is a plural message.
func hasPlurals(entries map[string]*gotext.Translation, contexts map[string]map[string]*gotext.Translation) bool {
	for _, entry := range entries {
		if entry.PluralID != "" {
			return true
		}
	}

	for _, contextEntries := range contexts {
		for _, entry := range contextEntries {
			if entry.PluralID != "" {
				return true
			}
		}
	}

	return false
}

// addEntries collects the default entries without a context that the loaded
// catalogue misses or has blank. Appended entries are parsed after the
// existing ones, so a default that fills a blank entry replaces it in the
// merged catalogue. The header entry (msgid "") is never added.
func (m *merger) addEntries(entries map[string]*gotext.Translation) {
	for key, entry := range entries {
		if key == "" || !needsDefault(m.entries[key], entry) {
			continue
		}

		m.appendEntry("", entry)
		m.entries[key] = entry
	}
}

// addContexts collects the default entries with a context that the loaded
// catalogue misses or has blank, as addEntries does for those without.
func (m *merger) addContexts(contexts map[string]map[string]*gotext.Translation) {
	for context, entries := range contexts {
		for key, entry := range entries {
			if !needsDefault(m.contexts[context][key], entry) {
				continue
			}

			m.appendEntry(context, entry)

			if m.contexts[context] == nil {
				m.contexts[context] = make(map[string]*gotext.Translation)
			}

			m.contexts[context][key] = entry
		}
	}
}

// appendEntry writes entry as PO text to the additions.
func (m *merger) appendEntry(context string, entry *gotext.Translation) {
	additions := &m.additions

	additions.WriteByte('\n')

	if context != "" {
		fmt.Fprintf(additions, "msgctxt %q\n", context)
	}

	fmt.Fprintf(additions, "msgid %q\n", entry.ID)

	if entry.PluralID == "" {
		fmt.Fprintf(additions, "msgstr %q\n", entry.Trs[0])
		return
	}

	fmt.Fprintf(additions, "msgid_plural %q\n", entry.PluralID)

	for _, index := range slices.Sorted(maps.Keys(entry.Trs)) {
		fmt.Fprintf(additions, "msgstr[%d] %q\n", index, entry.Trs[index])
	}
}

// sameBaseLanguage reports whether two Language headers name the same base
// language, such as nl_NL and nl.
func sameBaseLanguage(a, b string) bool {
	return baseLanguage(a) == baseLanguage(b)
}

// samePluralRules reports whether two Plural-Forms headers choose the same
// form for every count from 0 to 1000, with the same number of forms.
func samePluralRules(a, b string) bool {
	if strings.Join(strings.Fields(a), "") == strings.Join(strings.Fields(b), "") {
		return true
	}

	formsA, ruleA := pluralRule(a)
	formsB, ruleB := pluralRule(b)

	if formsA != formsB {
		return false
	}

	for n := range uint32(1001) {
		if ruleA(n) != ruleB(n) {
			return false
		}
	}

	return true
}

// pluralRule reads a Plural-Forms header as gotext applies it: the number of
// forms and the rule. A header without a valid rule gets the rule gotext
// falls back to, the singular for 1 and the plural otherwise.
func pluralRule(header string) (int, func(n uint32) int) {
	forms := 2
	rule := func(n uint32) int {
		if n == 1 {
			return 0
		}

		return 1
	}

	for part := range strings.SplitSeq(header, ";") {
		key, value, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}

		switch strings.TrimSpace(key) {
		case "nplurals":
			if number, err := strconv.Atoi(strings.TrimSpace(value)); err == nil {
				forms = number
			}
		case "plural":
			if expression, err := plurals.Compile(strings.TrimSpace(value)); err == nil {
				rule = expression.Eval
			}
		}
	}

	return forms, rule
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
