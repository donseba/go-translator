package translator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractTemplateTranslationCalls(t *testing.T) {
	dir := t.TempDir()
	pot := filepath.Join(dir, DefaultPotFile)
	if err := os.WriteFile(pot, []byte("msgid \"\"\nmsgstr \"\"\n\"Content-Type: text/plain; charset=UTF-8\\n\"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	source := `
{{- tl localizer "Hello" -}}
{{ tn localizer "One item" "Many items" 2 }}
{{ tl localizer "One item" }}
{{ ctl localizer "button" "Save" }}
{{ ctn localizer "cart" "One product" "Many products" 2 }}
{{ ctl localizer "cart" "One product" }}
{{ tl .Session.Locale "Nested field" }}
{{ tl $ "Root localizer" }}
{{ tl (localizer) "Helper expression" }}
{{ if eq (tl localizer "Nested call") "value" }}
  {{ tl localizer "Inside if" }}
{{ else }}{{ tl localizer "Inside else" }}{{ end }}
{{ range .Items }}{{ tl $.Loc "Inside range" }}{{ end }}
{{ with .Loc }}{{ tl . "Inside with" }}{{ end }}
{{ define "row" }}{{ tn .Loc "One row" "Many rows" 2 }}{{ end }}
{{ template "row" (tl localizer "Template argument") }}
{{ tl localizer "A \"quoted\" word" }}
{{ $text := tl localizer "Assigned call" }}
{{ unknownHelper . }}
{{ tl localizer .DynamicKey }}
{{/* {{ tl localizer "Commented call" }} */}}
`
	if err := os.WriteFile(filepath.Join(dir, "page.gohtml"), []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	tr := NewTranslator(dir, dir)
	for range 2 {
		if err := tr.CheckMissingTranslations(); err != nil {
			t.Fatal(err)
		}
	}
	for _, key := range []string{"Hello", "Nested field", "Root localizer", "Helper expression", "Nested call", "Inside if", "Inside else", "Inside range", "Inside with", "Template argument", `A "quoted" word`, "Assigned call"} {
		if _, found := tr.uniqueKeys[key]; !found {
			t.Errorf("missing extracted key %q", key)
		}
	}
	for _, item := range []struct{ ctx, key, plural string }{
		{"", "One item", "Many items"},
		{"", "One row", "Many rows"},
		{"button", "Save", ""},
		{"cart", "One product", "Many products"},
	} {
		keys := tr.uniqueKeys
		if item.ctx != "" {
			keys = tr.uniqueKeysCtx[item.ctx]
		}
		if got, ok := keys[item.key]; !ok || got.plural != item.plural {
			t.Errorf("missing translation %q/%q/%q: %#v", item.ctx, item.key, item.plural, got)
		}
	}
	if _, ok := tr.uniqueKeys["Commented call"]; ok {
		t.Error("extracted a template comment")
	}
	content, err := os.ReadFile(pot)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(content), `msgid "Hello"`) != 1 || !strings.Contains(string(content), `msgctxt "cart"`) || !strings.Contains(string(content), `msgid_plural "Many products"`) {
		t.Fatalf("incorrect POT output: %s", content)
	}
}

func TestExtractMalformedTemplateReturnsFilename(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "broken.gohtml")
	if err := os.WriteFile(filename, []byte(`{{ tl localizer "unterminated }}`), 0600); err != nil {
		t.Fatal(err)
	}
	err := NewTranslator(dir, dir).ScanFiles(dir)
	if err == nil || !strings.Contains(err.Error(), filename) {
		t.Fatalf("expected parse error with filename, got %v", err)
	}
}
