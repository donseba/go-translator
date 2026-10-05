package translator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLookupCreatesMissingPotFile(t *testing.T) {
	dir := t.TempDir()
	tr := NewTranslator(dir, "")
	loc := mockLocalizer{locale: "missing"}

	tr.Tl(loc, "fresh key")
	tr.Ctn(loc, "context", "one item", "many items", 2)

	content, err := os.ReadFile(filepath.Join(dir, DefaultPotFile))
	if err != nil {
		t.Fatalf("POT file was not created: %v", err)
	}

	text := string(content)
	if !strings.HasPrefix(text, "msgid \"\"\nmsgstr \"\"\n") {
		t.Fatalf("POT file does not start with a header entry:\n%s", text)
	}

	for _, want := range []string{
		"\"Content-Type: text/plain; charset=UTF-8\\n\"",
		"\"Plural-Forms: nplurals=2; plural=(n != 1);\\n\"",
		"msgid \"fresh key\"\nmsgstr \"\"",
		"msgctxt \"context\"\nmsgid \"one item\"\nmsgid_plural \"many items\"",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("POT file misses %q:\n%s", want, text)
		}
	}

	// A second translator must see the recorded keys and not append them again.
	again := NewTranslator(dir, "")
	again.Tl(loc, "fresh key")

	content, err = os.ReadFile(filepath.Join(dir, DefaultPotFile))
	if err != nil {
		t.Fatal(err)
	}

	if count := strings.Count(string(content), "msgid \"fresh key\""); count != 1 {
		t.Errorf("key appears %d times, want once", count)
	}
}

func TestCheckMissingTranslationsCreatesPotFile(t *testing.T) {
	dir := t.TempDir()
	tr := NewTranslator(dir, templateDir)

	if err := tr.CheckMissingTranslations(); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filepath.Join(dir, DefaultPotFile))
	if err != nil {
		t.Fatalf("POT file was not created: %v", err)
	}

	if !strings.HasPrefix(string(content), "msgid \"\"\n") {
		t.Fatalf("POT file has no header:\n%s", content)
	}

	if strings.Count(string(content), "msgid \"\"\n") != 1 {
		t.Fatalf("POT file has more than one header:\n%s", content)
	}

	if len(tr.uniqueKeys) == 0 {
		t.Fatal("no keys were extracted from the test templates")
	}
}

func TestMissingTranslationsDirectoryIsNotCreated(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "absent")
	tr := NewTranslator(dir, "")

	tr.Tl(mockLocalizer{locale: "missing"}, "key")

	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("translations directory should not be created, stat error = %v", err)
	}
}

func TestSetRecordMissingDisablesRuntimeRecording(t *testing.T) {
	dir := t.TempDir()
	tr := NewTranslator(dir, "")

	if !tr.RecordMissing() {
		t.Fatal("runtime recording should be on by default")
	}

	tr.SetRecordMissing(false)

	if tr.RecordMissing() {
		t.Fatal("RecordMissing() = true after SetRecordMissing(false)")
	}

	loc := mockLocalizer{locale: "missing"}

	if got := tr.Tl(loc, "unrecorded"); got != "*unrecorded*" {
		t.Errorf("Tl() = %q, want the missing-translation marker", got)
	}

	tr.Tn(loc, "one", "many", 2)
	tr.Ctl(loc, "context", "unrecorded")
	tr.Ctn(loc, "context", "one", "many", 2)

	if _, err := os.Stat(filepath.Join(dir, DefaultPotFile)); !os.IsNotExist(err) {
		t.Fatalf("POT file should not be written, stat error = %v", err)
	}

	if len(tr.uniqueKeys) != 0 || len(tr.uniqueKeysCtx) != 0 {
		t.Fatal("keys were recorded in memory while recording is off")
	}

	tr.SetRecordMissing(true)
	tr.Tl(loc, "recorded")

	content, err := os.ReadFile(filepath.Join(dir, DefaultPotFile))
	if err != nil {
		t.Fatalf("POT file was not created after re-enabling recording: %v", err)
	}

	if !strings.Contains(string(content), "msgid \"recorded\"") {
		t.Fatalf("POT file misses the recorded key:\n%s", content)
	}
}
