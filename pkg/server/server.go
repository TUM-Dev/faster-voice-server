package server

import (
	"context"
	"github.com/google/uuid"
	"log/slog"

	"net/http"
	"time"
)

type Server struct {
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Run() {
	mux := http.NewServeMux()
	stack := newStack(mux)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log(r.Context(), slog.LevelInfo, "request executed")
	})
	err := http.ListenAndServe(":8772", stack)

	if err != nil {
		panic(err)
	}
}

func newStack(mux *http.ServeMux) http.Handler {
	return newVerifier(
		newTracer(
			newLogger(
				newDownloader(
					newMetadataDetector(
						newVoiceDetector(
							newSlicer(
								newTranscriber(
									mux,
								),
							),
						),
					),
				),
			),
		),
	)
}

// logger is a middleware that logs all requests.
type logger struct {
	handler http.Handler
}

func newLogger(handler http.Handler) *logger {
	return &logger{handler}
}

func (l *logger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	l.handler.ServeHTTP(w, r)
	log(r.Context(), slog.LevelInfo, "http request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(start).String())
}

type tracer struct {
	handler http.Handler
}

func newTracer(handler http.Handler) *tracer {
	return &tracer{handler}
}

func (t *tracer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log(r.Context(), slog.LevelInfo, "tracer")
	newCtx := r.WithContext(context.WithValue(r.Context(), "traceId", uuid.New().String()))
	t.handler.ServeHTTP(w, newCtx)
}

type verifier struct {
	handler http.Handler
}

func newVerifier(handler http.Handler) *verifier {
	return &verifier{handler}
}

func (v *verifier) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log(r.Context(), slog.LevelInfo, "verifier")
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		//r.Context().Value("cancelFunc").(context.CancelFunc)()
		return
	}
	v.handler.ServeHTTP(w, r)
}
