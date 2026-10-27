package translator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCheckMissingTranslationsOrder checks that new keys are appended to the
// POT file in the order of the templates, and that an empty key is not
// written.
func TestCheckMissingTranslationsOrder(t *testing.T) {
	dir := t.TempDir()
	templates := t.TempDir()

	files := map[string]string{
		"a" + TemplateExtension: `{{tl . "zebra"}}{{tl . "apple"}}{{ctl . "ctx" "mango"}}{{tl . ""}}`,
		"b" + TemplateExtension: `{{tn . "%d kiwi" "%d kiwis" 2}}{{tl . "banana"}}{{tl . "apple"}}`,
	}

	for name, text := range files {
		if err := os.WriteFile(filepath.Join(templates, name), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	tr := New(WithTranslationsDir(dir))
	tr.SetTemplateDir(templates)

	if err := tr.CheckMissingTranslations(); err != nil {
		t.Fatal(err)
	}

	pot, err := os.ReadFile(filepath.Join(dir, DefaultPotFile))
	if err != nil {
		t.Fatal(err)
	}

	var ids []string

	for line := range strings.SplitSeq(string(pot), "\n") {
		if strings.HasPrefix(line, "msgid ") && line != `msgid ""` {
			ids = append(ids, strings.TrimPrefix(line, "msgid "))
		}
	}

	want := `"zebra" "apple" "mango" "%d kiwi" "banana"`
	if got := strings.Join(ids, " "); got != want {
		t.Errorf("POT keys = %s, want %s", got, want)
	}

	if strings.Count(string(pot), `msgid ""`) != 1 {
		t.Errorf("the empty key was written:\n%s", pot)
	}

	if !strings.Contains(string(pot), "msgid \"%d kiwi\"\nmsgid_plural \"%d kiwis\"\n") {
		t.Errorf("the plural is missing:\n%s", pot)
	}
}

// TestScanFilesPlural checks that a key used with and without a plural keeps
// the plural, and that two plurals keep the first.
func TestScanFilesPlural(t *testing.T) {
	templates := t.TempDir()
	text := `{{tl . "%d cat"}}{{tn . "%d cat" "%d cats" 1}}{{tn . "%d cat" "%d kitties" 2}}`

	if err := os.WriteFile(filepath.Join(templates, "a"+TemplateExtension), []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}

	tr := New()

	if err := tr.ScanFiles(templates); err != nil {
		t.Fatal(err)
	}

	if got := tr.uniqueKeys["%d cat"].plural; got != "%d cats" {
		t.Errorf("plural = %q, want the first one", got)
	}

	// A later scan without the plural keeps it.
	if err := os.WriteFile(filepath.Join(templates, "a"+TemplateExtension), []byte(`{{tl . "%d cat"}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := tr.ScanFiles(templates); err != nil {
		t.Fatal(err)
	}

	if got := tr.uniqueKeys["%d cat"].plural; got != "%d cats" {
		t.Errorf("plural after a second scan = %q", got)
	}
}
