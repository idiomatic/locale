package locale

import (
	"fmt"
	"io/ioutil"
	"strings"

	"golang.org/x/text/language"
)

// Languages (offical and otherwise) likely to be used by clients of this package.
var SuggestedTags = []language.Tag{
	language.AmericanEnglish,
	language.BritishEnglish,
	language.MustParse("en-CA"), // HACK
	language.CanadianFrench,
	language.French,
	language.German,
	language.Spanish,
	language.Italian,
	language.Japanese,
	language.Dutch,
	language.Russian,
	language.MustParse("es-MX"), // HACK should be language.LatinAmericanSpanish, es-419
	language.MustParse("en-AU"), // IETF BCP 47
	language.Polish,
	language.BrazilianPortuguese,
	language.EuropeanPortuguese,
}

// Create a language matcher from a directory.
func NewMatcherFromDir(dir string) language.Matcher {
	files, _ := ioutil.ReadDir("localized")
	var locales []string
	for _, file := range files {
		locales = append(locales, file.Name())
	}
	return NewMatcher(locales)
}

// Create a language matcher from an array from a set of "likely" languages.
func NewMatcher(locales []string) language.Matcher {
	var tags []language.Tag

	for _, tag := range SuggestedTags {
		if contains(locales, BaseRegion(tag)) {
			tags = append(tags, tag)
		}
	}

	return language.NewMatcher(tags)
}

// Map certain country whom group with their motherland (or empire).
func RemapCountryCode(code string) string {
	code = strings.ToUpper(code)
	switch code {
	case "GB", "UK", "IE", "JE", "IM", "GG", "PN", "AG", "BM", "IO", "VG", "KY", "FK", "GI", "MS", "SH", "GS", "TC", "BQ", "BZ":
		return "GB"
	case "US", "GU", "PR", "AS", "MP", "VI", "UM", "PU", "MI":
		return "US"
	case "AU", "NZ", "CX", "CC", "HM", "NF":
		return "AU"
	default:
	}
	return code
}

// Encode simplified string representation of a locale.
func BaseRegion(tag language.Tag) string {
	tagBase, _ := tag.Base()
	tagRegion, _ := tag.Region()

	return fmt.Sprintf("%s-%s", tagBase, tagRegion)
}
