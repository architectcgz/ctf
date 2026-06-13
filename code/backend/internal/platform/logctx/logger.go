package logctx

import (
	"context"

	"go.uber.org/zap"

	"ctf-platform/internal/platform/requestctx"
)

func Info(ctx context.Context, logger *zap.Logger, msg string, fields ...zap.Field) {
	withContext(ctx, logger).Info(msg, fields...)
}

func Warn(ctx context.Context, logger *zap.Logger, msg string, fields ...zap.Field) {
	withContext(ctx, logger).Warn(msg, fields...)
}

func Error(ctx context.Context, logger *zap.Logger, msg string, fields ...zap.Field) {
	withContext(ctx, logger).Error(msg, fields...)
}

func withContext(ctx context.Context, logger *zap.Logger) *zap.Logger {
	if logger == nil {
		logger = zap.NewNop()
	}
	requestID := requestctx.RequestIDFromContext(ctx)
	if requestID == "" {
		return logger
	}
	return logger.With(zap.String("request_id", requestID))
}
