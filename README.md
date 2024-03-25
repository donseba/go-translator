# go-translator

Overview
--
The Translator package is a Go module designed to facilitate easy and dynamic localization of applications. It provides a robust and flexible way to manage translations, including support for plural forms and customizable prefix handling for translation keys. This package utilizes the gotext library for managing PO files and integrates seamlessly with Go templates, making it an ideal solution for applications requiring multi-language support.

Features
--
- **Dynamic Language Support**: Add new languages by parsing PO files.
- **Template Integration**: Works with Go HTML templates, extracting translation keys directly from them.
- **Pluralization Support**: Handles singular and plural forms for languages with complex plural rules.
- **Contextual Translations**: Supports context-based translations for more accurate localization.
- **Prefix Handling**: Customizable prefix handling in translation keys, allowing for organized and readable translation files.
- **Missing Translation Detection**: Scans for and logs missing translations, simplifying the translation management process.

Installation
--
To install the Translator package, use the following go get command:

```bash
go get github.com/donseba/go-translator
```

Usage
--
Initializing the Translator

```go
package main 

import "github.com/donseba/go-translator"

func main() {
    translationsDir := "path/to/translations"
    templateDir := "path/to/templates"

    tr := translator.NewTranslator(translationsDir, templateDir)
    tr.SetPrefixSeparator("__.") // Set the prefix separator if different from the default

    // load languages
    tr.AddLanguage("en_US")
    tr.AddLanguage("nl_NL")

    // check for missing translations and add them to the pot file
    err = app.Translation.CheckMissingTranslations("translations.pot")
    if err != nil {
        log.Fatal(err)
    }   
	
    // add template functions
    // this will add the tl, tn, ctl and ctn functions to the template
    var yourTemplateFunctions = make(template.FuncMap) 
    for k, v := range tr.FuncMap() {
        yourTemplateFunctions[k] = v
    }
```

Using the Translator in Templates
--
Use tl and ctl for translating singular texts and tn and ctn for plural forms.

```html
<!-- Singular -->
<p>{{ tl .Loc "Hello, World!" }}</p>

<!-- Plural -->
<p>{{ tn .Loc "You have one message." "You have %d messages." 5, 5 }}</p>
```

Localizer interface
--

The Localizer interface is used to provide the translator with the current language. 
It is used to determine the correct translation for the given key.

```go
	//Localizer interface contains the methods that are needed for the translator
Localizer interface {
   // GetLocale returns the locale of the localizer, ie. "en_US"
   GetLocale() string
}
``` 


Adding a New Language
--
```go
tr.AddLanguage("en_US")
tr.AddLanguage("nl_NL")
```

Scanning for Missing Translations
--
To check for missing translations in your templates:

```go
err := tr.CheckMissingTranslations("messages.pot")
if err != nil {
log.Fatal(err)
}
```
Customizing Prefix Separator
--
You can customize the prefix separator used in translation keys:

```go
tr.SetPrefixSeparator("__CUSTOM__")
```
The default prefix separator is `__.`

Removing Prefixes from Translations
--
The package automatically handles the removal of prefixes from translations at runtime:

```go
translatedText := tr.Tl(localizer, "prefix__.your_translation_key") // output your_translation_key
```

TODO
--
- add caching functionality of the loaded translated keys.
- add more unit tests
- live reload translations when the file changes
- fallback language


Contributing
--
Contributions to the Translator package are welcome! Please submit a pull request or open an issue for any bugs, features, or improvements.

License
--
This package is licensed under MIT. Please see the LICENSE file for more details.
