package app

import (
	"encoding/json"
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

	// API endpoint to assemble code. Expects a JSON object with a single key "assembly" containing the assembly code.
	// Returns a JSON object with the assembled binary as a uint16 array and the length of the output program.
	r.Post("/api/assemble", func(w http.ResponseWriter, r *http.Request) {
		// get the body of the request (the assembly code)
		var assembly string
		if err := json.NewDecoder(r.Body).Decode(&assembly); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// asswemble that shit :3 rawr xD uwu owo nyaaa 【=◈︿◈=】
		var binary []uint16
		length, err := Assemble(assembly, binary)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// return the output program length and the assembled binary as a uint16 array
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"length": length,
			"binary": binary,
		})
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
