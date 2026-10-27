package extract_test

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/donseba/go-translator/v2/extract"
)

const layout = `{{define "header"}}<h1>{{tl localizer "Welcome"}}</h1>{{end}}
<p>{{tl localizer "Hello, %s" .Name}}</p>
{{if .Items}}
	{{range .Items}}{{tn localizer "%d item" "%d items" .Count .Count}}{{end}}
{{else}}
	{{with .Empty}}{{ctl localizer "button" "Open"}}{{end}}
{{end}}
{{printf "%s!" (tl localizer "Nested")}}
{{template "header" .}}
{{/* {{tl localizer "Commented out"}} */}}
{{tl localizer .Dynamic}}
{{"Piped" | tl localizer}}
{{ctn localizer "inbox" "%d message" "%d messages" .N .N}}
{{tl localizer "Quote \" and\nnewline"}}
{{unknownHelper "not a translation"}}
`

const page = `{{tl localizer "Welcome"}}
{{ctl localizer "door" "Open"}}
{{tl localizer "%d item"}}
`

func sources() []extract.Source {
	return []extract.Source{
		{Name: "default/layout.gohtml", Text: layout},
		{Name: "default/page.gohtml", Text: page},
	}
}

// wantMessages are the messages of sources, in order.
var wantMessages = []extract.Message{
	{ID: "Welcome", References: []string{"default/layout.gohtml:1", "default/page.gohtml:1"}},
	{ID: "Hello, %s", References: []string{"default/layout.gohtml:2"}},
	{ID: "%d item", Plural: "%d items", References: []string{"default/layout.gohtml:4", "default/page.gohtml:3"}},
	{Context: "button", ID: "Open", References: []string{"default/layout.gohtml:6"}},
	{ID: "Nested", References: []string{"default/layout.gohtml:8"}},
	{Context: "inbox", ID: "%d message", Plural: "%d messages", References: []string{"default/layout.gohtml:13"}},
	{ID: "Quote \" and\nnewline", References: []string{"default/layout.gohtml:14"}},
	{Context: "door", ID: "Open", References: []string{"default/page.gohtml:2"}},
}

// sameMessages reports whether got equals want, references included.
func sameMessages(t *testing.T, got, want []extract.Message) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("found %d messages, want %d: %+v", len(got), len(want), got)
	}

	for i, message := range got {
		if message.String() != want[i].String() ||
			strings.Join(message.References, " ") != strings.Join(want[i].References, " ") {
			t.Errorf("message %d = %+v, want %+v", i, message, want[i])
		}
	}
}

func TestExtract(t *testing.T) {
	catalog, err := extract.Extract(sources()...)
	if err != nil {
		t.Fatal(err)
	}

	sameMessages(t, catalog.Messages, wantMessages)

	if catalog.Sources != 2 {
		t.Errorf("Sources = %d, want 2", catalog.Sources)
	}

	if len(catalog.Warnings) != 2 {
		t.Fatalf("warnings = %q, want the dynamic and the piped key", catalog.Warnings)
	}

	dynamic := catalog.Warnings[0]
	if dynamic.Reference != "default/layout.gohtml:11" || !strings.HasPrefix(dynamic.Problem, "tl ") ||
		!strings.Contains(dynamic.Problem, ".Dynamic") {
		t.Errorf("warning = %+v, want the dynamic key on line 11", dynamic)
	}

	if !strings.HasPrefix(dynamic.String(), "default/layout.gohtml:11: tl ") {
		t.Errorf("warning text = %q", dynamic.String())
	}

	if !strings.HasPrefix(catalog.Warnings[1].String(), "default/layout.gohtml:12: tl has its text from a pipeline") {
		t.Errorf("warning = %q, want the piped key on line 12", catalog.Warnings[1])
	}
}

func TestExtractIsDeterministic(t *testing.T) {
	// Many {{define}}s are many parse trees, kept in a map by the parser.
	var b strings.Builder

	for _, name := range []string{"a", "b", "c", "d", "e", "f", "g", "h"} {
		b.WriteString(`{{define "` + name + `"}}{{tl localizer "` + name + `"}}{{tl localizer .` + name + `}}{{end}}` + "\n")
	}

	source := extract.Source{Name: "defines.gohtml", Text: b.String()}

	first, err := extract.Extract(source)
	if err != nil {
		t.Fatal(err)
	}

	for range 20 {
		again, err := extract.Extract(source)
		if err != nil {
			t.Fatal(err)
		}

		sameMessages(t, again.Messages, first.Messages)

		for i, warning := range again.Warnings {
			if warning != first.Warnings[i] {
				t.Fatalf("warning %d = %v, was %v", i, warning, first.Warnings[i])
			}
		}
	}

	var ids []string

	for _, message := range first.Messages {
		ids = append(ids, message.ID)
	}

	if got := strings.Join(ids, ""); got != "abcdefgh" {
		t.Errorf("messages in order %q, want abcdefgh", got)
	}

	for i, warning := range first.Warnings {
		if !strings.Contains(warning.Problem, "."+ids[i]) {
			t.Errorf("warning %d = %q, want the one of .%s", i, warning, ids[i])
		}
	}
}

func TestExtractParseError(t *testing.T) {
	_, err := extract.Extract(extract.Source{Name: "broken.gohtml", Text: "{{if}}"})
	if err == nil || !strings.Contains(err.Error(), "broken.gohtml") {
		t.Fatalf("error = %v, want one naming broken.gohtml", err)
	}
}

