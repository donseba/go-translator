package translator

import (
	"fmt"

	"github.com/leonelquinteros/gotext"
)

// tl translates a string based on the given language tag and key.
func (t *Translator) tl(loc Localizer, key string, args ...any) string {
	text, _ := t.Lookup(loc, key, args...)

	return text
}

// ctl translates a string with a context. Without a context it is tl.
func (t *Translator) ctl(loc Localizer, ctx, key string, args ...any) string {
	text, _ := t.LookupC(loc, ctx, key, args...)

	return text
}

// tn method for handling plurals
func (t *Translator) tn(loc Localizer, singular, plural string, n int, args ...any) string {
	text, _ := t.LookupN(loc, singular, plural, n, args...)

	return text
}

// ctn translates a string with plural forms and a context.
func (t *Translator) ctn(loc Localizer, ctx, singular, plural string, n int, args ...any) string {
	text, _ := t.LookupCN(loc, ctx, singular, plural, n, args...)

	return text
}

// lookup records the message (see SetRecordMissing) and translates it in the
// language of loc. Without translation it returns the text set with
// SetUntranslated and false.
func (t *Translator) lookup(loc Localizer, m Miss) (string, bool) {
	// A message without a context is looked up as one, also by ctl and ctn.
	if m.Context == "" {
		switch m.kind {
		case kindCTL:
			m.kind = kindTL
		case kindCTN:
			m.kind = kindTN
		}
	}

	m.Locale = loc.GetLocale()

	t.recordKey(m.Context, uniqueKey{singular: m.Key, plural: m.Plural})

	text, ok := t.find(m)
	if ok {
		return text, true
	}

	return t.miss(m), false
}

// find translates m in the catalogue of its locale or, with
// SetLanguageFallback, in the first catalogue of a similar language that
// translates it.
func (t *Translator) find(m Miss) (string, bool) {
	if !t.fallback {
		po, exists := t.languages[m.Locale]
		if !exists {
			return "", false
		}

		return t.translate(po, m)
	}

	for _, name := range t.resolve(m.Locale) {
		text, ok := t.translate(t.languages[name], m)
		if ok {
			return text, true
		}
	}

	return "", false
}

// translate translates m in one catalogue. An empty translation counts as
// none.
func (t *Translator) translate(po *gotext.Po, m Miss) (string, bool) {
	var translated string

	switch m.kind {
	case kindTN:
		if !po.IsTranslatedN(m.Key, m.N) {
			return "", false
		}

		translated = po.GetN(m.Key, m.Plural, m.N, m.Args...)
	case kindCTL:
		if !po.IsTranslatedC(m.Key, m.Context) {
			return "", false
		}

		//nolint:staticcheck // Keep dynamic catalogue keys from being inferred as printf format parameters.
		translated = po.GetC(fmt.Sprintf("%s", m.Key), m.Context, m.Args...)
	case kindCTN:
		if !po.IsTranslatedNC(m.Key, m.N, m.Context) {
			return "", false
		}

		translated = po.GetNC(m.Key, m.Plural, m.N, m.Context, m.Args...)
	default:
		if !po.IsTranslated(m.Key) {
			return "", false
		}

		//nolint:staticcheck // Keep dynamic catalogue keys from being inferred as printf format parameters.
		translated = po.Get(fmt.Sprintf("%s", m.Key), m.Args...)
	}

	return t.removePrefix(translated), true
}

// Tl translates a string based on the given language tag and key.
func (t *Translator) Tl(loc Localizer, key string, args ...any) string {
	return t.tl(loc, key, args...)
}

// Tn translates a message with plural forms, choosing the form for n by the
// plural rules of the language. args format the chosen form, so n is
// usually passed again:
//
//	tr.Tn(loc, "%d message", "%d messages", count, count)
func (t *Translator) Tn(loc Localizer, singular, plural string, n int, args ...any) string {
	return t.tn(loc, singular, plural, n, args...)
}

// Ctl translates a message with a context (msgctxt). Without a context it is
// Tl.
func (t *Translator) Ctl(loc Localizer, ctx, key string, args ...any) string {
	return t.ctl(loc, ctx, key, args...)
}

// Ctn is Tn for a message with a context (msgctxt). Without a context it is
// Tn.
func (t *Translator) Ctn(loc Localizer, ctx, singular, plural string, n int, args ...any) string {
	return t.ctn(loc, ctx, singular, plural, n, args...)
}

// Lookup is Tl that also reports whether a translation was found. Without
// one it returns the text set with SetUntranslated (by default the marker
// "*key*") and false. Like Tl it records the key (see SetRecordMissing).
func (t *Translator) Lookup(loc Localizer, key string, args ...any) (string, bool) {
	return t.lookup(loc, Miss{kind: kindTL, Key: key, Args: args})
}

// LookupN is Tn that also reports whether a translation was found, see
// Lookup.
func (t *Translator) LookupN(loc Localizer, singular, plural string, n int, args ...any) (string, bool) {
	return t.lookup(loc, Miss{kind: kindTN, Key: singular, Plural: plural, N: n, Args: args})
}

// LookupC is Ctl that also reports whether a translation was found, see
// Lookup.
func (t *Translator) LookupC(loc Localizer, ctx, key string, args ...any) (string, bool) {
	return t.lookup(loc, Miss{kind: kindCTL, Context: ctx, Key: key, Args: args})
}

// LookupCN is Ctn that also reports whether a translation was found, see
// Lookup.
func (t *Translator) LookupCN(loc Localizer, ctx, singular, plural string, n int, args ...any) (string, bool) {
	return t.lookup(loc, Miss{kind: kindCTN, Context: ctx, Key: singular, Plural: plural, N: n, Args: args})
}

// Details returns the headers of the catalogue of loc, also those set with
// NewLanguage and SetDetails; nil for a language that is not loaded.
func (t *Translator) Details(loc Localizer) map[string][]string {
	po, exists := t.languages[loc.GetLocale()]
	if !exists {
		return nil
	}

	return po.GetDomain().Headers
}
