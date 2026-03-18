package handler

import (
	"testing"
)

func TestValidateSourceURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		// Valid
		{"github https", "https://github.com/acme/myapp", false},
		{"github https with .git", "https://github.com/acme/myapp.git", false},
		{"gitlab https", "https://gitlab.com/acme/myapp", false},
		{"bitbucket https", "https://bitbucket.org/acme/myapp", false},

		// Scheme violations
		{"http scheme", "http://github.com/acme/myapp", true},
		{"ssh scheme", "ssh://git@github.com/acme/myapp.git", true},
		{"file scheme", "file:///etc/passwd", true},
		{"no scheme", "github.com/acme/myapp", true},

		// Internal cluster services (SSRF)
		{"k8s svc.cluster.local", "https://soft-serve.lucity-system.svc.cluster.local:23232/repo.git", true},
		{"k8s svc short", "https://soft-serve.lucity-system.svc:23232/repo.git", true},
		{"localhost", "https://localhost:5000/v2/_catalog", true},
		{"dot local", "https://registry.local/repo", true},
		{"dot internal", "https://metadata.internal/latest", true},
		{"dot localhost", "https://whatever.localhost/foo", true},

		// Cloud metadata (link-local IP)
		{"aws metadata", "https://169.254.169.254/latest/meta-data/", true},

		// Private IPs
		{"loopback", "https://127.0.0.1/repo", true},
		{"private 10.x", "https://10.96.100.100:5000/v2/_catalog", true},
		{"private 192.168", "https://192.168.1.1/repo", true},
		{"private 172.16", "https://172.16.0.1/repo", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSourceURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSourceURL(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
			}
		})
	}
}
