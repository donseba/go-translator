package main

import (
	"log"

	"github.com/donseba/go-translator/v2"
)

func main() {
	err := translator.GenerateLanguageHeaderTemplatesFromJSON("./plurals.json", "./generated_plural_templates.go")
	if err != nil {
		log.Fatalf("Error generating templates: %v", err)
	}
}
