package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	Web      *slog.Logger
	Telegram *slog.Logger
	Executor *slog.Logger
	WebWriter io.Writer // Exported for Gin
)

// InitLoggers 初始化多個 Logger，分別輸出到不同檔案
func InitLoggers(logDir string, debug bool) error {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	// Helper to create a logger
	createLogger := func(filename string) (*slog.Logger, io.Writer) {
		fileLogger := &lumberjack.Logger{
			Filename:   filepath.Join(logDir, filename),
			MaxSize:    10,   // megabytes
			MaxBackups: 30,   // max files
			MaxAge:     30,   // days
			LocalTime:  true, // use local time for filenames
			Compress:   true, // compress rotated files
		}
		// 同時輸出到檔案與控制台
		w := io.MultiWriter(os.Stdout, fileLogger)
		return slog.New(slog.NewJSONHandler(w, opts)), w
	}

	var w io.Writer
	Web, w = createLogger("web.log")
	WebWriter = w // Assign to exported variable

	Telegram, _ = createLogger("telegram.log")
	Executor, _ = createLogger("executor.log")

	// 設定預設 Logger 為 Web
	slog.SetDefault(Web)

	return nil
}
