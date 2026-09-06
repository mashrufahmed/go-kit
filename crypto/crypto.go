package crypto

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
)

func RandomBytes(n int) ([]byte, error) { b := make([]byte, n); _, err := rand.Read(b); return b, err }
func RandomToken(n int) (string, error) {
	b, err := RandomBytes(n)
	return base64.RawURLEncoding.EncodeToString(b), err
}
func SHA256(value []byte) []byte { sum := sha256.Sum256(value); return sum[:] }
func SHA512(value []byte) []byte { sum := sha512.Sum512(value); return sum[:] }
func HMACSHA256(key, value []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(value)
	return h.Sum(nil)
}
func Equal(a, b []byte) bool { return hmac.Equal(a, b) }
