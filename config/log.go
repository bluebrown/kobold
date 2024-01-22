package config

import (
	"io"
	"log/slog"
)

func SetLog(w io.Writer, fmt string, lvl slog.Leveler) {
	ho := &slog.HandlerOptions{Level: lvl}

	var handler slog.Handler
	switch fmt {
	case "json":
		handler = slog.NewJSONHandler(w, ho)
	case "text":
		handler = slog.NewTextHandler(w, ho)
	default:
		handler = slog.NewTextHandler(w, ho)
		defer slog.Warn("unknown log format", "fmt", fmt, "using", "text")
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
}
