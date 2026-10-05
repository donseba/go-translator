package translator

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestMergeDefaultsPreservesTranslatedOverridesContextsAndEveryPluralForm(t *testing.T) {
	dir := t.TempDir()
	header := `msgid ""
msgstr ""
"Language: plural6\n"
"Content-Type: text/plain; charset=UTF-8\n"
"Plural-Forms: nplurals=6; plural=(n==1 ? 0 : n==0 ? 1 : n==2 ? 2 : n%100>=3 && n%100<=10 ? 3 : n%100>=11 && n%100<=99 ? 4 : 5);\n"

`
	original := []byte(header + `msgid "Overridden"
msgstr "Application value"

msgid "Blank"
msgstr ""

msgctxt "menu"
msgid "Open"
msgstr "Application context"
`)
	path := filepath.Join(dir, "plural6.po")
	if err := os.WriteFile(path, original, 0o600); err != nil {
		t.Fatal(err)
	}
	defaults := fstest.MapFS{"defaults.po": &fstest.MapFile{Data: []byte(header + `msgid "Overridden"
msgstr "Module value"

msgid "Blank"
msgstr "Module blank replacement"

msgid "Message"
msgid_plural "Messages"
msgstr[0] "one"
msgstr[1] "zero"
msgstr[2] "two"
msgstr[3] "few"
msgstr[4] "many"
msgstr[5] "other"

msgctxt "menu"
msgid "Open"
msgstr "Module context replacement"

msgctxt "action"
msgid "Open"
msgstr "New context"

msgctxt "inbox"
msgid "Message"
msgid_plural "Messages"
msgstr[0] "context one"
msgstr[1] "context zero"
msgstr[2] "context two"
msgstr[3] "context few"
msgstr[4] "context many"
msgstr[5] "context other"
`)}}
	if err := WritePOTFile(filepath.Join(dir, DefaultPotFile)); err != nil {
		t.Fatal(err)
	}
	tr := NewTranslator(dir, "")
	if err := tr.SetLanguage("plural6"); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err := tr.MergeDefaults("plural6", defaults, "defaults.po"); err != nil {
			t.Fatal(err)
		}
	}
	loc := mockLocalizer{locale: "plural6"}
	if tr.Tl(loc, "Overridden") != "Application value" || tr.Tl(loc, "Blank") != "Module blank replacement" || tr.Ctl(loc, "menu", "Open") != "Application context" || tr.Ctl(loc, "action", "Open") != "New context" {
		t.Fatal("merge replaced an application entry or lost a context", tr.Tl(loc, "Overridden"), tr.Tl(loc, "Blank"), tr.Ctl(loc, "menu", "Open"), tr.Ctl(loc, "action", "Open"))
	}
	for _, test := range []struct {
		count int
		want  string
	}{{0, "zero"}, {1, "one"}, {2, "two"}, {3, "few"}, {11, "many"}, {100, "other"}} {
		if got := tr.Tn(loc, "Message", "Messages", test.count); got != test.want {
			t.Errorf("plural %d: %q, want %q", test.count, got, test.want)
		}
		if got := tr.Ctn(loc, "inbox", "Message", "Messages", test.count); got != "context "+test.want {
			t.Errorf("context plural %d: %q", test.count, got)
		}
	}
	data, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(data, original) {
		t.Fatal("merge rewrote the application's PO file", err)
	}
	if err := tr.MergeDefaults("missing", defaults, "defaults.po"); !errors.Is(err, ErrorLanguageNotFound) {
		t.Fatal(err)
	}
	if err := tr.MergeDefaults("plural6", defaults, "missing.po"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatal(err)
	}
	invalid := fstest.MapFS{"invalid.po": &fstest.MapFile{Data: []byte("invalid")}}
	if err := tr.MergeDefaults("plural6", invalid, "invalid.po"); !errors.Is(err, ErrorInvalidCatalogue) {
		t.Fatal(err)
	}
}

func TestMergeDefaultsFillsBlankEntries(t *testing.T) {
	dir := t.TempDir()
	header := `msgid ""
msgstr ""
"Language: nl_NL\n"
"Content-Type: text/plain; charset=UTF-8\n"
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

`
	// This is what a catalogue looks like after "Update from POT": every new
	// message is present with an empty translation.
	original := header + `msgid "Save"
msgstr ""

msgid "Cancel"
msgstr "Afbreken"

msgid "Untranslated everywhere"
msgstr ""

msgctxt "menu"
msgid "Open"
msgstr ""

msgid "file"
msgid_plural "files"
msgstr[0] ""
msgstr[1] ""

msgid "page"
msgid_plural "pages"
msgstr[0] "pagina"
msgstr[1] ""

msgctxt "inbox"
msgid "message"
msgid_plural "messages"
msgstr[0] ""
msgstr[1] ""
`
	if err := os.WriteFile(filepath.Join(dir, "nl_NL.po"), []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := WritePOTFile(filepath.Join(dir, DefaultPotFile)); err != nil {
		t.Fatal(err)
	}

	defaults := fstest.MapFS{"defaults.po": &fstest.MapFile{Data: []byte(header + `msgid "Save"
msgstr "Opslaan"

msgid "Cancel"
msgstr "Annuleren"

msgid "Untranslated everywhere"
msgstr ""

msgctxt "menu"
msgid "Open"
msgstr "Openen"

msgid "file"
msgid_plural "files"
msgstr[0] "bestand"
msgstr[1] "bestanden"

msgid "page"
msgid_plural "pages"
msgstr[0] "standaardpagina"
msgstr[1] "standaardpagina's"

msgctxt "inbox"
msgid "message"
msgid_plural "messages"
msgstr[0] "bericht"
msgstr[1] "berichten"
`)}}

	tr := NewTranslator(dir, "")
	if err := tr.SetLanguage("nl_NL"); err != nil {
		t.Fatal(err)
	}

	loc := mockLocalizer{locale: "nl_NL"}

	// Before merging, blank entries fall back exactly like missing ones.
	if got := tr.Tl(loc, "Save"); got != "*Save*" {
		t.Fatalf("blank singular before merge: %q", got)
	}

	if got := tr.Tn(loc, "file", "files", 2); got != fmt.Sprintf(DefaultNoTranslationTN, "file", "files") {
		t.Fatalf("blank plural before merge: %q", got)
	}

	if err := tr.MergeDefaults("nl_NL", defaults, "defaults.po"); err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct {
		got  string
		want string
	}{
		{tr.Tl(loc, "Save"), "Opslaan"},
		{tr.Tl(loc, "Cancel"), "Afbreken"},
		{tr.Tl(loc, "Untranslated everywhere"), "*Untranslated everywhere*"},
		{tr.Ctl(loc, "menu", "Open"), "Openen"},
		{tr.Tn(loc, "file", "files", 1), "bestand"},
		{tr.Tn(loc, "file", "files", 2), "bestanden"},
		{tr.Ctn(loc, "inbox", "message", "messages", 1), "bericht"},
		{tr.Ctn(loc, "inbox", "message", "messages", 2), "berichten"},
		// A partly translated plural is the application's own work and wins;
		// its missing form falls back like a missing message.
		{tr.Tn(loc, "page", "pages", 1), "pagina"},
		{tr.Tn(loc, "page", "pages", 2), fmt.Sprintf(DefaultNoTranslationTN, "page", "pages")},
	} {
		if test.got != test.want {
			t.Errorf("got %q, want %q", test.got, test.want)
		}
	}

	data, err := os.ReadFile(filepath.Join(dir, "nl_NL.po"))
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != original {
		t.Fatal("merge rewrote the application's PO file")
	}
}
