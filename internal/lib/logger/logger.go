package logger

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/mickamy/employee-management/internal/config"
)

const callerSkipDepth = 2

var moduleRoot string

func Init(cfg config.App) {
	slog.SetDefault(slog.New(jsonHandler(cfg)))
	moduleRoot = cfg.ModuleRoot
}

func jsonHandler(cfg config.App) *slog.JSONHandler {
	var opts = &slog.HandlerOptions{
		Level: logLevel(cfg),
	}
	return slog.NewJSONHandler(os.Stdout, opts)
}

func logLevel(cfg config.App) slog.Level {
	switch cfg.LogLevel {
	case config.LogLevelDebug:
		return slog.LevelDebug
	case config.LogLevelInfo:
		return slog.LevelInfo
	case config.LogLevelWarn:
		return slog.LevelWarn
	case config.LogLevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func internalHandle(ctx context.Context, level slog.Level, msg string, args ...any) {
	_, f, l, _ := runtime.Caller(callerSkipDepth)
	source := f + ":" + strconv.Itoa(l)

	source = strings.TrimPrefix(source, moduleRoot+"/")
	args = append(args, slog.String("source", source))

	slog.Default().Log(ctx, level, msg, args...)
}

func Debug(ctx context.Context, msg string, args ...any) {
	internalHandle(ctx, slog.LevelDebug, msg, args...)
}

func Info(ctx context.Context, msg string, args ...any) {
	internalHandle(ctx, slog.LevelInfo, msg, args...)
}

func Warn(ctx context.Context, msg string, args ...any) {
	internalHandle(ctx, slog.LevelWarn, msg, args...)
}

func Error(ctx context.Context, msg string, args ...any) {
	internalHandle(ctx, slog.LevelError, msg, args...)
}
