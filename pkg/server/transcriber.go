package server

import (
	"fmt"
	"github.com/joschahenningsen/openai-go"
	"github.com/joschahenningsen/openai-go/option"
	"io"
	"log/slog"
	"net/http"
	"os"
)

type transcriber struct {
	handler http.Handler
}

func newTranscriber(handler http.Handler) *transcriber {
	return &transcriber{handler}
}

func (t *transcriber) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log(r.Context(), slog.LevelInfo, "transcriber")

	var openAIResponse *string
	client := openai.NewClient(
		option.WithBaseURL("http://tailscale-stoffi:8000/v1/"),
		option.WithAPIKey("cant-be-empty"),
		option.WithResponseBodyInto(&openAIResponse),
	)
	segments := r.Context().Value(voiceSegmentsKey).([]slice)

	for i, segment := range segments {
		file, err := os.Open(segment.File)
		if err != nil {
			log(r.Context(), slog.LevelError, "open file", "error", err)
			http.Error(w, fmt.Errorf("open file: %w", err).Error(), http.StatusInternalServerError)
			return
		}
		_, err = client.Audio.Transcriptions.New(r.Context(), openai.AudioTranscriptionNewParams{
			File:                    openai.F(io.Reader(file)),
			Model:                   openai.F(openai.AudioModelWhisper1),
			ResponseFormat:          openai.F(openai.AudioResponseFormatSRT),
			Temperature:             openai.F(0.0),
			ConditionOnPreviousText: openai.F(false),
			VadFilter:               openai.F(true),
		})

		if err != nil {
			log(r.Context(), slog.LevelError, "transcribe", "error", err)
			http.Error(w, fmt.Errorf("transcribe: %w", err).Error(), http.StatusInternalServerError)
			return
		}
		log(r.Context(), slog.LevelInfo, "transcribed", "index", i, "text", *openAIResponse)
	}

	t.handler.ServeHTTP(w, r)
}
