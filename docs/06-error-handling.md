## ออกแบบระบบจัดการ Error ให้ตรวจสอบและแก้ไขง่าย

ในโปรเจกต์ช่วงแรก เรามัก return error กลับไปยัง client ด้วย HTTP status code เดียว เช่น `500 Internal Server Error` โดยไม่แยกประเภท error อย่างเหมาะสม ทำให้ client ไม่สามารถแยกแยะได้ว่าเกิดอะไรขึ้น เช่น input ผิด หรือ server พัง

**ในหัวข้อนี้ เราจะมาออกแบบระบบ error handling ให้เป็นระบบ รองรับทั้ง developer และ client อย่างเหมาะสม**

- วางมาตรฐาน HTTP Status Code ของระบบ
- สร้าง Custome Error
- การจัดการ Error ใน Repository Layer
- การจัดการ Error ใน Service Layer
- การจัดการ Error ใน Presentation Layer
- สร้าง ErrorHandler Middleware

### วางมาตรฐาน HTTP Status Code ของระบบ

ก่อนอื่นต้องกำหนดว่า error แต่ละประเภทจะ map กับ status code อะไร

| ประเภท | สถานะ | ใช้เมื่อ | หมายเหตุ |
| --- | --- | --- | --- |
| Input Validation | 400 Bad Request | ข้อมูลไม่ครบ, รูปแบบผิด | เกิดจาก **Client-Side**
ตรวจจับได้ที่ Handler / DTO |
| Authorization | 401 Unauthorized | ยังไม่ login / token ผิด | ตรวจจับได้ที่ Middleware |
|  | 403 Forbidden | login แล้ว แต่ไม่มีสิทธิ์ | ตรวจจับได้ที่ Middleware |
| Business Logic | 404 Not Found | ไม่พบข้อมูล | ตรวจจับได้ที่ Service |
|  | 409 Conflict | ข้อมูลซ้ำกัน, ขัดแย้ง เช่น email ซ้ำ, order ถูก cancel ไปแล้ว | ตรวจจับได้ที่ Service |
|  | 422 Unprocessable Entity | ข้อมูลมีรูปแบบถูก แต่ logic ผิด เช่น เครดิตไม่พอ, วันที่ย้อนหลัง | ตรวจจับได้ที่ Service |
| Database | 500 Internal Server Error | เกิด database connection error | ตรวจจับได้ที่ Repository |
| Exception | 500 Internal Server Error | เกิด exception หรือ panic ใน server code | เกิดได้ทุกที่ |

### สร้าง Custome Error

เมื่อได้ error ทั้งหมดที่จะเกิดขึ้นได้แล้วนั้น ก็มาสร้าง custome error เพื่อใช้จัดการ error ทั้งหมดที่จะเกิดขึ้นในระบบ

- สร้างไฟล์ `util/errs/types.go` ไว้สำหรับกำหนดประเภท error ทั้งหมดก่อน

    ```go
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
    ```

- สร้าง Custom Error ให้สร้างไฟล์ `util/errs/errs.go`

    ```go
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
    ```

### การจัดการ Error ใน Repository Layer

ในชั้นของ repository เมื่อเชื่อมต่อกับฐานข้อมูล PostgreSQL จะเกิด error ได้ ดังนี้

- 23502: Not null violation → **ErrDataIntegrity**
- 23503: Foreign key violation → **ErrDataIntegrity**
- 23505: Unique constraint violation → **ErrConflict**
- อื่นๆ → **ErrDatabaseFailure**

ขั้นตอนการ implement

- สร้างไฟล์ `util/errs/helpers.go` สำหรับเป็นตัวช่วย Map error code กับ error type

    ```go
    package errs
    
    import (
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
    ```

