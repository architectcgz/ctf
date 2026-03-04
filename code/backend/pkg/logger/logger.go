package logger

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"ctf-platform/internal/config"
)

func New(cfg config.LogConfig) (*zap.Logger, error) {
	level := zap.NewAtomicLevel()
	if err := level.UnmarshalText([]byte(strings.ToLower(cfg.Level))); err != nil {
		level.SetLevel(zap.InfoLevel)
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	encoding := strings.ToLower(cfg.Format)
	if encoding != "json" {
		encoding = "console"
	}

	zapConfig := zap.Config{
		Level:            level,
		Development:      encoding == "console",
		Encoding:         encoding,
		EncoderConfig:    encoderConfig,
		OutputPaths:      nonEmpty(cfg.OutputPaths, []string{"stdout"}),
		ErrorOutputPaths: nonEmpty(cfg.ErrorOutputPaths, []string{"stderr"}),
	}

	return zapConfig.Build(zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
}

func nonEmpty(values []string, fallback []string) []string {
	if len(values) == 0 {
		return fallback
	}
	return values
}
