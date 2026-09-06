package main

import (
	"time"

	"github.com/mashrufahmed/go-kit/httpx"
)

type User struct {
	Name  string `json:"name" validate:"required,min=3"`
	Email string `json:"email" validate:"required,email"`
}
type Config struct {
	Port        string        `env:"PORT" default:"8080"`
	JWTSecret   string        `env:"JWT_SECRET,required"`
	JWTExpiry   time.Duration `env:"JWT_EXPIRATION" default:"24h"`
	DatabaseURL string        `env:"DATABASE_URL,required"`
	Debug       bool          `env:"DEBUG" default:"false"`
}

func main() {
	app := httpx.New()

	app.Use(
		app.Logger(),
		httpx.Recovery(),
		httpx.RequestID(),
	)

	app.Run(":3000")
}
