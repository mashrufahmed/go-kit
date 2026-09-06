# go-kit

Reusable, modular Go building blocks for backend services.

`go-kit` is a thin toolkit—not a full-stack framework and not an ORM. It provides small, composable packages that work with Go’s standard library and mature, focused dependencies.

Module path: `github.com/mashrufahmed/go-kit`

## Features

- HTTP application toolkit with chi routing
- Express-style handlers that return errors
- Route groups and nested route groups
- JSON responses and request decoding
- Validation with JSON field names, including nested fields
- Consistent HTTP error responses
- Request IDs, structured logging and panic recovery
- Request timeout, CORS, security headers and rate limiting
- Configurable request body limits
- Bearer-token and cookie-token authentication helpers
- Secure cookie helpers and opt-in CSRF protection
- Generic JWT signing and verification with custom claims
- Environment loading from OS variables and `.env`
- Production-ready password hashing with configurable bcrypt cost
- UUID helpers
- Random bytes/tokens, hashing, HMAC and constant-time comparison
- Page/limit/offset pagination helpers
- Redis cache adapter and in-memory test cache
- Structured text/JSON logging with `log/slog`
- Small generic pointer, slice, map, string and time utilities

## Installation

```sh
go get github.com/mashrufahmed/go-kit
```

The project currently targets Go `1.26.5`.

## Quick start

```go
package main

import "github.com/mashrufahmed/go-kit/httpx"

func main() {
	app := httpx.New()
	app.Use(
		httpx.Recovery(),
		httpx.RequestID(),
		httpx.SecurityHeaders(),
		httpx.BodyLimit(1<<20),
		app.Logger(),
	)

	app.GET("/health", func(w httpx.Res, r *httpx.Req) error {
		return httpx.OK(w, map[string]string{"status": "ok"})
	})

	_ = app.Run(":3000")
}
```

## HTTPX

`httpx` provides a small HTTP layer on top of `net/http` and chi.

### Routes and groups

```go
app.GET("/users/{id}", getUser)
app.POST("/users", createUser)

api := app.Group("/api")
api.Use(httpx.RequestID())
api.GET("/profile", profile)

admin := api.Group("/admin")
admin.Use(httpx.RequireAuth[Claims]())
admin.DELETE("/users/{id}", deleteUser)
```

Handlers return an error and `httpx` maps it to a predictable JSON response:

```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "User not found",
    "timestamp": "2026-01-01T00:00:00Z"
  }
}
```

### Request decoding and validation

```go
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

func login(w httpx.Res, r *httpx.Req) error {
	input, err := httpx.DecodeAndValidate[LoginRequest](r)
	if err != nil {
		return err
	}
	return httpx.OK(w, input)
}
```

Validation errors expose JSON names such as `email` and `password`, not Go struct field names. `Decode` rejects empty bodies, malformed JSON, multiple JSON values and unsupported content types.

### Middleware

```go
app.Use(
	httpx.Recovery(),
	httpx.RequestID(),
	httpx.SecurityHeaders(),
	httpx.CORS("https://example.com"),
	httpx.Timeout(15*time.Second),
	httpx.BodyLimit(1<<20),
	httpx.RateLimit(100, time.Minute),
)
```

All middleware is independently composable. Applications only enable what they need.

### Authentication, cookies and CSRF

```go
app.Use(httpx.RequireAuth[MyClaims](httpx.FromBearer()))
claims, ok := httpx.Claims[MyClaims](r)
```

Cookie authentication is also supported with `httpx.FromCookie`. Cookie configuration includes `Secure`, `HttpOnly`, `SameSite`, `Domain`, `Path`, `MaxAge` and `Expires`.

Cookie-based authentication can opt into CSRF protection:

```go
token, err := httpx.NewCSRFToken()
if err != nil { return err }
httpx.SetCSRFCookie(w, token, httpx.CSRFConfig{})
app.Use(httpx.CSRF(httpx.CSRFConfig{}))
```

CSRF is intentionally not applied automatically to bearer-token authentication.

## JWT

JWT supports custom claim structures without hardcoded application claims:

