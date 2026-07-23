package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

// TODO(stage-6b): delete this entire file. The sealed AES-GCM CLI handoff
// (sealCLICode/openCLICode/handleCLIExchange + validCLIPort/validCLIState/
// cliHandoff/cliCookieName) is replaced by the CLI's native OIDC PKCE login.
// Remove alongside the /auth/cli/exchange route and the cli_port/cli_state
// branch in handleLogin/handleCallback (oidc.go).

const cliCookieName = "lucity_cli"

var cliStatePattern = regexp.MustCompile(`^[A-Za-z0-9_-]{8,128}$`)

type cliHandoff struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refreshToken"`
	LogtoToken   string    `json:"logtoToken"`
	ExpiresAt    time.Time `json:"expiresAt"`
	IssuedAt     int64     `json:"iat"`
}

func validCLIPort(port string) bool {
	n, err := strconv.Atoi(port)
	if err != nil {
		return false
	}

	return n >= 1024 && n <= 65535
}

func validCLIState(state string) bool {
	return cliStatePattern.MatchString(state)
}

func cliHandoffKey(sessionSecret string) [32]byte {
	return sha256.Sum256([]byte("lucity-cli-handoff:" + sessionSecret))
}

func sealCLICode(sessionSecret string, handoff cliHandoff) (string, error) {
	handoff.IssuedAt = time.Now().Unix()

	plaintext, err := json.Marshal(handoff)
	if err != nil {
		return "", err
	}

	key := cliHandoffKey(sessionSecret)
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	sealed := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.RawURLEncoding.EncodeToString(sealed), nil
}

func openCLICode(sessionSecret, code string) (*cliHandoff, error) {
	sealed, err := base64.RawURLEncoding.DecodeString(code)
	if err != nil {
		return nil, errors.New("invalid code")
	}

	key := cliHandoffKey(sessionSecret)
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, errors.New("invalid code")
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.New("invalid code")
	}

	if len(sealed) < gcm.NonceSize() {
		return nil, errors.New("invalid code")
	}

	plaintext, err := gcm.Open(nil, sealed[:gcm.NonceSize()], sealed[gcm.NonceSize():], nil)
	if err != nil {
		return nil, errors.New("invalid code")
	}

	var handoff cliHandoff
	if err := json.Unmarshal(plaintext, &handoff); err != nil {
		return nil, errors.New("invalid code")
	}

	now := time.Now().Unix()
	if handoff.IssuedAt < now-120 || handoff.IssuedAt > now+30 {
		return nil, errors.New("code expired")
	}

	return &handoff, nil
}

func handleCLIExchange(sessionSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var body struct {
			Code string `json:"code"`
		}
		if err := json.NewDecoder(io.LimitReader(r.Body, 16<<10)).Decode(&body); err != nil || body.Code == "" {
			http.Error(w, "missing code", http.StatusBadRequest)
			return
		}

		handoff, err := openCLICode(sessionSecret, body.Code)
		if err != nil {
			http.Error(w, "invalid code", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"token":        handoff.Token,
			"refreshToken": handoff.RefreshToken,
			"logtoToken":   handoff.LogtoToken,
			"expiresAt":    handoff.ExpiresAt.Format(time.RFC3339),
		})
	}
}
