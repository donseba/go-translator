package translator

import (
	"fmt"
	"github.com/donseba/gotext"
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
		languages       map[string]*gotext.Po
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

			tls := tr.languages[tt.lang].GetTranslations()
			for _, tl := range tls {
				fmt.Printf("%+v\n", tl)
			}

			err = tr.Write(tt.loc)
			if err != nil {
				t.Errorf("error = %v", err)
			}
		})
	}

}

func NewLocalizer(s string) Localizer {
	return mockLocalizer{locale: s}
}
