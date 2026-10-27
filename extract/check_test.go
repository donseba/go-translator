package extract_test

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/donseba/go-translator/v2/extract"
)

const dutch = `msgid ""
msgstr ""
"Language: nl\n"
"Content-Type: text/plain; charset=UTF-8\n"
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

msgid "Welcome"
msgstr "Welkom"

msgid "Hello, %s"
msgstr ""

msgid "%d item"
msgid_plural "%d items"
msgstr[0] "%d ding"
msgstr[1] ""

msgctxt "button"
msgid "Open"
msgstr "Openen"

msgid "Nested"
msgstr "Genest"

msgctxt "inbox"
msgid "%d message"
msgid_plural "%d messages"
msgstr[0] "%d bericht"
msgstr[1] "%d berichten"

msgid "Quote \" and\nnewline"
msgstr "Citaat \" en\nregel"

msgid "No longer used"
msgstr "Niet meer gebruikt"
`

const english = `msgid ""
msgstr ""
"Language: en\n"
"Content-Type: text/plain; charset=UTF-8\n"
`

// keys lists messages as context|text.
func keys(messages []extract.Message) string {
	var keys []string

	for _, message := range messages {
		keys = append(keys, message.Context+"|"+message.ID)
	}

	return strings.Join(keys, " ")
}

func TestCheck(t *testing.T) {
	catalog, err := extract.Extract(sources()...)
	if err != nil {
		t.Fatal(err)
	}

	files := fstest.MapFS{
		"nl.po":            {Data: []byte(dutch)},
		"en.po":            {Data: []byte(english)},
		"translations.pot": {Data: []byte("ignored")},
		"sub/de.po":        {Data: []byte("not in the root")},
	}

	reports, err := extract.Check(catalog, files)
	if err != nil {
		t.Fatal(err)
	}

	if len(reports) != 2 || reports[0].File != "en.po" || reports[1].File != "nl.po" {
		t.Fatalf("reports = %+v, want en.po and nl.po", reports)
	}

	en := reports[0]

	if len(en.Missing) != len(catalog.Messages) || len(en.Untranslated) != len(catalog.Messages) || len(en.Empty) != 0 {
		t.Errorf("en.po misses %d, want all %d", len(en.Missing), len(catalog.Messages))
	}

	if en.Language != "en" || en.Forms != 2 || en.Complete() {
		t.Errorf("en.po = %+v", en)
	}

	nl := reports[1]

	if got, want := keys(nl.Untranslated), "|Hello, %s |%d item door|Open"; got != want {
		t.Errorf("nl.po untranslated %q, want %q", got, want)
	}

	if got, want := keys(nl.Missing), "door|Open"; got != want {
		t.Errorf("nl.po missing %q, want %q", got, want)
	}

	if got, want := keys(nl.Empty), "|Hello, %s |%d item"; got != want {
		t.Errorf("nl.po empty %q, want %q", got, want)
	}

	if nl.Untranslated[1].References[0] != "default/layout.gohtml:4" {
		t.Errorf("references = %q, want those of the catalog", nl.Untranslated[1].References)
	}
}

// TestCheckPluralForms checks that a plural message needs every form of its
// language, and no more.
func TestCheckPluralForms(t *testing.T) {
	var catalog extract.Catalog

	catalog.Add(
		extract.Message{ID: "%d file", Plural: "%d files"},
		extract.Message{ID: "Save"},
	)

	polish := `msgid ""
msgstr ""
"Plural-Forms: nplurals=3; plural=(n==1 ? 0 : n%10>=2 && n%10<=4 && (n%100<10 || n%100>=20) ? 1 : 2);\n"

msgid "%d file"
msgid_plural "%d files"
msgstr[0] "%d plik"
msgstr[1] "%d pliki"
msgstr[2] ""

msgid "Save"
msgstr "Zapisz"
`

	polishComplete := strings.Replace(polish, `msgstr[2] ""`, `msgstr[2] "%d plików"`, 1)

	japanese := `msgid ""
msgstr ""
"Language: ja\n"
"Plural-Forms: nplurals=1; plural=0;\n"

msgid "%d file"
msgid_plural "%d files"
msgstr[0] "%d ファイル"

msgid "Save"
msgstr "保存"
`

	files := fstest.MapFS{
		"pl.po":    {Data: []byte(polish)},
		"pl_PL.po": {Data: []byte(polishComplete)},
		"ja.po":    {Data: []byte(japanese)},
	}

	reports, err := extract.Check(&catalog, files)
	if err != nil {
		t.Fatal(err)
	}

	byFile := map[string]extract.Report{}

	for _, report := range reports {
		byFile[report.File] = report
	}

	pl := byFile["pl.po"]

	if pl.Forms != 3 || pl.Language != "pl" || keys(pl.Empty) != "|%d file" || len(pl.Missing) != 0 {
		t.Errorf("pl.po = %+v, want %%d file empty in its third form", pl)
	}

	if report := byFile["pl_PL.po"]; !report.Complete() || report.Language != "pl_PL" {
		t.Errorf("pl_PL.po = %+v, want complete", report)
	}

	if report := byFile["ja.po"]; !report.Complete() || report.Forms != 1 || report.Language != "ja" {
		t.Errorf("ja.po = %+v, want complete with one form", report)
	}
}

func TestCheckWithoutCatalogOrFiles(t *testing.T) {
	reports, err := extract.Check(nil, fstest.MapFS{"nl.po": {Data: []byte(dutch)}})
	if err != nil {
		t.Fatal(err)
	}

	if len(reports) != 1 || !reports[0].Complete() || reports[0].Language != "nl" {
		t.Errorf("reports = %+v, want one complete", reports)
	}

	reports, err = extract.Check(&extract.Catalog{}, fstest.MapFS{})
	if err != nil || len(reports) != 0 {
		t.Errorf("Check without PO files = %+v, %v", reports, err)
	}

	if _, err := extract.Check(&extract.Catalog{}, nil); err == nil {
		t.Error("Check(nil file system) gave no error")
	}
}
