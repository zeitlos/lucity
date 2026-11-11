// Package shellwords splits a command line into words the way a POSIX shell
// would, without running one: words break at unquoted whitespace, single
// quotes keep everything literal, and double quotes and backslashes escape.
// No variables, globs or operators are expanded.
package shellwords

import (
	"errors"
	"strings"
)

var ErrUnclosedQuote = errors.New("unclosed quote")

func Split(line string) ([]string, error) {
	var (
		words   []string
		word    strings.Builder
		inWord  bool
		quote   rune
		escaped bool
	)

	for _, r := range line {
		switch {
		case escaped:
			if quote == '"' && r != '"' && r != '\\' && r != '$' && r != '`' {
				word.WriteRune('\\')
			}
			word.WriteRune(r)
			escaped = false
		case quote == '\'':
			if r == '\'' {
				quote = 0
			} else {
				word.WriteRune(r)
			}
		case r == '\\' && quote != '\'':
			escaped = true
			inWord = true
		case quote == '"':
			if r == '"' {
				quote = 0
			} else {
				word.WriteRune(r)
			}
		case r == '\'' || r == '"':
			quote = r
			inWord = true
		case r == ' ' || r == '\t' || r == '\n':
			if inWord {
				words = append(words, word.String())
				word.Reset()
				inWord = false
			}
		default:
			word.WriteRune(r)
			inWord = true
		}
	}

	if quote != 0 || escaped {
		return nil, ErrUnclosedQuote
	}

	if inWord {
		words = append(words, word.String())
	}

	return words, nil
}

func Join(words []string) string {
	quoted := make([]string, len(words))

	for i, word := range words {
		quoted[i] = quote(word)
	}

	return strings.Join(quoted, " ")
}

func quote(word string) string {
	if word == "" {
		return "''"
	}

	if !strings.ContainsAny(word, " \t\n'\"\\$`&|;<>()*?[]{}~#!") {
		return word
	}

	return "'" + strings.ReplaceAll(word, "'", `'\''`) + "'"
}
