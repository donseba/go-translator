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
go get github.com/donseba/go-translator
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

	"github.com/donseba/go-translator"
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

Normalize locale names in your application (for example, map `en` to your loaded `en_US` catalogue). Missing languages or translations produce visible markers such as `*key*`; automatic fallback to another language is not provided.

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

Go code can use `Tl`, `Tn`, `Ctl`, and `Ctn`. `Tl` and `Ctl` accept formatting arguments; public `Tn` and `Ctn` accept the count only. Use the template helpers when plural text also needs formatting arguments.

## Extracting translation keys

`CheckMissingTranslations()` scans `.gohtml` files below the configured template directory and adds literal keys to `translations/translations.pot`. Run it during development or before serving requests.

Extraction parses Go template syntax without requiring application helper functions to be registered. It supports `localizer` helpers, dot and variable localizers, nested expressions, template definitions, and escaped literal keys. Comments and dynamic keys are skipped. Malformed templates return an error containing the source filename.

Runtime lookups also record previously unseen keys, including keys constructed in Go code. Those POT updates are serialized for concurrent lookups. Allow the catalogue directory to be writable when using runtime discovery; keep language loading, configuration, and catalogue editing outside request handling.

## Creating and editing languages

`EnsureLanguage("fr")` creates `fr.po` with language and plural-form headers if missing, then loads it. The translations directory must already exist. Repeated calls preserve an existing catalogue. `AddLanguage` is deprecated; use `SetLanguage`.

Use `SetTL`, `SetTLN`, `SetCTL`, `SetCTN`, and `Write` for catalogue editing before requests start. If updating `plurals.json`, regenerate the plural rules with:

```sh
go run tools/generate_templates.go
```

## Translation contexts and prefixes

Prefer explicit `ctl` / `ctn` contexts when the same text needs different translations. Legacy prefixed keys remain supported: `SetPrefixSeparator` changes the default `__.` separator, and the prefix is removed from translated output.

## License

Distributed under the MIT License. See [LICENSE](LICENSE).
