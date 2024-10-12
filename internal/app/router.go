package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func disableCacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0")
		next.ServeHTTP(w, r)
	})
}

// NewRouter creates and returns a new Chi router.
func NewRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(disableCacheMiddleware)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "public/emulator.html")
	})

	r.Get("/public/*", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, r.URL.Path[1:])
	})

	r.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "public/logo.svg")
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 - Not Found"))
	})

	return r
}
