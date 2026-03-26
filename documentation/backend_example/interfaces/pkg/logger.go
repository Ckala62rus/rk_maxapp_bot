package pkg

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"runtime"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

var (
	MainLogger *slog.Logger
)

func init() {
	// Используем ENV переменную или значение по умолчанию
	env := os.Getenv("ENV")
	if env == "" {
		env = envDev
	}

	MainLogger = SetupNewLogger(env)
	appName := os.Getenv("APP_NAME")
	if appName == "" {
		appName = "interfaces"
	}
	appVersion := os.Getenv("APP_VERSION")
	if appVersion == "" {
		appVersion = "0.1.0"
	}
	MainLogger = MainLogger.With("app_name", appName, "app_version", appVersion)
	MainLogger.Info("*** LOGGER INITIALIZED RUN ***")
}

type PrettyJSONHandler struct {
	slog.Handler
	enc    *json.Encoder
	opts   *slog.HandlerOptions
	writer io.Writer
}

// NewPrettyJSONHandler для красивого отображения JSON с отступами
func NewPrettyJSONHandler(w io.Writer, opts *slog.HandlerOptions) *PrettyJSONHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}

	return &PrettyJSONHandler{
		Handler: slog.NewJSONHandler(w, opts),
		enc:     json.NewEncoder(w),
		opts:    opts,
		writer:  w,
	}
}

func (h *PrettyJSONHandler) Handle(ctx context.Context, r slog.Record) error {
	// Собираем атрибуты
	attrs := make(map[string]any, r.NumAttrs()+3)

	// Добавляем стандартные поля
	attrs["time"] = r.Time.Format("2006-01-02T15:04:05.000Z07:00")
	attrs["level"] = r.Level.String()
	attrs["msg"] = r.Message

	// Добавляем source если включено
	if h.opts.AddSource && r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()
		attrs["source"] = map[string]any{
			"function": f.Function,
			"file":     f.File,
			"line":     f.Line,
		}
	}

	// Добавляем остальные атрибуты
	r.Attrs(func(attr slog.Attr) bool {
		attrs[attr.Key] = attr.Value.Any()
		return true
	})

	// Настраиваем encoder для каждого вызова
	enc := json.NewEncoder(h.writer)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)

	// Кодируем с отступами
	return enc.Encode(attrs)
}

func SetupNewLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev, "development":
		log = slog.New(NewPrettyJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:     slog.LevelDebug,
			AddSource: true,
		}))
	case envProd, "production":
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		// Fallback to dev mode for unknown environments
		log = slog.New(NewPrettyJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}

	return log
}
