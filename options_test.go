package translator

import (
	"errors"
	"os"
	"testing"
	"testing/fstest"
)

// TestNewDefaults checks that New records nothing and otherwise has the
// defaults of NewTranslator.
func TestNewDefaults(t *testing.T) {
	tr := New()
	old := NewTranslator("", "")

	if tr.RecordMissing() {
		t.Error("New records missing keys")
	}

	if !old.RecordMissing() {
		t.Error("NewTranslator no longer records missing keys")
	}

	if tr.PotFile() != DefaultPotFile || tr.PrefixSeparator() != DefaultSeparator || tr.LanguageFallback() {
		t.Errorf("defaults: pot %q, separator %q, fallback %v", tr.PotFile(), tr.PrefixSeparator(), tr.LanguageFallback())
	}

	if got := tr.Tl(mockLocalizer{locale: "nl"}, "Hello %s", "Bob"); got != "*Hello %s*" {
		t.Errorf("untranslated = %q, want the marker", got)
	}
}

// TestNewOptions checks that every option does what its setter does.
func TestNewOptions(t *testing.T) {
	dir := t.TempDir()

	tr := New(
		WithTranslationsDir(dir),
		WithUntranslated(ShowOriginal),
		WithLanguageFallback(true),
	)

	if tr.TranslationsDir() != dir || !tr.LanguageFallback() {
		t.Errorf("dir %q, fallback %v", tr.TranslationsDir(), tr.LanguageFallback())
	}

	if got := tr.Tn(mockLocalizer{locale: "nl"}, "%d file", "%d files", 0, 0); got != "0 files" {
		t.Errorf("Tn = %q, want the original", got)
	}
}

// TestUntranslated checks the translator without catalogues: messages show
// as written, and nothing is recorded or read.
func TestUntranslated(t *testing.T) {
	t.Chdir(t.TempDir())

	tr := Untranslated()
	loc := mockLocalizer{locale: "nl_NL"}

	if got := tr.Tl(loc, "Hello %s", "Bob"); got != "Hello Bob" {
		t.Errorf("Tl = %q", got)
	}

	if got := tr.Tn(loc, "%d file", "%d files", 2, 2); got != "2 files" {
		t.Errorf("Tn = %q", got)
	}

	if got := tr.Ctl(loc, "verb", "100% done"); got != "100% done" {
		t.Errorf("Ctl = %q", got)
	}

	if text, ok := tr.LookupC(loc, "month", "May"); ok || text != "May" {
		t.Errorf("LookupC = %q, %v", text, ok)
	}

	if tr.RecordMissing() || len(tr.languages) != 0 {
		t.Errorf("record %v, languages %v", tr.RecordMissing(), sortedKeys(tr.languages))
	}

	if _, err := os.Stat(DefaultPotFile); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("a POT file was written to the working directory: %v", err)
	}

	if Untranslated() == tr {
		t.Error("Untranslated returned the same translator twice")
	}

	withFallback := Untranslated(WithLanguageFallback(true))

	if !withFallback.LanguageFallback() {
		t.Error("Untranslated did not apply its options")
	}
}

// TestUntranslatedLoadsCatalogues checks that a translator from Untranslated
// still translates once catalogues are loaded.
func TestUntranslatedLoadsCatalogues(t *testing.T) {
	tr := Untranslated()

	if err := tr.LoadLanguages(fstest.MapFS{"nl.po": {Data: []byte(nlCatalogue)}}); err != nil {
		t.Fatal(err)
	}

	if got := tr.Tl(mockLocalizer{locale: "nl"}, "Hello %s", "Bob"); got != "Hallo Bob" {
		t.Errorf("Tl = %q", got)
	}

	if got := tr.Tl(mockLocalizer{locale: "nl"}, "Unknown"); got != "Unknown" {
		t.Errorf("Tl without translation = %q", got)
	}
}
