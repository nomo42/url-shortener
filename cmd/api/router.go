package api

import "github.com/go-chi/chi/v5"

func newMuxer() chi.Router {
	mux := chi.NewRouter()
	mux.Get("/{hash}", resolveShortcutHandler)
	mux.Post("/", createShortcutHandler)
	mux.Post("/api/shorten", createShortcutJSONHandler)
	return mux
}
