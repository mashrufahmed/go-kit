package httpx

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type RouterGroup struct {
	router chi.Router
}

func (a *App) Group(prefix string) *RouterGroup {
	r := chi.NewRouter()

	a.mux.Mount(prefix, r)

	return &RouterGroup{
		router: r,
	}
}

func (a *App) GET(path string, handler Handler) {
	a.mux.Get(path, a.wrap(handler))
}

func (a *App) POST(path string, handler Handler) {
	a.mux.Post(path, a.wrap(handler))
}

func (a *App) PUT(path string, handler Handler) {
	a.mux.Put(path, a.wrap(handler))
}

func (a *App) PATCH(path string, handler Handler) {
	a.mux.Patch(path, a.wrap(handler))
}

func (a *App) DELETE(path string, handler Handler) {
	a.mux.Delete(path, a.wrap(handler))
}

func (a *App) wrap(handler Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			writeError(w, err)
		}
	}
}

// RouterGroup middleware

func (g *RouterGroup) Use(middlewares ...Middleware) {
	g.router.Use(middlewares...)
}

// Nested groups

func (g *RouterGroup) Group(prefix string) *RouterGroup {
	r := chi.NewRouter()

	g.router.Mount(prefix, r)

	return &RouterGroup{
		router: r,
	}
}

// RouterGroup routes

func (g *RouterGroup) GET(path string, handler Handler) {
	g.router.Get(path, handlerAdapter(handler))
}

func (g *RouterGroup) POST(path string, handler Handler) {
	g.router.Post(path, handlerAdapter(handler))
}

func (g *RouterGroup) PUT(path string, handler Handler) {
	g.router.Put(path, handlerAdapter(handler))
}

func (g *RouterGroup) PATCH(path string, handler Handler) {
	g.router.Patch(path, handlerAdapter(handler))
}

func (g *RouterGroup) DELETE(path string, handler Handler) {
	g.router.Delete(path, handlerAdapter(handler))
}

func handlerAdapter(handler Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			writeError(w, err)
		}
	}
}
