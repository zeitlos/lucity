package platform

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// IDMaxLength is the longest id a project, environment, service or resource can have.
const IDMaxLength = 16

var idPattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

// ValidID reports whether id follows the id rule: 2 to IDMaxLength lowercase
// letters, digits and hyphens, starting and ending with a letter or digit.
func ValidID(id string) bool {
	return len(id) >= 2 && len(id) <= IDMaxLength && idPattern.MatchString(id)
}

// DeriveID turns a readable name into an id. Names without two usable
// characters fall back to fallback. While taken reports the id as used, a
// counter is added and the rest is shortened to make room.
func DeriveID(name, fallback string, taken func(string) bool) string {
	base := slug(name)

	if len(base) < 2 {
		base = fallback
	}

	id := cut(base, IDMaxLength)

	for n := 2; taken(id); n++ {
		suffix := "-" + strconv.Itoa(n)
		id = cut(base, IDMaxLength-len(suffix)) + suffix
	}

	return id
}

func slug(name string) string {
	var b strings.Builder

	for _, r := range norm.NFKD.String(strings.ToLower(name)) {
		switch {
		case unicode.Is(unicode.Mn, r):
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}

	s := b.String()

	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}

	return strings.Trim(s, "-")
}

func cut(s string, n int) string {
	if len(s) > n {
		s = s[:n]
	}

	return strings.TrimRight(s, "-")
}
