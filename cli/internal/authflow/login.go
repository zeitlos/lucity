package authflow

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/zeitlos/lucity/cli/internal/api"
)

const loginTimeout = 5 * time.Minute

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

func Login(ctx context.Context, baseURL string) (*api.Session, error) {
	base, err := url.Parse(strings.TrimSuffix(baseURL, "/"))
	if err != nil || (base.Scheme != "http" && base.Scheme != "https") || base.Host == "" {
		return nil, fmt.Errorf("invalid base URL %q: expected an http(s) URL", baseURL)
	}

	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		return nil, fmt.Errorf("generate login state: %w", err)
	}
	state := hex.EncodeToString(stateBytes)

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("listen for login callback: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port

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

	loginQuery := url.Values{}
	loginQuery.Set("cli_port", strconv.Itoa(port))
	loginQuery.Set("cli_state", state)
	loginURL := base.String() + "/auth/login?" + loginQuery.Encode()

	fmt.Fprintln(os.Stderr, "Opening your browser to sign in…")
	fmt.Fprintln(os.Stderr, "If nothing opens, visit: "+loginURL)
	openBrowser(loginURL)

	var code string
	select {
	case result := <-results:
		if result.err != nil {
			return nil, result.err
		}
		code = result.code
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(loginTimeout):
		return nil, errors.New("login timed out after 5 minutes")
	}

	session, err := api.ExchangeCode(ctx, &http.Client{Timeout: 30 * time.Second}, baseURL, code)
	if err != nil {
		return nil, fmt.Errorf("exchange login code: %w", err)
	}
	return session, nil
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
