package server

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
)

type VoiceDetector struct {
	handler http.Handler
}

func newVoiceDetector(handler http.Handler) *VoiceDetector {
	return &VoiceDetector{handler}
}

type slice struct {
	Start uint
	End   uint
	File  string
}

const voiceSegmentsKey = "voiceSegments"

func (v *VoiceDetector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log(r.Context(), slog.LevelInfo, "voice detector")
	stream := r.Context().Value(streamFileKey).(string) //"a361f8b3-f2df-4e9a-881c-ba79a11c5f56.mp3"
	cmd := exec.Command("ffmpeg", "-nostats", "-i", stream, "-af", "silencedetect=n=-15dB:d=30", "-f", "null", "-")
	output, err := cmd.CombinedOutput()
	if err != nil {
		http.Error(w, fmt.Errorf("voice detector: %w", err).Error(), http.StatusInternalServerError)
		return
	}
	l := strings.Split(string(output), "\n")
	var silences []slice
	for _, str := range l {
		if strings.Contains(str, "silence_start:") {
			start, err := strconv.ParseFloat(strings.Split(str, "silence_start: ")[1], 32)
			if err != nil {
				http.Error(w, fmt.Errorf("voice detector: ParseFloat: %w", err).Error(), http.StatusInternalServerError)
				return
			}
			silences = append(silences, slice{
				Start: uint(start),
				End:   0,
			})
		} else if strings.Contains(str, "silence_end:") {
			end, err := strconv.ParseFloat(strings.Split(strings.Split(str, "silence_end: ")[1], " |")[0], 32)
			if err != nil || silences == nil || len(silences) == 0 {
				http.Error(w, fmt.Errorf("voice detector: ParseFloat: %v", err).Error(), http.StatusInternalServerError)
				return
			}
			silences[len(silences)-1].End = uint(end)
		}
	}
	voiceSegments := invertSlices(postProcessSilences(silences), r.Context().Value(durationKey).(float64))
	log(r.Context(), slog.LevelInfo, "silences", "voiceSegments", voiceSegments)

	v.handler.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), voiceSegmentsKey, voiceSegments)))
}

func postProcessSilences(oldSilences []slice) []slice {
	if len(oldSilences) < 2 {
		return oldSilences
	}
	if oldSilences[0].Start < 30 {
		oldSilences[0].Start = 0
	}
	newSilences := []slice{{Start: oldSilences[0].Start, End: oldSilences[0].Start}}
	oldPtr := 0
	for oldPtr < len(oldSilences) {
		if oldSilences[oldPtr].Start-newSilences[len(newSilences)-1].End < 30 { // Ignore sound that's shorter than 30 seconds
			newSilences[len(newSilences)-1].End = oldSilences[oldPtr].End
		} else {
			newSilences = append(newSilences, oldSilences[oldPtr])
		}
		oldPtr++
	}
	return newSilences
}

func invertSlices(slices []slice, duration float64) []slice {
	var newSilences []slice
	for i, s := range slices {
		if i == 0 && s.Start > 0 {
			newSilences = append(newSilences, slice{Start: 0, End: s.Start})
		} else if i > 0 {
			newSilences = append(newSilences, slice{Start: slices[i-1].End, End: s.Start})
		}
	}
	// check if last slice is longer than 30 seconds away from end of file
	if math.Abs(duration-float64(slices[len(slices)-1].End)) > 30 {
		newSilences = append(newSilences, slice{Start: slices[len(slices)-1].End, End: uint(duration)})
	}
	return newSilences
}
