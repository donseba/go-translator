package translator

import (
	"fmt"
)

// Tl translates a string based on the given language tag and key.
func (t *Translator) tl(loc Localizer, key string, args ...any) string {
	// Always add to POT if not exists
	if _, ok := t.uniqueKeys[key]; !ok {
		t.uniqueKeys[key] = uniqueKey{singular: key}
		err := t.addToPotFileIfNotExists(translationKey{"", key, false})
		if err != nil {
			fmt.Println(err)
		}
	}

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
	// Always add to POT if not exists
	if _, ok := t.uniqueKeys[singular]; !ok {
		t.uniqueKeys[singular] = uniqueKey{singular: singular, plural: plural}
		err := t.addToPotFileIfNotExists(translationKey{"", singular, true})
		if err != nil {
			fmt.Println(err)
		}
	}

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
