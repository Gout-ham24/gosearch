package indexer

import (
	"strings"
	"unicode"
)

// stopWords are common words we exclude from the index since they add no search value
var stopWords = map[string]bool{
	"the": true, "is": true, "at": true, "a": true, "an": true,
	"and": true, "or": true, "of": true, "to": true, "in": true,
	"it": true, "for": true, "on": true, "with": true, "as": true,
	"this": true, "that": true, "be": true, "are": true, "was": true,
}

// Tokenize breaks raw text into a clean list of lowercase words,
// stripping punctuation and removing common stopwords.
func Tokenize(text string) []string {
	var tokens []string
	var current strings.Builder

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current.WriteRune(unicode.ToLower(r))
		} else {
			if current.Len() > 0 {
				word := current.String()
				if !stopWords[word] {
					tokens = append(tokens, word)
				}
				current.Reset()
			}
		}
	}

	// Catch the last word if text doesn't end in punctuation/space
	if current.Len() > 0 {
		word := current.String()
		if !stopWords[word] {
			tokens = append(tokens, word)
		}
	}

	return tokens
}
