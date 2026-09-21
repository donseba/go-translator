package translator

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestConcurrentLookupsAddEachKeyOnce(t *testing.T) {
	dir := t.TempDir()
	potPath := filepath.Join(dir, DefaultPotFile)
	initial, err := os.ReadFile(filepath.Join(translationsDir, DefaultPotFile))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(potPath, initial, 0644); err != nil {
		t.Fatal(err)
	}

	tr := NewTranslator(dir, "")
	loc := mockLocalizer{locale: "missing"}
	var workers sync.WaitGroup
	for range 32 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for range 100 {
				tr.Tl(loc, "missing singular")
				tr.Ctl(loc, "context", "missing context")
				tr.Tn(loc, "missing one", "missing many", 2)
				tr.Ctn(loc, "context", "missing one", "missing many", 2)
			}
		}()
	}
	workers.Wait()

	content, err := os.ReadFile(potPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range []string{
		"msgid \"missing singular\"\nmsgstr \"\"",
		"msgctxt \"context\"\nmsgid \"missing context\"\nmsgstr \"\"",
		"\n\nmsgid \"missing one\"\nmsgid_plural \"missing many\"",
		"msgctxt \"context\"\nmsgid \"missing one\"\nmsgid_plural \"missing many\"",
	} {
		if count := strings.Count(string(content), entry); count != 1 {
			t.Errorf("POT entry %q appears %d times, want once", entry, count)
		}
	}
	if len(tr.uniqueKeys) != 2 || len(tr.uniqueKeysCtx["context"]) != 2 {
		t.Fatal("lookup did not record all discovered keys")
	}
}

func TestConcurrentLoadedLanguageLookups(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{DefaultPotFile, lang + DefaultPoExtension} {
		content, err := os.ReadFile(filepath.Join(translationsDir, name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), content, 0644); err != nil {
			t.Fatal(err)
		}
	}
	tr := NewTranslator(dir, "")
	if err := tr.SetLanguage(lang); err != nil {
		t.Fatal(err)
	}
	loc := mockLocalizer{locale: lang}
	var workers sync.WaitGroup
	for range 32 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for range 100 {
				if got := tr.Tl(loc, "First translation text"); got != "First translation text" {
					t.Errorf("unexpected translation: %q", got)
					return
				}
				tr.Ctl(loc, "context", "missing context")
				tr.Tn(loc, "missing one", "missing many", 2)
				tr.Ctn(loc, "context", "missing one", "missing many", 2)
			}
		}()
	}
	workers.Wait()
}
