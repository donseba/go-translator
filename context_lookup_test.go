package translator

import (
	"os"
	"testing"
	"testing/fstest"
)

// finnish has month names that differ between the long and the short form,
// and date patterns, each told apart by a context, as a calendar uses them.
const finnish = `msgid ""
msgstr ""
"Language: fi\n"
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

msgctxt "month"
msgid "May"
msgstr "toukokuu"

msgctxt "month-short"
msgid "May"
msgstr "toukok."

msgctxt "month"
msgid "June"
msgstr ""

msgctxt "date"
msgid "short"
msgstr "d.M.yyyy"

msgid "May"
msgstr "touko"
`

// TestLookupCalendarByContext checks the lookups a calendar needs: the same
// English text in different contexts, and whether a translation exists, with
// no comparison of strings.
func TestLookupCalendarByContext(t *testing.T) {
	t.Chdir(t.TempDir())

	tr := New(WithUntranslated(ShowOriginal), WithLanguageFallback(true))

	if err := tr.LoadLanguages(fstest.MapFS{"fi.po": {Data: []byte(finnish)}}); err != nil {
		t.Fatal(err)
	}

	for _, locale := range []string{"fi", "fi_FI"} {
		loc := mockLocalizer{locale: locale}

		for _, test := range []struct {
			context, key, want string
			translated         bool
		}{
			{"month", "May", "toukokuu", true},
			{"month-short", "May", "toukok.", true},
			{"date", "short", "d.M.yyyy", true},
			// An empty translation is no translation.
			{"month", "June", "June", false},
			// No entry: the key itself, so a pattern name is never
			// mistaken for a pattern.
			{"date", "long", "long", false},
			// A context does not fall back to the entry without one.
			{"month", "July", "July", false},
			// Without a context, the entry without one.
			{"", "May", "touko", true},
		} {
			got, ok := tr.LookupC(loc, test.context, test.key)
			if got != test.want || ok != test.translated {
				t.Errorf("%s LookupC(%q, %q) = %q, %v; want %q, %v",
					locale, test.context, test.key, got, ok, test.want, test.translated)
			}

			if text := tr.Ctl(loc, test.context, test.key); text != test.want {
				t.Errorf("%s Ctl(%q, %q) = %q, want %q", locale, test.context, test.key, text, test.want)
			}
		}
	}

	// A language without a catalogue gets the English text and false.
	if got, ok := tr.LookupC(mockLocalizer{locale: "sv"}, "month-short", "May"); ok || got != "May" {
		t.Errorf("sv LookupC = %q, %v", got, ok)
	}

	// A text from a context without translation keeps the marker of
	// SetUntranslated.
	marked := New()

	if err := marked.LoadLanguages(fstest.MapFS{"fi.po": {Data: []byte(finnish)}}); err != nil {
		t.Fatal(err)
	}

	if got, ok := marked.LookupC(mockLocalizer{locale: "fi"}, "date", "long"); ok || got != "*date/long*" {
		t.Errorf("marker LookupC = %q, %v", got, ok)
	}

	if _, err := os.Stat(DefaultPotFile); !os.IsNotExist(err) {
		t.Errorf("a lookup wrote a POT file: %v", err)
	}
}
