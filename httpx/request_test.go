package httpx

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDecodeRejectsMultipleValuesAndContentType(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{} {}`))
	r.Header.Set("Content-Type", "application/json")
	if _, err := Decode[map[string]any](r); err == nil {
		t.Fatal("expected multiple JSON values to fail")
	}
	r = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
	r.Header.Set("Content-Type", "text/plain")
	if _, err := Decode[map[string]any](r); err != ErrUnsupportedContentType {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCSRF(t *testing.T) {
	token, err := NewCSRFToken()
	if err != nil {
		t.Fatal(err)
	}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })
	handler := CSRF(CSRFConfig{})(next)
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	r.AddCookie(&http.Cookie{Name: "csrf_token", Value: token})
	r.Header.Set("X-CSRF-Token", token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected success, got %d", w.Code)
	}
	r = httptest.NewRequest(http.MethodPost, "/", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden, got %d", w.Code)
	}
}

func TestRateLimit(t *testing.T) {
	handler := RateLimit(1, time.Minute)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "client:1"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusNoContent {
		t.Fatal(w.Code)
	}
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusTooManyRequests {
		t.Fatal(w.Code)
	}
}
