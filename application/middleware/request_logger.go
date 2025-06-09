package middleware

import (
	"fmt"
	"go-mma/util/logger"
	"runtime/debug"
	"time"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

func RequestLogger() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()

		log := logger.With(
			zap.String("requestId", c.GetRespHeader("X-Request-ID")),
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
		)

		// catch panic
		defer func() {
			if r := recover(); r != nil {
				printAccessLog(log, c.Method(), c.Path(), start, fiber.StatusInternalServerError, r)
				panic(r) // throw panic to recover middleware
			}
		}()

		err := c.Next()

		status := c.Response().StatusCode()
		if err != nil {
			switch e := err.(type) {
			case *fiber.Error:
				status = e.Code
			default: // case error
				status = fiber.StatusInternalServerError
			}
		}

		printAccessLog(log, c.Method(), c.Path(), start, status, err)

		return err
	}
}

func printAccessLog(log *zap.Logger, method string, uri string, start time.Time, status int, err any) {
	if err != nil {
		// log unhandle error
		log.Error("an error occurred",
			zap.Any("error", err),
			zap.ByteString("stack", debug.Stack()),
		)
	}

	msg := fmt.Sprintf("%d - %s %s", status, method, uri)

	log.Info(msg,
		zap.Int("status", status),
		zap.Duration("latency", time.Since(start)))
}
