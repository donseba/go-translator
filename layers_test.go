package translator

import (
	"errors"
	"io/fs"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
)

// poCatalogue builds a PO catalogue: a header with language (none when empty)
// and English plural rules, then msgid/msgstr pairs.
func poCatalogue(language string, pairs ...string) string {
	var b strings.Builder

	b.WriteString("msgid \"\"\nmsgstr \"\"\n")

	if language != "" {
		b.WriteString("\"Language: " + language + "\\n\"\n")
	}

	b.WriteString("\"Plural-Forms: nplurals=2; plural=(n != 1);\\n\"\n")

	for i := 0; i+1 < len(pairs); i += 2 {
		b.WriteString("\nmsgid \"" + pairs[i] + "\"\nmsgstr \"" + pairs[i+1] + "\"\n")
	}

	return b.String()
}

// TestLoadLanguagesLayered checks that the first source with a catalogue
// loads it and later sources fill in what it misses.
func TestLoadLanguagesLayered(t *testing.T) {
	site := fstest.MapFS{
		"nl.po": {Data: []byte(poCatalogue("nl", "Save", "Bewaren", "Delete", ""))},
	}

	admin := fstest.MapFS{
		"nl.po":    {Data: []byte(poCatalogue("nl", "Save", "Opslaan", "Delete", "Verwijderen", "Edit", "Bewerken"))},
		"de.po":    {Data: []byte(poCatalogue("de", "Save", "Speichern"))},
		"nl.pot":   {Data: []byte("not a catalogue")},
		"sub/x.po": {Data: []byte("not in the root")},
	}

	section := fstest.MapFS{
		"nl.po": {Data: []byte(poCatalogue("", "Edit", "Wijzigen", "Publish", "Publiceren"))},
		"de.po": {Data: []byte(poCatalogue("de_DE", "Save", "Sichern", "Publish", "Veröffentlichen"))},
	}

	// A source without the file is skipped, as a nil source is.
	empty := fstest.MapFS{}

	tr := New()

	if err := tr.LoadLanguages(site, nil, empty, admin, section); err != nil {
		t.Fatal(err)
	}

	if got := sortedKeys(tr.languages); !slices.Equal(got, []string{"de", "nl"}) {
		t.Fatalf("languages = %v", got)
	}

	nl := mockLocalizer{locale: "nl"}
	de := mockLocalizer{locale: "de"}

	for _, test := range []struct {
		loc        Localizer
		key, want  string
		translated bool
	}{
		{nl, "Save", "Bewaren", true},       // the site wins
		{nl, "Delete", "Verwijderen", true}, // an empty entry is filled
		{nl, "Edit", "Bewerken", true},      // the admin wins over the section
		{nl, "Publish", "Publiceren", true}, // a header-less file counts as nl
		{de, "Save", "Speichern", true},     // loaded from the admin
		{de, "Publish", "Veröffentlichen", true},
		{nl, "Unknown", "*Unknown*", false},
	} {
		got, ok := tr.Lookup(test.loc, test.key)
		if got != test.want || ok != test.translated {
			t.Errorf("%s %q = %q, %v; want %q, %v", test.loc.GetLocale(), test.key, got, ok, test.want, test.translated)
		}
	}
}

// TestLoadLanguagesLayeredReplaces checks that a catalogue loaded before is
// replaced, as with one source.
func TestLoadLanguagesLayeredReplaces(t *testing.T) {
	tr := New()

	if err := tr.LoadLanguages(fstest.MapFS{"nl.po": {Data: []byte(poCatalogue("nl", "Old", "Oud"))}}); err != nil {
		t.Fatal(err)
	}

	if err := tr.LoadLanguages(fstest.MapFS{"nl.po": {Data: []byte(poCatalogue("nl", "New", "Nieuw"))}}); err != nil {
		t.Fatal(err)
	}

	if _, ok := tr.Lookup(mockLocalizer{locale: "nl"}, "Old"); ok {
		t.Error("the earlier catalogue was kept")
	}

	if err := tr.LoadLanguages(); err != nil {
		t.Errorf("LoadLanguages without sources: %v", err)
	}
}

// TestLoadLanguagesLayeredErrors checks that a later catalogue in another
// language, or with other plural rules, is an error naming the file.
func TestLoadLanguagesLayeredErrors(t *testing.T) {
	site := fstest.MapFS{"nl.po": {Data: []byte(poCatalogue("nl", "Save", "Bewaren"))}}

	german := fstest.MapFS{"nl.po": {Data: []byte(poCatalogue("de", "Save", "Speichern"))}}

	err := New().LoadLanguages(site, german)
	if !errors.Is(err, ErrorInvalidCatalogue) || !strings.Contains(err.Error(), "nl.po") {
		t.Errorf("another language: %v", err)
	}

	plurals := fstest.MapFS{"nl.po": {Data: []byte(`msgid ""
msgstr ""
"Language: nl\n"
"Plural-Forms: nplurals=3; plural=(n==1 ? 0 : n==2 ? 1 : 2);\n"

msgid "%d file"
msgid_plural "%d files"
msgstr[0] "%d bestand"
msgstr[1] "%d bestanden"
msgstr[2] "%d bestanden"
`)}}

	err = New().LoadLanguages(site, plurals)
	if !errors.Is(err, ErrorInvalidCatalogue) {
		t.Errorf("other plural rules: %v", err)
	}

	unreadable := fstest.MapFS{"nl.po": {Mode: fs.ModeDir}}

	if err := New().LoadLanguages(site, unreadable); err == nil {
		t.Error("an unreadable later file gave no error")
	}
}
