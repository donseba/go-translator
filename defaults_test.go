package translator

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestMergeDefaultsPreservesOverridesContextsAndEveryPluralForm(t *testing.T) {
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
	if tr.Tl(loc, "Overridden") != "Application value" || tr.Tl(loc, "Blank") != "*Blank*" || tr.Ctl(loc, "menu", "Open") != "Application context" || tr.Ctl(loc, "action", "Open") != "New context" {
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
