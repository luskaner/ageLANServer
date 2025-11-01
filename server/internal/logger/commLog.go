package logger

import (
	"log/slog"
	"time"
)

func LogMessage(name string, args ...any) {
	if SlogEnabled {
		serverUptime := time.Since(StartTime)
		allArgs := []any{slog.Int64("uptime", serverUptime.Milliseconds())}
		allArgs = append(allArgs, args...)
		slog.Info(name, allArgs...)
	}
}
