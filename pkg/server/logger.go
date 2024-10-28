package server

import (
	"context"
	"log/slog"
)

func log(ctx context.Context, level slog.Level, msg string, args ...any) {
	if traceID := ctx.Value("traceId"); traceID != nil {
		args = append([]any{"traceId", traceID}, args...)
	}
	slog.Log(ctx, level, msg, args...)
}
