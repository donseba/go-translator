package translator

import (
	"fmt"
)

// Tl translates a string based on the given language tag and key.
func (t *Translator) tl(loc Localizer, key string, args ...any) string {
	t.recordKey("", uniqueKey{singular: key})
	translator, exists := t.languages[loc.GetLocale()]
	if !exists {
		return fmt.Sprintf(DefaultNoTranslationTL, key)
	}

	if !translator.IsTranslated(key) {
		return fmt.Sprintf(DefaultNoTranslationTL, key)
	}

	translated := translator.Get(fmt.Sprintf("%s", key), args...) //nolint:gosimple
	return t.removePrefix(translated)
}

func (t *Translator) ctl(loc Localizer, ctx, key string, args ...any) string {
	t.recordKey(ctx, uniqueKey{singular: key})
	translator, exists := t.languages[loc.GetLocale()]
	if !exists {
		return fmt.Sprintf(DefaultNoTranslationCTL, ctx, key)
	}

	if ctx == "" {
		return t.tl(loc, key, args...)
	}

	if !translator.IsTranslatedC(key, ctx) {
		return fmt.Sprintf(DefaultNoTranslationCTL, ctx, key)
	}

	translated := translator.GetC(fmt.Sprintf("%s", key), ctx, args...) //nolint:gosimple
	return t.removePrefix(translated)
}

// tn method for handling plurals
func (t *Translator) tn(loc Localizer, singular, plural string, n int, args ...any) string {
	t.recordKey("", uniqueKey{singular: singular, plural: plural})
	translator, exists := t.languages[loc.GetLocale()]
	if !exists {
		return fmt.Sprintf(DefaultNoTranslationTN, singular, plural)
	}

	if !translator.IsTranslatedN(singular, n) {
		return fmt.Sprintf(DefaultNoTranslationTN, singular, plural)
	}

	translated := translator.GetN(singular, plural, n, args...)
	return t.removePrefix(translated)
}

func (t *Translator) ctn(loc Localizer, ctx, singular, plural string, n int, args ...any) string {
	t.recordKey(ctx, uniqueKey{singular: singular, plural: plural})
	translator, exists := t.languages[loc.GetLocale()]
	if !exists {
		return fmt.Sprintf(DefaultNoTranslationCTN, ctx, singular, plural)
	}

	if !translator.IsTranslatedNC(singular, n, ctx) {
		return fmt.Sprintf(DefaultNoTranslationCTN, ctx, singular, plural)
	}

	translated := translator.GetNC(singular, plural, n, ctx, args...)
	return t.removePrefix(translated)
}

// Tl translates a string based on the given language tag and key.
func (t *Translator) Tl(loc Localizer, key string, args ...any) string {
	return t.tl(loc, key, args...)
}

// Tn method for handling plurals
func (t *Translator) Tn(loc Localizer, singular, plural string, n int) string {
	return t.tn(loc, singular, plural, n)
}

// Ctl method for handling string translation with context
func (t *Translator) Ctl(loc Localizer, ctx, key string, args ...any) string {
	return t.ctl(loc, ctx, key, args...)
}

// Ctn method for handling plurals with context
func (t *Translator) Ctn(loc Localizer, ctx, singular, plural string, n int) string {
	return t.ctn(loc, ctx, singular, plural, n)
}

// Details return the header of the language file
func (t *Translator) Details(loc Localizer) map[string][]string {
	return t.languages[loc.GetLocale()].Headers
}
