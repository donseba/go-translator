package translator

import (
	"fmt"
	"io/fs"
	"slices"
	"strings"

	"github.com/leonelquinteros/gotext"
)

// SetLanguageFS loads the catalogue of lang from the PO file name in fsys,
// such as an embed.FS, like SetLanguage does from the translations
// directory. A catalogue already loaded under lang is replaced.
//
// When the file cannot be read, the error matches both ErrorLanguageNotFound
// and the cause, such as fs.ErrNotExist, with errors.Is. A catalogue without
// a Language header is loaded as lang. Configure catalogues before serving
// requests.
func (t *Translator) SetLanguageFS(lang string, fsys fs.FS, name string) error {
	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrorLanguageNotFound, err)
	}

	t.setCatalogue(lang, data)

	return nil
}

// LoadLanguages loads every PO file (DefaultPoExtension) in the root of each
// of sources, each as the language its file name names: nl.po as "nl",
// en_US.po as "en_US". Use fs.Sub for a folder of a source; nil sources are
// skipped. The POT file is not loaded.
//
// Sources are layers, the first one on top: the first source that has
// <lang>.po loads it, replacing a catalogue already loaded under lang, and
// the same file in the sources after it fills in the messages it misses. An
// entry left empty counts as missing. A later file without a Language header
// counts as the language of its name.
//
//	// The site's own catalogues win over the defaults of its modules.
//	err := tr.LoadLanguages(os.DirFS("translations"), moduleA, moduleB)
//
// It stops at the first file that cannot be loaded or merged, with an error
// naming it.
func (t *Translator) LoadLanguages(sources ...fs.FS) error {
	sources = slices.DeleteFunc(slices.Clone(sources), func(source fs.FS) bool {
		return source == nil
	})

	loaded := map[string]bool{}

	for index, source := range sources {
		names, err := fs.Glob(source, "*"+DefaultPoExtension)
		if err != nil {
			return err
		}

		for _, name := range names {
			lang := strings.TrimSuffix(name, DefaultPoExtension)
			if loaded[lang] {
				continue
			}

			loaded[lang] = true

			if err := t.SetLanguageFS(lang, source, name); err != nil {
				return fmt.Errorf("load %s: %w", name, err)
			}

			// The same file in the later sources fills in the messages it misses.
			if err := t.mergeFiles(lang, sources[index+1:], name); err != nil {
				return fmt.Errorf("load %s: %w", name, err)
			}
		}
	}

	return nil
}

// setCatalogue parses data as the catalogue of lang.
func (t *Translator) setCatalogue(lang string, data []byte) {
	po := gotext.NewPo()
	po.Parse(data)

	if po.Language == "" {
		po.Language = lang
	}

	t.languages[lang] = po
}
