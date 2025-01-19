package internal

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var titleCaser = cases.Title(language.English)

func TitleCase(content string) string {
	return titleCaser.String(content)
}

func TitleCaseBytes(content []byte) []byte {
	return titleCaser.Bytes(content)
}
