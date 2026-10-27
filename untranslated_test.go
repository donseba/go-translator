package translator

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// nlCatalogue is a small Dutch catalogue for the tests of this file.
const nlCatalogue = `msgid ""
msgstr ""
"Language: nl\n"
"Content-Type: text/plain; charset=UTF-8\n"
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

msgid "Hello %s"
msgstr "Hallo %s"

msgid "%d file"
msgid_plural "%d files"
msgstr[0] "%d bestand"
msgstr[1] "%d bestanden"

msgctxt "month"
msgid "May"
msgstr "mei"

msgctxt "inbox"
msgid "%d message"
msgid_plural "%d messages"
msgstr[0] "%d bericht"
msgstr[1] "%d berichten"

msgid "form__.Save"
msgstr "form__.Opslaan"
`

// newTestTranslator returns a translator over a temporary directory with the
// given files, recording nothing.
func newTestTranslator(t *testing.T, files map[string]string) *Translator {
	t.Helper()

	dir := t.TempDir()

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	tr := NewTranslator(dir, "")
	tr.SetRecordMissing(false)

	return tr
}

// newDutchTranslator returns a translator with nlCatalogue loaded as "nl".
func newDutchTranslator(t *testing.T) *Translator {
	t.Helper()

	tr := newTestTranslator(t, map[string]string{"nl.po": nlCatalogue})

	if err := tr.SetLanguage("nl"); err != nil {
		t.Fatal(err)
	}

	return tr
}

// TestDefaultUntranslatedMarkers checks that without SetUntranslated every
// lookup without translation shows the same markers as before.
func TestDefaultUntranslatedMarkers(t *testing.T) {
	tr := newDutchTranslator(t)

	for _, locale := range []string{"nl", "missing"} {
		loc := mockLocalizer{locale: locale}

		for _, test := range []struct {
			got  string
			want string
		}{
			{tr.Tl(loc, "Bye %s", "Bob"), "*Bye %s*"},
			{tr.Tn(loc, "%d dog", "%d dogs", 2, 2), "*%d dog/%d dogs*"},
			{tr.Ctl(loc, "menu", "Open", "x"), "*menu/Open*"},
			{tr.Ctn(loc, "inbox", "%d mail", "%d mails", 1, 1), "*inbox/%d mail/%d mails*"},
		} {
			if test.got != test.want {
				t.Errorf("%s: got %q, want %q", locale, test.got, test.want)
			}
		}
	}

	// Translated messages are not affected.
	loc := mockLocalizer{locale: "nl"}

	if got := tr.Tl(loc, "Hello %s", "Bob"); got != "Hallo Bob" {
		t.Errorf("translated: %q", got)
	}
}

// TestMarkersFollowDefaultVariables checks that the markers still read the
// DefaultNoTranslation variables at lookup time.
func TestMarkersFollowDefaultVariables(t *testing.T) {
	previous := DefaultNoTranslationTL
	DefaultNoTranslationTL = "[%s]"

	defer func() { DefaultNoTranslationTL = previous }()

	tr := newDutchTranslator(t)

	if got := tr.Tl(mockLocalizer{locale: "nl"}, "Bye"); got != "[Bye]" {
		t.Errorf("got %q", got)
	}
}

// TestCtlWithoutContextIsTl checks that ctl without a context gives what tl
// gives, also for a language that is not loaded.
func TestCtlWithoutContextIsTl(t *testing.T) {
	tr := newDutchTranslator(t)

	for _, locale := range []string{"nl", "missing"} {
		loc := mockLocalizer{locale: locale}

		for _, key := range []string{"Hello %s", "Bye %s"} {
			want := tr.Tl(loc, key, "Bob")

			if got := tr.Ctl(loc, "", key, "Bob"); got != want {
				t.Errorf("%s %q: Ctl %q, Tl %q", locale, key, got, want)
			}
		}
	}

	if got := tr.Ctl(mockLocalizer{locale: "missing"}, "", "key"); got != "*key*" {
		t.Errorf("missing language: %q", got)
	}

	tr.SetUntranslated(ShowOriginal)

	if got := tr.Ctl(mockLocalizer{locale: "missing"}, "", "Bye %s", "Bob"); got != "Bye Bob" {
		t.Errorf("missing language, ShowOriginal: %q", got)
	}
}

