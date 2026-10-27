package translator

import (
	"fmt"
)

// lookupKind is the lookup a Miss comes from. It chooses the marker of
// Miss.Marker and whether Miss.Original picks a plural form.
type lookupKind int

const (
	kindTL lookupKind = iota
	kindTN
	kindCTL
	kindCTN
)

// Miss describes a lookup that found no translation: the language is not
// loaded, the message is not in its catalogue, or its translation is empty.
// Handlers set with SetUntranslated receive it and return the text to show.
type Miss struct {
	// Locale is the locale of the Localizer, as passed to the lookup.
	Locale string

	// Context is the message context (msgctxt), "" without one.
	Context string

	// Key is the message (msgid): the key of tl and ctl, the singular of
	// tn and ctn.
	Key string

	// Plural is the plural message (msgid_plural) of tn and ctn, "" for tl
	// and ctl.
	Plural string

	// N is the count of tn and ctn that chooses the plural form, 0 for tl
	// and ctl.
	N int

	// Args are the formatting arguments of the lookup.
	Args []any

	kind      lookupKind
	separator string
}

// Marker returns the marker go-translator shows for a message without
// translation by default, such as "*key*" for tl. It is formatted with the
// DefaultNoTranslation variables at the time of the call; the arguments of
// the lookup are not applied.
func (m Miss) Marker() string {
	switch m.kind {
	case kindTN:
		return fmt.Sprintf(DefaultNoTranslationTN, m.Key, m.Plural)
	case kindCTL:
		return fmt.Sprintf(DefaultNoTranslationCTL, m.Context, m.Key)
	case kindCTN:
		return fmt.Sprintf(DefaultNoTranslationCTN, m.Context, m.Key, m.Plural)
	default:
		return fmt.Sprintf(DefaultNoTranslationTL, m.Key)
	}
}

// Original returns the message as written in the source, as gettext shows
// it: for tn and ctn the singular for a count of 1 and the plural for every
// other count, formatted with the arguments of the lookup, without a prefix
// before the prefix separator (see SetPrefixSeparator). Without arguments the
// text is returned as is, so a literal % stays.
func (m Miss) Original() string {
	text := m.Key

	counted := m.kind == kindTN || m.kind == kindCTN
	if counted && m.N != 1 {
		text = m.Plural
	}

	if len(m.Args) > 0 {
		text = fmt.Sprintf(text, m.Args...)
	}

	return removeSeparator(text, m.separator)
}

// ShowOriginal shows a message without translation as written in the source,
// formatted with its arguments, as gettext does (see Miss.Original).
func ShowOriginal(m Miss) string {
	return m.Original()
}

// SetUntranslated sets what lookups show for a message without translation:
// by default a marker such as "*key*" (see Miss.Marker), with ShowOriginal
// the message as written, with its arguments. Any other function of a Miss
// can decide for itself, for example to log the miss and return
// m.Original(). nil restores the marker.
//
// It applies to the template functions tl, tn, ctl and ctn, to Tl, Tn, Ctl
// and Ctn, and to the text the Lookup functions return with false. Configure
// it before serving requests.
func (t *Translator) SetUntranslated(handler func(Miss) string) {
	t.untranslated = handler
}

// miss returns the text for a message without translation, as set with
// SetUntranslated.
func (t *Translator) miss(m Miss) string {
	m.separator = t.prefixSeparator

	if t.untranslated == nil {
		return m.Marker()
	}

	return t.untranslated(m)
}
