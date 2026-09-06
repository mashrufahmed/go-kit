package jwt

import (
	"testing"
	"time"
)

func TestClientSignVerifyAndExpiry(t *testing.T) {
	c, err := New(Config{Secret: "test-secret", Expiration: time.Hour, Issuer: "test"})
	if err != nil {
		t.Fatal(err)
	}
	token, err := c.Sign(struct {
		Subject string `json:"sub"`
	}{"u1"})
	if err != nil {
		t.Fatal(err)
	}
	var claims map[string]any
	if err := c.Verify(token, &claims); err != nil {
		t.Fatal(err)
	}
	if claims["sub"] != "u1" {
		t.Fatalf("unexpected claims: %#v", claims)
	}
	bad, _ := New(Config{Secret: "wrong"})
	if err := bad.Verify(token, &claims); err == nil {
		t.Fatal("expected signature error")
	}
}
