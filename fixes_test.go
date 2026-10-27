package translator

import (
	"bytes"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCtnWithoutContextIsTn checks that ctn without a context gives what tn
// gives, translated or not, and records the key once.
func TestCtnWithoutContextIsTn(t *testing.T) {
	tr := newDutchTranslator(t)

	for _, locale := range []string{"nl", "missing"} {
		loc := mockLocalizer{locale: locale}

		for _, singular := range []string{"%d file", "%d dog"} {
			want := tr.Tn(loc, singular, singular+"s", 2, 2)

			if got := tr.Ctn(loc, "", singular, singular+"s", 2, 2); got != want {
				t.Errorf("%s %q: Ctn %q, Tn %q", locale, singular, got, want)
			}

			text, _ := tr.LookupCN(loc, "", singular, singular+"s", 2, 2)
			if text != want {
				t.Errorf("%s %q: LookupCN %q, Tn %q", locale, singular, text, want)
			}
		}
	}

	if got := tr.Ctn(mockLocalizer{locale: "nl"}, "", "%d file", "%d files", 2, 2); got != "2 bestanden" {
		t.Errorf("translated: %q", got)
	}

	if got := tr.Ctn(mockLocalizer{locale: "missing"}, "", "a", "b", 2); got != "*a/b*" {
		t.Errorf("marker: %q", got)
	}
}

// TestContextlessLookupsRecordOnce checks that ctl and ctn without a context
// record their key once, without a context.
func TestContextlessLookupsRecordOnce(t *testing.T) {
	dir := t.TempDir()

	if err := WritePOTFile(filepath.Join(dir, DefaultPotFile)); err != nil {
		t.Fatal(err)
	}

	tr := NewTranslator(dir, "")
	loc := mockLocalizer{locale: "missing"}

	tr.Ctl(loc, "", "single")
	tr.Ctn(loc, "", "one", "many", 2)

	data, err := os.ReadFile(filepath.Join(dir, DefaultPotFile))
	if err != nil {
		t.Fatal(err)
	}

	for _, entry := range []string{"msgid \"single\"", "msgid \"one\"\nmsgid_plural \"many\""} {
		if count := strings.Count(string(data), entry); count != 1 {
			t.Errorf("%q recorded %d times", entry, count)
		}
	}

	if strings.Contains(string(data), "msgctxt") {
		t.Error("recorded with an empty context")
	}
}

// TestHeaderForRegionalLanguage checks that a locale without its own entry
// gets the plural rules of its base language.
func TestHeaderForRegionalLanguage(t *testing.T) {
	for _, test := range []struct {
		language string
		forms    string
	}{
		{"pl_PL", "nplurals=3;"},
		{"pl-PL", "nplurals=3;"},
		{"ru_RU", "nplurals=3;"},
		{"ja_JP", "nplurals=1;"},
		{"xx_XX", "nplurals=2;"},
	} {
		h := GetHeaderForLanguage(test.language)

		if !strings.HasPrefix(h.PluralForms, test.forms) || h.Language != test.language {
			t.Errorf("%s: %q, language %q", test.language, h.PluralForms, h.Language)
		}
	}
}

// TestDetailsOfMissingLanguage checks that Details does not panic for a
// language that is not loaded.
func TestDetailsOfMissingLanguage(t *testing.T) {
	tr := newDutchTranslator(t)

	if got := tr.Details(mockLocalizer{locale: "missing"}); got != nil {
		t.Errorf("got %v", got)
	}

	if got := tr.Details(mockLocalizer{locale: "nl"}); len(got["Language"]) != 1 || got["Language"][0] != "nl" {
		t.Errorf("loaded: %v", got)
	}
}

// TestNewLanguageAndWrite checks the X-Generator header, the file name and
// the revision date of a written catalogue.
func TestNewLanguageAndWrite(t *testing.T) {
	tr := newTestTranslator(t, nil)
	loc := mockLocalizer{locale: "fr"}

	if err := tr.NewLanguage(loc); err != nil {
		t.Fatal(err)
	}

	if err := tr.SetTL(loc, "Yes", "Oui"); err != nil {
		t.Fatal(err)
	}

	if err := tr.Write(loc); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(tr.TranslationsDir(), "fr"+DefaultPoExtension))
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)

	if !strings.Contains(content, `"X-Generator: github.com/donseba/go-translator\n"`) ||
		strings.Contains(content, "X-Generator::") {
		t.Errorf("X-Generator header:\n%s", content)
	}

	revision := tr.Details(loc)["PO-Revision-Date"]
	if len(revision) != 1 || len(revision[0]) != len(headerDateFormat) || strings.Contains(revision[0], "T") {
		t.Errorf("revision date %q, want the gettext format", revision)
	}
}

// TestRecordErrors checks that errors of runtime recording are logged with
// log/slog and that CheckMissingTranslations returns its errors.
func TestRecordErrors(t *testing.T) {
	var logged bytes.Buffer

	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logged, nil)))
	t.Cleanup(func() {
		slog.SetDefault(previous)
	})

	// A translations directory that does not exist makes every recording fail.
	missing := filepath.Join(t.TempDir(), "missing")

	tr := NewTranslator(missing, "")
	tr.Tl(mockLocalizer{locale: "nl"}, "Hello")

	if !strings.Contains(logged.String(), "record missing translation") ||
		!strings.Contains(logged.String(), `record \"Hello\"`) {
		t.Errorf("the recording error was not logged:\n%s", logged.String())
	}

	templates := t.TempDir()
	if err := os.WriteFile(
		filepath.Join(templates, "page.gohtml"),
		[]byte(`{{tl .L "Bye"}}{{ctl .L "menu" "Open"}}`),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	tr = NewTranslator(missing, templates)

	if err := tr.CheckMissingTranslations(); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("CheckMissingTranslations: %v", err)
	}
}
