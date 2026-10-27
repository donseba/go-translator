package extract

import (
	"bytes"
	"cmp"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// DefaultGenerator is the X-Generator header of a POT file whose catalog
// names no generator.
const DefaultGenerator = "go-translator"

// Message is a translatable text.
type Message struct {
	// Context tells identical texts apart (msgctxt, as with ctl and ctn);
	// usually empty.
	Context string

	// ID is the text as written in the source (msgid).
	ID string

	// Plural is the plural text of tn and ctn (msgid_plural); empty for a
	// text without plural.
	Plural string

	// References are the places that use the text, such as
	// "default/layout.gohtml:12", written as #: comments in the POT file.
	References []string
}

// String describes the message for people: its context, the text, its
// plural and where it is first used, such as
// `[button] "Open" (default/layout.gohtml:6)`.
func (m Message) String() string {
	text := strconv.Quote(m.ID)
	if m.Plural != "" {
		text += " / " + strconv.Quote(m.Plural)
	}

	if m.Context != "" {
		text = fmt.Sprintf("[%s] %s", m.Context, text)
	}

	if len(m.References) > 0 {
		text += " (" + m.References[0] + ")"
	}

	return text
}

// Texts returns ids as messages without context or plural, each referenced by
// reference (no reference when it is empty). Use it with Catalog.Add for texts
// that Go code translates with Tl, such as a list of error titles.
func Texts(reference string, ids ...string) []Message {
	messages := make([]Message, 0, len(ids))

	for _, id := range ids {
		message := Message{ID: id}

		if reference != "" {
			message.References = []string{reference}
		}

		messages = append(messages, message)
	}

	return messages
}

// Warning is a problem found while extracting or adding texts, such as a
// call whose text is not a literal string.
type Warning struct {
	// Reference is the place of the problem, such as
	// "default/layout.gohtml:12"; empty when it is not known.
	Reference string

	// Problem describes the problem and what to do about it.
	Problem string
}

// String returns the warning as "reference: problem", or the problem alone
// without a reference.
func (w Warning) String() string {
	if w.Reference == "" {
		return w.Problem
	}

	return w.Reference + ": " + w.Problem
}

// Header is what a POT file says about itself, above its messages.
type Header struct {
	// Comment is written above the header entry, each line as a "# "
	// comment, for example to tell people how the file is made.
	Comment string

	// ProjectID is the Project-Id-Version header, empty by default.
	ProjectID string

	// Generator is the X-Generator header; empty means DefaultGenerator.
	Generator string
}

// Catalog is a set of translatable texts, from Extract and Add.
type Catalog struct {
	// Messages are the texts, one per context and text, in the order they
	// were first found or added.
	Messages []Message

	// Warnings are the problems found while extracting and adding, in the
	// order they were found.
	Warnings []Warning

	// Sources is the number of templates Extract scanned.
	Sources int

	// Header is the header WritePOT writes.
	Header Header
}

// key identifies a message: its context and text.
type key struct {
	context string
	id      string
}

// Add adds messages to the catalog, such as texts that Go code translates
// (see Texts). A message whose context and text are already in the catalog is
// merged into it: its references are added, and its plural is used when the
// message in the catalog has none. A different plural gives a warning and the
// first one is kept. A message with an empty text is skipped with a warning,
// since msgid "" is the header of a PO file.
func (c *Catalog) Add(messages ...Message) {
	index := make(map[key]int, len(c.Messages))

	for position, message := range c.Messages {
		index[key{message.Context, message.ID}] = position
	}

	for _, message := range messages {
		c.add(index, message)
	}
}

// add adds one message, keeping index up to date.
func (c *Catalog) add(index map[key]int, message Message) {
	reference := ""
	if len(message.References) > 0 {
		reference = message.References[0]
	}

	if message.ID == "" {
		c.Warnings = append(c.Warnings, Warning{
			Reference: reference,
			Problem:   "a text is empty and is left out",
		})

		return
	}

	id := key{message.Context, message.ID}

	position, exists := index[id]
	if !exists {
		message.References = slices.Clone(message.References)
		index[id] = len(c.Messages)
		c.Messages = append(c.Messages, message)

		return
	}

	existing := &c.Messages[position]

	for _, ref := range message.References {
		if !slices.Contains(existing.References, ref) {
			existing.References = append(existing.References, ref)
		}
	}

	// A text used with tl and with tn gets the plural of tn.
	if message.Plural == "" || message.Plural == existing.Plural {
		return
	}

	if existing.Plural == "" {
		existing.Plural = message.Plural

		return
	}

	c.Warnings = append(c.Warnings, Warning{
		Reference: reference,
		Problem: fmt.Sprintf(
			"plural %q of %q differs from %q; the first one is used",
			message.Plural, message.ID, existing.Plural,
		),
	})
}

var escaper = strings.NewReplacer(
	`\`, `\\`,
	`"`, `\"`,
	"\n", `\n`,
	"\r", `\r`,
	"\t", `\t`,
)

// quote writes s as a PO string.
func quote(s string) string {
	return `"` + escaper.Replace(s) + `"`
}

// WritePOT writes the catalog as a gettext template: the header, then every
// message in order with its references as #: comments and empty
// translations. The header has no creation date, so the same catalog always
// gives the same file, and a POT file under version control only changes
// when its texts do.
func (c *Catalog) WritePOT(w io.Writer) error {
	var b strings.Builder

	c.writeHeader(&b)

	for _, message := range c.Messages {
		b.WriteString("\n")

		for _, ref := range message.References {
			b.WriteString("#: " + ref + "\n")
		}

		if message.Context != "" {
			b.WriteString("msgctxt " + quote(message.Context) + "\n")
		}

		b.WriteString("msgid " + quote(message.ID) + "\n")

		if message.Plural == "" {
			b.WriteString("msgstr \"\"\n")

			continue
		}

		b.WriteString("msgid_plural " + quote(message.Plural) + "\n")
		b.WriteString("msgstr[0] \"\"\n")
		b.WriteString("msgstr[1] \"\"\n")
	}

	_, err := io.WriteString(w, b.String())

	return err
}

// writeHeader writes the comment and the header entry.
func (c *Catalog) writeHeader(b *strings.Builder) {
	if c.Header.Comment != "" {
		for line := range strings.SplitSeq(strings.TrimRight(c.Header.Comment, "\n"), "\n") {
			if line == "" {
				b.WriteString("#\n")

				continue
			}

			b.WriteString("# " + line + "\n")
		}
	}

	headers := []string{
		"Project-Id-Version: " + c.Header.ProjectID,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"Content-Transfer-Encoding: 8bit",
		"X-Generator: " + cmp.Or(c.Header.Generator, DefaultGenerator),
	}

	b.WriteString("msgid \"\"\nmsgstr \"\"\n")

	for _, header := range headers {
		b.WriteString(quote(header+"\n") + "\n")
	}
}

// WritePOTFile writes the catalog to the POT file at path (see WritePOT),
// replacing it, and creates its folder when needed.
func (c *Catalog) WritePOTFile(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	var b bytes.Buffer

	if err := c.WritePOT(&b); err != nil {
		return err
	}

	return os.WriteFile(path, b.Bytes(), 0o644)
}
