package server

import (
	"log/slog"
	"net/http"
	"os"
)

type cleaner struct {
	handler http.Handler
}

func newCleaner(handler http.Handler) *cleaner {
	return &cleaner{handler}
}

func (c *cleaner) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log(r.Context(), slog.LevelInfo, "cleaner")
	if stream := r.Context().Value(streamFileKey); stream != nil {
		log(r.Context(), slog.LevelInfo, "removing stream", "file", stream)
		err := os.Remove(stream.(string))
		if err != nil {
			log(r.Context(), slog.LevelError, "remove stream", "error", err)
		}
	}
	c.handler.ServeHTTP(w, r)
}
