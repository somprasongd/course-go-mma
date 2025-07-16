package logger

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type closeLog func() error

var baseLogger *zap.Logger

func Init(serviceName string) (closeLog, error) {
	config := zap.NewProductionConfig() // ใช้ production config → output เป็น JSON

	// ตั้งค่า key ของ field ให้ตรงกับ OTel semantic convention
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.LevelKey = "severity"
	config.EncoderConfig.MessageKey = "message"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	var err error
	zlog, err := config.Build()

	if err != nil {
		return nil, err
	}

	baseLogger = zlog.With(zap.String("service.name", serviceName))

	return func() error {
		return baseLogger.Sync()
	}, nil
}

func Log() *zap.Logger {
	return baseLogger
}

func With(fields ...zap.Field) *zap.Logger {
	return baseLogger.With(fields...)
}

type loggerKey struct{}

func NewContext(parent context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(parent, loggerKey{}, logger)
}

func FromContext(ctx context.Context) *zap.Logger {
	log, ok := ctx.Value(loggerKey{}).(*zap.Logger)
	if !ok {
		log = baseLogger
	}

	// ดึง trace_id + span_id จาก OTel context
	span := trace.SpanContextFromContext(ctx)

	if span.IsValid() {
		return log.With(
			zap.String("trace_id", span.TraceID().String()), // เพิ่ม trace_id เพื่อเชื่อมโยง log กับ trace
			zap.String("span_id", span.SpanID().String()),   // เพิ่ม span_id เพื่อเชื่อมโยง log กับ trace
		)
	}

	return log
}
