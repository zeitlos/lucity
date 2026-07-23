package session

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
)

var ErrInvalid = errors.New("invalid session")

type Codec struct {
	key    [32]byte
	name   string
	secure bool
	maxAge int
}

func NewCodec(secret, cookieName string, secure bool, maxAge int) *Codec {
	return &Codec{
		key:    sha256.Sum256([]byte(secret)),
		name:   cookieName,
		secure: secure,
		maxAge: maxAge,
	}
}

func (c *Codec) Seal(v any) (string, error) {
	plaintext, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	gcm, err := c.gcm()
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(gcm.Seal(nonce, nonce, plaintext, nil)), nil
}

func (c *Codec) Open(value string, v any) error {
	sealed, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return ErrInvalid
	}
	gcm, err := c.gcm()
	if err != nil {
		return ErrInvalid
	}
	if len(sealed) < gcm.NonceSize() {
		return ErrInvalid
	}
	plaintext, err := gcm.Open(nil, sealed[:gcm.NonceSize()], sealed[gcm.NonceSize():], nil)
	if err != nil {
		return ErrInvalid
	}
	if err := json.Unmarshal(plaintext, v); err != nil {
		return ErrInvalid
	}
	return nil
}

func (c *Codec) gcm() (cipher.AEAD, error) {
	block, err := aes.NewCipher(c.key[:])
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func (c *Codec) Read(r *http.Request) (string, bool) {
	cookie, err := r.Cookie(c.name)
	if err != nil {
		return "", false
	}
	return cookie.Value, true
}

func (c *Codec) SetCookie(w http.ResponseWriter, sealed string) {
	http.SetCookie(w, &http.Cookie{
		Name:     c.name,
		Value:    sealed,
		Path:     "/",
		MaxAge:   c.maxAge,
		HttpOnly: true,
		Secure:   c.secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (c *Codec) ClearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: c.name, Path: "/", MaxAge: -1})
}
