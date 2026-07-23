package authflow

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/zeitlos/lucity/cli/internal/oidc"
)

const loginTimeout = 5 * time.Minute

// callbackPorts are the loopback ports the CLI binds for the OAuth redirect.
// They are fixed because the identity provider rejects wildcard and dynamic
// loopback redirect URIs; each must be registered on the native client.
var callbackPorts = []int{8765, 8766, 8767}

const successPage = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Signed in to Lucity</title>
<style>
body { display: flex; align-items: center; justify-content: center; min-height: 100vh; margin: 0; font-family: -apple-system, "Segoe UI", Helvetica, Arial, sans-serif; background: #fafafa; color: #18181b; }
main { text-align: center; padding: 2rem; }
h1 { font-size: 1.25rem; font-weight: 600; margin: 0 0 0.5rem; }
p { margin: 0; color: #71717a; }
@media (prefers-color-scheme: dark) {
body { background: #09090b; color: #fafafa; }
p { color: #a1a1aa; }
}
</style>
</head>
<body>
<main>
<h1>Signed in to Lucity</h1>
<p>You can close this tab and return to your terminal.</p>
</main>
</body>
</html>
`

type callbackResult struct {
	code string
	err  error
}

// Login runs a native Authorization-Code + PKCE flow against the identity
// provider using a loopback redirect, and returns the resulting refresh token.
func Login(ctx context.Context, provider *oidc.Provider) (string, error) {
	if provider.ClientID == "" {
		return "", errors.New("the platform did not advertise a CLI client id — the maintainer must register a native CLI client and set OIDC_CLI_CLIENT_ID")
	}

	listener, port, err := listenLoopback()
	if err != nil {
		return "", err
	}
	redirectURI := fmt.Sprintf("http://127.0.0.1:%d/callback", port)

	state, err := randomToken()
	if err != nil {
		return "", fmt.Errorf("generate login state: %w", err)
	}
	verifier, err := randomToken()
	if err != nil {
		return "", fmt.Errorf("generate PKCE verifier: %w", err)
	}
	challenge := codeChallenge(verifier)

	results := make(chan callbackResult, 1)

	var (
		mu        sync.Mutex
		delivered bool
	)
	claim := func() bool {
		mu.Lock()
		defer mu.Unlock()
		if delivered {
			return false
		}
		delivered = true
		return true
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /callback", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if query.Get("state") != state {
			http.Error(w, "state mismatch", http.StatusBadRequest)
			if claim() {
				results <- callbackResult{err: errors.New("login callback state mismatch")}
			}
			return
		}

		if errParam := query.Get("error"); errParam != "" {
			http.Error(w, "sign-in failed", http.StatusBadRequest)
			if claim() {
				results <- callbackResult{err: fmt.Errorf("sign-in failed: %s", errParam)}
			}
			return
		}

		code := query.Get("code")
		if code == "" {
			http.Error(w, "missing code", http.StatusBadRequest)
			return
		}

		if !claim() {
			http.Error(w, "login already completed", http.StatusGone)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(successPage))
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
		results <- callbackResult{code: code}
	})

	server := &http.Server{Handler: mux}
	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			if claim() {
				results <- callbackResult{err: fmt.Errorf("login callback server: %w", err)}
			}
		}
	}()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	loginURL := provider.AuthCodeURL(redirectURI, state, challenge)
	fmt.Fprintln(os.Stderr, "Opening your browser to sign in…")
	fmt.Fprintln(os.Stderr, "If nothing opens, visit: "+loginURL)
	openBrowser(loginURL)

	var code string
	select {
	case result := <-results:
		if result.err != nil {
			return "", result.err
		}
		code = result.code
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(loginTimeout):
		return "", errors.New("login timed out after 5 minutes")
	}

	tokens, err := provider.Exchange(ctx, code, redirectURI, verifier)
	if err != nil {
		return "", fmt.Errorf("exchange login code: %w", err)
	}
	if tokens.RefreshToken == "" {
		return "", errors.New("the identity provider did not return a refresh token")
	}
	return tokens.RefreshToken, nil
}

// listenLoopback binds the first available fixed callback port.
func listenLoopback() (net.Listener, int, error) {
	var lastErr error
	for _, port := range callbackPorts {
		listener, err := net.Listen("tcp4", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			return listener, port, nil
		}
		lastErr = err
	}
	return nil, 0, fmt.Errorf("no free login callback port (tried %v): %w", callbackPorts, lastErr)
}

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func codeChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func openBrowser(target string) {
	var command *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		command = exec.Command("open", target)
	case "windows":
		command = exec.Command("rundll32", "url.dll,FileProtocolHandler", target)
	default:
		command = exec.Command("xdg-open", target)
	}
	_ = command.Start()
}
