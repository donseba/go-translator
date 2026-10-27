package translator

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

const sentinelPot = "msgid \"\"\nmsgstr \"\"\n\nmsgid \"sentinel key\"\nmsgstr \"\"\n"

// writeSentinelPot writes a POT file with the entry "sentinel key" to dir.
func writeSentinelPot(t *testing.T, dir string) string {
	t.Helper()

	path := filepath.Join(dir, DefaultPotFile)
	if err := os.WriteFile(path, []byte(sentinelPot), 0o644); err != nil {
		t.Fatal(err)
	}

	return path
}

// countEntries counts the entries of key in the POT file at path.
func countEntries(t *testing.T, path, key string) int {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	return strings.Count(string(content), "msgid \""+key+"\"")
}

// lookUpAll calls every lookup function once with keys the POT file does
// not have.
func lookUpAll(tr *Translator) {
	loc := mockLocalizer{locale: "nl"}

	tr.Tl(loc, "lazy tl")
	tr.Tn(loc, "lazy tn", "lazy tns", 2)
	tr.Ctl(loc, "lazy", "lazy ctl")
	tr.Ctn(loc, "lazy", "lazy ctn", "lazy ctns", 1)
}

func TestPotNotReadWithoutRecording(t *testing.T) {
	dir := t.TempDir()
	path := writeSentinelPot(t, dir)

	tr := NewTranslator(dir, "")
	tr.SetRecordMissing(false)
	tr.SetUntranslated(ShowOriginal)

	if tr.pot != nil {
		t.Fatal("NewTranslator read the POT file")
	}

	lookUpAll(tr)

	if tr.pot != nil {
		t.Fatal("a lookup with recording off read the POT file")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if string(content) != sentinelPot {
		t.Fatalf("a lookup with recording off changed the POT file:\n%s", content)
	}
}

func TestPotReadWhenRecording(t *testing.T) {
	dir := t.TempDir()

	// The POT file appears after the translator is made: recording must
	// still find its keys, so it is read when needed, not at construction.
	tr := NewTranslator(dir, "")
	path := writeSentinelPot(t, dir)
	loc := mockLocalizer{locale: "nl"}

	tr.Tl(loc, "sentinel key")

	if tr.pot == nil {
		t.Fatal("recording did not read the POT file")
	}

	if got := countEntries(t, path, "sentinel key"); got != 1 {
		t.Fatalf("sentinel key appears %d times, want once", got)
	}

	tr.Tl(loc, "new key")
	tr.Tl(loc, "newer key")

	for _, key := range []string{"new key", "newer key"} {
		if got := countEntries(t, path, key); got != 1 {
			t.Errorf("%q appears %d times, want once", key, got)
		}
	}
}

func TestPotRecordingConcurrent(t *testing.T) {
	dir := t.TempDir()
	path := writeSentinelPot(t, dir)
	tr := NewTranslator(dir, "")

	var workers sync.WaitGroup

	for range 16 {
		workers.Add(1)

		go func() {
			defer workers.Done()

			loc := mockLocalizer{locale: "nl"}
			tr.Tl(loc, "sentinel key")
			tr.Tl(loc, "shared key")
		}()
	}

	workers.Wait()

	for _, key := range []string{"sentinel key", "shared key"} {
		if got := countEntries(t, path, key); got != 1 {
			t.Errorf("%q appears %d times, want once", key, got)
		}
	}
}

func TestSetTranslationsDirReadsItsPot(t *testing.T) {
	first := t.TempDir()
	second := t.TempDir()
	path := writeSentinelPot(t, second)

	tr := NewTranslator(first, "")
	tr.Tl(mockLocalizer{locale: "nl"}, "in first")

	tr.SetTranslationsDir(second)
	tr.Tl(mockLocalizer{locale: "nl"}, "sentinel key")

	if got := countEntries(t, path, "sentinel key"); got != 1 {
		t.Errorf("sentinel key appears %d times, want once", got)
	}
}

func TestEmptyTranslationsDirLeavesWorkingDirPot(t *testing.T) {
	t.Run("existing", func(t *testing.T) {
		t.Chdir(t.TempDir())
		path := writeSentinelPot(t, ".")

		tr := NewTranslator("", "")
		tr.SetRecordMissing(false)
		tr.SetUntranslated(ShowOriginal)
		lookUpAll(tr)

		if tr.pot != nil {
			t.Fatal("NewTranslator(\"\", \"\") read ./translations.pot")
		}

		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		if string(content) != sentinelPot {
			t.Fatalf("./translations.pot changed:\n%s", content)
		}
	})

	t.Run("missing", func(t *testing.T) {
		t.Chdir(t.TempDir())

		tr := NewTranslator("", "")
		tr.SetRecordMissing(false)
		lookUpAll(tr)

		_, err := os.Stat(DefaultPotFile)
		if !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("./translations.pot was created or cannot be checked: %v", err)
		}
	})
}

func TestCheckMissingTranslationsAddsEachKeyOnce(t *testing.T) {
	dir := t.TempDir()
	templates := t.TempDir()
	path := writeSentinelPot(t, dir)

	template := `{{ tl . "sentinel key" }}{{ tl . "scanned key" }}{{ ctl . "ctx" "scanned ctx key" }}`
	if err := os.WriteFile(filepath.Join(templates, "page"+TemplateExtension), []byte(template), 0o644); err != nil {
		t.Fatal(err)
	}

	tr := NewTranslator(dir, templates)
	tr.SetRecordMissing(false)

	for range 2 {
		if err := tr.CheckMissingTranslations(); err != nil {
			t.Fatal(err)
		}
	}

	for _, key := range []string{"sentinel key", "scanned key", "scanned ctx key"} {
		if got := countEntries(t, path, key); got != 1 {
			t.Errorf("%q appears %d times, want once", key, got)
		}
	}
}

func TestRecordingSkipsAnEmptyText(t *testing.T) {
	dir := t.TempDir()
	tr := NewTranslator(dir, "")
	loc := mockLocalizer{locale: "nl"}

	// Without a POT file yet, the empty text was written as a second
	// msgid "" next to the file's header.
	tr.Tl(loc, "")
	tr.Tl(loc, "a key")

	path := filepath.Join(dir, DefaultPotFile)
	if got := countEntries(t, path, ""); got != 1 {
		t.Fatalf(`msgid "" appears %d times, want once (the header)`, got)
	}

	if got := countEntries(t, path, "a key"); got != 1 {
		t.Fatalf("a key appears %d times, want once", got)
	}
}