- แก้ไฟล์ `repository/customer.go` เพื่อมาเรียกใช้งาน `HandleDBError`

    ```go
    func (r *CustomerRepository) Create(ctx context.Context, customer *model.Customer) error {
     // ...
     if err != nil {
       // ใช้ตรงนี้
      return errs.HandleDBError(fmt.Errorf("an error occurred while inserting customer: %w", err))
     }
     return nil // Return nil if the operation is successful
    }
    
    func (r *CustomerRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
     // ...
     if err != nil {
      if err == sql.ErrNoRows {
       return false, nil
      }
      // ใช้ตรงนี้
      return false, errs.HandleDBError(fmt.Errorf("an error occurred while checking email: %w", err))
     }
     return true, nil
    }
    ```

### การจัดการ Error ใน Service Layer

Service layer จะเป็นจุดที่ตัดสินใจว่า error ไหนควรถูกห่อด้วย custom error อะไร

แต่ถ้าเป็น error ที่ได้มาจาก repository layer เราจะคืนกลับ error นั้นๆ ได้เลย เพราะถูกจัดการมาแล้ว

ในไฟล์ `service/customer.go` มีแค่ error การตรวจ business logic ห้ามอีเมลซ้ำเท่านั้น

```go
var (
 ErrEmailExists = errs.ConflictError("email already exists")
)
```

### การจัดการ Error ใน Presentation Layer

ใน handler จะมี error ดังนี้

- **การแปลง JSON → DTO**: ต้องเปลี่ยนมาใช้ AppError
- **การตรวจสอบ DTO**: ต้องเปลี่ยนมาใช้ AppError
- **Error ที่ได้รับมาจาก Service Layer**: สามารถใช้ได้เลย

ขั้นตอนการ implement

- แก้ไขไฟล์ `handler/customer.go` เพื่อเปลี่ยนมาใช้ `AppError`

    ```go
    func (h *CustomerHandler) CreateCustomer(c fiber.Ctx) error {
     // แปลง request body -> dto
     var req dto.CreateCustomerRequest
     if err := c.Bind().Body(&req); err != nil {
      // แปลงไม่ได้ให้ส่ง error 400
      errResp := errs.InputValidationError(err.Error()) // <-- ปรับมาเป็น AppError
      return c.Status(fiber.StatusBadRequest).JSON(errResp)
     }
    
     // ตรวจสอบ input fields (e.g., value, format, etc.)
     if err := req.Validate(); err != nil {
      errResp := errs.InputValidationError(err.Error()) // <-- ปรับมาเป็น AppError
      return c.Status(fiber.StatusBadRequest).JSON(errResp)
     }
    
     // ส่งไปที่ Service Layer
     // จัดการ error จาก Service Layer หากเกิดขึ้น
     // ตอบกลับด้วย status code 201 (created) และข้อมูลแบบ JSON
    }
    ```

- แต่จะเห็นว่า status code ยังไม่ถูกต้อง ซึ่งจะต้องดึงมาจาก AppError ดังนั้น ให้สร้าง helper function สำหรับถอด status code มา โดยแก้ไขไฟล์ `util/errs/helpers.go` ให้เพิ่ม ตามนี้

    ```go
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
    ```

- แก้ไขไฟล์ `handler/customer.go` เพื่อใช้ status code ที่ถูกต้อง

    ```go
    func (h *CustomerHandler) CreateCustomer(c fiber.Ctx) error {
     // แปลง request body -> dto
     var req dto.CreateCustomerRequest
     if err := c.Bind().Body(&req); err != nil {
      // แปลงไม่ได้ให้ส่ง error 400
      errResp := errs.InputValidationError(err.Error())
      return c.Status(
       errs.GetHTTPStatus(errResp), // <-- ดึง status code จาก AppError
      ).JSON(errResp)
     }
    
     // ตรวจสอบ input fields (e.g., value, format, etc.)
     if err := req.Validate(); err != nil {
      errResp := errs.InputValidationError(err.Error())
      return c.Status(
       errs.GetHTTPStatus(errResp), // <-- ดึง status code จาก AppError
      ).JSON(errResp)
     }
    
     // ส่งไปที่ Service Layer
     // จัดการ error จาก Service Layer หากเกิดขึ้น
     // ตอบกลับด้วย status code 201 (created) และข้อมูลแบบ JSON
    }
    ```

