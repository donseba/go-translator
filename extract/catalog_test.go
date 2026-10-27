package extract_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/donseba/go-translator/v2/extract"
)

func TestAdd(t *testing.T) {
	catalog, err := extract.Extract(extract.Source{Name: "a.gohtml", Text: `{{tl localizer "Save"}}{{tl localizer "%d file"}}`})
	if err != nil {
		t.Fatal(err)
	}

	references := []string{"go:handlers"}

	catalog.Add(
		extract.Message{ID: "Save", References: references},
		extract.Message{ID: "%d file", Plural: "%d files"},
		extract.Message{ID: "Not found", References: references},
		extract.Message{Context: "verb", ID: "Save"},
		extract.Message{ID: "Not found", References: []string{"go:handlers", "go:pages"}},
		extract.Message{ID: "%d file", Plural: "%d documents", References: []string{"go:other"}},
		extract.Message{ID: "", References: []string{"go:empty"}},
	)

	want := []extract.Message{
		{ID: "Save", References: []string{"a.gohtml:1", "go:handlers"}},
		{ID: "%d file", Plural: "%d files", References: []string{"a.gohtml:1", "go:other"}},
		{ID: "Not found", References: []string{"go:handlers", "go:pages"}},
		{Context: "verb", ID: "Save"},
	}

	sameMessages(t, catalog.Messages, want)

	if len(catalog.Warnings) != 2 {
		t.Fatalf("warnings = %q, want the conflicting plural and the empty text", catalog.Warnings)
	}

	if catalog.Warnings[0].Reference != "go:other" || !strings.Contains(catalog.Warnings[0].Problem, "%d documents") {
		t.Errorf("warning = %+v, want the conflicting plural", catalog.Warnings[0])
	}

	if catalog.Warnings[1].Reference != "go:empty" {
		t.Errorf("warning = %+v, want the empty text", catalog.Warnings[1])
	}

	// The catalog has its own references: changing the caller's slice
	// changes nothing.
	references[0] = "changed"

	if catalog.Messages[2].References[0] != "go:handlers" {
		t.Errorf("references = %q, share the caller's slice", catalog.Messages[2].References)
	}
}

func TestAddToEmptyCatalog(t *testing.T) {
	var catalog extract.Catalog

	catalog.Add(extract.Texts("core:texts", "Not found", "Forbidden", "Not found")...)
	catalog.Add(extract.Texts("", "Read more")...)

	want := []extract.Message{
		{ID: "Not found", References: []string{"core:texts"}},
		{ID: "Forbidden", References: []string{"core:texts"}},
		{ID: "Read more"},
	}

	sameMessages(t, catalog.Messages, want)

	if catalog.Messages[2].References != nil {
		t.Errorf("references = %q, want none", catalog.Messages[2].References)
	}
}

func TestMessageString(t *testing.T) {
	for _, test := range []struct {
		message extract.Message
		want    string
	}{
		{extract.Message{ID: "Save"}, `"Save"`},
		{extract.Message{ID: "%d file", Plural: "%d files", References: []string{"a:1", "b:2"}}, `"%d file" / "%d files" (a:1)`},
		{extract.Message{Context: "door", ID: "Open", References: []string{"a:3"}}, `[door] "Open" (a:3)`},
	} {
		if got := test.message.String(); got != test.want {
			t.Errorf("String() = %s, want %s", got, test.want)
		}
	}
}

func TestWarningString(t *testing.T) {
	if got := (extract.Warning{Reference: "a:1", Problem: "oops"}).String(); got != "a:1: oops" {
		t.Errorf("String() = %q", got)
	}

	if got := (extract.Warning{Problem: "oops"}).String(); got != "oops" {
		t.Errorf("String() without reference = %q", got)
	}
}

