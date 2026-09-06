package httpx

import (
	"context"
	"net/http"
	"strings"

	"github.com/mashrufahmed/go-kit/jwt"
)

type TokenSource func(*http.Request) string

func FromBearer() TokenSource {
	return func(r *http.Request) string {
		header := r.Header.Get("Authorization")

		if header == "" {
			return ""
		}

		parts := strings.SplitN(
			header,
			" ",
			2,
		)

		if len(parts) != 2 {
			return ""
		}

		if !strings.EqualFold(
			parts[0],
			"Bearer",
		) {
			return ""
		}

		return strings.TrimSpace(parts[1])
	}
}

func FromCookie(name string) TokenSource {
	return func(r *http.Request) string {
		cookie, err := r.Cookie(name)

		if err != nil {
			return ""
		}

		return strings.TrimSpace(
			cookie.Value,
		)
	}
}

type claimsKey struct{}

func RequireAuth[T any](
	sources ...TokenSource,
) Middleware {
	if len(sources) == 0 {
		sources = []TokenSource{
			FromBearer(),
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {

				var token string

				for _, source := range sources {
					token = source(r)

					if token != "" {
						break
					}
				}

				if token == "" {
					writeError(
						w,
						Unauthorized(
							"Authentication required",
						),
					)

					return
				}

				claims, err := jwt.Verify[T](token)

				if err != nil {
					writeError(
						w,
						Unauthorized(
							"Invalid or expired token",
						),
					)

					return
				}

				ctx := context.WithValue(
					r.Context(),
					claimsKey{},
					claims,
				)

				next.ServeHTTP(
					w,
					r.WithContext(ctx),
				)
			},
		)
	}
}

func Claims[T any](
	r *http.Request,
) (T, bool) {
	value := r.Context().Value(
		claimsKey{},
	)

	if value == nil {
		var zero T

		return zero, false
	}

	claims, ok := value.(T)

	return claims, ok
}
