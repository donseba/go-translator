package translator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var (
	testSetLanguageTL = "en_TL"
	testSetLanguageTN = "en_TN"
)

func TestTranslator_SetTL(t *testing.T) {
	type fields struct {
		translationsDir string
		templateDir     string
		uniqueKeys      map[string]uniqueKey
		uniqueKeysCtx   map[string]map[string]uniqueKey
	}
	type args struct {
		key    string
		plural string
		n      int
		value  string
	}

	localizerTL := NewLocalizer(testSetLanguageTL)
	localizerTN := NewLocalizer(testSetLanguageTN)

	tests := []struct {
		name    string
		mode    string
		loc     Localizer
		lang    string
		fields  fields
		args    []args
		wantErr bool
	}{
		{
			name: "Test SetTL",
			mode: "SetTL",
			lang: testSetLanguageTL,
			fields: fields{
				translationsDir: "test_translations",
				templateDir:     "test_templates",
				uniqueKeys:      map[string]uniqueKey{},
				uniqueKeysCtx:   map[string]map[string]uniqueKey{},
			},
			loc: localizerTL,
			args: []args{
				{key: "test_key", value: "test_value"},
				{key: "test_key2", value: "test_value2"},
				{key: "test_key3", value: "test_value3"},
			},
			wantErr: false,
		},
		{
			name: "Test SetTN",
			mode: "SetTN",
			lang: testSetLanguageTN,
			fields: fields{
				translationsDir: "test_translations",
				templateDir:     "test_templates",
				uniqueKeys:      map[string]uniqueKey{},
				uniqueKeysCtx:   map[string]map[string]uniqueKey{},
			},
			loc: localizerTN,
			args: []args{
				{key: "test_key", plural: "test_plural", n: 0, value: "test_value 0"},
				{key: "test_key", plural: "test_plural", n: 1, value: "test_value 1"},
				{key: "test_key", plural: "test_plural", n: 2, value: "test_value 2"},
				{key: "test_key", plural: "test_plural", n: 3, value: "test_value 2"},
				{key: "test_key2", plural: "test_plural2", n: 2, value: "test_value2"},
				{key: "test_key3", plural: "test_plural3", n: 3, value: "test_value3"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Translator{
				translationsDir: tt.fields.translationsDir,
				templateDir:     tt.fields.templateDir,
				uniqueKeys:      tt.fields.uniqueKeys,
				uniqueKeysCtx:   tt.fields.uniqueKeysCtx,
			}

			err := tr.NewLanguage(tt.loc, map[string]string{
				"Language-Team":      fmt.Sprintf("%s %s", "English", tt.mode),
				"Last-Translator:":   "Don Seba <donseba[at]someemail.com>",
				"Language":           tt.lang,
				"Content-Type":       "text/plain; charset=UTF-8",
				"Plural-Forms":       "nplurals=2; plural=(n != 1)",
				"X-Generator":        "Poedit 2.2.4",
				"Project-Id-Version": "TEST VERSION",
			})

			if err != nil {
				t.Errorf("error = %v", err)
			}

			for _, arg := range tt.args {
				switch tt.mode {
				case "SetTL":
					if err := tr.SetTL(tt.loc, arg.key, arg.value); (err != nil) != tt.wantErr {
						t.Errorf("Translator.SetTL() error = %v, wantErr %v", err, tt.wantErr)
					}
				case "SetTN":
					if err := tr.SetTLN(tt.loc, arg.key, arg.plural, arg.n, arg.value); (err != nil) != tt.wantErr {
						t.Errorf("Translator.SetTL() error = %v, wantErr %v", err, tt.wantErr)
					}
				}
			}

			//tls := tr.languages[tt.lang].GetDomain().GetTranslations()
			//for _, tl := range tls {
			//	fmt.Printf("%+v\n", tl)
			//}

			err = tr.Write(tt.loc)
			if err != nil {
				t.Errorf("error = %v", err)
			}
		})
	}

}

func TestGeneratedPluralForms(t *testing.T) {
	cases := []struct {
		lang   string
		expect string
	}{
		{"en", "nplurals=2; plural=(n != 1);"},
		{"ar", "nplurals=6; plural=(n == 0 ? 0 : n == 1 ? 1 : n == 2 ? 2 : n % 100 >= 3 && n % 100 <= 10 ? 3 : n % 100 >= 11 && n % 100 <= 99 ? 4 : 5);"},
		{"ru", "nplurals=3; plural=(n % 10 == 1 && n % 100 != 11 ? 0 : n % 10 >= 2 && n % 10 <= 4 && (n % 100 < 12 || n % 100 > 14) ? 1 : 2);"},
	}
	for _, c := range cases {
		h, ok := LanguageHeaderTemplates[c.lang]
		if !ok {
			t.Errorf("Language %s not found in LanguageHeaderTemplates", c.lang)
			continue
		}
		if h.PluralForms != c.expect {
			t.Errorf("PluralForms for %s: got %q, want %q", c.lang, h.PluralForms, c.expect)
		}
	}
}

func TestEnsureLanguage(t *testing.T) {
	tempDir := t.TempDir()
	tr := NewTranslator(tempDir, "")
	lang := "fr"
	poPath := filepath.Join(tempDir, lang+DefaultPoExtension)

	// Ensure file does not exist
	if _, err := os.Stat(poPath); err == nil {
		os.Remove(poPath)
	}

	t.Cleanup(func() {
		if err := os.Remove(poPath); err != nil && !os.IsNotExist(err) {
			t.Errorf("Failed to clean up PO file: %v", err)
		}
	})

	err := tr.EnsureLanguage(lang)
	if err != nil {
		t.Fatalf("EnsureLanguage failed: %v", err)
	}

	// Check file was created
	if _, err := os.Stat(poPath); err != nil {
		t.Errorf("PO file was not created: %v", err)
	}

	// Check language is loaded
	if _, ok := tr.languages[lang]; !ok {
		t.Errorf("Language %s not loaded in translator", lang)
	}
}

func TestEnsureLanguage_Idempotent(t *testing.T) {
	tempDir := t.TempDir()
	tr := NewTranslator(tempDir, "")
	lang := "fr"
	poPath := filepath.Join(tempDir, lang+DefaultPoExtension)

	// Ensure file does not exist
	if _, err := os.Stat(poPath); err == nil {
		os.Remove(poPath)
	}

	t.Cleanup(func() {
		if err := os.Remove(poPath); err != nil && !os.IsNotExist(err) {
			t.Errorf("Failed to clean up PO file: %v", err)
		}
	}

	// First call should create the file
	err := tr.EnsureLanguage(lang)
	if err != nil {
		t.Fatalf("First EnsureLanguage failed: %v", err)
	}

	// Write a marker to the file
	marker := "# marker line\n"
	f, err := os.OpenFile(poPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("Failed to open PO file for appending: %v", err)
	}
	_, err = f.WriteString(marker)
	f.Close()
	if err != nil {
		t.Fatalf("Failed to write marker: %v", err)
	}

	// Second call should NOT overwrite the file
	err = tr.EnsureLanguage(lang)
	if err != nil {
		t.Fatalf("Second EnsureLanguage failed: %v", err)
	}

	// Check that marker is still present
	data, err := os.ReadFile(poPath)
	if err != nil {
		t.Fatalf("Failed to read PO file: %v", err)
	}
	if !contains(string(data), marker) {
		t.Errorf("Marker line was overwritten by EnsureLanguage: %s", string(data))
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func NewLocalizer(s string) Localizer {
	return mockLocalizer{locale: s}
}
