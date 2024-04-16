package slogdiscard

import (
	"context"
	"log/slog"
)

// slog handler for mock records handling, do literaly nothing
func NewDiscardLogger() *slog.Logger {
	return slog.New(NewDiscardHandler())

}

type DiscardHandler struct {
}

func NewDiscardHandler() *DiscardHandler {
	return &DiscardHandler{}
}

// Fake logging implementations for custom slog.Handler (DiscardHandler)

func (h *DiscardHandler) Handle(_ context.Context, _ slog.Record) error {
	return nil
}

func (h *DiscardHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *DiscardHandler) WithGroup(_ string) slog.Handler {
	return h
}

func (h *DiscardHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return false
}