```go
type Claims struct {
	Subject string `json:"sub"`
	Role    string `json:"role"`
}

client, err := jwt.New(jwt.Config{
	Secret:     os.Getenv("JWT_SECRET"),
	Expiration: 24 * time.Hour,
	Issuer:     "my-service",
})
if err != nil { return err }

token, err := client.Sign(Claims{Subject: "user-1", Role: "admin"})
var claims Claims
err = client.Verify(token, &claims)
```

Verification enforces HS256, signature, expiration, issuer and audience where configured. Instance-based clients are preferred; `jwt.Configure`, `jwt.Sign` and `jwt.Verify` remain available for simple applications.

## Environment configuration

```go
type Config struct {
	Port    string        `env:"PORT" default:"3000"`
	Debug   bool          `env:"DEBUG" default:"false"`
	Timeout time.Duration `env:"TIMEOUT" default:"15s"`
	JWT     struct {
		Secret string `env:"JWT_SECRET,required"`
	}
}

var cfg Config
if err := env.Load(&cfg); err != nil {
	return err
}
```

Supported values include strings, booleans, integers, floats, `time.Duration`, defaults, required variables and nested structs. OS environment variables take precedence over `.env` values. `env.MustLoad` is available for intentional panic-on-startup behavior.

## Passwords, UUIDs and crypto

```go
hash, err := password.Hash("plain text")
valid := password.Verify(hash, "plain text")

id := uuid.NewString()
token, err := crypto.RandomToken(32)
signature := crypto.HMACSHA256(secret, message)
same := crypto.Equal(signature, otherSignature)
```

Password hashing uses the mature bcrypt implementation from `x/crypto`; passwords are never logged or included in errors.

## Cache

The `cache.Cache` interface supports `Get`, `Set`, `Delete`, `Exists`, `TTL` and `Close`, with context-aware operations.

### Redis

```go
ctx := context.Background()
c, err := cache.NewRedis(ctx, cache.RedisConfig{
	Addr:     "localhost:6379",
	Password: os.Getenv("REDIS_PASSWORD"),
	Ping:     true,
})
if err != nil { return err }
defer c.Close()

_ = c.Set(ctx, "user:1", []byte(`{"name":"Rahim"}`), time.Hour)
```

Redis also provides `JSONSet` and `JSONGet` helpers. `cache.Memory` is useful for tests and local development. Missing keys return `cache.ErrNotFound`.

## Pagination

```go
params := pagination.Params{Page: 2, Limit: 25, MaxLimit: 100}
offset := params.Offset()
metadata := params.Metadata(totalRows)
```

Pagination is database-agnostic and only handles page, limit, offset and metadata calculations.

## Logger

The logger package uses Go’s `log/slog`:

```go
log := logger.New(logger.Config{
	Output: os.Stdout,
	JSON:   true,
	Level:  slog.LevelInfo,
})
app.SetLogger(log)
```

Text and JSON output, levels, structured fields, context fields and request ID fields are supported through `slog`.

## Utility packages

| Package | Purpose |
| --- | --- |
| `ptr` | Generic pointer creation and safe dereferencing |
| `slice` | Generic `Map`, `Filter` and `Contains` helpers |
| `maps` | Generic map key/value extraction |
| `stringx` | Blank checks and Unicode-safe truncation |
| `timeutil` | Small UTC and elapsed-time helpers |
| `validate` | Struct validation and JSON-aware error details |

## Design principles

- Small public APIs
- Explicit configuration
- Context-aware operations
- No global database or cache requirement
- No custom ORM
- No hidden application state where an instance can be used
- Standard-library-first implementation
- Composable middleware and adapters

## Testing

```sh
go test ./...
go test -race ./...
go vet ./...
```

Redis integration tests are skipped by default. Run them when Redis is available:

```sh
REDIS_ADDR=localhost:6379 go test ./cache
```

## Scope

`go-kit` intentionally does not provide a database package or ORM. Applications remain free to choose PostgreSQL, MySQL, SQLite or another database driver and keep their query/repository layer explicit.
