package translator

import (
	"strings"
	"testing"
)

// TestHeaderForLanguageIsNeutral checks that a language header names no
// project or team.
func TestHeaderForLanguageIsNeutral(t *testing.T) {
	for _, language := range []string{"nl", "pt_PT", "xx_XX"} {
		h := GetHeaderForLanguage(language)

		if h.ProjectID != "PACKAGE VERSION" || h.LanguageTeam != "" || h.MIMEVersion != "1.0" || h.Language != language {
			t.Errorf("%s: %+v", language, h)
		}

		if strings.Contains(h.HeaderString(), "Go-FORM") {
			t.Errorf("%s: header names Go-FORM", language)
		}
	}

	if got := GetHeaderForLanguage("pl").PluralForms; !strings.HasPrefix(got, "nplurals=3;") {
		t.Errorf("Polish plural forms: %q", got)
	}
}
