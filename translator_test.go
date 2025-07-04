package translator

import (
	"testing"
)

var (
	translationsDir = "test_translations"
	templateDir     = "test_templates"

	lang        = "en_US"
	langNL      = "nl_NL"
	multiplural = "multiplural"
)

func TestNewTranslator(t *testing.T) {
	translator := NewTranslator(translationsDir, templateDir)

	if translator.translationsDir != translationsDir {
		t.Errorf("NewTranslator() translationsDir = %v, want %v", translator.translationsDir, translationsDir)
	}

	if translator.templateDir != templateDir {
		t.Errorf("NewTranslator() templateDir = %v, want %v", translator.templateDir, templateDir)
	}

	if translator.languages == nil {
		t.Errorf("NewTranslator() languages map should not be nil")
	}

	if translator.uniqueKeys == nil {
		t.Errorf("NewTranslator() uniqueKeys map should not be nil")
	}
}

func TestSetAndGetPrefixSeparator(t *testing.T) {
	translator := NewTranslator(translationsDir, templateDir)

	defaultSeparator := "__."
	if sep := translator.PrefixSeparator(); sep != defaultSeparator {
		t.Errorf("PrefixSeparator() = %v, want %v", sep, defaultSeparator)
	}

	newSeparator := "--"
	translator.SetPrefixSeparator(newSeparator)
	if sep := translator.PrefixSeparator(); sep != newSeparator {
		t.Errorf("SetPrefixSeparator() = %v, want %v", sep, newSeparator)
	}
}

func TestSetLanguage(t *testing.T) {
	translator := NewTranslator(translationsDir, templateDir)

	err := translator.SetLanguage(lang)
	if err != nil {
		t.Errorf("SetLanguage() error = %v", err)
	}

	if _, exists := translator.languages[lang]; !exists {
		t.Errorf("SetLanguage() language %v not added", lang)
	}

	// Test with non-existent language file
	nonExistentLang := "non_existent"
	err = translator.SetLanguage(nonExistentLang)
	if err == nil {
		t.Errorf("SetLanguage() should return error for non-existent language %v", nonExistentLang)
	}

	if _, exists := translator.languages[nonExistentLang]; exists {
		t.Errorf("SetLanguage() should not add non-existent language %v", nonExistentLang)
	}
}

func TestTlFunction(t *testing.T) {
	translator := NewTranslator(translationsDir, templateDir)
	err := translator.SetLanguage(lang)
	if err != nil {
		t.Errorf("SetLanguage() error = %v", err)
	}

	loc := mockLocalizer{locale: lang} // Assume mockLocalizer implements Localizer interface

	tests := []struct {
		key      string
		expected string
	}{
		{"First translation text", "First translation text"},
		{"Second translation text", "Second translation text"},
		{"Third translation text", "Third translation text"},
		{"Non-existent text", "*Non-existent text*"}, // Assuming the behavior for non-translated text
	}

	for _, tt := range tests {
		result := translator.tl(loc, tt.key)
		if result != tt.expected {
			t.Errorf("tl() for key '%s' = %s, want %s", tt.key, result, tt.expected)
		}
	}
}

func TestTlFunctionWithArgs(t *testing.T) {
	translator := NewTranslator(translationsDir, templateDir)
	err := translator.SetLanguage(lang)
	if err != nil {
		t.Errorf("SetLanguage() error = %v", err)
	}

	loc := mockLocalizer{locale: lang}

	tests := []struct {
		key      string
		expected string
		args     []any
	}{
		{"First translation text with %s", "First translation text with args", []any{"args"}},
		{"Second translation text with %s", "Second translation text with args", []any{"args"}},
		{"Third translation text with %s", "Third translation text with args", []any{"args"}},
		{"Non-existent text with %s", "*Non-existent text with %s*", []any{"args"}}, // Assuming the behavior for non-translated text
	}

	for _, tt := range tests {
		result := translator.tl(loc, tt.key, tt.args...)
		if result != tt.expected {
			t.Errorf("tl() for key '%s' = %s, want %s", tt.key, result, tt.expected)
		}
	}
}

func TestTlFunctionNL(t *testing.T) {
	translator := NewTranslator(translationsDir, templateDir)
	err := translator.SetLanguage(langNL)
	if err != nil {
		t.Errorf("SetLanguage() error = %v", err)
	}

	loc := mockLocalizer{locale: langNL} // Assume mockLocalizer implements Localizer interface

	tests := []struct {
		key      string
		expected string
	}{
		{"First translation text", "Eerste vertalingstekst"},
		{"Second translation text", "Tweede vertalingstekst"},
		{"Third translation text", "Derde vertalingstekst"},
		{"Non-existent text", "*Non-existent text*"}, // Assuming the behavior for non-translated text
	}

	for _, tt := range tests {
		result := translator.tl(loc, tt.key)
		if result != tt.expected {
			t.Errorf("tl() for key '%s' = %s, want %s", tt.key, result, tt.expected)
		}
	}
}

func TestTnFunction(t *testing.T) {
	translator := NewTranslator("test_translations", "")
	err := translator.SetLanguage(lang)
	if err != nil {
		t.Errorf("SetLanguage() error = %v", err)
	}

	loc := mockLocalizer{locale: lang}

	// Assuming these translations are defined in your .po file
	singular := "There is one apple"
	plural := "There are many apples"

	// Singular case
	resultSingular := translator.tn(loc, singular, plural, 1)
	expectedSingular := "There is one apple" // expected translation for singular
	if resultSingular != expectedSingular {
		t.Errorf("tn() for 1 apple = %s, want %s", resultSingular, expectedSingular)
	}

	// Plural case
	resultPlural := translator.tn(loc, singular, plural, 5)
	expectedPlural := "There are many apples" // expected translation for plural
	if resultPlural != expectedPlural {
		t.Errorf("tn() for 5 apples = %s, want %s", resultPlural, expectedPlural)
	}
}

func TestTnFunctionMultiPlural(t *testing.T) {
	translator := NewTranslator("test_translations", "")
	err := translator.SetLanguage(multiplural)
	if err != nil {
		t.Errorf("SetLanguage() error = %v", err)
	}

	loc := mockLocalizer{locale: multiplural}

	tests := []struct {
		singular string
		plural   string
		count    int
		expected string
		vars     []any
	}{
		{"There is one apple", "There are many apples", 1, "C'è una mela", nil},
		{"There is one apple", "There are many apples", 2, "Ci sono due mele", nil},
		{"There is one apple", "There are %d apples", 0, "Ci sono 0 mele", []any{0}},
		{"There is one apple", "There are %d apples", 5, "Ci sono 5 mele", []any{5}},
		// Add more test cases as needed
	}

	for _, tt := range tests {
		result := translator.tn(loc, tt.singular, tt.plural, tt.count, tt.vars...)
		if result != tt.expected {
			t.Errorf("tn() with count %d = %s, want %s", tt.count, result, tt.expected)
		} else {
			t.Logf("tn() with count %d = %s", tt.count, result)
		}
	}
}

func TestCheckMissingTranslations(t *testing.T) {
	translator := NewTranslator(translationsDir, templateDir)
	err := translator.SetLanguage(lang)
	if err != nil {
		t.Errorf("SetLanguage() error = %v", err)
	}

	// Test with missing translations
	err = translator.CheckMissingTranslations()
	if err != nil {
		t.Error(err)
	}
}

// mockLocalizer is a mock implementation of the Localizer interface for testing
type mockLocalizer struct {
	locale string
}

func (m mockLocalizer) GetLocale() string {
	return m.locale
}
