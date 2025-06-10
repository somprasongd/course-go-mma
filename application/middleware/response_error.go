package middleware

import (
	"errors"
	"go-mma/util/errs"

	"github.com/gofiber/fiber/v3"
)

func ResponseError() fiber.Handler {
	return func(c fiber.Ctx) error {
		err := c.Next()
		if err == nil {
			return nil
		}

		return jsonError(c, err)
	}
}

// ย้ายจาก util/response มาไว้ที่นี่แทน เพราะใช้งานเฉพาะในนี้
func jsonError(c fiber.Ctx, err error) error {
	// Convert non-AppError to AppError with type ErrOperationFailed
	appErr, ok := err.(*errs.AppError)
	if !ok {
		appErr = errs.New(
			errs.ErrOperationFailed,
			err.Error(),
			err,
		)
	}

	// Get the appropriate HTTP status code
	statusCode := errs.GetHTTPStatus(err)

	// Retrieve the custom status code if it's a *fiber.Error
	var e *fiber.Error
	if errors.As(err, &e) {
		statusCode = e.Code
	}

	// Return structured response with error type and message
	return c.Status(statusCode).JSON(appErr)
}
