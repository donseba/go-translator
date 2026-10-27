package translator

// Option configures a Translator made with New or Untranslated. Each option
// does what the setter of the same name does.
type Option func(*Translator)

// New returns a translator configured by opts. It differs from NewTranslator
// in one default: lookups record nothing unless SetRecordMissing(true) is
// called, so serving requests never writes to disk and a translator without
// a translations directory never writes a POT file to the working directory.
// Without catalogues, every message is untranslated (see SetUntranslated);
// load them with SetLanguage, SetLanguageFS or LoadLanguages.
//
//	tr := translator.New(
//		translator.WithUntranslated(translator.ShowOriginal),
//		translator.WithLanguageFallback(true),
//	)
//	err := tr.LoadLanguages(siteFiles, moduleFiles)
func New(opts ...Option) *Translator {
	tr := NewTranslator("", "")
	tr.recordMissing.Store(false)

	for _, opt := range opts {
		opt(tr)
	}

	return tr
}

// Untranslated returns a translator without catalogues that shows every
// message as written in the source, formatted with its arguments (see
// ShowOriginal), and records nothing. Use it where texts go through tl or Tl
// but no language is set up, such as default templates rendered without
// translations. opts configure it further. Every call returns a new
// translator.
func Untranslated(opts ...Option) *Translator {
	options := append([]Option{WithUntranslated(ShowOriginal)}, opts...)

	return New(options...)
}

// WithTranslationsDir sets the directory of the catalogues and the POT file,
// see SetTranslationsDir.
func WithTranslationsDir(dir string) Option {
	return func(t *Translator) {
		t.SetTranslationsDir(dir)
	}
}

// WithUntranslated sets what lookups show for a message without
// translation, such as ShowOriginal, see SetUntranslated.
func WithUntranslated(handler func(Miss) string) Option {
	return func(t *Translator) {
		t.SetUntranslated(handler)
	}
}

// WithLanguageFallback turns language fallback on or off, see
// SetLanguageFallback.
func WithLanguageFallback(enabled bool) Option {
	return func(t *Translator) {
		t.SetLanguageFallback(enabled)
	}
}