func TestWritePOT(t *testing.T) {
	catalog, err := extract.Extract(sources()...)
	if err != nil {
		t.Fatal(err)
	}

	var b bytes.Buffer

	if err := catalog.WritePOT(&b); err != nil {
		t.Fatal(err)
	}

	pot := b.String()

	for _, want := range []string{
		"msgid \"\"\nmsgstr \"\"\n",
		"\"Content-Type: text/plain; charset=UTF-8\\n\"\n",
		"\"X-Generator: go-translator\\n\"\n",
		"\n#: default/layout.gohtml:1\n#: default/page.gohtml:1\nmsgid \"Welcome\"\nmsgstr \"\"\n",
		"msgid \"%d item\"\nmsgid_plural \"%d items\"\nmsgstr[0] \"\"\nmsgstr[1] \"\"\n",
		"msgctxt \"button\"\nmsgid \"Open\"\nmsgstr \"\"\n",
		"msgctxt \"door\"\nmsgid \"Open\"\nmsgstr \"\"\n",
		"msgctxt \"inbox\"\nmsgid \"%d message\"\nmsgid_plural \"%d messages\"\n",
		"msgid \"Quote \\\" and\\nnewline\"\n",
	} {
		if !strings.Contains(pot, want) {
			t.Errorf("POT misses %q:\n%s", want, pot)
		}
	}

	if strings.Count(pot, "msgid \"Welcome\"") != 1 {
		t.Errorf("Welcome is not written once:\n%s", pot)
	}

	if strings.Contains(pot, "Creation-Date") || strings.HasPrefix(pot, "#") {
		t.Errorf("POT has a creation date or a comment:\n%s", pot)
	}
}

// TestWritePOTHeader checks the whole header, as go-webthings writes it.
func TestWritePOTHeader(t *testing.T) {
	catalog := extract.Catalog{
		Header: extract.Header{
			Comment: "The texts of this website's templates, extracted by WebThings.\n" +
				"Do not edit: run the extraction again.\n\nSecond paragraph.\n",
			Generator: "WebThings",
		},
	}

	catalog.Add(extract.Message{ID: "Hi"})

	var b bytes.Buffer

	if err := catalog.WritePOT(&b); err != nil {
		t.Fatal(err)
	}

	want := `# The texts of this website's templates, extracted by WebThings.
# Do not edit: run the extraction again.
#
# Second paragraph.
msgid ""
msgstr ""
"Project-Id-Version: \n"
"MIME-Version: 1.0\n"
"Content-Type: text/plain; charset=UTF-8\n"
"Content-Transfer-Encoding: 8bit\n"
"X-Generator: WebThings\n"

msgid "Hi"
msgstr ""
`

	if got := b.String(); got != want {
		t.Errorf("POT =\n%s\nwant\n%s", got, want)
	}

	catalog.Header = extract.Header{ProjectID: "shop 1.0"}
	b.Reset()

	if err := catalog.WritePOT(&b); err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(b.String(), "msgid \"\"\nmsgstr \"\"\n\"Project-Id-Version: shop 1.0\\n\"\n") {
		t.Errorf("POT =\n%s", b.String())
	}
}

func TestWritePOTFile(t *testing.T) {
	catalog, err := extract.Extract(sources()...)
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), "new", "translations", "translations.pot")

	if err := catalog.WritePOTFile(path); err != nil {
		t.Fatal(err)
	}

	pot, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(pot), "msgid \"Welcome\"") {
		t.Errorf("POT file misses Welcome:\n%s", pot)
	}

	// Extracting again gives the same file, so it only changes with the
	// templates.
	again, err := extract.Extract(sources()...)
	if err != nil {
		t.Fatal(err)
	}

	if err := again.WritePOTFile(path); err != nil {
		t.Fatal(err)
	}

	second, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(pot, second) {
		t.Error("a second extraction changed the POT file")
	}

	// A smaller catalog replaces the file: texts no longer used are gone.
	var smaller extract.Catalog

	smaller.Add(extract.Message{ID: "Only"})

	if err := smaller.WritePOTFile(path); err != nil {
		t.Fatal(err)
	}

	third, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(third), "Welcome") || !strings.Contains(string(third), "msgid \"Only\"") {
		t.Errorf("POT file was not replaced:\n%s", third)
	}
}
