// Package extract finds the texts that templates translate with go-translator's
// template functions (tl, tn, ctl and ctn), writes them to a gettext template
// (a POT file) and checks which texts each language's PO file still misses.
//
// The workflow for translators: extract the texts and write the POT file, open
// a language's PO file in Poedit, choose Translation → Update from POT File,
// translate the new texts and save. Check lists what each language still
// misses.
//
//	sources, err := extract.Files(os.DirFS("templates"))
//	catalog, err := extract.Extract(sources...)
//	catalog.Add(extract.Texts("go:errors", "Not found", "Forbidden")...)
//	err = catalog.WritePOTFile("translations/translations.pot")
//	reports, err := extract.Check(catalog, os.DirFS("translations"))
//
// Extraction parses the templates instead of recording texts while pages are
// rendered, so it finds every text, also on pages nobody visited, and a
// running application never writes to its catalogues. Templates are parsed
// without their functions, so no helper needs to be registered. The result is
// deterministic: the same sources give the same catalog and the same POT file.
package extract

import (
	"cmp"
	"fmt"
	"io/fs"
	"slices"
	"strings"
	"text/template/parse"
)

// DefaultExtension is the file extension Files looks for when it is given
// none.
const DefaultExtension = ".gohtml"

// Source is one template to scan.
type Source struct {
	// Name names the template in references and errors, for example
	// "default/layout.gohtml".
	Name string

	// Text is the template's source.
	Text string
}

// Function describes which arguments of a translation function hold its
// texts. The arguments are, in order: Skip arguments that are not texts (such
// as the localizer), the context when Context is set, the text (msgid), and
// the plural text (msgid_plural) when Plural is set.
type Function struct {
	// Skip is the number of arguments before the texts: 1 for the localizer
	// of tl, tn, ctl and ctn.
	Skip int

	// Context reports whether the first text is a context (msgctxt), as with
	// ctl and ctn.
	Context bool

	// Plural reports whether the text is followed by its plural, as with tn
	// and ctn.
	Plural bool
}

// texts is the number of literal texts the function takes.
func (f Function) texts() int {
	count := 1

	if f.Context {
		count++
	}

	if f.Plural {
		count++
	}

	return count
}

// DefaultFunctions returns go-translator's template functions, which each take
// a localizer before their texts:
//
//	tl  localizer text
//	tn  localizer text plural
//	ctl localizer context text
//	ctn localizer context text plural
//
// The map is new on every call, so it can be changed before it is given to an
// Extractor.
func DefaultFunctions() map[string]Function {
	return map[string]Function{
		"tl":  {Skip: 1},
		"tn":  {Skip: 1, Plural: true},
		"ctl": {Skip: 1, Context: true},
		"ctn": {Skip: 1, Context: true, Plural: true},
	}
}

// Extractor finds translation calls in templates. The zero value looks for
// DefaultFunctions between "{{" and "}}".
type Extractor struct {
	// Functions are the translation functions to look for, by name. Nil
	// means DefaultFunctions.
	Functions map[string]Function

	// LeftDelim and RightDelim are the template's action delimiters, as
	// set with template.Delims. Empty means "{{" and "}}".
	LeftDelim  string
	RightDelim string
}

// Extract finds the tl, tn, ctl and ctn calls in sources, see
// Extractor.Extract.
func Extract(sources ...Source) (*Catalog, error) {
	return Extractor{}.Extract(sources...)
}

// Extract finds the translation calls in sources and returns their texts as a
// catalog: one message per context and text, in the order they first appear in
// the sources, each with the places that use it as references (name:line).
//
// A call whose text is not a literal string, such as {{tl localizer .Label}},
// or comes from a pipeline, such as {{"Save" | tl localizer}}, cannot be
// extracted and gives a warning. A text used with and without a plural gets
// the plural; two different plurals of one text give a warning, and the first
// one is used.
//
// A template that does not parse is an error naming its source.
func (x Extractor) Extract(sources ...Source) (*Catalog, error) {
	functions := x.Functions
	if functions == nil {
		functions = DefaultFunctions()
	}

	left := cmp.Or(x.LeftDelim, "{{")
	right := cmp.Or(x.RightDelim, "}}")

	scan := scanner{functions: functions}

	for index, source := range sources {
		if err := scan.source(index, source, left, right); err != nil {
			return nil, err
		}
	}

	return scan.catalog(len(sources)), nil
}

// Files returns the files in fsys with one of extensions as sources, named by
// their path in fsys, in lexical order. Without extensions it returns the
// DefaultExtension files. A nil fsys has no sources.
func Files(fsys fs.FS, extensions ...string) ([]Source, error) {
	if fsys == nil {
		return nil, nil
	}

	if len(extensions) == 0 {
		extensions = []string{DefaultExtension}
	}

	var sources []Source

	err := fs.WalkDir(fsys, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() || !hasExtension(path, extensions) {
			return nil
		}

		text, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}

		sources = append(sources, Source{Name: path, Text: string(text)})

		return nil
	})
	if err != nil {
		return nil, err
	}

	return sources, nil
}

