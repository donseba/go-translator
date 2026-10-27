package translator

import (
	"slices"
	"testing"
	"testing/fstest"
)

// catalogue returns a catalogue in language with one message translated.
func catalogue(language, key, value string) *fstest.MapFile {
	return &fstest.MapFile{Data: []byte(`msgid ""
msgstr ""
"Language: ` + language + `\n"
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

msgid "` + key + `"
msgstr "` + value + `"
`)}
}

// TestFallbackOffByDefault checks that a lookup only uses the catalogue of
// exactly its locale by default.
func TestFallbackOffByDefault(t *testing.T) {
	tr := newTestTranslator(t, nil)

	if err := tr.LoadLanguages(fstest.MapFS{"nl.po": catalogue("nl", "Yes", "Ja")}); err != nil {
		t.Fatal(err)
	}

	if tr.LanguageFallback() {
		t.Error("fallback is on by default")
	}

	if got := tr.Tl(mockLocalizer{locale: "nl_NL"}, "Yes"); got != "*Yes*" {
		t.Errorf("got %q", got)
	}
}

// TestLanguageFallback checks the catalogues a lookup falls back to.
func TestLanguageFallback(t *testing.T) {
	tr := newTestTranslator(t, nil)
	files := fstest.MapFS{
		"nl.po":    catalogue("nl", "Yes", "Ja"),
		"nl_BE.po": catalogue("nl_BE", "Bicycle", "Velo"),
		"nl_NL.po": catalogue("nl_NL", "Bicycle", "Fiets"),
		"de_AT.po": catalogue("de_AT", "Yes", "Jo"),
		"de_CH.po": catalogue("de_CH", "Yes", "Joo"),
	}

	if err := tr.LoadLanguages(files); err != nil {
		t.Fatal(err)
	}

	tr.SetLanguageFallback(true)

	for _, test := range []struct {
		locale string
		key    string
		want   string
	}{
		{"nl_NL", "Bicycle", "Fiets"},
		{"nl_NL", "Yes", "Ja"},
		{"nl-nl", "Bicycle", "Fiets"},
		{"nl_NL.UTF-8", "Bicycle", "Fiets"},
		{"nl_BE", "Bicycle", "Velo"},
		{"nl_BE", "Yes", "Ja"},
		{"nl", "Yes", "Ja"},
		{"nl", "Bicycle", "Fiets"},
		{"nl_SR", "Bicycle", "Fiets"},
		{"de", "Yes", "Jo"},
		{"de_DE", "Yes", "Jo"},
		{"fr", "Yes", "*Yes*"},
	} {
		if got := tr.Tl(mockLocalizer{locale: test.locale}, test.key); got != test.want {
			t.Errorf("%s %q: got %q, want %q", test.locale, test.key, got, test.want)
		}
	}

	if text, found := tr.Lookup(mockLocalizer{locale: "nl_BE"}, "Yes"); text != "Ja" || !found {
		t.Errorf("Lookup: (%q, %v)", text, found)
	}

	tr.SetLanguageFallback(false)

	if got := tr.Tl(mockLocalizer{locale: "nl_BE"}, "Yes"); got != "*Yes*" {
		t.Errorf("off again: %q", got)
	}
}

// TestResolveOrder checks the order of the catalogues a locale tries.
func TestResolveOrder(t *testing.T) {
	tr := newTestTranslator(t, nil)

	for _, name := range []string{"nl", "nl_BE", "nl_NL", "nl-nl", "nl_AW", "en"} {
		if err := tr.NewLanguage(mockLocalizer{locale: name}); err != nil {
			t.Fatal(err)
		}
	}

	for _, test := range []struct {
		locale string
		want   []string
	}{
		{"nl_NL", []string{"nl_NL", "nl-nl", "nl", "nl_AW", "nl_BE"}},
		{"nl", []string{"nl", "nl-nl", "nl_NL", "nl_AW", "nl_BE"}},
		{"NL-be", []string{"nl_BE", "nl", "nl-nl", "nl_NL", "nl_AW"}},
		{"fr", nil},
	} {
		if got := tr.resolve(test.locale); !slices.Equal(got, test.want) {
			t.Errorf("%s: %v, want %v", test.locale, got, test.want)
		}
	}
}
