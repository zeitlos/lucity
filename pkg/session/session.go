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

var (
	ErrNoCookie = errors.New("no session cookie")
	ErrInvalid  = errors.New("invalid session cookie")
)

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

func (c *Codec) Save(w http.ResponseWriter, v any) error {
	sealed, err := c.seal(v)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     c.name,
		Value:    sealed,
		Path:     "/",
		MaxAge:   c.maxAge,
		HttpOnly: true,
		Secure:   c.secure,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func (c *Codec) Load(r *http.Request, v any) error {
	cookie, err := r.Cookie(c.name)
	if err != nil {
		return ErrNoCookie
	}
	return c.open(cookie.Value, v)
}

func (c *Codec) Clear(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: c.name, Path: "/", MaxAge: -1})
}

func (c *Codec) seal(v any) (string, error) {
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

func (c *Codec) open(value string, v any) error {
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
