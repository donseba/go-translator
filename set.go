package translator

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/leonelquinteros/gotext"
)

func (t *Translator) SetTL(loc Localizer, key string, value string) error {
	translator, exists := t.languages[loc.GetLocale()]
	if !exists {
		return ErrorLanguageNotFound
	}

	translator.Set(key, value)
	return nil
}

func (t *Translator) SetTLN(loc Localizer, key, plural string, n int, value string) error {
	translator, exists := t.languages[loc.GetLocale()]
	if !exists {
		return ErrorLanguageNotFound
	}

	translator.SetN(key, plural, n, value)
	return nil
}

func (t *Translator) SetCTL(loc Localizer, key, ctx, value string) error {
	translator, exists := t.languages[loc.GetLocale()]
	if !exists {
		return ErrorLanguageNotFound
	}

	translator.SetC(key, ctx, value)
	return nil
}

func (t *Translator) SetCTN(loc Localizer, key, plural, ctx string, n int, value string) error {
	translator, exists := t.languages[loc.GetLocale()]
	if !exists {
		return ErrorLanguageNotFound
	}

	translator.SetNC(key, plural, ctx, n, value)
	return nil
}

func (t *Translator) SetRefs(loc Localizer, key string, refs []string) error {
	translator, exists := t.languages[loc.GetLocale()]
	if !exists {
		return ErrorLanguageNotFound
	}

	translator.SetRefs(key, refs)
	return nil
}

func (t *Translator) SetDetails(loc Localizer, key, value string) error {
	translator, exists := t.languages[loc.GetLocale()]
	if !exists {
		return ErrorLanguageNotFound
	}

	translator.GetDomain().Headers.Set(key, value)
	return nil
}

func (t *Translator) AddDetails(loc Localizer, key, value string) error {
	translator, exists := t.languages[loc.GetLocale()]
	if !exists {
		return ErrorLanguageNotFound
	}

	translator.GetDomain().Headers.Add(key, value)
	return nil
}

func (t *Translator) Write(loc Localizer) error {
	translator, exists := t.languages[loc.GetLocale()]
	if !exists {
		return ErrorLanguageNotFound
	}

	if translator.GetDomain().Headers.Get("PO-Revision-Date") == "" {
		translator.GetDomain().Headers.Set("PO-Revision-Date", time.Now().Format(time.RFC3339))
	}

	data, err := translator.MarshalText()
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(t.translationsDir, fmt.Sprintf("%s.po", loc.GetLocale())), data, 0644)
	if err != nil {
		return err
	}

	return nil
}

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
	t.languages[loc.GetLocale()].GetDomain().Headers.Set("X-Generator:", "github.com/donseba/go-translator")

	for _, header := range headers {
		for key, value := range header {
			t.languages[loc.GetLocale()].GetDomain().Headers.Set(key, value)
		}
	}

	return nil
}
