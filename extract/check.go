package extract

import (
	"errors"
	"io/fs"
	"regexp"
	"strconv"
	"strings"

	"github.com/leonelquinteros/gotext"
)

// PoExtension is the extension of the PO files Check reads.
const PoExtension = ".po"

// Report is one PO file compared with a catalog.
type Report struct {
	// File is the file's name in the checked file system, such as "nl.po".
	File string

	// Language is the file's Language header, or its name without the
	// extension when it has none.
	Language string

	// Forms is the number of plural forms of the language, from the
	// Plural-Forms header; 2 when the header has none.
	Forms int

	// Missing are the catalog's messages the file does not have.
	Missing []Message

	// Empty are the catalog's messages the file has without translation:
	// an empty msgstr, or for a plural message any of its Forms msgstr[n]
	// empty.
	Empty []Message

	// Untranslated are Missing and Empty together, in the catalog's order:
	// everything a translator still has to do.
	Untranslated []Message
}

// Complete reports whether the file translates every message of the catalog.
func (r Report) Complete() bool {
	return len(r.Untranslated) == 0
}

// Check compares catalog with every PO file in the root of fsys, in file
// name order; use fs.Sub for a folder of fsys. A POT file is not read. Texts
// in a PO file that the catalog does not have are not reported.
func Check(catalog *Catalog, fsys fs.FS) ([]Report, error) {
	if fsys == nil {
		return nil, errors.New("extract: check without a file system")
	}

	names, err := fs.Glob(fsys, "*"+PoExtension)
	if err != nil {
		return nil, err
	}

	var reports []Report

	for _, name := range names {
		data, err := fs.ReadFile(fsys, name)
		if err != nil {
			return nil, err
		}

		reports = append(reports, check(catalog, name, data))
	}

	return reports, nil
}

// check compares catalog with the PO file name, data.
func check(catalog *Catalog, name string, data []byte) Report {
	po := gotext.NewPo()
	po.Parse(data)

	domain := po.GetDomain()
	report := Report{
		File:     name,
		Language: po.Language,
		Forms:    pluralForms(domain.PluralForms),
	}

	if report.Language == "" {
		report.Language = strings.TrimSuffix(name, PoExtension)
	}

	if catalog == nil {
		return report
	}

	plain := domain.GetTranslations()
	contexts := domain.GetCtxTranslations()

	for _, message := range catalog.Messages {
		entry := plain[message.ID]
		if message.Context != "" {
			entry = contexts[message.Context][message.ID]
		}

		switch {
		case entry == nil:
			report.Missing = append(report.Missing, message)
		case !translated(entry, message, report.Forms):
			report.Empty = append(report.Empty, message)
		default:
			continue
		}

		report.Untranslated = append(report.Untranslated, message)
	}

	return report
}

var nplurals = regexp.MustCompile(`nplurals\s*=\s*(\d+)`)

// pluralForms reads the number of plural forms from a Plural-Forms header;
// two when it has none.
func pluralForms(header string) int {
	match := nplurals.FindStringSubmatch(header)
	if match == nil {
		return 2
	}

	count, err := strconv.Atoi(match[1])
	if err != nil || count < 1 {
		return 2
	}

	return count
}

// translated reports whether entry translates message: its msgstr, or every
// one of the forms of a plural message, is not empty.
func translated(entry *gotext.Translation, message Message, forms int) bool {
	if message.Plural == "" {
		return entry.Trs[0] != ""
	}

	for form := range forms {
		if entry.Trs[form] == "" {
			return false
		}
	}

	return true
}
