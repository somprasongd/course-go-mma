package errs

import "fmt"

type AppError struct {
	Type    ErrorType `json:"type"`    // สำหรับ client
	Message string    `json:"message"` // สำหรับ client
	Err     error     `json:"-"`       // สำหรับ log ภายใน
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s - %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// Unwrap allows for errors.Is and errors.As compatibility
func (e *AppError) Unwrap() error {
	return e.Err
}

func New(errorType ErrorType, message string, err ...error) *AppError {
	var underlyingErr error
	if len(err) > 0 {
		underlyingErr = err[0]
	}
	return &AppError{
		Type:    errorType,
		Message: message,
		Err:     underlyingErr,
	}
}

// Helper functions for each error type
func InputValidationError(message string, err ...error) *AppError {
	return New(ErrInputValidation, message, err...)
}

func AuthenticationError(message string, err ...error) *AppError {
	return New(ErrAuthentication, message, err...)
}

func NewAuthorizationError(message string, err ...error) *AppError {
	return New(ErrAuthorization, message, err...)
}

func ResourceNotFoundError(message string, err ...error) *AppError {
	return New(ErrResourceNotFound, message, err...)
}

func ConflictError(message string, err ...error) *AppError {
	return New(ErrConflict, message, err...)
}

func BusinessRuleError(message string, err ...error) *AppError {
	return New(ErrBusinessRule, message, err...)
}

func DataIntegrityError(message string, err ...error) *AppError {
	return New(ErrDataIntegrity, message, err...)
}

func DatabaseFailureError(message string, err ...error) *AppError {
	return New(ErrDatabaseFailure, message, err...)
}

func OperationFailedError(message string, err ...error) *AppError {
	return New(ErrOperationFailed, message, err...)
}

func ServiceDependencyError(message string, err ...error) *AppError {
	return New(ErrServiceDependency, message, err...)
}
