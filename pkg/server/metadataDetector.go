package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os/exec"
	"strconv"
)

const durationKey = "duration"

type metadataDetector struct {
	handler http.Handler
}

func newMetadataDetector(handler http.Handler) *metadataDetector {
	return &metadataDetector{handler}
}

func (m *metadataDetector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log(r.Context(), slog.LevelInfo, "metadata detector")

	stream := r.Context().Value(streamFileKey).(string) //"a361f8b3-f2df-4e9a-881c-ba79a11c5f56.mp3"
	cmd := exec.CommandContext(r.Context(), "ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", stream)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log(r.Context(), slog.LevelError, "metadata detector", "error", err)
		http.Error(w, fmt.Errorf("metadata detector: %w", err).Error(), http.StatusInternalServerError)
		return
	}

	var ffProbeOutput ffprobe
	err = json.Unmarshal(output, &ffProbeOutput)
	if err != nil {
		log(r.Context(), slog.LevelError, "metadata detector", "error", err)
		http.Error(w, fmt.Errorf("metadata detector: %w", err).Error(), http.StatusInternalServerError)
		return
	}
	duration, err := strconv.ParseFloat(ffProbeOutput.Format.Duration, 64)
	if err != nil {
		log(r.Context(), slog.LevelError, "metadata detector: parse duration", "error", err)
		http.Error(w, fmt.Errorf("metadata detector: %w", err).Error(), http.StatusInternalServerError)
		return
	}

	log(r.Context(), slog.LevelInfo, "metadata detector", durationKey, duration)

	m.handler.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "duration", duration)))
}

type ffprobe struct {
	Format struct {
		Filename       string `json:"filename"`
		NbStreams      int    `json:"nb_streams"`
		NbPrograms     int    `json:"nb_programs"`
		NbStreamGroups int    `json:"nb_stream_groups"`
		FormatName     string `json:"format_name"`
		FormatLongName string `json:"format_long_name"`
		StartTime      string `json:"start_time"`
		Duration       string `json:"duration"`
		Size           string `json:"size"`
		BitRate        string `json:"bit_rate"`
		ProbeScore     int    `json:"probe_score"`
		Tags           struct {
			MajorBrand       string `json:"major_brand"`
			MinorVersion     string `json:"minor_version"`
			CompatibleBrands string `json:"compatible_brands"`
			Encoder          string `json:"encoder"`
		} `json:"tags"`
	} `json:"format"`
}
