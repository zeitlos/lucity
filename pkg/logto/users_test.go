package logto

import "testing"

func TestSanitizeUsername(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"marcelhintermann", "marcelhintermann"},
		{"toni-bentini", "toni_bentini"},
		{"user.name", "user_name"},
		{"123numeric", "_123numeric"},
		{"---hyphens---", "hyphens"},
		{"normal_user", "normal_user"},
		{"MixedCase", "MixedCase"},
		{"a", "a"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := SanitizeUsername(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeUsername(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
