package handler

import (
	"testing"
)

func TestValidateRepository(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantOwner string
		wantRepo  string
		wantErr   bool
	}{
		// Valid
		{"simple", "acme/myapp", "acme", "myapp", false},
		{"with dots", "my.org/my.app", "my.org", "my.app", false},
		{"with hyphens", "my-org/my-app", "my-org", "my-app", false},
		{"with underscores", "my_org/my_app", "my_org", "my_app", false},

		// Invalid
		{"empty", "", "", "", true},
		{"no slash", "myapp", "", "", true},
		{"bare url", "https://github.com/acme/myapp", "", "", true},
		{"triple segment", "github.com/acme/myapp", "", "", true},
		{"trailing slash", "acme/myapp/", "", "", true},
		{"empty owner", "/myapp", "", "", true},
		{"empty repo", "acme/", "", "", true},
		{"spaces", "acme/my app", "", "", true},
		{"special chars", "acme/my@app", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := validateRepository(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRepository(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if owner != tt.wantOwner {
					t.Errorf("validateRepository(%q) owner = %q, want %q", tt.input, owner, tt.wantOwner)
				}
				if repo != tt.wantRepo {
					t.Errorf("validateRepository(%q) repo = %q, want %q", tt.input, repo, tt.wantRepo)
				}
			}
		})
	}
}
