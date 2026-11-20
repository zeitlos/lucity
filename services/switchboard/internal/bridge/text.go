package bridge

import (
	"strings"
	"unicode/utf8"
)

// chunks splits text at line breaks into parts of at most size bytes.
func chunks(text string, size int) []string {
	var parts []string

	for len(text) > size {
		cut := strings.LastIndex(text[:size], "\n")

		if cut <= 0 {
			cut = size

			for cut > 0 && !utf8.RuneStart(text[cut]) {
				cut--
			}
		}

		parts = append(parts, text[:cut])
		text = strings.TrimLeft(text[cut:], "\n")
	}

	if strings.TrimSpace(text) != "" {
		parts = append(parts, text)
	}

	return parts
}

func truncate(text string, size int) string {
	if len(text) <= size {
		return text
	}

	cut := size

	for cut > 0 && !utf8.RuneStart(text[cut]) {
		cut--
	}

	return text[:cut] + "…"
}
