package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os/exec"
)

type downloader struct {
	handler http.Handler
}

func newDownloader(handler http.Handler) *downloader {
	return &downloader{handler}
}

const streamFileKey = "streamFile"

func (d *downloader) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log(r.Context(), slog.LevelInfo, "downloader")
	stream, err := url.Parse(r.URL.Query().Get("url"))

	if err != nil {
		http.Error(w, fmt.Errorf("parse url: %w", err).Error(), http.StatusBadRequest)
		return
	}
	log(r.Context(), slog.LevelInfo, "downloading stream", "url", stream)

	outfile := fmt.Sprintf("%s.mp3", r.Context().Value("traceId").(string))
	output, err := exec.CommandContext(r.Context(), "ffmpeg", "-i", stream.String(), outfile).CombinedOutput()
	if err != nil {
		http.Error(w, fmt.Errorf("download stream: %w", err).Error(), http.StatusInternalServerError)
		log(r.Context(), slog.LevelError, "download stream", "error", err, "output", string(output))
		return
	}
	d.handler.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), streamFileKey, outfile)))
}
