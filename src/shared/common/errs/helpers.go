package errs

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/lib/pq"
)

// HandleDBError maps PostgreSQL errors to custom application errors
func HandleDBError(err error) *AppError {
	if pgErr, ok := err.(*pq.Error); ok {
		switch pgErr.Code {
		case "23505": // Unique constraint violation
			return New(ErrConflict, "duplicate entry detected: "+pgErr.Message)
		case "23503": // Foreign key violation
			return New(ErrDataIntegrity, "foreign key constraint violation: "+pgErr.Message)
		case "23502": // Not null violation
			return New(ErrDataIntegrity, "not null constraint violation: "+pgErr.Message)
		default:
			return New(ErrDatabaseFailure, "database error: "+pgErr.Message)
		}
	}
	// Fallback for unknown DB errors
	return New(ErrDatabaseFailure, err.Error())
}

// GetErrorType extracts the error type from an errorAdd commentMore actions
func GetErrorType(err error) ErrorType {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type
	}
	return ErrOperationFailed // Default error type if not recognized
}

// Map error type to HTTP status code
func GetHTTPStatus(err error) int {
	switch GetErrorType(err) {
	case ErrInputValidation:
		return fiber.StatusBadRequest // 400
	case ErrAuthentication:
		return fiber.StatusUnauthorized // 401
	case ErrAuthorization:
		return fiber.StatusForbidden // 403
	case ErrResourceNotFound:
		return fiber.StatusNotFound // 404
	case ErrConflict:
		return fiber.StatusConflict // 409
	case ErrBusinessRule:
		return fiber.StatusUnprocessableEntity // 422
	case ErrDataIntegrity, ErrDatabaseFailure:
		return fiber.StatusInternalServerError // 500
	case ErrOperationFailed:
		return fiber.StatusInternalServerError // 500
	case ErrServiceDependency:
		return fiber.StatusServiceUnavailable // 503
	default: // Default: Unknown errors, fallback to internal server error
		return fiber.StatusInternalServerError // 500
	}
}
