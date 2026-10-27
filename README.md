<p align="center">
    <a href="https://docs.gowebthings.com/go-translator">
        <img src="./assets/go-translator-logo.png" alt="go-translator" height="70">
    </a>
</p>

# go-translator

[Documentation](https://docs.gowebthings.com/go-translator) · Part of [go-webthings](https://gowebthings.com/components).

go-translator provides gettext PO catalogues, plural forms, translation contexts, and helpers for Go `html/template`. A small `Localizer` interface supplies the locale for each request, so the application controls locale selection.

## Installation

```sh
go get github.com/donseba/go-translator/v2
```

## Catalogue layout

Create a `translations` directory and a `templates` directory containing your `.gohtml` files. Put one PO file per locale in `translations`, for example `en_US.po` and `nl_NL.po`. Locale strings must match the names used with `SetLanguage`.

A minimal `translations/en_US.po` for the example:

```po
msgid ""
msgstr ""
"Language: en_US\n"
"Content-Type: text/plain; charset=UTF-8\n"
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

msgid "Hello, World!"
msgstr "Hello, World!"

msgid "%d message"
msgid_plural "%d messages"
msgstr[0] "%d message"
msgstr[1] "%d messages"
```

## Quick start

```go
package main

import (
	"html/template"
	"log"
	"os"

	"github.com/donseba/go-translator/v2"
)

type Locale string

func (l Locale) GetLocale() string { return string(l) }

func main() {
	tr := translator.NewTranslator("translations", "templates")
	if err := tr.SetLanguage("en_US"); err != nil {
		log.Fatal(err)
	}
	if err := tr.CheckMissingTranslations(); err != nil {
		log.Fatal(err)
	}
	page := template.Must(template.New("welcome").Funcs(tr.FuncMap()).Parse(
		`<h1>{{ tl .Loc "Hello, World!" }}</h1><p>{{ tn .Loc "%d message" "%d messages" .Count .Count }}</p>`,
	))
	if err := page.Execute(os.Stdout, struct {
		Loc   Locale
		Count int
	}{Loc: "en_US", Count: 5}); err != nil {
		log.Fatal(err)
	}
}
```

`SetLanguage` loads a catalogue; it does not change a global current locale. Load every supported language and configure the translator before serving concurrent requests. Supply a request-specific `Localizer` to each lookup or template render:

```go
type Localizer interface {
    GetLocale() string
}
```

By default a lookup only uses the catalogue loaded under exactly the locale of the `Localizer`, and missing languages or translations produce visible markers such as `*key*`. See [Language fallback](#language-fallback) and [Messages without translation](#messages-without-translation) to change either.

## Creating a translator with options

`New` configures a translator in one call. Each option does what the setter of the same name does:

```go
tr := translator.New(
	translator.WithTranslationsDir("translations"),
	translator.WithUntranslated(translator.ShowOriginal),
	translator.WithLanguageFallback(true),
)
err := tr.LoadLanguages(os.DirFS("translations"))
```

| Option | Setter |
|--------|--------|
| `WithTranslationsDir(dir)` | `SetTranslationsDir` |
| `WithUntranslated(fn)` | `SetUntranslated` |
| `WithLanguageFallback(bool)` | `SetLanguageFallback` |

`New` differs from `NewTranslator` in one default: it records nothing while translating (see [Extracting translation keys](#extracting-translation-keys)) unless you call `SetRecordMissing(true)`. Serving requests then never writes to disk, and a translator without a translations directory never writes `translations.pot` to the working directory.

`Untranslated()` is a translator without catalogues that shows every message as written, with its arguments (`ShowOriginal`), and records nothing. Use it where texts go through `tl` or `Tl` but no language is set up, such as default templates rendered without translations. It takes the same options and returns a new translator on every call.

```go
tr := translator.Untranslated()
tr.Tn(loc, "%d file", "%d files", 2, 2) // "2 files"
```

## Template functions

| Helper | Arguments after the localizer |
|--------|------------------------------|
| `tl` | key, optional formatting arguments |
| `tn` | singular key, plural key, count, optional formatting arguments |
| `ctl` | context, key, optional formatting arguments |
| `ctn` | context, singular key, plural key, count, optional formatting arguments |

```gotemplate
{{ tl .Loc "Hello, %s" .Name }}
{{ tn .Loc "%d message" "%d messages" .Count .Count }}
{{ ctl .Loc "navigation" "Home" }}
{{ ctn .Loc "inbox" "%d message" "%d messages" .Count .Count }}
```

Template arguments are separated by spaces. In the plural examples, the first `.Count` chooses the plural form; the second formats `%d`.

Go code can use `Tl`, `Tn`, `Ctl`, and `Ctn`, with the same arguments as the template helpers:

```go
tr.Tl(loc, "Hello, %s", name)
tr.Tn(loc, "%d message", "%d messages", count, count)
tr.Ctl(loc, "navigation", "Home")
tr.Ctn(loc, "inbox", "%d message", "%d messages", count, count)
```

`ctl`/`Ctl` and `ctn`/`Ctn` without a context (`""`) behave exactly like `tl`/`Tl` and `tn`/`Tn`.

To know whether a translation exists, use `Lookup`, `LookupN`, `LookupC`, and `LookupCN`. They take the same arguments as `Tl`, `Tn`, `Ctl`, and `Ctn`, return the same text, and also report whether a translation was found:

```go
text, found := tr.Lookup(loc, "Hello, %s", name)
if !found {
	// text is what SetUntranslated shows, by default "*Hello, %s*"
}
```

## Messages without translation

By default a message without translation shows as a marker: `*key*` for `tl`, `*singular/plural*` for `tn`, `*context/key*` for `ctl`, and `*context/singular/plural*` for `ctn`. The arguments are not applied. That makes missing translations easy to spot during development.

`SetUntranslated` changes what is shown, for the template helpers, for `Tl`, `Tn`, `Ctl`, and `Ctn`, and for the text the `Lookup` functions return:

```go
tr.SetUntranslated(translator.ShowOriginal)
```

| Mode | `tl "Hello, %s" "Bob"` | `tn "%d file" "%d files" 2 2` |
|------|------------------------|-------------------------------|
| Default | `*Hello, %s*` | `*%d file/%d files*` |
| `translator.ShowOriginal` | `Hello, Bob` | `2 files` |

`ShowOriginal` shows the message as written in the source, as gettext does: formatted with its arguments (a message without arguments is shown as is, so a literal `%` stays), without a prefix before the prefix separator, and for plurals the singular for 1 and the plural for every other count.

A mode is a function of a `translator.Miss`, so you can write your own, for example to log what is missing. `Miss` holds the locale, context, key, plural, count and arguments of the lookup; `Marker()` and `Original()` return what the two built-in modes show:

```go
tr.SetUntranslated(func(m translator.Miss) string {
	slog.Debug("missing translation", "locale", m.Locale, "key", m.Key)

	return m.Original()
})
```

`SetUntranslated(nil)` restores the markers. Configure this before serving requests.

## Extracting translation keys

`CheckMissingTranslations()` scans the files with the `TemplateExtension` (`.gohtml` by default) below the configured template directory and adds literal keys to `translations/translations.pot`. Run it during development or before serving requests. When the POT file does not exist yet, it is created with a gettext header first; the translations directory itself must exist. A translator made with `New` has no template directory: call `SetTemplateDir` before `CheckMissingTranslations`, or write the POT file with the [extract package](#the-extract-package).

Extraction parses Go template syntax without requiring application helper functions to be registered. It supports `localizer` helpers, dot and variable localizers, nested expressions, template definitions, and escaped literal keys. Comments and dynamic keys are skipped. Malformed templates return an error containing the source filename.

Runtime lookups also record previously unseen keys, including keys constructed in Go code. An empty key is not recorded: `msgid ""` is the header of the POT file. Those POT updates are serialized for concurrent lookups. Allow the catalogue directory to be writable when using runtime discovery; keep language loading, configuration, and catalogue editing outside request handling.

Runtime recording is on by default for `NewTranslator`; `New` and `Untranslated` record nothing unless you call `SetRecordMissing(true)`. Turn it off where serving requests must never write to disk, for example in production:

```go
tr.SetRecordMissing(false)
```

With recording off, lookups only translate; `CheckMissingTranslations` and `ScanFiles` still extract keys when you call them. `RecordMissing()` reports the current setting.

The POT file is read the first time it is needed: when a lookup records a key, or when `CheckMissingTranslations` runs. A translator with recording off that only translates never reads it. `NewTranslator` records by default, so the first lookup of `NewTranslator("", "")`, a translator without catalogues, reads `translations.pot` in the working directory and records the key there, creating the file if needed; after `SetRecordMissing(false)` it does not look for that file. `New` and `Untranslated` record nothing by default. `SetTranslationsDir` and `SetPotFile` take effect for the next key recorded.

`CheckMissingTranslations` returns the errors of every key it could not add to the POT file, joined. A lookup has no error to return, so errors of runtime recording, such as a POT file that cannot be written, are logged with `log/slog`.

A typical workflow: extract keys into `translations.pot`, open a language `.po` file in Poedit, choose *Translation → Update from POT File*, translate the new entries, and save.

### The extract package

`CheckMissingTranslations` only appends to the POT file. The `github.com/donseba/go-translator/v2/extract` package writes the whole file instead, with source references, and checks which texts each language still misses. It reads templates from any `fs.FS` or from memory, and it never needs a `Translator`:

```go
import "github.com/donseba/go-translator/v2/extract"

sources, err := extract.Files(os.DirFS("templates")) // every .gohtml file
if err != nil {
	return err
}

catalog, err := extract.Extract(sources...)
if err != nil {
	return err
}

for _, warning := range catalog.Warnings {
	log.Println("warning:", warning)
}

// Texts that Go code translates with Tl are added by hand.
catalog.Add(extract.Texts("go:errors", "Not found", "Forbidden")...)

err = catalog.WritePOTFile("translations/translations.pot")
```

- **Messages.** `Extract` finds the literal texts of `tl`, `tn`, `ctl` and `ctn` calls: one `Message` per context and text, with its plural and its references (`name:line`), in the order they first appear in the sources. The same sources always give the same catalog.
- **Warnings.** A call whose text is not a literal, such as `{{tl localizer .Label}}`, or comes from a pipeline, such as `{{"Save" | tl localizer}}`, cannot be extracted and gives a `Warning`. Name each possible text in a literal call, for example in a `{{define}}` that is never used, or updating a `.po` file from the POT file drops its translations. A text with two different plurals also gives a warning, and the first plural is used; an empty text is left out with a warning.
- **Sources.** A `Source` is a name and a template text, so generated templates can be extracted too. `Files(fsys, extensions...)` reads every file with one of the extensions (`extract.DefaultExtension`, `.gohtml`, by default) from an `fs.FS`, named by its path.
- **Other functions.** An `Extractor` looks for other function names or argument shapes, and other delimiters:

  ```go
  functions := extract.DefaultFunctions() // tl, tn, ctl, ctn
  functions["t"] = extract.Function{}     // [[t "Text"]], no localizer
  x := extract.Extractor{Functions: functions, LeftDelim: "[[", RightDelim: "]]"}
  catalog, err := x.Extract(sources...)
  ```

- **Adding texts.** `Catalog.Add(messages...)` adds texts that do not come from templates, merged with the extracted ones as `Extract` merges calls. `extract.Texts(reference, ids...)` turns a list of plain texts into messages.
- **The POT file.** `Catalog.WritePOT(w)` and `Catalog.WritePOTFile(path)` replace the file: texts no longer used are gone, and the header has no creation date, so the file only changes when the texts do. `Catalog.Header` sets a comment above it, the `Project-Id-Version` and the `X-Generator` (`extract.DefaultGenerator`, `go-translator`, by default).

`extract.Check(catalog, fsys)` compares the catalog with every `.po` file in the root of `fsys` and returns a `Report` per file, in name order:

```go
reports, err := extract.Check(catalog, os.DirFS("translations"))
for _, report := range reports {
	for _, message := range report.Untranslated {
		fmt.Printf("%s: %s\n", report.File, message) // nl.po: "Hello, %s" (default/layout.gohtml:2)
	}
}
```

`Missing` are the texts the file does not have, `Empty` those it has without translation, and `Untranslated` both, in the catalog's order. A plural text needs every plural form of its language (`Forms`, from the `Plural-Forms` header): three for Polish, one for Japanese. `Complete()` reports whether nothing is left to translate.

## Loading catalogues

`SetLanguage("nl_NL")` loads `nl_NL.po` from the translations directory. Catalogues can also come from any `fs.FS`, such as an `embed.FS`:

```go
//go:embed translations
var files embed.FS

// One catalogue, under a name of your choice:
err := tr.SetLanguageFS("nl_NL", files, "translations/nl_NL.po")

// Every *.po file in the root of a file system, named after the file:
// translations/nl.po as "nl", translations/en_US.po as "en_US".
sub, _ := fs.Sub(files, "translations")
err = tr.LoadLanguages(sub)
```

`LoadLanguages` takes several file systems as layers, the first one on top. The first source that has `<lang>.po` loads it, and the same file in the sources after it fills in the messages it misses; an entry left empty counts as missing. A later file without a `Language` header counts as the language of its name, and nil sources are skipped. A later file whose `Language` header names a different base language, or whose plural messages choose different forms, is refused: `LoadLanguages` stops with `load <file>: invalid translation catalogue: …`:

```go
// The application's catalogues win over those of its modules.
err := tr.LoadLanguages(os.DirFS("translations"), moduleA, moduleB)
```

A catalogue without a `Language` header is loaded as the language it was loaded as. When a file cannot be read, the error matches both `translator.ErrorLanguageNotFound` and the cause, such as `fs.ErrNotExist`, with `errors.Is`. Load every language before serving requests.

## Language fallback

Language fallback is off by default. Turn it on to let a lookup use the catalogue of a similar language:

```go
tr.SetLanguageFallback(true)
```

`LanguageFallback()` reports the current setting.

A lookup then tries, message by message, and uses the first catalogue that translates the message:

1. the catalogue loaded under the locale itself, such as `nl_NL`;
2. one loaded under the same locale written differently: case, `-` or `_`, and an encoding or modifier (`.UTF-8`, `@euro`) do not count, so `nl-nl` finds `nl_NL`;
3. the base language, such as `nl` for `nl_NL`;
4. another region of the same language, such as `nl_NL` for `nl` or `nl_BE`: first the region named like the language (`nl_NL`, `de_DE`), then the others in alphabetical order.

So a site with only `nl.po` serves `nl_NL` and `nl_BE` visitors, and an `nl_BE.po` with a few regional words falls back to `nl.po` for everything else. Fallback applies to lookups only; functions that edit or write a catalogue still need its exact name.

## Creating and editing languages

`EnsureLanguage("fr")` creates `fr.po` with language and plural-form headers (`GetHeaderForLanguage`) if missing, then loads it. A regional locale without its own plural rules gets those of its base language: `pl_PL` gets the Polish rules of `pl`. The translations directory must already exist. Repeated calls preserve an existing catalogue. `AddLanguage` is deprecated; use `SetLanguage`.

`MergeDefaults(language, files, path)` adds module-provided PO messages to an already loaded language. Translated application entries win. Blank entries (`msgstr ""`, or a plural whose `msgstr[n]` forms are all empty, as left by an "Update from POT" in Poedit) count as missing, so the defaults fill them; a partly translated plural is kept as is. Contexts and all plural forms are preserved. Catalogues must be in the same base language: `nl_NL` defaults merge into `nl`, and `nl` into `nl_NL`. The incoming catalogue needs a `Language` header; `LoadLanguages` layers take the file name instead. When the incoming catalogue contains plural messages, its plural rules must choose the same forms; they are compared by evaluating them, so `plural=n != 1` and `plural=(n != 1)` match. A catalogue that does not fit returns an error matching `ErrorInvalidCatalogue`. Merging does not rewrite the application PO file and must run before serving requests.

At lookup time an empty translation behaves like a missing one: `Tl`, `Ctl`, `Tn` and `Ctn` return the untranslated text (by default the marker, for example `*key*`) instead of an empty string.

Use `SetTL`, `SetTLN`, `SetCTL`, `SetCTN`, and `Write` for catalogue editing before requests start. `SetCTL` and `SetCTN` take the key before the context, unlike `Ctl` and `Ctn`: `SetCTL(loc, key, ctx, value)`, `SetCTN(loc, key, plural, ctx, n, value)`. `Write` saves `<locale>.po` in the translations directory. `Details` returns a loaded catalogue's headers, or nil for a language that is not loaded. `NewLanguage` creates a catalogue in memory, `SetDetails` and `AddDetails` set and add headers, and `SetRefs` records source references for a key. `WritePOTFile(path)` and `WritePOFile(path, language)` create an empty POT or PO file with a header (`DefaultHeader`) when it does not exist. The markers for missing translations are the variables `DefaultNoTranslationTL`, `DefaultNoTranslationTN`, `DefaultNoTranslationCTL` and `DefaultNoTranslationCTN`. If updating `plurals.json`, regenerate the plural rules with:

```sh
go run tools/generate_templates.go
```

## Translation contexts and prefixes

Prefer explicit `ctl` / `ctn` contexts when the same text needs different translations. A calendar is a typical case: English writes the long and the short name of May the same, but most languages do not. With `LookupC` it uses a translation only when there is one, without comparing strings:

```go
long, _ := tr.LookupC(loc, "month", "May")        // "toukokuu" in Finnish
short, _ := tr.LookupC(loc, "month-short", "May") // "toukok."

pattern := "d MMM y"
if own, ok := tr.LookupC(loc, "date", "medium"); ok {
	pattern = own
}
```

A lookup with a context only finds entries with that context (`msgctxt`), never the entry without one. Legacy prefixed keys remain supported: `SetPrefixSeparator` changes the default `__.` separator, and the prefix is removed from translated output.

## Upgrading

### From v1.7.0 to v2

The module path is now `github.com/donseba/go-translator/v2`. Change your imports from `github.com/donseba/go-translator` to `github.com/donseba/go-translator/v2`; the package name stays `translator`. The new `extract` package is `github.com/donseba/go-translator/v2/extract`.

**Compile break.** `Tn` and `Ctn` take formatting arguments, as the template helpers `tn` and `ctn` already did: `Tn(loc, singular, plural, n, args...)` and `Ctn(loc, ctx, singular, plural, n, args...)`. Existing calls compile and return the same text. A method value or method expression of `Tn` or `Ctn` used as a func type with the old signature, and an interface that declares the old signature, no longer compile; add `args ...any` there.

These existing behaviours change:

**Lookups**

| What | Before | After |
|------|--------|-------|
| `ctl`/`Ctl` with context `""` for a language that is not loaded | `*/key*` | as `tl`: `*key*` |
| `ctn`/`Ctn` with context `""` | never translated, marker `*/singular/plural*` | as `tn`/`Tn`: translated, marker `*singular/plural*` |

**Loading**

| What | Before | After |
|------|--------|-------|
| `SetLanguage` with a missing file | returns `ErrorLanguageNotFound` itself | returns an error wrapping `ErrorLanguageNotFound` and the cause; use `errors.Is`, not `==` |
| `SetLanguage` with a catalogue without `Language` header, also an empty file | `ErrorLanguageNotFound` | loaded as the language it was loaded as |

**Headers and files**

| What | Before | After |
|------|--------|-------|
| `GetHeaderForLanguage`, also in the files `EnsureLanguage` creates | project and team `Go-FORM`; `pl_PL` got English plural rules | `PACKAGE VERSION`, no team, `MIME-Version: 1.0`; `pl_PL` gets the rules of `pl` |
| `Details` for a language that is not loaded | panics | returns nil |
| `Details` after `NewLanguage` | nil | the headers set by `NewLanguage` and `SetDetails` |
| `NewLanguage` header | `X-Generator:` (with a colon in the key) | `X-Generator` |
| `Write` | `PO-Revision-Date` in RFC 3339 | gettext format, `2006-01-02 15:04-0700`; the file extension is `DefaultPoExtension` |

**`MergeDefaults`**

| What | Before | After |
|------|--------|-------|
| Languages | `Language` headers must be equal | same base language is enough: `nl_NL` merges into `nl` and back |
| Plural rules | compared as text | compared by what they choose for n = 0..1000 |
| Into a catalogue made with `NewLanguage` | always `ErrorInvalidCatalogue` | merged when the defaults are in the same base language |

**Extracting and recording keys**

| What | Before | After |
|------|--------|-------|
| Reading the POT file | at `NewTranslator`, also with recording off; `NewTranslator("", "")` read `./translations.pot` | when a lookup first records a key or `CheckMissingTranslations` runs; with recording off, a translator that only translates never reads it (`NewTranslator` records by default) |
| `SetTranslationsDir`, `SetPotFile` after `NewTranslator` | new keys were checked against the POT file read at construction | new keys are checked against the POT file of the new location |
| `CheckMissingTranslations` with a key it cannot add to the POT file | the error printed with `fmt.Println`, nil returned | the errors returned, joined; a `log.Fatal(err)` on it now stops the program |
| Errors of runtime recording | printed with `fmt.Println` | logged with `log/slog` |
| `CheckMissingTranslations` order | new keys appended in random order | in the order they first appear in the templates, then keys recorded at runtime, sorted |
| `ScanFiles`, `CheckMissingTranslations` with an empty key in a template, `{{tl . ""}}`, and a lookup of an empty key with recording on, `Tl(loc, "")` | `msgid ""` appended when the POT file did not exist yet or had no header; with a context, `msgctxt` and `msgid ""` appended also when it had one | left out |
| `ScanFiles` with two different plurals of one key | the last one scanned | the first one in the templates |
| `ScanFiles` with a template that does not parse | error `scan translations in <file>: …`; the keys of the files scanned before it were kept | error `extract <file>: …`; no key of that scan is kept |
| `ScanFiles` with a call in a chain expression, `(…).Field` | not found | extracted |

## License

Distributed under the MIT License. See [LICENSE](LICENSE).
