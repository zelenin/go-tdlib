package codegen

import (
	"strings"
	"unicode"
)

func firstUpper(str string) string {
	for i, r := range str {
		return string(unicode.ToUpper(r)) + str[i+1:]
	}

	return str
}

func firstLower(str string) string {
	for i, r := range str {
		return string(unicode.ToLower(r)) + str[i+1:]
	}

	return str
}

func underscoreToCamelCase(s string) string {
	return strings.Replace(strings.Title(strings.Replace(strings.ToLower(s), "_", " ", -1)), " ", "", -1)
}
