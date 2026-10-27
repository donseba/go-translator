package translator

import (
	"os"
	"path/filepath"
	"time"

	"github.com/leonelquinteros/gotext"
)

// SetTL sets the translation of key in the catalogue of loc's language. It
// returns ErrorLanguageNotFound when that language is not loaded, as do the
// other setters.
func (t *Translator) SetTL(loc Localizer, key string, value string) error {
	translator, exists := t.languages[loc.GetLocale()]
	if !exists {
		return ErrorLanguageNotFound
	}

	translator.Set(key, value)
	return nil
}

// SetTLN sets, for the message key with plural, the form that the plural
// rules of the language choose for the count n.
func (t *Translator) SetTLN(loc Localizer, key, plural string, n int, value string) error {
	translator, exists := t.languages[loc.GetLocale()]
	if !exists {
		return ErrorLanguageNotFound
	}

	translator.SetN(key, plural, n, value)
	return nil
}

// SetCTL sets the translation of key in context ctx. It takes the key before
// the context, unlike Ctl.
func (t *Translator) SetCTL(loc Localizer, key, ctx, value string) error {
	translator, exists := t.languages[loc.GetLocale()]
	if !exists {
		return ErrorLanguageNotFound
	}

	translator.SetC(key, ctx, value)
	return nil
}

// SetCTN sets, for the message key with plural in context ctx, the form that
// the plural rules choose for the count n. It takes the key before the
// context, unlike Ctn.
func (t *Translator) SetCTN(loc Localizer, key, plural, ctx string, n int, value string) error {
	translator, exists := t.languages[loc.GetLocale()]
	if !exists {
		return ErrorLanguageNotFound
	}

	translator.SetNC(key, plural, ctx, n, value)
	return nil
}

// SetRefs sets the source references of key.
func (t *Translator) SetRefs(loc Localizer, key string, refs []string) error {
	translator, exists := t.languages[loc.GetLocale()]
	if !exists {
		return ErrorLanguageNotFound
	}

	translator.SetRefs(key, refs)
	return nil
}

// SetDetails sets the header key of the catalogue.
func (t *Translator) SetDetails(loc Localizer, key, value string) error {
	translator, exists := t.languages[loc.GetLocale()]
	if !exists {
		return ErrorLanguageNotFound
	}

	translator.GetDomain().Headers.Set(key, value)
	return nil
}

// AddDetails adds a value to the header key of the catalogue.
func (t *Translator) AddDetails(loc Localizer, key, value string) error {
	translator, exists := t.languages[loc.GetLocale()]
	if !exists {
		return ErrorLanguageNotFound
	}

	translator.GetDomain().Headers.Add(key, value)
	return nil
}

// Write saves the catalogue of loc's language as <locale>.po in the
// translations directory. It sets the PO-Revision-Date header when the
// catalogue has none.
func (t *Translator) Write(loc Localizer) error {
	translator, exists := t.languages[loc.GetLocale()]
	if !exists {
		return ErrorLanguageNotFound
	}

	if translator.GetDomain().Headers.Get("PO-Revision-Date") == "" {
		translator.GetDomain().Headers.Set("PO-Revision-Date", time.Now().Format(headerDateFormat))
	}

	data, err := translator.MarshalText()
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(t.translationsDir, loc.GetLocale()+DefaultPoExtension), data, 0o644)
	if err != nil {
		return err
	}

	return nil
}

// NewLanguage creates an empty catalogue for loc's language in memory, with
// the Language, Content-Type and X-Generator headers and then headers. It
// returns ErrorLanguageAlreadyExists when the language is loaded.
func (t *Translator) NewLanguage(loc Localizer, headers ...map[string]string) error {
	_, exists := t.languages[loc.GetLocale()]
	if exists {
		return ErrorLanguageAlreadyExists
	}

	if t.languages == nil {
		t.languages = make(map[string]*gotext.Po)
	}

	t.languages[loc.GetLocale()] = gotext.NewPo()
	t.languages[loc.GetLocale()].GetDomain().Headers.Set("Language", loc.GetLocale())
	t.languages[loc.GetLocale()].GetDomain().Headers.Set("Content-Type", "text/plain; charset=UTF-8")
	t.languages[loc.GetLocale()].GetDomain().Headers.Set("X-Generator", "github.com/donseba/go-translator")

	for _, header := range headers {
		for key, value := range header {
			t.languages[loc.GetLocale()].GetDomain().Headers.Set(key, value)
		}
	}

	return nil
}
