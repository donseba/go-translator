package translator

import (
	"fmt"
)

// Tl translates a string based on the given language tag and key.
func (t *Translator) tl(loc Localizer, key string, args ...interface{}) string {
	translator, exists := t.languages[loc.GetLocale()]
	if !exists {
		return fmt.Sprintf(DefaultNoTranslationTL, key)
	}

	if !translator.IsTranslated(key) {
		// check if translation is in the pot file, if not, add it
		if _, ok := t.uniqueKeys[key]; !ok {
			t.uniqueKeys[key] = uniqueKey{singular: key}

			err := t.addToPotFileIfNotExists(translationKey{"", key, false})
			if err != nil {
				fmt.Println(err)
			}
		}
		return fmt.Sprintf(DefaultNoTranslationTL, key)
	}

	translated := translator.Get(key, args...)
	return t.removePrefix(translated)
}

func (t *Translator) ctl(loc Localizer, ctx, key string, args ...interface{}) string {
	translator, exists := t.languages[loc.GetLocale()]
	if !exists {
		return fmt.Sprintf(DefaultNoTranslationCTL, ctx, key)
	}

	if ctx == "" {
		return t.tl(loc, key, args...)
	}

	if !translator.IsTranslatedC(key, ctx) {
		if t.uniqueKeysCtx[ctx] == nil {
			t.uniqueKeysCtx[ctx] = make(map[string]uniqueKey)
		}

		if _, ok := t.uniqueKeysCtx[ctx][key]; !ok {
			t.uniqueKeysCtx[ctx][key] = uniqueKey{singular: key}

			err := t.addToPotFileIfNotExists(translationKey{ctx, key, false})
			if err != nil {
				fmt.Println(err)
			}
		}

		return fmt.Sprintf(DefaultNoTranslationCTL, ctx, key)
	}

	translated := translator.GetC(key, ctx, args...)
	return t.removePrefix(translated)
}

// tn method for handling plurals
func (t *Translator) tn(loc Localizer, singular, plural string, n int, args ...any) string {
	translator, exists := t.languages[loc.GetLocale()]
	if !exists {
		return fmt.Sprintf(DefaultNoTranslationTN, singular, plural)
	}

	if !translator.IsTranslatedN(singular, n) {
		// check if translation is in the pot file, if not, add it
		if _, ok := t.uniqueKeys[singular]; !ok {
			t.uniqueKeys[singular] = uniqueKey{singular: singular, plural: plural}

			err := t.addToPotFileIfNotExists(translationKey{"", singular, true})
			if err != nil {
				fmt.Println(err)
			}
		}
		return fmt.Sprintf(DefaultNoTranslationTN, singular, plural)
	}

	translated := translator.GetN(singular, plural, n, args...)
	return t.removePrefix(translated)
}

func (t *Translator) ctn(loc Localizer, ctx, singular, plural string, n int, args ...any) string {
	translator, exists := t.languages[loc.GetLocale()]
	if !exists {
		return fmt.Sprintf(DefaultNoTranslationCTN, ctx, singular, plural)
	}

	if !translator.IsTranslatedNC(singular, n, ctx) {
		if t.uniqueKeysCtx[ctx] == nil {
			t.uniqueKeysCtx[ctx] = make(map[string]uniqueKey)
		}

		if _, ok := t.uniqueKeysCtx[ctx][singular]; !ok {
			t.uniqueKeysCtx[ctx][singular] = uniqueKey{singular: singular, plural: plural}

			err := t.addToPotFileIfNotExists(translationKey{ctx, singular, true})
			if err != nil {
				fmt.Println(err)
			}
		}

		return fmt.Sprintf(DefaultNoTranslationCTN, ctx, singular, plural)
	}

	translated := translator.GetNC(singular, plural, n, ctx, args...)
	return t.removePrefix(translated)
}

// Tl translates a string based on the given language tag and key.
func (t *Translator) Tl(loc Localizer, key string, args ...interface{}) string {
	return t.tl(loc, key, args...)
}

// Tn method for handling plurals
func (t *Translator) Tn(loc Localizer, singular, plural string, n int) string {
	return t.tn(loc, singular, plural, n)
}

// Ctl method for handling string translation with context
func (t *Translator) Ctl(loc Localizer, ctx, key string, args ...interface{}) string {
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