- สร้าง Standard Error Response เพื่อมาจัดการส่ง error response ให้สร้างไฟล์ `util/response/response.go`

    ```go
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
    ```

- แก้ไขไฟล์ `handler/customer.go` เพื่อใช้เรียกใช้ `JSONError`

    ```go
    func (h *CustomerHandler) CreateCustomer(c fiber.Ctx) error {
     // 1. รับ request body มาเป็น DTO
     var req dto.CreateCustomerRequest
     if err := c.Bind().Body(&req); err != nil {
      return response.JSONError(c, errs.InputValidationError(err.Error()))
     }
    
     // 2. ตรวจสอบความถูกต้อง (validate)
     if err := req.Validate(); err != nil {
      return response.JSONError(c, errs.InputValidationError(strings.Join(strings.Split(err.Error(), "\n"), ", ")))
     }
    
     // 3. ส่งไปที่ Service Layer
     resp, err := h.custService.CreateCustomer(c.Context(), &req)
    
     // 4. จัดการ error จาก Service Layer หากเกิดขึ้น
     if err != nil {
      return response.JSONError(c, err)
     }
    
     // 5. ตอบกลับ client
     return c.Status(fiber.StatusCreated).JSON(resp)
    }
    ```

### สร้าง Middleware สำหรับจัดการตอบกลับ Error

อีกวิธีการหนึ่งในการตอบกลับ error แทนที่จะเรียก `response.JSONError` ในทุกๆ ที่ ที่เกิด error ขึ้นใน handler คือ ให้ `return error` กลับออกไปเลย แล้วสร้าง middleware ใหม่ ขึ้นมาจัดการแทน ดังนี้

- สร้างไฟล์ `application/middleware/response_error.go`

    ```go
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
    ```

- แก้ไขไฟล์ `application/http.go` เพื่อเรียกใช้ middleware `ResponseError`

    ```go
    func newFiber() *fiber.App {
     app := fiber.New(fiber.Config{
      AppName: fmt.Sprintf("Go MMA version %s", build.Version),
     })
    
     // global middleware
     app.Use(cors.New())                 // CORS ลำดับแรก เพื่อให้ OPTIONS request ผ่านได้เสมอ
     app.Use(requestid.New())            // สร้าง request id ใน request header สำหรับการ debug
     app.Use(recover.New())              // auto-recovers from panic (internal only)
     app.Use(middleware.RequestLogger()) // logs HTTP request
     app.Use(middleware.ResponseError()) // <-- เพิ่มตรงนี้
    
     app.Get("/", func(c fiber.Ctx) error {
      return c.JSON(map[string]string{"version": build.Version, "time": build.Time})
     })
    
     return app
    }
    ```

- แก้ไขไฟล์ `handler/customer.go` ให้ `return errror` กลับออกไปให้ middleware จัดการ

    ```go
    func (h *CustomerHandler) CreateCustomer(c fiber.Ctx) error {
     // แปลง request body -> dto
     var req dto.CreateCustomerRequest
     if err := c.Bind().Body(&req); err != nil {
      // <-- return error ออกไปเลย
      return errs.InputValidationError(err.Error())
     }
    
     logger.Log.Info(fmt.Sprintf("Received customer: %v", req))
    
     // ตรวจสอบ input fields (e.g., value, format, etc.)
     if err := req.Validate(); err != nil {
      // <-- return error ออกไปเลย
      return errs.InputValidationError(err.Error())
     }
    
     // ส่งไปที่ Service Layer
     resp, err := h.custService.CreateCustomer(c.Context(), &req)
    
     // จัดการ error จาก Service Layer หากเกิดขึ้น
     if err != nil {
      // <-- return error ออกไปเลย
      return err
     }
    
     // ตอบกลับด้วย status code 201 (created) และข้อมูลแบบ JSON
     return c.Status(fiber.StatusCreated).JSON(resp)
    }
    ```

