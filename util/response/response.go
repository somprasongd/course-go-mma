package response

import (
	"errors"
	"go-mma/util/errs"

	"github.com/gofiber/fiber/v3"
)

func JSONError(c fiber.Ctx, err error) error {
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