// hasExtension reports whether path ends in one of extensions.
func hasExtension(path string, extensions []string) bool {
	return slices.ContainsFunc(extensions, func(extension string) bool {
		return strings.HasSuffix(path, extension)
	})
}

// found is one translation call with literal texts, or a warning about one
// without, at a position in a source.
type found struct {
	source  int
	pos     parse.Pos
	message Message
	warning *Warning
}

// scanner collects the calls of all sources.
type scanner struct {
	functions map[string]Function
	found     []found
}

// source parses one source and collects its calls.
func (s *scanner) source(index int, source Source, left, right string) error {
	trees := map[string]*parse.Tree{}
	tree := parse.New(source.Name)
	tree.Mode = parse.SkipFuncCheck

	if _, err := tree.Parse(source.Text, left, right, trees); err != nil {
		return fmt.Errorf("extract %s: %w", source.Name, err)
	}

	// Every {{define}} becomes its own tree. Positions are offsets in the
	// whole source and the calls are sorted by them, so the order of the
	// trees does not matter.
	for _, tree := range trees {
		s.walk(index, source, tree.Root)
	}

	return nil
}

// walk collects the calls in node and the nodes below it.
func (s *scanner) walk(index int, source Source, node parse.Node) {
	switch node := node.(type) {
	case *parse.ListNode:
		if node == nil {
			return
		}

		for _, child := range node.Nodes {
			s.walk(index, source, child)
		}

	case *parse.ActionNode:
		s.walk(index, source, node.Pipe)

	case *parse.IfNode:
		s.branch(index, source, &node.BranchNode)

	case *parse.RangeNode:
		s.branch(index, source, &node.BranchNode)

	case *parse.WithNode:
		s.branch(index, source, &node.BranchNode)

	case *parse.TemplateNode:
		s.walk(index, source, node.Pipe)

	case *parse.PipeNode:
		if node == nil {
			return
		}

		for _, command := range node.Cmds {
			s.walk(index, source, command)
		}

	case *parse.ChainNode:
		s.walk(index, source, node.Node)

	case *parse.CommandNode:
		s.command(index, source, node)

		for _, arg := range node.Args {
			s.walk(index, source, arg)
		}
	}
}

// branch collects the calls of an if, range or with: in its pipeline, its
// body and its else.
func (s *scanner) branch(index int, source Source, node *parse.BranchNode) {
	s.walk(index, source, node.Pipe)
	s.walk(index, source, node.List)
	s.walk(index, source, node.ElseList)
}

// command records a translation call: the function, the arguments it skips,
// then the literal texts.
func (s *scanner) command(index int, source Source, command *parse.CommandNode) {
	if len(command.Args) == 0 {
		return
	}

	identifier, ok := command.Args[0].(*parse.IdentifierNode)
	if !ok {
		return
	}

	function, ok := s.functions[identifier.Ident]
	if !ok {
		return
	}

	pos := command.Position()
	ref := fmt.Sprintf("%s:%d", source.Name, line(source.Text, pos))
	first := 1 + function.Skip
	count := function.texts()

	if len(command.Args) < first+count {
		s.warn(index, pos, ref, identifier.Ident, "has its text from a pipeline")

		return
	}

	texts := make([]string, count)

	for i := range texts {
		literal, ok := command.Args[first+i].(*parse.StringNode)
		if !ok {
			s.warn(index, pos, ref, identifier.Ident, "has a text that is not a literal string: "+command.Args[first+i].String())

			return
		}

		texts[i] = literal.Text
	}

	var message Message

	if function.Context {
		message.Context = texts[0]
		texts = texts[1:]
	}

	message.ID = texts[0]

	if function.Plural {
		message.Plural = texts[1]
	}

	message.References = []string{ref}

	s.found = append(s.found, found{source: index, pos: pos, message: message})
}

// warn records a call whose texts cannot be extracted.
func (s *scanner) warn(index int, pos parse.Pos, ref, function, problem string) {
	warning := Warning{
		Reference: ref,
		Problem: function + " " + problem +
			"; name each possible text in a literal call, for example in a {{define}} that is never used," +
			" or updating a PO file from the POT file drops its translations",
	}

	s.found = append(s.found, found{source: index, pos: pos, warning: &warning})
}

// catalog merges the calls, in the order they appear, into a catalog.
func (s *scanner) catalog(sources int) *Catalog {
	slices.SortStableFunc(s.found, func(a, b found) int {
		return cmp.Or(cmp.Compare(a.source, b.source), cmp.Compare(a.pos, b.pos))
	})

	catalog := &Catalog{Sources: sources}

	var messages []Message

	for _, call := range s.found {
		if call.warning != nil {
			catalog.Warnings = append(catalog.Warnings, *call.warning)

			continue
		}

		messages = append(messages, call.message)
	}

	catalog.Add(messages...)

	return catalog
}

// line is the line number of pos in text.
func line(text string, pos parse.Pos) int {
	end := min(int(pos), len(text))

	return 1 + strings.Count(text[:end], "\n")
}