// TestTnAndCtnFormatArguments checks the arguments of Tn and Ctn.
func TestTnAndCtnFormatArguments(t *testing.T) {
	tr := newDutchTranslator(t)
	loc := mockLocalizer{locale: "nl"}

	for _, test := range []struct {
		got  string
		want string
	}{
		{tr.Tn(loc, "%d file", "%d files", 1, 1), "1 bestand"},
		{tr.Tn(loc, "%d file", "%d files", 3, 3), "3 bestanden"},
		{tr.Tn(loc, "%d file", "%d files", 3), "%d bestanden"},
		{tr.Ctn(loc, "inbox", "%d message", "%d messages", 1, 1), "1 bericht"},
		{tr.Ctn(loc, "inbox", "%d message", "%d messages", 4, 4), "4 berichten"},
	} {
		if test.got != test.want {
			t.Errorf("got %q, want %q", test.got, test.want)
		}
	}
}

// TestShowOriginal checks the source text of every lookup without
// translation.
func TestShowOriginal(t *testing.T) {
	tr := newDutchTranslator(t)
	tr.SetUntranslated(ShowOriginal)

	for _, locale := range []string{"nl", "missing"} {
		loc := mockLocalizer{locale: locale}

		for _, test := range []struct {
			got  string
			want string
		}{
			{tr.Tl(loc, "Bye %s", "Bob"), "Bye Bob"},
			{tr.Tl(loc, "100%"), "100%"},
			{tr.Tn(loc, "%d dog", "%d dogs", 1, 1), "1 dog"},
			{tr.Tn(loc, "%d dog", "%d dogs", 0, 0), "0 dogs"},
			{tr.Tn(loc, "%d dog", "%d dogs", 2, 2), "2 dogs"},
			{tr.Ctl(loc, "menu", "Open %s", "x"), "Open x"},
			{tr.Ctn(loc, "inbox", "%d mail", "%d mails", 1, 1), "1 mail"},
			{tr.Ctn(loc, "inbox", "%d mail", "%d mails", 5, 5), "5 mails"},
			{tr.Tl(loc, "admin__.Cancel"), "Cancel"},
		} {
			if test.got != test.want {
				t.Errorf("%s: got %q, want %q", locale, test.got, test.want)
			}
		}
	}

	// Translations are not affected.
	loc := mockLocalizer{locale: "nl"}

	if got := tr.Tn(loc, "%d file", "%d files", 2, 2); got != "2 bestanden" {
		t.Errorf("translated: %q", got)
	}

	tr.SetUntranslated(nil)

	if got := tr.Tl(loc, "Bye"); got != "*Bye*" {
		t.Errorf("nil restores the markers: %q", got)
	}
}

// TestUntranslatedHandler checks what a custom handler receives.
func TestUntranslatedHandler(t *testing.T) {
	tr := newDutchTranslator(t)

	var misses []Miss

	tr.SetUntranslated(func(m Miss) string {
		misses = append(misses, m)

		return "<" + m.Marker() + "|" + m.Original() + ">"
	})

	loc := mockLocalizer{locale: "nl"}

	if got := tr.Ctn(loc, "inbox", "%d mail", "%d mails", 3, 3); got != "<*inbox/%d mail/%d mails*|3 mails>" {
		t.Errorf("got %q", got)
	}

	if got := tr.Tl(loc, "Hello %s", "Bob"); got != "Hallo Bob" {
		t.Errorf("translated message reached the handler: %q", got)
	}

	if len(misses) != 1 {
		t.Fatalf("misses: %+v", misses)
	}

	m := misses[0]

	if m.Locale != "nl" || m.Context != "inbox" || m.Key != "%d mail" || m.Plural != "%d mails" || m.N != 3 ||
		fmt.Sprint(m.Args) != "[3]" {
		t.Errorf("miss: %+v", m)
	}
}

