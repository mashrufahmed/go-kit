package httpx

import (
	"context"
	"net/http"
	"sync"
	"time"

	kitcrypto "github.com/mashrufahmed/go-kit/crypto"

	"github.com/google/uuid"
)

type Middleware = func(http.Handler) http.Handler
type requestIDKey struct{}

func (a *App) Use(
	middlewares ...Middleware,
) {
	a.middlewares = append(
		a.middlewares,
		middlewares...,
	)
}

func (a *App) buildHandler() http.Handler {
	var handler http.Handler = a.mux

	for i := len(a.middlewares) - 1; i >= 0; i-- {
		handler = a.middlewares[i](handler)
	}
	if a.config.MaxBodySize > 0 {
		handler = BodyLimit(a.config.MaxBodySize)(handler)
	}

	return handler
}

func (a *App) Logger() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := newResponseWriter(w)

			next.ServeHTTP(rw, r)

			a.logger.Info(
				"request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.status,
				"bytes", rw.written,
				"request_id", GetRequestID(r),
				"duration", time.Since(start),
			)
		})
	}
}

func Recovery() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					writeError(
						w,
						InternalServerError("Internal server error"),
					)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

func GetRequestID(r *http.Request) string {
	value := r.Context().Value(requestIDKey{})

	if value == nil {
		return ""
	}

	result, _ := value.(string)
	return result
}

func Timeout(timeout time.Duration) Middleware {
	return func(next http.Handler) http.Handler {
		return http.TimeoutHandler(next, timeout, `{"error":{"code":"TIMEOUT","message":"Request timed out"}}`)
	}
}

func BodyLimit(maxBytes int64) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if maxBytes > 0 && r.Body != nil {
				r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func SecurityHeaders() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("X-Frame-Options", "DENY")
			h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
			next.ServeHTTP(w, r)
		})
	}
}

func CORS(origins ...string) Middleware {
	allowed := map[string]bool{}
	for _, origin := range origins {
		allowed[origin] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if allowed[origin] {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Add("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

type rateEntry struct {
	start time.Time
	count int
}

func RateLimit(limit int, window time.Duration) Middleware {
	if limit < 1 {
		limit = 60
	}
	if window <= 0 {
		window = time.Minute
	}
	var mu sync.Mutex
	entries := map[string]rateEntry{}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.RemoteAddr
			now := time.Now()
			mu.Lock()
			for client, old := range entries {
				if !old.start.IsZero() && now.Sub(old.start) >= window*2 {
					delete(entries, client)
				}
			}
			e := entries[key]
			if e.start.IsZero() || now.Sub(e.start) >= window {
				e = rateEntry{start: now}
			}
			e.count++
			entries[key] = e
			allowed := e.count <= limit
			mu.Unlock()
			if !allowed {
				w.Header().Set("Retry-After", "1")
				writeError(w, NewError(http.StatusTooManyRequests, "RATE_LIMITED", "Too many requests"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

type CSRFConfig struct {
	CookieName string
	HeaderName string
	Secure     bool
	SameSite   http.SameSite
}

func CSRF(cfg CSRFConfig) Middleware {
	if cfg.CookieName == "" {
		cfg.CookieName = "csrf_token"
	}
	if cfg.HeaderName == "" {
		cfg.HeaderName = "X-CSRF-Token"
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}
			cookie, err := r.Cookie(cfg.CookieName)
			token := r.Header.Get(cfg.HeaderName)
			if err != nil || cookie.Value == "" || token == "" || !kitcrypto.Equal([]byte(cookie.Value), []byte(token)) {
				writeError(w, Forbidden("CSRF validation failed"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func NewCSRFToken() (string, error) { return kitcrypto.RandomToken(32) }

func SetCSRFCookie(w http.ResponseWriter, token string, cfg CSRFConfig) {
	if cfg.CookieName == "" {
		cfg.CookieName = "csrf_token"
	}
	if cfg.SameSite == 0 {
		cfg.SameSite = http.SameSiteLaxMode
	}
	http.SetCookie(w, &http.Cookie{Name: cfg.CookieName, Value: token, Path: "/", Secure: cfg.Secure, HttpOnly: false, SameSite: cfg.SameSite})
}
func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")

			if requestID == "" {
				requestID = uuid.NewString()
			}

			w.Header().Set("X-Request-ID", requestID)

			ctx := context.WithValue(
				r.Context(),
				requestIDKey{},
				requestID,
			)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func LimitBody(
	r *http.Request,
	maxBytes int64,
) {
	r.Body = http.MaxBytesReader(
		nil,
		r.Body,
		maxBytes,
	)
}