func TestExtractPluralConflict(t *testing.T) {
	catalog, err := extract.Extract(extract.Source{
		Name: "a.gohtml",
		Text: `{{tl localizer "%d cat"}}{{tn localizer "%d cat" "%d cats" 1}}{{tn localizer "%d cat" "%d kitties" 2}}`,
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(catalog.Messages) != 1 || catalog.Messages[0].Plural != "%d cats" {
		t.Fatalf("messages = %+v, want one with the first plural", catalog.Messages)
	}

	if len(catalog.Warnings) != 1 || !strings.Contains(catalog.Warnings[0].Problem, "%d kitties") ||
		catalog.Warnings[0].Reference != "a.gohtml:1" {
		t.Fatalf("warnings = %q, want the conflicting plural", catalog.Warnings)
	}
}

func TestExtractEmptyText(t *testing.T) {
	catalog, err := extract.Extract(extract.Source{Name: "a.gohtml", Text: `{{tl localizer ""}}{{tl localizer "Kept"}}`})
	if err != nil {
		t.Fatal(err)
	}

	if len(catalog.Messages) != 1 || catalog.Messages[0].ID != "Kept" {
		t.Fatalf("messages = %+v, want only Kept", catalog.Messages)
	}

	if len(catalog.Warnings) != 1 || catalog.Warnings[0].Reference != "a.gohtml:1" {
		t.Fatalf("warnings = %q, want the empty text", catalog.Warnings)
	}
}

func TestExtractChainsAndArguments(t *testing.T) {
	text := `{{(tl localizer "Chained").Field}}
{{template "row" (tn localizer "%d row" "%d rows" 2)}}
{{$text := ctl localizer "verb" "Save"}}
{{if eq (tl localizer "In a condition") "x"}}{{end}}
{{ctl localizer "" "No context"}}
{{.Name | tl localizer "Hello %s"}}
`

	catalog, err := extract.Extract(extract.Source{Name: "a.gohtml", Text: text})
	if err != nil {
		t.Fatal(err)
	}

	want := []extract.Message{
		{ID: "Chained", References: []string{"a.gohtml:1"}},
		{ID: "%d row", Plural: "%d rows", References: []string{"a.gohtml:2"}},
		{Context: "verb", ID: "Save", References: []string{"a.gohtml:3"}},
		{ID: "In a condition", References: []string{"a.gohtml:4"}},
		{ID: "No context", References: []string{"a.gohtml:5"}},
		{ID: "Hello %s", References: []string{"a.gohtml:6"}},
	}

	sameMessages(t, catalog.Messages, want)

	if len(catalog.Warnings) != 0 {
		t.Errorf("warnings = %q, want none", catalog.Warnings)
	}
}

func TestExtractorFunctionsAndDelims(t *testing.T) {
	functions := extract.DefaultFunctions()
	functions["T"] = extract.Function{}
	functions["P"] = extract.Function{Skip: 2, Context: true, Plural: true}
	delete(functions, "tl")

	x := extract.Extractor{Functions: functions, LeftDelim: "[[", RightDelim: "]]"}

	catalog, err := x.Extract(extract.Source{
		Name: "a.html",
		Text: `[[T "Plain"]] [[P .Loc .N "cart" "%d product" "%d products"]] ` +
			`[[tl localizer "Not a function now"]] {{T "Not an action"}}
[[ctl localizer "menu" "Home"]] [[T .Dynamic]]`,
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []extract.Message{
		{ID: "Plain", References: []string{"a.html:1"}},
		{Context: "cart", ID: "%d product", Plural: "%d products", References: []string{"a.html:1"}},
		{Context: "menu", ID: "Home", References: []string{"a.html:2"}},
	}

	sameMessages(t, catalog.Messages, want)

	if len(catalog.Warnings) != 1 ||
		!strings.HasPrefix(catalog.Warnings[0].Problem, "T has a text that is not a literal string: .Dynamic") {
		t.Errorf("warnings = %q, want the dynamic T", catalog.Warnings)
	}
}

func TestDefaultFunctionsIsACopy(t *testing.T) {
	functions := extract.DefaultFunctions()
	delete(functions, "tl")

	if _, ok := extract.DefaultFunctions()["tl"]; !ok {
		t.Fatal("changing the returned map changed the defaults")
	}
}

func TestFiles(t *testing.T) {
	files := fstest.MapFS{
		"default/layout.gohtml":        {Data: []byte(`{{tl localizer "A"}}`)},
		"default/mail/x/note.gohtml":   {Data: []byte(`{{tl localizer "B"}}`)},
		"default/assets/site.css":      {Data: []byte(`body{}`)},
		"night/partials/card.gohtml":   {Data: []byte(`{{tl localizer "C"}}`)},
		"night/theme.json":             {Data: []byte(`{}`)},
		"default/partials/menu.gohtml": {Data: []byte(``)},
		"night/page.tmpl":              {Data: []byte(`{{tl localizer "D"}}`)},
	}

	names := func(sources []extract.Source) string {
		var names []string

		for _, source := range sources {
			names = append(names, source.Name)
		}

		return strings.Join(names, " ")
	}

	sources, err := extract.Files(files)
	if err != nil {
		t.Fatal(err)
	}

	want := "default/layout.gohtml default/mail/x/note.gohtml default/partials/menu.gohtml night/partials/card.gohtml"
	if got := names(sources); got != want {
		t.Fatalf("sources = %s, want %s", got, want)
	}

	sources, err = extract.Files(files, ".tmpl", ".css")
	if err != nil {
		t.Fatal(err)
	}

	if got := names(sources); got != "default/assets/site.css night/page.tmpl" {
		t.Fatalf("sources with extensions = %s", got)
	}

	none, err := extract.Files(nil)
	if err != nil || none != nil {
		t.Fatalf("Files(nil) = %v, %v", none, err)
	}
}
