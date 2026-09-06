package httpx

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/mashrufahmed/go-kit/env"
)

type ShutdownHook func(context.Context) error

type App struct {
	server     *http.Server
	mux        *chi.Mux
	logger     *slog.Logger
	authCookie AuthCookieConfig

	config        Config
	middlewares   []Middleware
	shutdownHooks []ShutdownHook
}

func New(configs ...Config) *App {
	config := defaultConfig()

	if len(configs) > 0 {
		config = mergeConfig(configs[0])
	}

	logger := slog.New(
		slog.NewTextHandler(
			os.Stdout,
			&slog.HandlerOptions{
				Level: slog.LevelInfo,
			},
		),
	)

	mux := chi.NewRouter()

	return &App{
		mux:    mux,
		logger: logger,
		config: config,

		authCookie: AuthCookieConfig{
			Name:     "access_token",
			Path:     "/",
			MaxAge:   86400,
			Secure:   true,
			HTTPOnly: true,
			SameSite: http.SameSiteLaxMode,
		},

		server: &http.Server{
			Handler: mux,

			ReadTimeout:  config.ReadTimeout,
			WriteTimeout: config.WriteTimeout,
			IdleTimeout:  config.IdleTimeout,
		},
	}
}

func (a *App) LoadEnv(config any) {
	if err := env.Load(config); err != nil {
		a.logger.Error("environment loading failed", "error", err)
	}
}

func (a *App) SetLogger(logger *slog.Logger) {
	if logger != nil {
		a.logger = logger
	}
}

func (a *App) OnShutdown(hook ShutdownHook) {
	a.shutdownHooks = append(
		a.shutdownHooks,
		hook,
	)
}

func (a *App) Run(addr string) error {
	a.server.Addr = addr
	a.server.Handler = a.buildHandler()

	a.logger.Info(
		"server starting",
		"addr", addr,
	)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)

	defer stop()

	serverErr := make(chan error, 1)

	go func() {
		err := a.server.ListenAndServe()

		if err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}

		serverErr <- nil
	}()

	select {
	case err := <-serverErr:
		return err

	case <-ctx.Done():
		a.logger.Info(
			"shutdown signal received",
		)
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		a.config.ShutdownTimeout,
	)

	defer cancel()

	a.logger.Info(
		"server shutting down",
		"timeout", a.config.ShutdownTimeout,
	)

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		a.logger.Error(
			"server shutdown failed",
			"error", err,
		)

		return err
	}
	var shutdownErr error
	for i := len(a.shutdownHooks) - 1; i >= 0; i-- {
		if err := a.shutdownHooks[i](shutdownCtx); err != nil {
			a.logger.Error(
				"shutdown hook failed",
				"error", err,
			)

			shutdownErr = errors.Join(
				shutdownErr,
				err,
			)
		}
	}

	if shutdownErr != nil {
		return shutdownErr
	}

	a.logger.Info("server stopped")

	return nil
}
