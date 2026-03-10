## เชื่อมโยง Log ด้วย Request ID

เนื่องจากตอนนี้มีเมื่อ request เข้ามาจะมีการกระจายไปทำงานในหลายๆ ชั้น หรือมีการเรียก service เข้าโมดูล เพื่อช่วยในการตรวจสอบ log เราจึงควรเชื่อมโยง log ทั้งหมดที่เกิดจาก request ครั้งเดียวกันเข้าด้วยกัน ซึ่งสามารถทำได้ง่าย โดยใช้ `requestId` ที่มีอยู่แล้วใน middleware ของ `RequestLogger`

- ปรับให้ logger สามารถใส่ใน context ได้

    > แก้ไขไฟล์ `common/logger/logger.go` โดยเพิ่มโค้ดตามนี้
    >

    ```go
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
    ```

- แก้ไข middleware ของ `RequestLogger` เพื่อใส่ `logger` เข้าไปใน context

    > แก้ไขไฟล์ `app/application/middleware/request_logger.go`
    >

    ```go
    func RequestLogger() fiber.Handler {
     return func(c fiber.Ctx) error {
      start := time.Now()
    
      log := logger.With(
       // ...
      )
    
      // สร้าง Context ใหม่ที่มี logger (จะได้ใช้ logger ตัวเดียวใน handler)
      ctx := logger.NewContext(c.Context(), log)
      // แทน Context เดิม
      c.SetContext(ctx)
    
      // catch panic
      // ...
     }
    }
    ```

- เมื่อต้องการใช้งาน logger ใน http handler, mediator handler, event handler, service หรือ repository โดยเปลี่ยนมาเรียกใช้ `logger.FromContext()` แทน ตัวอย่าง

    > แก้ไขไฟล์ `customer/internal/feature/create/endpoint.go`
    >

    ```diff
    func createCustomerHTTPHandler(c fiber.Ctx) error {
     // ...
    
    --  logger.Log.Info(fmt.Sprintf("Received customer: %v", req))
    ++ logger.FromContext(c.Context()).Info(fmt.Sprintf("Received customer: %v", req))
    
     // ...
    }
    
    ```

    > แก้ไขไฟล์ `notification/service/notification.go`
    >

    ```diff
    func (s *notificationService) SendEmail(ctx context.Context, to string, subject string, payload map[string]any) error {
     // implement email sending logic here
    -- logger.Log.Info(fmt.Sprintf("Sending email to %s with subject: %s and payload: %v", to, subject, payload))
    ++ logger.FromContext(ctx).Info(fmt.Sprintf("Sending email to %s with subject: %s and payload: %v", to, subject, payload))
     return nil
    }
    ```

- เมื่อรันใหม่อีกครั้ง และทดสอบสร้าง customer ให้ จะเห็นว่าการแสดง log ทั้งหมด ของ request ครั้งนี้ จะมีเลข requestId เดียวกัน

    ```bash
    2025-07-14T18:48:58.017+0700    INFO    map[file.line:34 file.name:create/endpoint.go function:go-mma/modules/customer/internal/feature/create.createCustomerHTTPHandler]       Received customer: {cust9@example.com 1000}     {"requestId": "a74318af-9bee-45fe-86f3-8a9c39636826", "method": "POST", "path": "/api/v1/customers", "ecs.version": "1.6.0"}
    2025-07-14T18:48:58.029+0700    INFO    map[file.line:65 file.name:middleware/request_logger.go function:go-mma/application/middleware.printAccessLog] 201 - POST /api/v1/customers     {"requestId": "a74318af-9bee-45fe-86f3-8a9c39636826", "method": "POST", "path": "/api/v1/customers", "status": 201, "latency": "12.802959ms", "ecs.version": "1.6.0"}
    2025-07-14T18:48:58.029+0700    INFO    map[file.line:26 file.name:service/notification.go function:go-mma/modules/notification/service.(*notificationService).SendEmail]       Sending email to cust9@example.com with subject: Welcome to our service! and payload: map[message:Thank you for joining us! We are excited to have you as a member.]    {"requestId": "a74318af-9bee-45fe-86f3-8a9c39636826", "method": "POST", "path": "/api/v1/customers", "ecs.version": "1.6.0"}
    ```

---