// TestUntranslatedTemplateFunctions checks that the template functions use
// SetUntranslated too.
func TestUntranslatedTemplateFunctions(t *testing.T) {
	tr := newDutchTranslator(t)

	page := template.Must(template.New("page").Funcs(tr.FuncMap()).Parse(
		`{{tl .L "Bye %s" "Bob"}}|{{tn .L "%d dog" "%d dogs" 2 2}}|{{ctl .L "" "Hi"}}|` +
			`{{ctn .L "c" "%d cat" "%d cats" 1 1}}|{{tl .L "Hello %s" "Bob"}}`,
	))

	render := func() string {
		var out strings.Builder

		if err := page.Execute(&out, map[string]any{"L": mockLocalizer{locale: "nl"}}); err != nil {
			t.Fatal(err)
		}

		return out.String()
	}

	if got := render(); got != "*Bye %s*|*%d dog/%d dogs*|*Hi*|*c/%d cat/%d cats*|Hallo Bob" {
		t.Errorf("markers: %q", got)
	}

	tr.SetUntranslated(ShowOriginal)

	if got := render(); got != "Bye Bob|2 dogs|Hi|1 cat|Hallo Bob" {
		t.Errorf("ShowOriginal: %q", got)
	}
}

// TestLookup checks the Lookup functions: the text of the matching T
// function and whether a translation was found.
func TestLookup(t *testing.T) {
	tr := newDutchTranslator(t)
	loc := mockLocalizer{locale: "nl"}

	for _, test := range []struct {
		name   string
		lookup func() (string, bool)
		want   string
		found  bool
	}{
		{"Lookup", func() (string, bool) { return tr.Lookup(loc, "Hello %s", "Bob") }, "Hallo Bob", true},
		{"Lookup missing", func() (string, bool) { return tr.Lookup(loc, "Bye %s", "Bob") }, "*Bye %s*", false},
		{"LookupN", func() (string, bool) { return tr.LookupN(loc, "%d file", "%d files", 2, 2) }, "2 bestanden", true},
		{
			"LookupN missing",
			func() (string, bool) { return tr.LookupN(loc, "%d dog", "%d dogs", 2, 2) },
			"*%d dog/%d dogs*",
			false,
		},
		{"LookupC", func() (string, bool) { return tr.LookupC(loc, "month", "May") }, "mei", true},
		{"LookupC without context", func() (string, bool) { return tr.LookupC(loc, "", "Hello %s", "Bob") }, "Hallo Bob", true},
		{"LookupC missing", func() (string, bool) { return tr.LookupC(loc, "month", "June") }, "*month/June*", false},
		{
			"LookupCN",
			func() (string, bool) { return tr.LookupCN(loc, "inbox", "%d message", "%d messages", 1, 1) },
			"1 bericht",
			true,
		},
		{
			"LookupCN missing",
			func() (string, bool) { return tr.LookupCN(loc, "inbox", "%d mail", "%d mails", 1, 1) },
			"*inbox/%d mail/%d mails*",
			false,
		},
	} {
		text, found := test.lookup()

		if text != test.want || found != test.found {
			t.Errorf("%s: (%q, %v), want (%q, %v)", test.name, text, found, test.want, test.found)
		}
	}

	tr.SetUntranslated(ShowOriginal)

	if text, found := tr.LookupN(loc, "%d dog", "%d dogs", 1, 1); text != "1 dog" || found {
		t.Errorf("ShowOriginal: (%q, %v)", text, found)
	}

	if text, found := tr.Lookup(mockLocalizer{locale: "missing"}, "Hello %s", "Bob"); text != "Hello Bob" || found {
		t.Errorf("missing language: (%q, %v)", text, found)
	}
}

// TestLookupRemovesPrefix checks that a translation loses its prefix, as
// with Tl.
func TestLookupRemovesPrefix(t *testing.T) {
	tr := newDutchTranslator(t)

	if text, found := tr.Lookup(mockLocalizer{locale: "nl"}, "form__.Save"); text != "Opslaan" || !found {
		t.Errorf("(%q, %v)", text, found)
	}
}
