package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os/exec"
	"strconv"
)

type slicer struct {
	handler http.Handler
}

func newSlicer(handler http.Handler) *slicer {
	return &slicer{handler}
}

func (s *slicer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log(r.Context(), slog.LevelInfo, "slicer")
	file := r.Context().Value(streamFileKey).(string) // "a361f8b3-f2df-4e9a-881c-ba79a11c5f56.mp3"
	segments := r.Context().Value(voiceSegmentsKey).([]slice)
	traceId := r.Context().Value("traceId").(string)

	for i, segment := range segments {
		segmentName := fmt.Sprintf("%s-%d.mp3", traceId, i)
		cmd := exec.Command("ffmpeg",
			"-nostats", "-i", file, "-ss", strconv.Itoa(int(segment.Start)), "-t", strconv.Itoa(int(segment.End)), "-c", "copy", segmentName)
		log(r.Context(), slog.LevelInfo, "creating slice", "index", i, "start", segment.Start, "end", segment.End, "cmd", cmd)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log(r.Context(), slog.LevelError, "create slice", "error", err, "output", string(output))
			http.Error(w, fmt.Errorf("create slice: %w", err).Error(), http.StatusInternalServerError)
			return
		}
		log(r.Context(), slog.LevelInfo, "created slice", "index", i, "start", segment.Start, "end", segment.End)
		segments[i].File = segmentName
	}
	s.handler.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), voiceSegmentsKey, segments)))
}
