package translator

import (
	"errors"
	"io/fs"
	"slices"
	"testing"
	"testing/fstest"
)

// TestSetLanguageFS checks loading a catalogue from an fs.FS.
func TestSetLanguageFS(t *testing.T) {
	tr := newTestTranslator(t, nil)
	files := fstest.MapFS{"locales/nl.po": &fstest.MapFile{Data: []byte(nlCatalogue)}}

	if err := tr.SetLanguageFS("nl", files, "locales/nl.po"); err != nil {
		t.Fatal(err)
	}

	if got := tr.Tl(mockLocalizer{locale: "nl"}, "Hello %s", "Bob"); got != "Hallo Bob" {
		t.Errorf("got %q", got)
	}

	err := tr.SetLanguageFS("de", files, "locales/de.po")
	if !errors.Is(err, ErrorLanguageNotFound) || !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("missing file: %v", err)
	}

	if _, loaded := tr.languages["de"]; loaded {
		t.Error("a missing catalogue was loaded")
	}
}

// TestCatalogueWithoutLanguageHeader checks that a catalogue without a
// Language header loads as the language it was loaded as.
func TestCatalogueWithoutLanguageHeader(t *testing.T) {
	catalogue := "msgid \"Yes\"\nmsgstr \"Ja\"\n"
	tr := newTestTranslator(t, map[string]string{"nl.po": catalogue})

	if err := tr.SetLanguage("nl"); err != nil {
		t.Fatal(err)
	}

	files := fstest.MapFS{"de.po": &fstest.MapFile{Data: []byte("msgid \"Yes\"\nmsgstr \"Ja\"\n")}}

	if err := tr.SetLanguageFS("de", files, "de.po"); err != nil {
		t.Fatal(err)
	}

	for _, locale := range []string{"nl", "de"} {
		if got := tr.Tl(mockLocalizer{locale: locale}, "Yes"); got != "Ja" {
			t.Errorf("%s: %q", locale, got)
		}

		if got := tr.languages[locale].Language; got != locale {
			t.Errorf("%s: language %q", locale, got)
		}
	}
}

// TestSetLanguageErrors checks the error of a catalogue that cannot be read.
func TestSetLanguageErrors(t *testing.T) {
	tr := newTestTranslator(t, nil)

	err := tr.SetLanguage("missing")
	if !errors.Is(err, ErrorLanguageNotFound) || !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("missing file: %v", err)
	}

	if len(tr.languages) != 0 {
		t.Error("a missing catalogue was loaded")
	}
}

// TestLoadLanguages checks loading every catalogue of a folder.
func TestLoadLanguages(t *testing.T) {
	tr := newTestTranslator(t, nil)
	files := fstest.MapFS{
		"nl.po":            &fstest.MapFile{Data: []byte(nlCatalogue)},
		"en_US.po":         &fstest.MapFile{Data: []byte("msgid \"\"\nmsgstr \"\"\n\"Language: en_US\\n\"\n")},
		"translations.pot": &fstest.MapFile{Data: []byte("msgid \"\"\nmsgstr \"\"\n")},
		"sub/de.po":        &fstest.MapFile{Data: []byte("msgid \"\"\nmsgstr \"\"\n")},
		"notes.txt":        &fstest.MapFile{Data: []byte("not a catalogue")},
	}

	if err := tr.LoadLanguages(files); err != nil {
		t.Fatal(err)
	}

	if got := sortedKeys(tr.languages); !slices.Equal(got, []string{"en_US", "nl"}) {
		t.Errorf("languages %v", got)
	}

	if got := tr.Tl(mockLocalizer{locale: "nl"}, "Hello %s", "Bob"); got != "Hallo Bob" {
		t.Errorf("got %q", got)
	}

	if err := tr.LoadLanguages(fstest.MapFS{}); err != nil {
		t.Errorf("empty folder: %v", err)
	}
}
