// Package translator translates Go templates and code with gettext PO
// catalogues, with plural forms and translation contexts.
//
// A Translator loads one catalogue per language. Each lookup takes a
// Localizer that names the locale of the request, so one translator serves
// every language without a shared current locale. FuncMap gives templates
// tl, tn, ctl and ctn; Go code uses Tl, Tn, Ctl and Ctn. The extract package
// writes the POT file from the templates.
package translator
