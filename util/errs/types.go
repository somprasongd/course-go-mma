package errs

type ErrorType string

const (
	ErrInputValidation   ErrorType = "input_validation_error"   // Invalid input (e.g., missing fields, format issues)
	ErrAuthentication    ErrorType = "authentication_error"     // Wrong credentials, not logged in
	ErrAuthorization     ErrorType = "authorization_error"      // No permission to access resource
	ErrResourceNotFound  ErrorType = "resource_not_found"       // Entity does not exist
	ErrConflict          ErrorType = "conflict"                 // Conflict, already exists
	ErrBusinessRule      ErrorType = "business_rule_error"      // Business rule violation
	ErrDataIntegrity     ErrorType = "data_integrity_error"     // Foreign key, constraint violations
	ErrDatabaseFailure   ErrorType = "database_failure"         // Generic DB error
	ErrOperationFailed   ErrorType = "operation_failed"         // General failure case
	ErrServiceDependency ErrorType = "service_dependency_error" // External service unavailable
)
