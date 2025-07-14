package logger

import (
	"context"

	"go.elastic.co/ecszap"
	"go.uber.org/zap"
)

type closeLog func() error

var baseLogger *zap.Logger

func Init() (closeLog, error) {
	config := zap.NewDevelopmentConfig()
	// ใช้ zap ร่วมกับ ecszap เพื่อให้รองรับการส่ง log ไปยัง Elastic Stack ได้ในอนาคต
	config.EncoderConfig = ecszap.ECSCompatibleEncoderConfig(config.EncoderConfig)

	var err error
	baseLogger, err = config.Build(ecszap.WrapCoreOption())

	if err != nil {
		return nil, err
	}

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
	if ok {
		return log
	}
	return baseLogger
}
