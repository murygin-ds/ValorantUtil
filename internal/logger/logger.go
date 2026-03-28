package logger

import (
	"log"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

type Logger struct {
	*zap.SugaredLogger
}

func New(env string, filePath string) *Logger {
	var cfg zap.Config

	if env == "production" {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
	}

	outputs := []string{"stdout"}
	if filePath != "" {
		dir := filepath.Dir(filePath)
		if dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				log.Fatalf("failed to create log directory %s: %v", dir, err)
			}
		}
		outputs = append(outputs, filePath)
	}
	cfg.OutputPaths = outputs
	cfg.ErrorOutputPaths = outputs

	logger, err := cfg.Build()
	if err != nil {
		log.Fatal(err)
	}

	return &Logger{logger.Sugar()}
}

func (l *Logger) Sync() {
	_ = l.SugaredLogger.Sync()
}
