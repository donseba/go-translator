package translator

import (
	"sort"
	"strings"
)

// SetLanguageFallback turns language fallback on or off; it is off by
// default, so a lookup only uses the catalogue loaded under exactly the
// locale of the Localizer.
//
// With fallback on, a lookup tries the catalogues of similar languages in
// this order, message by message, and uses the first that translates it:
//
//  1. the catalogue loaded under the locale itself, such as "nl_NL";
//  2. one loaded under the same locale written differently: case, "-" or
//     "_", and an encoding or modifier (".UTF-8", "@euro") do not count, so
//     "nl-nl" finds "nl_NL";
//  3. the base language, such as "nl" for "nl_NL";
//  4. another region of the same language, such as "nl_NL" for "nl" or for
//     "nl_BE": first the region named like the language ("nl_NL", "de_DE"),
//     then the others in alphabetical order.
//
// Only lookups fall back (tl, tn, ctl, ctn, Tl, Tn, Ctl, Ctn and the Lookup
// functions); functions that edit or write a catalogue still need its exact
// name. Configure it before serving requests.
func (t *Translator) SetLanguageFallback(enabled bool) {
	t.fallback = enabled
}

// LanguageFallback reports whether lookups fall back to similar languages,
// see SetLanguageFallback.
func (t *Translator) LanguageFallback() bool {
	return t.fallback
}

// resolve returns the names of the loaded catalogues a lookup in locale
// tries, in the order of SetLanguageFallback.
func (t *Translator) resolve(locale string) []string {
	var names []string

	if _, exists := t.languages[locale]; exists {
		names = append(names, locale)
	}

	want := normalizeLocale(locale)
	base := baseLanguage(locale)

	var same, bases, regions []string

	for name := range t.languages {
		if name == locale {
			continue
		}

		normalized := normalizeLocale(name)

		switch {
		case normalized == want:
			same = append(same, name)
		case normalized == base:
			bases = append(bases, name)
		case strings.HasPrefix(normalized, base+"_"):
			regions = append(regions, name)
		}
	}

	sort.Strings(same)
	sort.Strings(bases)

	// The region named like the language, such as nl_NL, comes first.
	sort.Slice(regions, func(i, j int) bool {
		first := normalizeLocale(regions[i]) == base+"_"+base
		second := normalizeLocale(regions[j]) == base+"_"+base

		if first != second {
			return first
		}

		return regions[i] < regions[j]
	})

	names = append(names, same...)
	names = append(names, bases...)

	return append(names, regions...)
}

// normalizeLocale writes a locale in one way: lower case, "_" between
// language and region, without an encoding or modifier. "nl-NL.UTF-8"
// becomes "nl_nl".
func normalizeLocale(locale string) string {
	locale, _, _ = strings.Cut(locale, ".")
	locale, _, _ = strings.Cut(locale, "@")

	return strings.ToLower(strings.ReplaceAll(locale, "-", "_"))
}

// baseLanguage returns the language of a locale without its region, written
// as normalizeLocale does: "nl" for "nl-NL.UTF-8".
func baseLanguage(locale string) string {
	base, _, _ := strings.Cut(normalizeLocale(locale), "_")

	return base
}
