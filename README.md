# แนวทางการสร้างโปรเจกต์แบบ Modular Monolith

> หากคุณเคยเริ่มพัฒนาแอปพลิเคชันด้วยโครงสร้างแบบ Monolith แล้วพบว่าระบบเริ่มยุ่งเหยิงเมื่อทีมเติบโตขึ้น หรือเมื่อฟีเจอร์ใหม่ๆ เริ่มทับซ้อนกัน — Modular Monolith อาจเป็นคำตอบที่เหมาะสม
>

บทความนี้จะพาคุณไล่เรียงตั้งแต่พื้นฐานของการพัฒนาเว็บแอปในรูปแบบ Monolith ไปสู่การออกแบบระบบแบบ **Modular Monolith** อย่างเป็นระบบ พร้อมแนวคิดและแนวปฏิบัติที่นำไปใช้ได้จริง โดยไม่ต้องกระโดดไปสู่ Microservices ตั้งแต่แรก

## สารบัญ

- โปรเจกต์นี้มีเป้าหมายอะไร
- ออกแบบ Web Server ให้ดีตั้งแต่แรก
- การแยก logic ออกจาก routing ด้วย Handlers
- เชื่อมต่อฐานข้อมูลอย่างปลอดภัยและยืดหยุ่น
- เริ่มต้นด้วย Layered Architecture ที่เข้าใจง่าย
- ออกแบบระบบจัดการ Error ให้ตรวจสอบและแก้ไขง่าย
- สร้างระบบส่งอีเมลแบบ Reusable ด้วย Notification Service
- สร้างระบบจัดการออเดอร์ ด้วย Layered Architecture

---

## โปรเจกต์นี้มีเป้าหมายอะไร

เป็นโปรเจกต์ Monolith App แบบง่ายๆ ที่มีฟีเจอร์ ดังนี้

1. สร้างลูกค้าใหม่ พร้อมส่งอีเมลต้อนรับ
2. สั่งออเดอร์ พร้อมส่งอีเมลยืนยันคำสั่งออเดอร์
3. ยกเลิกออเดอร์

### ภาพรวมโปรเจกต์

```markdown
+------------+        +----------------------+        +-----------+
|   Client   | <----> |    Monolith App      | <----> | Database  |
+------------+        |----------------------|        +-----------+
                      |  Modules:            |
                      |    - customer        |
                      |    - order           |
                      |    - email           |
                      +----------------------+

1. สร้างลูกค้าใหม่ (POST /customers)
---------------------------------------
Client ----> Monolith: POST /customers {email, credit}
Monolith.customer --> Database: ตรวจสอบ email ซ้ำ?
  └─ ซ้ำ --> Monolith.customer --> Client: 409 Conflict (email already exists)
  └─ ไม่ซ้ำ:
      Monolith.customer --> Database: INSERT INTO customers
      Monolith.email --> ส่งอีเมลต้อนรับ
      Monolith.customer --> Client: 201 Created

2. สั่งออเดอร์ (POST /orders)
-------------------------------
Client ----> Monolith: POST /orders {customer_id, order_total}
Monolith.order --> Database: ตรวจสอบ customer_id
  └─ ไม่พบ --> Monolith.order --> Client: 404 Not Found (customer not found)
  └─ พบ:
      Monolith.order --> Database: ตรวจสอบ credit เพียงพอ?
          └─ ไม่พอ --> Monolith.order --> Client: 422 Unprocessable Entity (insufficient credit)
          └─ พอ:
              Monolith.order --> Database: INSERT INTO orders, UPDATE credit (หักยอด)
              Monolith.email --> ส่งอีเมลยืนยันออเดอร์
              Monolith.order --> Client: 201 Created

3. ยกเลิกออเดอร์ (DELETE /orders/:orderID)
---------------------------------------------
Client ----> Monolith: DELETE /orders/:orderID
Monolith.order --> Database: ตรวจสอบ orderID
  └─ ไม่พบ --> Monolith.order --> Client: 404 Not Found (order not found)
  └─ พบ:
      Monolith.order --> Database: DELETE order, UPDATE credit (คืนยอด)
      Monolith.order --> Client: 204 No Content
```

### API endpoint

ระบบนี้มี 3 API endpoint

- `POST /customers` – สร้างลูกค้าใหม่

    | JSON Field | Type | Required | Description |
    | --- | --- | --- | --- |
    | `email` | string | ✅ | อีเมลลูกค้า |
    | `credit` | integer | ✅ | เครดิตเริ่มต้น |

    **Response**

    | Status Code | Description |
    | --- | --- |
    | `201` | สร้างสำเร็จ |
    | `400` | payload ไม่ครบ หรือข้อมูลไม่ถูกต้อง เช่น ไม่ส่ง `email` หรือ `email` ผิดรูปแบบ |
    | `422` | ไม่ผ่าน business rule เช่น `credit` ≤ 0 |
    | `409` | อีเมลนี้มีอยู่แล้วในระบบ (Conflict) |

- `POST /orders` – สร้างออเดอร์

    | JSON Field | Type | Required | Description |
    | --- | --- | --- | --- |
    | `customer_id` | integer | ✅ | ID ลูกค้า |
    | `order_total` | integer | ✅ | ยอดรวมออเดอร์ |

    **Response**

    | Status Code | Description |
    | --- | --- |
    | `201` | สร้างออเดอร์เรียบร้อย |
    | `400` | ไม่ส่ง `customer_id` หรือ `order_total` ≤ 0 |
    | `404` | ไม่พบลูกค้า (`customer_id` ไม่ถูกต้อง) |
    | `422` | เครดิตไม่เพียงพอในการสั่งออเดอร์ |

- `DELETE /orders/:orderID` – ยกเลิกออเดอร์

    | Path Param | Type | Required | Description |
    | --- | --- | --- | --- |
    | `orderID` | integer | ✅ | ID ออเดอร์ |

    **Response**

    | Status Code | Description |
    | --- | --- |
    | `204` | ลบออเดอร์สำเร็จ (No Content) |
    | `404` | ไม่พบออเดอร์นี้ในระบบ |

---

## ออกแบบ Web Server ให้ดีตั้งแต่แรก

> การออกแบบ Web Server ที่ดีตั้งแต่เริ่มต้น จะช่วยให้ระบบพร้อมขยาย รองรับการดูแล และทดสอบได้ง่ายในระยะยาว
>

เนื้อหาในส่วนนี้ประกอบด้วย

- เริ่มต้นสร้าง Web Server
- การจัดการ route ของ fiber
- การทดสอบ Rest API ด้วย REST Client
- รองรับ Graceful Shutdown
- สร้าง Logger กลางที่ใช้ได้ทั่วทั้งแอป
- การจัดการ Configurations
- ใช้ Makefile จัดการคำสั่ง
- Refactor เพื่อความสะอาดของโค้ด
- การ Build Web Server

### เริ่มต้นสร้าง Web Server

เริ่มจากสร้าง Web Server ขึ้นมาก่อน โดยในบทความนี้จะภาษา Go และใช้ Fiber v3

**เตรียมเครื่องมือ**

- [Go](https://go.dev/dl/) version 1.24.4 ขึ้นไป
- [VS Code](https://code.visualstudio.com/download) พร้อมติดตั้งส่วนเสริม [Go](https://marketplace.visualstudio.com/items?itemName=golang.Go), [Error Lens](https://marketplace.visualstudio.com/items?itemName=usernamehw.errorlens), [REST Client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client)

**ขั้นตอน**

- สร้างโปรเจคใหม่

    ```bash
    mkdir go-mma
    cd go-mma
    git init
    go mod init go-mma
    touch main.go
    ```

- จะได้แบบนี้

    ```bash
    tree
    .
    ├── go.mod
    └── main.go
    ```

- สร้าง Web Server ด้วย Fiber v3 ในไฟล์ `main.go`

    ```go
    package main
    
    import (
     "fmt"
    
     "github.com/gofiber/fiber/v3"
    )
    
    var (
     Version = "local-dev"
     Time    = "n/a"
    )
    
    func main() {
     app := fiber.New(fiber.Config{
      AppName: fmt.Sprintf("Go MMA version %s", Version),
     })
    
     app.Get("/", func(c fiber.Ctx) error {
       // การตอบกลับด้วย JSON
      return c.JSON(map[string]string{"version": Version, "time": Time})
     })
    
     app.Listen(":8090")
    }
    ```

- รันคำสั่ง `go mod tidy` เพื่อติดตั้ง package
- รันคำสั่ง `go run main.go` รันโปรแกรม

    ```bash
    go run main.go
    
        _______ __             
       / ____(_) /_  ___  _____
      / /_  / / __ \/ _ \/ ___/
     / __/ / / /_/ /  __/ /    
    /_/   /_/_.___/\___/_/          v3.0.0-beta.4
    --------------------------------------------------
    INFO Server started on:         http://127.0.0.1:8090 (bound on host 0.0.0.0 and port 3000)
    INFO Application name:          Go MMA v0.0.1
    INFO Total handlers count:      1
    INFO Prefork:                   Disabled
    INFO PID:                       47664
    INFO Total process count:       1
    ```

- ทดสอบเปิด <http://127.0.0.1:8090> ผ่านเบราว์เซอร์

    ```json
    {
    "time": "n/a",
    "version": "local-dev"
    }
    ```

### การจัดการ route ของ fiber v3

ควรแยกวางระบบ Routing ให้ชัดเจน แยก concerns อย่างเป็นระบบ พร้อมกับใช้งาน middlewares ที่จำเป็น

- แก้ไขไฟล์ `main.go`

    ```go
    package main
    
    import (
      "fmt"
      
     "github.com/gofiber/fiber/v3"
     "github.com/gofiber/fiber/v3/middleware/cors"
     "github.com/gofiber/fiber/v3/middleware/logger"
     "github.com/gofiber/fiber/v3/middleware/recover"
     "github.com/gofiber/fiber/v3/middleware/requestid"
    )
    
    var (
     Version = "local-dev"
     Time    = "n/a"
    )
    
    func main() {
     app := fiber.New(fiber.Config{
      AppName: fmt.Sprintf("Go MMA version %s", Version),
     })
    
     // กำหนด global middleware
     app.Use(cors.New())      // CORS ลำดับแรก เพื่อให้ OPTIONS request ผ่านได้เสมอ
     app.Use(requestid.New()) // สร้าง request id ใน request header สำหรับการ debug
     app.Use(recover.New())   // auto-recovers from panic (internal only)
     app.Use(logger.New())    // logs HTTP request
     
     app.Get("/", func(c fiber.Ctx) error {
      return c.JSON(map[string]string{"version": Version, "time": Time})
     })
    
      // แยกการทำ routing ให้ชัดเจน
     v1 := app.Group("/api/v1")
    
      // สร้างกลุ่มของ customer
     customers := v1.Group("/customers")
     {
      customers.Post("", func(c fiber.Ctx) error {
        // เพิ่มการกำหนด status code ด้วย Status()
       return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": 1})
      })
     }
    
      // สร้างกลุ่มของ order
     orders := v1.Group("/orders")
     {
      orders.Post("", func(c fiber.Ctx) error {
       return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": 1})
      })
    
      orders.Delete("/:orderID", func(c fiber.Ctx) error {
        // การตอบกลับแค่ status code เพียงอย่างเดียว
       return c.SendStatus(fiber.StatusNoContent)
      })
     }
    
     app.Listen(":8090")
    }
    ```

- รันโปรแกรมใหม่ `go run main.go`

### การทดสอบ Rest API ด้วย REST Client

ในบทความนี้จะใช้ VS Code Extensions ชื่อ [REST Client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client)

- ติดตั้ง [REST Client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client) หากยังไม่มี
- สร้างไฟล์ใหม่ 2 ไฟล์ `test/customers.http` กับ `test/orders.http`
- จะได้แบบนี้

    ```bash
    tree
    .
    ├── go.mod
    ├── main.go
    └── test
      ├── customers.http
        └── orders.http
    ```

- แก้ไขไฟล์ `test/customers.http`

    ```bash
    @host = http://localhost:8090
    @base_url = api/v1/customers
    ### Create Customer
    POST {{host}}/{{base_url}} HTTP/1.1
    content-type: application/json
    
    {
      "email": "cust@example.com",
      "credit": 1000
    }
    ```

- แก้ไขไฟล์ `tests/orders.http`

    ```bash
    @host = http://localhost:8090
    @base_url = api/v1/orders
    @customer_id = 1
    @order_id = 1
    ### Create Order
    POST {{host}}/{{base_url}} HTTP/1.1
    content-type: application/json
    
    {
      "customer_id": {{customer_id}},
      "order_total": 100
    }
    
    ### Cancel Order
    DELETE {{host}}/{{base_url}}/{{order_id}} HTTP/1.1
    ```

- การทดสอบเรียก API ให้กดที่คำว่า `Send Request`

    ```bash
    @host = http://localhost:8090
    @base_url = api/v1/customers
    ### Create Customer
    Send Request
    POST {{host}}/{{base_url}} HTTP/1.1
    content-type: application/json
    
    {
      "email": "cust@example.com",
      "credit": 1000
    }
    ```

- จะได้ผลลัพธ์แบบนี้

    ```bash
    HTTP/1.1 201 Created
    Date: Thu, 29 May 2025 04:07:11 GMT
    Content-Type: application/json
    Content-Length: 11
    Connection: close
    
    {
      "id": 1
    }
    ```

### รองรับ Graceful Shutdown

ในการออกแบบ Web Server ให้พร้อมใช้งานในระดับ production สิ่งหนึ่งที่ไม่ควรมองข้ามคือ **Graceful Shutdown — หรือการปิดระบบโดยไม่กระทบต่อการให้บริการที่กำลังดำเนินอยู่**

> หลักการง่ายๆ คือ: “อย่าตัดทุกอย่างทันที ให้โอกาสระบบได้ปิดตัวเองอย่างเรียบร้อยก่อน”
>

ทำไมถึงสำคัญ?

- ป้องกันการสูญเสียข้อมูลหรือการทำงานครึ่งๆ กลางๆ
- ปิดการเชื่อมต่อกับฐานข้อมูลและ service อื่นๆ อย่างปลอดภัย
- ช่วยให้ระบบสามารถ deploy หรือ restart ได้อย่างราบรื่น โดยไม่กระทบต่อผู้ใช้งาน

การทำ Graceful Shutdown ด้วยภาษา Go

- เริ่มจาก **ย้ายการ start server** ไปรันใน goroutine เพื่อให้ main thread สามารถรอรับสัญญาณ shutdown ได้

    ```go
    // ย้ายมา run server ใน goroutine
    go func() {
      if err := app.Listen(":8090"); err != nil && err != http.ErrServerClosed {
        log.Fatalf("Error starting server: %v", err)
      }
    }()
    ```

- จากนั้น **รอสัญญาณ OS** เช่น `SIGINT` หรือ `SIGTERM` เพื่อเริ่มกระบวนการ shutdown

    ```go
    // รอสัญญาณการปิด
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
    <-stop
    ```

- เมื่อได้รับสัญญาณ ให้เริ่ม **หยุดรับ request ใหม่** แล้ว **รอให้ request เดิมทำงานเสร็จ** ภายใน timeout ที่กำหนด (เช่น 5 วินาที) แล้วค่อยปิด resource อื่นๆ เช่น DB connection

    ```go
    log.Println("Shutting down...")
    
    // **หยุดรับ** request **ใหม่** แล้ว **รอให้** request **เดิมทำงานเสร็จ** ภายใน timeout ที่กำหนด (เช่น 5 วินาที)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := app.ShutdownWithContext(ctx); err != nil {
      log.Fatalf("Error shutting down server: %v", err)
    }
    
    // Optionally: แล้วค่อยปิด resource อื่นๆ เช่น DB connection, cleanup, etc.
    
    log.Println("Shutdown complete.")
    ```

- เพื่อทดสอบ: แก้ endpoint สร้าง customer ให้มีการหน่วงเวลา 3 วินาที

    ```go
    customers.Post("", func(c fiber.Ctx) error {
      // เพิ่มหน่วงเวลา 3 วินาที สำหรับทดสอบ Graceful Shutdown
     time.Sleep(3 * time.Second)
     return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": 1})
    })
    ```

- รัน server แล้วเรียก `/customers` → จากนั้นกด `Ctrl+C` ระหว่างที่ request ยังไม่จบ จะเห็นว่า server **รอจนจบ** ก่อนจะ shutdown

    ```go
    2025/05/29 12:00:07 Shutting down...
    12:00:06 | 201 |  3.001456625s |       127.0.0.1 | POST    | /api/v1/customers       
    2025/05/29 12:00:09 Shutdown complete.
    ```

### สร้าง Logger กลางที่ใช้ได้ทั่วทั้งแอป

จากหัวข้อก่อนหน้า จะเห็นว่า log ที่พิมพ์ออกมามีรูปแบบไม่สม่ำเสมอ การจัดการให้ **รูปแบบ log เป็นมาตรฐานเดียวกัน** จึงเป็นสิ่งสำคัญ โดยสามารถทำได้ด้วยการสร้าง **centralized logger** ดังนี้

- สร้างไฟล์ `util/logger/logger.go` เพื่อกำหนด logger เพียงตัวเดียวให้ทั้งระบบใช้ร่วมกัน
มีการใช้ `zap` ร่วมกับ `ecszap` เพื่อให้รองรับการส่ง log ไปยัง Elastic Stack ได้ในอนาคต

    ```go
    package logger
    
    import (
     "go.elastic.co/ecszap"
     "go.uber.org/zap"
    )
    
    type closeLog func() error
    
    var Log *zap.Logger
    
    func Init() (closeLog, error) {
     config := zap.NewDevelopmentConfig()
     // ใช้ zap ร่วมกับ ecszap เพื่อให้รองรับการส่ง log ไปยัง Elastic Stack ได้ในอนาคต
     config.EncoderConfig = ecszap.ECSCompatibleEncoderConfig(config.EncoderConfig)
    
     var err error
     Log, err = config.Build(ecszap.WrapCoreOption())
    
     if err != nil {
      return nil, err
     }
    
     return func() error {
      return Log.Sync()
     }, nil
    }
    
    func With(fields ...zap.Field) *zap.Logger {
     return Log.With(fields...)
    }
    ```

- ติดตั้ง dependencies ด้วย `go mod tidy`
- เรียกใช้งาน logger จาก `main.go` โดยทำการ `init` และ `defer` ปิดให้เรียบร้อย เพื่อรองรับ `log.Sync()`

    ```go
    func main() {
     closeLog, err := logger.Init()
     if err != nil {
      panic(err.Error())
     }
     defer closeLog()
    
     // ...
    }
    ```

- ปรับทุกจุดที่ใช้ `log.Println` หรือ `log.Fatal` ให้เปลี่ยนมาใช้ `logger.Log` แทน และหากต้องการแนบ context เพิ่มเติม ให้ใช้ `logger.With(...)`

    ```go
     func main() {
     // ...
     
     // ย้ายมา run server ใน goroutine
     go func() {
       if err := app.Listen(":8090"); err != nil && err != http.ErrServerClosed {
         logger.Log.Fatal(fmt.Sprintf("Error starting server: %v", err))
       }
     }()
     
     // รอสัญญาณการปิด
     stop := make(chan os.Signal, 1)
     signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
     <-stop
     
     logger.Log.Info("Shutting down...")
     
     // หยุดรับ request ใหม่ แล้ว รอให้ request เดิมทำงานเสร็จ ภายใน timeout ที่กำหนด (เช่น 5 วินาที)
     ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
     defer cancel()
     if err := app.ShutdownWithContext(ctx); err != nil {
       logger.Log.Fatal(fmt.Sprintf("Error shutting down server: %v", err))
     }
     
     // Optionally: แล้วค่อยปิด resource อื่นๆ เช่น DB connection, cleanup, etc.
    
     logger.Log.Info("Shutdown complete.")
    }
    ```

- เพิ่ม middleware `RequestLogger` เพื่อ log ข้อมูลการเรียกใช้งาน HTTP เช่น method, path, status code และระยะเวลา พร้อมทั้ง log error และ stack trace ในกรณีที่เกิด panic หรือ unhandled error

    > สร้างไฟล์ `application/middleware/request_logger.go`
    >

    ```go
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
    
    ```

- ใน `main.go` ให้เปลี่ยนมาใช้ middleware `RequestLogger` แทน log เดิม เพื่อให้ทุก request ถูก log ด้วย format เดียวกั

    ```go
    // global middleware
    app.Use(cors.New())                 // CORS ลำดับแรก เพื่อให้ OPTIONS request ผ่านได้เสมอ
    app.Use(requestid.New())            // สร้าง request id ใน request header สำหรับการ debug
    app.Use(recover.New())              // auto-recovers from panic (internal only)
    app.Use(middleware.RequestLogger()) // logs HTTP request
    ```

เมื่อลองรันระบบใหม่ จะเห็นว่า log ทั้งหมดมีรูปแบบที่เป็นมาตรฐาน และมีข้อมูลเพียงพอสำหรับการ debug ได้อย่างมีประสิทธิภาพ

### การจัดการ Configurations

จากโค้ดก่อนหน้านี้ จะเห็นว่ามีค่าหลายจุดที่ถูกเขียนค่าคงที่ไว้ในโค้ดโดยตรง เช่น พอร์ตของ HTTP Server และระยะเวลาสำหรับปิดระบบอย่างปลอดภัย (graceful shutdown) ซึ่งการเขียนค่าดังกล่าวลงไปในโค้ด ทำให้ทุกครั้งที่ต้องการเปลี่ยนค่าเหล่านี้ เช่น เปลี่ยนพอร์ตตามแต่ละ environment จำเป็นต้องแก้ไขโค้ดและทำการ build ใหม่

แนวทางที่เหมาะสมกว่าคือ การแยกค่าคอนฟิกเหล่านี้ออกจากโค้ด และให้สามารถกำหนดค่าผ่าน environment variables ได้ ทำให้ระบบมีความยืดหยุ่นและสะดวกต่อการ deploy ในแต่ละสภาพแวดล้อม

เพื่อให้จัดการ environment variables ได้สะดวกและมีมาตรฐานเดียวกัน จึงได้สร้าง utility ชื่อ `env` สำหรับใช้ดึงค่าจาก environment variables พร้อม fallback เป็นค่า default กรณีที่ไม่มีการกำหนด

- สร้างไฟล์ `util/env/env.go`

    ```go
    package env
    
    import (
      "os"
      "strconv"
      "time"
    )
    
    func Get(key string) string {
      v, ok := os.LookupEnv(key)
      if !ok {
        return ""
      }
      return v
    }
    
    func GetDefault(key string, defaultValue string) string {
      v, ok := os.LookupEnv(key)
      if !ok {
        return defaultValue
      }
      return v
    }
    
    func GetInt(key string) int {
      v, err := strconv.Atoi(Get(key))
      if err != nil {
        return 0
      }
      return v
    }
    
    func GetIntDefault(key string, defaultValue int) int {
      v, err := strconv.Atoi(Get(key))
      if err != nil {
        return defaultValue
      }
      return v
    }
    func GetFloat(key string) float64 {
      v, err := strconv.ParseFloat(Get(key), 64)
      if err != nil {
        return 0.0
      }
      return v
    }
    
    func GetFloatDefault(key string, defaultValue float64) float64 {
      v, err := strconv.ParseFloat(Get(key), 64)
      if err != nil {
        return defaultValue
      }
      return v
    }
    
    func GetBool(key string) bool {
      v := Get(key)
      switch v {
      case "true", "yes":
        return true
      case "false", "no":
        return false
      default:
        return false
      }
    }
    func GetBoolDefault(key string, defaultValue bool) bool {
      v := Get(key)
      switch v {
      case "true", "yes":
        return true
      case "false", "no":
        return false
      default:
        return defaultValue
      }
    }
    
    func GetDuration(key string) time.Duration {
      v := Get(key)
      if len(v) == 0 {
        return 0
      }
      d, err := time.ParseDuration(v)
      if err != nil {
        return 0
      }
      return d
    }
    
    func GetDurationDefault(key string, defaultValue time.Duration) time.Duration {
      v := Get(key)
      if len(v) == 0 {
        return defaultValue
      }
      d, err := time.ParseDuration(v)
      if err != nil {
        return defaultValue
      }
      return d
    }
    ```

- แก้โค้ดในไฟล์ `main.go` ให้ใช้ค่าจาก environment แทน

    ```go
    func main() {
     // ...
     go func() {
       // ถ้าไม่กำหนด env มาให้ default 8090
      if err := app.Listen(fmt.Sprintf(":%d", env.GetIntDefault("HTTP_PORT", 8090))); err != nil && err != http.ErrServerClosed {
       // ...
      }
     }()
     // ...
     
     // ถ้าไม่กำหนด env มาให้ default 5 วินาที
     ctx, cancel := context.WithTimeout(context.Background(), env.GetDurationDefault("GRACEFUL_TIMEOUT", 5*time.Second))
     // ...
    }
    ```

- การรันระบบใหม่พร้อมกำหนดค่าต่างๆ ผ่าน environment variables

    ```bash
    HTTP_PORT=8091 GRACEFUL_TIMEOUT=10s go run main.go
    ```

หลังจากนั้น เพื่อรวมการโหลดค่าคอนฟิกทั้งหมดไว้ในจุดเดียวอย่างเป็นระบบ จึงสร้าง package `config` ขึ้นมา ซึ่งจะดึงค่าจาก environment ผ่าน `env` และมีการตรวจสอบความถูกต้องเบื้องต้น (validation) ก่อนใช้งาน เช่น ตรวจสอบว่า HTTP_PORT ต้องเป็นค่าที่มากกว่า 0 เป็นต้น

- สร้างไฟล์ `config/config.go`

    ```go
    package config
    
    import (
     "errors"
     "go-mma/util/env"
     "time"
    )
    
    var (
     ErrInvalidHTTPPort = errors.New("HTTP_PORT must be a positive integer")
     ErrGracefulTimeout = errors.New("GRACEFUL_TIMEOUT must be a positive duration")
    )
    
    // รวมการโหลดค่าคอนฟิกทั้งหมดไว้ในจุดเดียว
    type Config struct {
     HTTPPort        int
     GracefulTimeout time.Duration
    }
    
    func Load() (*Config, error) {
     config := &Config{
      HTTPPort:        env.GetIntDefault("HTTP_PORT", 8090),
      GracefulTimeout: env.GetDurationDefault("GRACEFUL_TIMEOUT", 5*time.Second),
     }
     err := config.Validate()
     if err != nil {
      return nil, err
     }
     return config, err
    }
    
    func (c *Config) Validate() error {
     if c.HTTPPort <= 0 {
      return ErrInvalidHTTPPort
     }
     if c.GracefulTimeout <= 0 {
      return ErrGracefulTimeout
     }
    
     return nil
    }
    ```

- การเรียกใช้งาน `config` ใน `main.go`

    ```go
    func main() {
      // logger
     config, err := config.Load()
     if err != nil {
      panic(err.Error())
     }
     
      // ...
      
     go func() {
      if err := app.Listen(fmt.Sprintf(":%d", config.HTTPPort)); err != nil && err != http.ErrServerClosed {
        // ...
      }
     }()
     
     // ...
     
     ctx, cancel := context.WithTimeout(context.Background(), config.GracefulTimeout)
     
     // ...
    }
    ```

### ใช้ Makefile จัดการคำสั่ง

เนื่องจากการรันโปรแกรมแบบปกติจำเป็นต้องกำหนด environment variables ทุกครั้ง ซึ่งอาจทำให้ไม่สะดวกในการพิมพ์คำสั่ง จึงแนะนำให้ใช้ Makefile เพื่อช่วยจัดการขั้นตอนนี้ให้ง่ายและสะดวกขึ้น โดยมีขั้นตอนดังนี้:

- สร้างไฟล์ `.env` เพื่อเก็บค่าคอนฟิกที่ใช้ขณะรันโปรแกรม

    ```
    HTTP_PORT=8090
    GRACEFUL_TIMEOUT=5s
    ```

- สร้างไฟล์ `.gitignore` เพื่อไม่ให้ไฟล์ `.env` ถูกนำเข้า git

    ```
    *.env
    ```

    <aside>
    💡

    หากต้องการให้มีตัวอย่าง config สำหรับผู้ใช้งาน ให้คัดลอกไฟล์ `.env` ไปเป็น `.env.example` แทน และอย่าใส่ข้อมูลที่เป็นความลับลงในตัวอย่าง

    </aside>

- สร้างไฟล์ `Makefile` เพื่อกำหนดคำสั่งสำหรับรันโปรแกรม

    ```makefile
    include .env
    export
    
    .PHONY: run
    run:
     go run main.go
    ```

- รันโปรแกรมด้วยคำสั่ง `make run`

    ```bash
    make run
    
        _______ __             
       / ____(_) /_  ___  _____
      / /_  / / __ \/ _ \/ ___/
     / __/ / / /_/ /  __/ /    
    /_/   /_/_.___/\___/_/          v3.0.0-beta.4
    --------------------------------------------------
    INFO Server started on:         http://127.0.0.1:8090 (bound on host 0.0.0.0 and port 8090)
    INFO Application name:          Go MMA v0.0.1
    INFO Total handlers count:      6
    INFO Prefork:                   Disabled
    INFO PID:                       31427
    INFO Total process count:       1
    ```

เมื่อรันคำสั่งนี้ ระบบจะอ่านค่าคอนฟิกจาก `.env` แล้วใช้ในการรันโปรแกรม ทำให้ไม่ต้องพิมพ์ค่าคอนฟิกทุกครั้งที่รัน ช่วยให้สะดวกและลดความผิดพลาดจากการพิมพ์ซ้ำ

### Refactor เพื่อความสะอาดของโค้ด

เมื่อโปรเจกต์เติบโตขึ้น ไฟล์ `main.go` มักจะเริ่มมีขนาดใหญ่และทำหลายหน้าที่เกินไป การแยกส่วนความรับผิดชอบออกเป็นโมดูลต่าง ๆ จะช่วยให้โค้ดอ่านง่าย ดูแลรักษาง่าย และมีโครงสร้างชัดเจนมากขึ้น โดยขั้นตอนมีดังนี้

- ย้ายไฟล์ `main.go` ไปไว้ที่ `cmd/api/main.go` เพื่อแยก entrypoint ออกจาก business logic
- ปรับ `Makefile` ให้ชี้ไปที่ตำแหน่งใหม่ของ `main.go`

    ```makefile
    include .env
    export
    
    .PHONY: run
    run:
     go run cmd/api/main.go
    ```

- สร้างไฟล์ `build/build.go` เพื่อจัดการเรื่อง build version และ build time

    ```go
    package build
    
    var (
     Version = "local-dev"
     Time    = "n/a"
    )
    ```

- สร้างไฟล์ `application/http.go` เพื่อรับผิดชอบเรื่อง HTTP server, middleware และ route registration

    ```go
    package application
    
    import (
     "context"
     "fmt"
     "go-mma/application/middleware"
     "go-mma/build"
     "go-mma/config"
     "go-mma/util/logger"
     "net/http"
     "time"
    
     "github.com/gofiber/fiber/v3"
     "github.com/gofiber/fiber/v3/middleware/cors"
     "github.com/gofiber/fiber/v3/middleware/recover"
     "github.com/gofiber/fiber/v3/middleware/requestid"
    )
    
    type HTTPServer interface {
     Start()
     Shutdown() error
     RegisterRoutes()
    }
    
    type httpServer struct {
     config config.Config
     app    *fiber.App
    }
    
    func newHTTPServer(config config.Config) HTTPServer {
     return &httpServer{
      config: config,
      app:    newFiber(),
     }
    }
    
    func newFiber() *fiber.App {
     app := fiber.New(fiber.Config{
      AppName: fmt.Sprintf("Go MMA version %s", build.Version),
     })
    
     // global middleware
      app.Use(cors.New())                 // CORS ลำดับแรก เพื่อให้ OPTIONS request ผ่านได้เสมอ
     app.Use(requestid.New())            // สร้าง request id ใน request header สำหรับการ debug
     app.Use(recover.New())              // auto-recovers from panic (internal only)
     app.Use(middleware.RequestLogger()) // logs HTTP request
     
     app.Get("/", func(c fiber.Ctx) error {
      return c.JSON(map[string]string{"version": build.Version, "time": build.Time})
     })
    
     return app
    }
    
    func (s *httpServer) Start() {
     go func() {
      logger.Log.Info(fmt.Sprintf("Starting server on port %d", s.config.HTTPPort))
      if err := s.app.Listen(fmt.Sprintf(":%d", s.config.HTTPPort)); err != nil && err != http.ErrServerClosed {
       logger.Log.Fatal(fmt.Sprintf("Error starting server: %v", err))
      }
     }()
    }
    
    func (s *httpServer) Shutdown() error {
     ctx, cancel := context.WithTimeout(context.Background(), s.config.GracefulTimeout)
     defer cancel()
     return s.app.ShutdownWithContext(ctx)
    }
    
    func (s *httpServer) RegisterRoutes() {
     v1 := s.app.Group("/api/v1")
    
     customers := v1.Group("/customers")
     {
      customers.Post("", func(c fiber.Ctx) error {
       time.Sleep(3 * time.Second)
       return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": 1})
      })
     }
    
     orders := v1.Group("/orders")
     {
      orders.Post("", func(c fiber.Ctx) error {
       return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": 1})
      })
    
      orders.Delete("/:orderID", func(c fiber.Ctx) error {
       return c.SendStatus(fiber.StatusNoContent)
      })
     }
    }
    ```

- สร้างไฟล์ `application/application.go` เพื่อจัดการ lifecycle ของแอป (เช่น start/shutdown)

    ```go
    package application
    
    import (
     "go-mma/config"
     "log"
    )
    
    type Application struct {
     config     config.Config
     httpServer HTTPServer
    }
    
    func New(config config.Config) *Application {
     return &Application{
      config:     config,
      httpServer: newHTTPServer(config),
     }
    }
    
    func (app *Application) Run() error {
     app.httpServer.Start()
    
     return nil
    }
    
    func (app *Application) Shutdown() error {
     // Gracefully close fiber server
     logger.Log.Info("Shutting down server")
     if err := app.httpServer.Shutdown(); err != nil {
      logger.Log.Fatal(fmt.Sprintf("Error shutting down server: %v", err))
     }
     logger.Log.Info("Server stopped")
    
     return nil
    }
    
    func (app *Application) RegisterRoutes() {
     app.httpServer.RegisterRoutes()
    }
    ```

- ปรับ `main.go` ให้เหลือเฉพาะการ bootstrap ระบบ เช่น init logger, โหลด config, start app, handle graceful shutdown

    ```go
    package main
    
    import (
     "go-mma/application"
     "go-mma/config"
     "log"
     "os"
     "os/signal"
     "syscall"
    )
    
    func main() {
     // logger
     // config
    
     app := application.New(*config)
     app.RegisterRoutes()
     app.Run()
    
     // Wait for shutdown signal
     // stop
     
     app.Shutdown()
    
     // Optionally: close DB, cleanup, etc.
    
     logger.Log.Info("Shutdown complete.")
    }
    ```

การจัดโครงสร้างแบบนี้ช่วยให้โค้ดใน `main.go` สะอาดขึ้น และแยกความรับผิดชอบได้อย่างเหมาะสมตามหลัก Single Responsibility Principle

### การ Build Web Server

สำหรับการนำแอปไปใช้งานจริง (deploy) ในภาษา Go สามารถ build แอปเป็น binary เพื่อนำไปรันโดยตรง

- ปรับ `Makefile` เพื่อเพิ่มคำสั่ง build

    ```bash
    # ...
    
    ROOT_DIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))
    # ถ้า BUILD_VERSION ไม่ถูกเซ็ตใน .env, ให้ใช้ git tag ล่าสุด (ถ้าไม่มี tag จะ fallback เป็น "unknown")
    BUILD_VERSION := $(or ${BUILD_VERSION}, $(shell git describe --tags --abbrev=0 2>/dev/null || echo "unknown"))
    BUILD_TIME := $(shell date +"%Y-%m-%dT%H:%M:%S%z")
    
    .PHONY: build
    build:
     go build -ldflags \
     "-X 'go-mma/build.Version=${BUILD_VERSION}' \
     -X 'go-mma/build.Time=${BUILD_TIME}'" \
     -o app cmd/api/main.go
    ```

- รันคำสั่ง build พร้อมกำหนด build version

    ```bash
    BUILD_VERSION=0.0.1 make build
    ```

อีกทางเลือกหนึ่งคือการ package แอปเป็น Docker image เพื่อให้รันได้แบบ isolated และพร้อม deploy

- สร้าง `Dockerfile`

    ```docker
    FROM golang:1.24-alpine AS base
    WORKDIR /app
    COPY go.mod go.sum ./
    RUN go mod download
    COPY . .
    
    FROM base AS builder
    ENV GOARCH=amd64
    
    # ตั้งค่า default สำหรับ VERSION
    ARG VERSION=latest
    ENV IMAGE_VERSION=${VERSION}
    RUN echo "Build version: $IMAGE_VERSION"
    RUN go build -ldflags \
     "-X 'go-mma/build.Version=${IMAGE_VERSION}' \
     -X 'go-mma/build.Time=$(date +"%Y-%m-%dT%H:%M:%S%z")'" \
     -o app cmd/api/main.go
    
    FROM alpine:latest
    WORKDIR /root/
    EXPOSE 8090
    ENV TZ=Asia/Bangkok
    RUN apk --no-cache add ca-certificates tzdata
    
    COPY --from=builder /app/app .
    
    CMD ["./app"]
    ```

- สร้าง `.dockerignore` เพื่อ exclude ไฟล์ที่ไม่จำเป็นในการ build

    ```
    .git
    .env
    Dockerfile
    Makefile
    *.md
    ```

- ปรับ `Makefile` เพื่อเพิ่มคำสั่งสร้าง image

    ```bash
    # build:
    
    .PHONY: image
    image:
     docker build \
     -t go-mma:${BUILD_VERSION} \
     --build-arg VERSION=${BUILD_VERSION} \
     .
    ```

- รันคำสั่ง build พร้อมกำหนด build version

    ```bash
    BUILD_VERSION=0.0.1 make image
    ```

- รัน container จาก image

    ```bash
    docker run --name go-mma --env-file .env -p 8090:8090 go-mma:0.0.1
    ```

---

## การแยก logic ออกจาก routing ด้วย Handlers

เพื่อแยก concerns ระหว่าง routing และ business logic สามารถจัดโครงสร้าง handler แยกตาม feature

- `handler/customer.go` สำหรับจัดการลูกค้า
- `handler/order.go` สำหรับจัดการออเดอร์

### Customer Handler

การทำงานของ customer handler

```markdown
สร้างลูกค้าใหม่ (POST /customers)
---------------------------------------
Client ----> Routing: POST /customers {email, credit}
Handler.customer --> Database: ตรวจสอบ email ซ้ำ?
  └─ ซ้ำ --> Handler.customer --> Client: 409 Conflict (email already exists)
  └─ ไม่ซ้ำ:
      Handler.customer --> Database: INSERT INTO customers
      Module.email --> ส่งอีเมลต้อนรับ
      Handler.customer --> Client: 201 Created
```

สร้าง customer handler

- สร้างไฟล์ `handler/customer.go`

    ```go
    package handler
    
    import (
     "fmt"
     "go-mma/util/logger"
     "net/mail"
    
     "github.com/gofiber/fiber/v3"
    )
    
    type CustomerHandler struct {
    }
    
    func NewCustomerHandler() *CustomerHandler {
     return &CustomerHandler{}
    }
    
    func (h *CustomerHandler) CreateCustomer(c fiber.Ctx) error {
     // กำหนด payload structure (DTO: Request)
     type CreateCustomerRequest struct {
      Email  string `json:"email"`
      Credit int    `json:"credit"`
     }
     // แปลง request body -> dto
     var req CreateCustomerRequest
     if err := c.Bind().Body(&req); err != nil {
      // แปลงไม่ได้ให้ส่ง error 400
      return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
     }
    
     logger.Log.Info(fmt.Sprintf("Received customer: %v", req))
    
     // ตรวจสอบ input fields (e.g., value, format, etc.)
     if req.Email == "" {
      return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email is required"})
     }
     if _, err := mail.ParseAddress(req.Email); err != nil {
      return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email is invalid"})
     }
    
     // ตรวจสอบเงื่อนไขของ business rules
     // Rule 1: credit ต้องมากกว่า 0
     if req.Credit <= 0 {
      return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "credit must be greater than 0"})
     }
    
     // TODO: Rule 2: ตรวจสอบ email ต้องไม่ซ้ำ
     // if exists {
     //  return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "email already exists"})
     // }
    
     // TODO: บันทึกลงฐานข้อมูล
     var id int64 // id ในฐานข้อมูล
    
     // กำหนดโครงสร้างข้อมูลที่จะส่งกลับไป (DTO: Response)
     type CreateCustomerResponse struct {
      ID int64 `json:"id"`
     }
     resp := &CreateCustomerResponse{ID: id}
    
     // ตอบกลับด้วย status code 201 (created) และข้อมูลแบบ JSON
     return c.Status(fiber.StatusCreated).JSON(resp)
    }
    ```

- แก้ `application/http.go` เพื่อ bind handler

    ```go
    customers := v1.Group("/customers")
    {
     hdl := handler.NewCustomerHandler()
     customers.Post("", hdl.CreateCustomer)
    }
    ```

### Order Handler

การทำงานของ order handler

```markdown
สั่งออเดอร์ (POST /orders)
-------------------------------
Client ----> Routing: POST /orders {customer_id, order_total}
Handler.order --> Database: ตรวจสอบ customer_id
  └─ ไม่พบ --> Handler.order --> Client: 404 Not Found (customer not found)
  └─ พบ:
      Handler.order --> Database: ตรวจสอบ credit เพียงพอ?
          └─ ไม่พอ --> Monolith.order --> Client: 422 Unprocessable Entity (insufficient credit)
          └─ พอ:
              Handler.order --> Database: INSERT INTO orders, UPDATE credit (หักยอด)
              Module.email --> ส่งอีเมลยืนยันออเดอร์
              Handler.order --> Client: 201 Created

ยกเลิกออเดอร์ (DELETE /orders/:orderID)
---------------------------------------------
Client ----> Routing: DELETE /orders/:orderID
Handler.order --> Database: ตรวจสอบ orderID
  └─ ไม่พบ --> Handler.order --> Client: 404 Not Found (order not found)
  └─ พบ:
      Handler.order --> Database: DELETE order, UPDATE credit (คืนยอด)
      Handler.order --> Client: 204 No Content
```

สร้าง order handler

- ใน `handler/order.go`

    ```go
    package handler
    
    import (
     "fmt"
     "go-mma/util/logger"
     "strconv"
    
     "github.com/gofiber/fiber/v3"
    )
    
    type OrderHandler struct {
    }
    
    func NewOrderHandler() *OrderHandler {
     return &OrderHandler{}
    }
    
    func (h *OrderHandler) CreateOrder(c fiber.Ctx) error {
     // กำหนด payload structure (DTO: Request)
     type CreateOrderRequest struct {
      CustomerID string `json:"customer_id"`
      OrderTotal int    `json:"order_total"`
     }
     // แปลง request body -> dto
     var req CreateOrderRequest
     if err := c.Bind().Body(&req); err != nil {
      // แปลงไม่ได้ให้ส่ง error 400
      return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
     }
    
     logger.Log.Info(fmt.Sprintf("Received Order: %v", req))
    
     // ตรวจสอบ input fields (e.g., value, format, etc.)
     if req.CustomerID == "" {
      return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "customer_id is required"})
     }
    
     // ตรวจสอบ business rules (e.g., order_total must be greater than 0)
     if req.OrderTotal <= 0 {
      return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "order_total must be greater than 0"})
     }
    
     // TODO: ตรวจสอบว่ามี customer id อยู่ในฐานข้อมูล หรือไม่
     // customer := getCustomer(order.CustomerID)
     // if customer == nil {
     //  return return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "the customer with given id was not found"})
     // }
    
     // TODO: ตรวจสอบ credit เพียงพอ หรือไม่
     // if credit < payload.OrderTotal {
     //  return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "insufficient credit"})
     // }
    
     // TODO: หักยอด credit ของ customer
    
     // TODO: update customer's credit balance ในฐานข้อมูล
    
     // TODO: บันทึกรายการออเดอร์ใหม่ลงในฐานข้อมูล
     var id int64
    
     // กำหนดโครงสร้างข้อมูลที่จะส่งกลับไป (DTO: Response)
     type CreateOrderResponse struct {
      ID int6464 `json:"id"`
     }
     resp := &CreateOrderResponse{ID: id}
    
     // ตอบกลับด้วย status code 201 (created) และข้อมูลแบบ JSON
     return c.Status(fiber.StatusCreated).JSON(resp)
    }
    
    func (h *OrderHandler) CancelOrder(c fiber.Ctx) error {
     // ตรวจสอบรูปแบบ orderID
     orderID, err := strconv.Atoi(c.Params("orderID"))
     if err != nil {
      // ถ้าไม่ถูกต้อง error 400
      return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid order id"})
     }
    
     logger.Log.Info(fmt.Sprintf("Cancelling order: %v", orderID))
    
     // TODO: ตรวจสอบ orderID ในฐานข้อมูล
     // order := getOrder(orderID)
     // if order == nil {
     //  return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "the order with given id was not found"})
     // }
    
     // TODO: ค้นหา customer จาก customerID ของ order
     // customer := getCustomer(order.CustomerID)
     // if customer == nil {
     //  return return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "the customer with given id was not found"})
     // }
    
     // TODO: คืนยอด credit กลับให้ customer
     // creditLimit += CreateOrderRequest.OrderTotal
    
     // TODO: update customer's credit balance ในฐานข้อมูล
    
     // TODO: update สถานะของ order ในฐานข้อมูลให้เป็น cancel
    
     // ตอบกลับด้วย status code 204 (no content)
     return c.SendStatus(fiber.StatusNoContent)
    }
    ```

- แก้ `application/http.go` เพื่อ bind handler

    ```go
    orders := v1.Group("/orders")
    {
     hdl := handler.NewOrderHandler()
     orders.Post("", hdl.CreateOrder)
     orders.Delete("/:orderID", hdl.CancelOrder)
    }
    ```

## เชื่อมต่อฐานข้อมูลอย่างปลอดภัยและยืดหยุ่น

จากโค้ดก่อนหน้า เรายังไม่สามารถดำเนินการได้สมบูรณ์ เพราะจำเป็นต้องเชื่อมต่อและบันทึกข้อมูลลงฐานข้อมูลก่อน ในส่วนนี้จะประกอบด้วยขั้นตอนสำคัญดังนี้

- ติดตั้ง PostgreSQL ด้วย Docker
- ออกแบบ Schema ให้เหมาะสม
- จัดการ Schema ด้วย Migration Tool
- ตั้งค่า Database Connection อย่างปลอดภัย
- สร้างฟังก์ชันสำหรับ Gen ID
- เชื่อมต่อ Database ด้วย Dependency Injection
- เพิ่มความสามารถในการสร้างข้อมูลลูกค้า (Insert Customer)

### ติดตั้ง PostgreSQL ด้วย Docker

ในบทความนี้จะใช้ PostgreSQL และติดตั้งด้วย docker

- สร้างไฟล์ `docker-compose.yml` สำหรับสร้าง postgres service

    ```yaml
    services:
      db:
        image: postgres:17-alpine
        container_name: go-mma-db
        volumes:
          - pg_data:/var/lib/postgresql/data
    
    volumes:
      pg_data:
    ```

- สร้างไฟล์ `docker-compose.dev.yml` สำหรับกำหนด environtment ของ `dev`

    ```yaml
    services:
      db:
        environment:
          POSTGRES_DB: go-mma-db
          POSTGRES_USER: postgres
          POSTGRES_PASSWORD: postgres
        ports:
          - 5433:5432
    ```

    <aside>
    💡

    สามารถเพิ่มสำหรับ environtment ของ `test` และ `production` ได้ภายหลัง

    </aside>

- รัน PostgreSQL Server ด้วย `Makefile` โดยเพิ่มคำสั่ง ดังนี้

    ```bash
    # ...
    
    .PHONY: devup
    devup:
     docker compose -f docker-compose.yml -f docker-compose.dev.yml up -d
    
    .PHONY: devdown
    devdown:
     docker compose -f docker-compose.yml -f docker-compose.dev.yml down
    ```

- รันคำสั่ง `make devup`

### ออกแบบ Schema ให้เหมาะสม

```sql
+------------------+           +------------------+
|   customers      |           |     orders       |
+------------------+           +------------------+
| id (PK)          |<--------+ | id (PK)          |
| email (UNIQUE)   |         | | customer_id (FK) |
| credit           |         | | order_total      |
| created_at       |         | | created_at       |
| updated_at       |         | | canceled_at       |
+------------------+         | +------------------+
                             |
      [1] ---------------- [*]
     1 customer      →   many orders
```

### จัดการ Schema ด้วย Migration Tool

**Database Migration** คือกระบวนการจัดการและติดตามการเปลี่ยนแปลงโครงสร้างฐานข้อมูล (เช่น ตาราง, คอลัมน์, ดัชนี ฯลฯ) ด้วยสคริปต์ที่สามารถรันซ้ำได้อย่างปลอดภัยในทุกสภาพแวดล้อม (dev, staging, production) เพื่อให้ทีมพัฒนาและระบบสามารถทำงานร่วมกันได้อย่างราบรื่น

- เพิ่ม Environment Variable สำหรับการเชื่อมต่อ

    เพิ่มตัวแปร `DB_DSN` ในไฟล์ `.env` เพื่อเก็บ connection string ของ PostgreSQL

    ```
    DB_DSN=postgres://postgres:postgres@localhost:5433/go-mma-db?sslmode=disable
    ```

- เพิ่มคำสั่ง Migration ใน Makefile
แก้ไข `Makefile` เพื่อเพิ่มคำสั่งสำหรับสร้างและรัน migration ผ่าน Docker

    ```makefile
    # ...
    
    .PHONY: mgc
    # Example: make mgc filename=create_customer
    mgc:
     docker run --rm -v $(ROOT_DIR)migrations:/migrations migrate/migrate -verbose create -ext sql -dir /migrations $(filename)
    
    .PHONY: mgu
    mgu:
     docker run --rm --network host -v $(ROOT_DIR)migrations:/migrations migrate/migrate -verbose -path=/migrations/ -database "$(DB_DSN)" up
    
    .PHONY: mgd
    mgd:
     docker run --rm --network host -v $(ROOT_DIR)migrations:/migrations migrate/migrate -verbose -path=/migrations/ -database $(DB_DSN) down 1
    ```

    <aside>
    💡

    หากใช้ Docker Desktop ต้องเปิดใช้งาน host networking โดยไปที่ `Settings → Resources → Network` และเลือก "Enable host networking" จากนั้นกด Apply & Restart

    </aside>

- สร้าง Migration สำหรับ Customer
ใช้คำสั่งด้านล่างเพื่อสร้าง migration สำหรับตาราง `customers`

    ```bash
    make mgc filename=create_customer
    ```

    ระบบจะสร้างไฟล์ 2 ไฟล์

    ```bash
    ./migrations/20250529103238_create_customer.up.sql
    ./migrations/20250529103238_create_customer.down.sql
    ```

    แก้ไขไฟล์ `create_customer.up.sql`

    ```sql
    CREATE TABLE public.customers (
     id BIGINT NOT NULL,
     email text NOT NULL,
     credit int4 NOT NULL,
     created_at timestamp DEFAULT CURRENT_TIMESTAMP NULL,
     updated_at timestamp DEFAULT CURRENT_TIMESTAMP NULL,
     CONSTRAINT customers_pkey PRIMARY KEY (id),
     CONSTRAINT customers_unique UNIQUE (email)
    );
    ```

    และไฟล์ `create_customer.down.sql`

    ```sql
    drop table public.customers;
    ```

- สร้าง Migration สำหรับ Order
ใช้คำสั่ง

    ```bash
    make mgc filename=create_order
    ```

    จะได้ไฟล์

    ```bash
    ./migrations/20250529103715_create_order.up.sql
    ./migrations/20250529103715_create_order.down.sql
    ```

    แก้ไขไฟล์ `create_order.up.sql`

    ```sql
    CREATE TABLE public.orders (
     id BIGINT NOT NULL,
     customer_id BIGINT NOT NULL,
     order_total int4 NOT NULL,
     created_at timestamp DEFAULT CURRENT_TIMESTAMP NULL,
     canceled_at timestamp NULL,
     CONSTRAINT orders_pkey PRIMARY KEY (id),
     CONSTRAINT fk_customer FOREIGN KEY (customer_id) REFERENCES public.customers(id)
    );
    ```

    และ `create_order.down.sql`

    ```sql
    drop table public.orders;
    ```

- รัน Migration เพื่อสร้างตาราง
ใช้คำสั่ง

    ```bash
    make mgu
    ```

    ระบบจะแสดงผลลัพธ์การรัน migration เช่น

    ```bash
    2025/05/29 10:39:20 Start buffering 20250529103238/u create_customer
    2025/05/29 10:39:20 Start buffering 20250529103715/u create_order
    2025/05/29 10:39:20 Read and execute 20250529103238/u create_customer
    2025/05/29 10:39:20 Finished 20250529103238/u create_customer (read 906.667µs, ran 2.125583ms)
    2025/05/29 10:39:20 Read and execute 20250529103715/u create_order
    2025/05/29 10:39:20 Finished 20250529103715/u create_order (read 3.458625ms, ran 1.860583ms)
    2025/05/29 10:39:20 Finished after 7.190625ms
    2025/05/29 10:39:20 Closing source and database
    ```

### ตั้งค่า Database Connection อย่างปลอดภัย

ก่อนจะเริ่มพัฒนา API จำเป็นต้องตั้งค่าการเชื่อมต่อกับฐานข้อมูลให้เรียบร้อย เพื่อให้ระบบสามารถใช้งานฐานข้อมูล PostgreSQL ได้อย่างมีประสิทธิภาพและปลอดภัย

- เพิ่มคอนฟิก `DB_DSN` ในไฟล์ `config/config.go` เพื่อระบุข้อมูลการเชื่อมต่อฐานข้อมูล

    ```go
    package config
    
    import (
     "errors"
     "go-mma/util/env"
     "time"
    )
    
    var (
     ErrInvalidHTTPPort = errors.New("HTTP_PORT must be a positive integer")
     ErrGracefulTimeout = errors.New("GRACEFUL_TIMEOUT must be a positive duration")
     ErrDSN             = errors.New("DB_DSN must be set") // เพิ่ม
    )
    
    type Config struct {
     HTTPPort        int
     GracefulTimeout time.Duration
     DSN             string  // เพิ่ม
    }
    
    func Load() (*Config, error) {
     config := &Config{
      HTTPPort:        env.GetIntDefault("HTTP_PORT", 8090),
      GracefulTimeout: env.GetDurationDefault("GRACEFUL_TIMEOUT", 5*time.Second),
      DSN:             env.Get("DB_DSN"),  // เพิ่ม
     }
     err := config.Validate()
     if err != nil {
      return nil, err
     }
     return config, err
    }
    
    func (c *Config) Validate() error {
     if c.HTTPPort <= 0 {
      return ErrInvalidHTTPPort
     }
     if c.GracefulTimeout <= 0 {
      return ErrGracefulTimeout
     }
     // เพิ่ม
     if len(c.DSN) == 0 {
      return ErrDSN
     }
    
     return nil
    }
    ```

- สร้างไฟล์ `util/storage/sqldb/sqldb.go` สำหรับจัดการการเชื่อมต่อกับฐานข้อมูล PostgreSQL โดยใช้ `sqlx`

    ```go
    package sqldb
    
    import (
     "github.com/jmoiron/sqlx"
     _ "github.com/lib/pq"
    )
    
    type closeDB func() error
    
    type DBContext interface {
     DB() *sqlx.DB
    }
    
    type dbContext struct {
     db *sqlx.DB
    }
    
    var _ DBContext = (*dbContext)(nil)
    
    func NewDBContext(dsn string) (DBContext, closeDB, error) {
     // this Pings the database trying to connect
     db, err := sqlx.Connect("postgres", dsn)
     if err != nil {
      return nil, nil, err
     }
     return &dbContext{db: db},
      func() error {
       return db.Close()
      },
      nil
    }
    
    func (c *dbContext) DB() *sqlx.DB {
     return c.db
    }
    ```

- รันคำสั่ง `go mod tidy` เพื่อดึง dependencies ที่จำเป็น
- ปรับปรุงไฟล์ `application/application.go` เพื่อเก็บอินสแตนซ์ของ database connection

    ```go
    type Application struct {
     config     config.Config
     httpServer HTTPServer
     dbCtx      sqldb.DBContext
    }
    
    func New(config config.Config, dbCtx sqldb.DBContext) *Application {
     return &Application{
      config:     config,
      httpServer: newHTTPServer(config),
      dbCtx:         dbCtx,
     }
    }
    ```

- เชื่อมต่อฐานข้อมูลใน `cmd/api/main.go` และส่งผ่านเข้าไปใน Application

    ```go
    func main() {
     // config
     
     dbCtx, closeDB, err := sqldb.NewDBContext(config.DSN)
     if err != nil {
      panic(err.Error())
     }
     defer func() { // ใช่ท่า IIFE เพราะต้องการแสดง error ถ้าปิดไม่ได้
      if err := closeDB(); err != nil {
       logger.Log.Error(fmt.Sprintf("Error closing database: %v", err))
      }
     }()
    
     app := application.New(*config, dbCtx)
     // ...
    }
    ```

### สร้างฟังก์ชันสำหรับ Gen ID

เพื่อให้สามารถสร้าง ID สำหรับลูกค้า (หรือ entity อื่น ๆ) ได้อย่างยืดหยุ่น จึงต้องแยกฟังก์ชันสำหรับการ generate ID ออกเป็น utility

- สร้างไฟล์ `util/idgen/idgen.go` เพื่อรวมฟังก์ชันสร้าง ID หลายรูปแบบ เช่น
  - `GenerateTimeRandomID`: สร้างเลขสุ่มแบบ `int64` โดยอิงจาก timestamp (ใช้ตัวนี้)
  - `GenerateTimeID`: สร้างเลข `int` โดยใช้เวลาปัจจุบัน
  - `GenerateTimeRandomIDBase36`: แปลง ID เป็น string แบบ base36
  - `GenerateUUIDLikeID`: สร้าง string ที่มีรูปแบบคล้าย UUID

    ```go
    package idgen
    
    import (
     "fmt"
     "math/rand"
     "strconv"
     "strings"
     "time"
    )
    
    // GenerateTimeRandomID สร้าง ID แบบ int64
    func GenerateTimeRandomID() int64 {
     timestamp := time.Now().UnixNano() >> 32
     randomPart := rand.Int63() & 0xFFFFFFFF
     return (timestamp << 32) | randomPart
    }
    
    // GenerateTimeID สร้าง ID แบบ int (ใช้ timestamp เป็นหลัก)
    func GenerateTimeID() int {
     // ใช้ timestamp Unix วินาที (int64) แปลงเป็น int (int32/64 ขึ้นกับระบบ)
     return int(time.Now().Unix())
    }
    
    // GenerateTimeRandomIDBase36 คืนค่า ID เป็น base36 string
    func GenerateTimeRandomIDBase36() string {
     id := GenerateTimeRandomID()
     return strconv.FormatInt(id, 36) // แปลงเลขฐาน 10 -> 36
    }
    
    // GenerateUUIDLikeID คืนค่าเป็น string แบบ UUID-like (แต่ไม่ใช่ UUID จริง)
    func GenerateUUIDLikeID() string {
     id := GenerateTimeRandomID()
    
     // แปลง int64 เป็น hex string ยาว 16 ตัว (64 bit)
     hex := fmt.Sprintf("%016x", uint64(id))
    
     // สร้าง UUID-like string รูปแบบ 8-4-4-4-12
     // แต่มีแค่ 16 hex chars แบ่งคร่าวๆ: 8-4-4 (เหลือไม่พอจริงๆ)
     // ดังนั้นเราจะเติม random เพิ่มเพื่อครบ 32 hex (128 bit) เหมือน UUID
    
     randPart := fmt.Sprintf("%016x", rand.Uint64())
    
     uuidLike := strings.Join([]string{
      hex[0:8],
      hex[8:12],
      hex[12:16],
      randPart[0:4],
      randPart[4:16],
     }, "-")
    
     return uuidLike
    }
    
    // ก่อน Go 1.20 ต้องเรียก เพื่อให้ได้เลขสุ่มไม่ซ้ำ
    // func init() {
    //  rand.Seed(time.Now().UnixNano())
    // }
    ```

### เชื่อมต่อ Database ด้วย Dependency Injection

เพื่อให้ระบบมีความยืดหยุ่นและทดสอบง่ายขึ้น เราใช้หลักการ Dependency Injection (DI) ในการส่ง database connection เข้าไปยัง handler

- แก้ไขไฟล์ `handlers/customer.go` โดยปรับโครงสร้าง `CustomerHandler` เพื่อรับ `DBContext` ผ่าน constructor

    ```go
    type CustomerHandler struct {
     dbCtx sqldb.DBContext
    }
    
    func NewCustomerHandler(db sqldb.DBContext) *CustomerHandler {
     return &CustomerHandler{dbCtx: db}
    }
    ```

- แก้ไข `application/http.go` เพื่อส่ง `DBContext` ไปยัง handler

    ```go
    type HTTPServer interface {
     Start()
     Shutdown() error
     RegisterRoutes(dbCtx sqldb.DBContext)
    }
    
    // ...
    
    func (s *httpServer) RegisterRoutes(dbCtx sqldb.DBContext) {
     v1 := s.app.Group("/api/v1")
    
     customers := v1.Group("/customers")
     {
      hdlr := handlers.NewCustomerHandler(dbCtx)
      customers.Post("", hdlr.CreateCustomer)
     }
    
     // orders
    }
    ```

- เพิ่มเมธอด `RegisterRoutes()` ใน `application/application.go` เพื่อจัดการเส้นทาง HTTP โดยส่ง db context เข้าไป

    ```go
    func (app *Application) RegisterRoutes() {
     app.httpServer.RegisterRoutes(app.db)
    }
    ```

### เพิ่มความสามารถในการสร้างข้อมูลลูกค้า (Insert Customer)

หลังจากเชื่อมต่อฐานข้อมูลเรียบร้อยแล้ว ขั้นตอนต่อไปคือการเพิ่มความสามารถในการสร้างข้อมูลลูกค้าผ่าน HTTP POST

- แก้ไขไฟล์ `handlers/customer.go` เพื่อบันทึกข้อมูลลงตาราง customers

    ```go
    func (h *CustomerHandler) CreateCustomer(c fiber.Ctx) error {
     // กำหนด payload structure (DTO: Request)
     // แปลง request body -> dto
     // ตรวจสอบ input fields (e.g., value, format, etc.)
     // ตรวจสอบเงื่อนไขของ business rules
     // Rule 1: credit ต้องมากกว่า 0
    
     // Rule 2: ตรวจสอบ email ต้องไม่ซ้ำ
     query := "SELECT 1 FROM public.customers where email = $1 LIMIT 1"
     ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
     defer cancel()
    
     var exists int
     if err := h.dbCtx.DB().QueryRowxContext(ctx, query, req.Email).Scan(&exists); err != nil {
      if err != sql.ErrNoRows {
       logger.Log.Error(fmt.Sprintf("error checking email: %v", err))
       return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "an error occurred while checking email"})
      }
     }
     if exists == 1 {
      return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "email already exists"})
     }
    
     // บันทึกลงฐานข้อมูล
     var id int64
     query = "INSERT INTO customers (id, email, credit) VALUES ($1, $2, $3) RETURNING id"
     ctxIns, cancelIns := context.WithTimeout(c.Context(), 10*time.Second)
     defer cancelIns()
     if err := h.dbCtx.DB().QueryRowxContext(ctxIns, query, idgen.GenerateTimeRandomID(), req.Email, req.Credit).Scan(&id); err != nil {
      logger.Log.Error(fmt.Sprintf("error insert customer: %v", err))
      return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "an error occurred while inserting customer"})
     }
    
     // กำหนดโครงสร้างข้อมูลที่จะส่งกลับไป (DTO: Response)
     // ตอบกลับด้วย status code 201 (created) และข้อมูลแบบ JSON
    }
    
    ```

## เริ่มต้นด้วย Layered Architecture ที่เข้าใจง่าย

ในการพัฒนา API หากเราเขียนโค้ดทั้งหมดไว้รวมกันในที่เดียว เช่น การเชื่อมต่อฐานข้อมูล การตรวจสอบข้อมูล และการจัดการ HTTP request ทั้งหมดไว้ใน handler เดียว จะทำให้ระบบ scale ได้ยาก ทดสอบลำบาก และเสี่ยงต่อ bug ที่ไม่คาดคิด

**ทางออกคือการใช้ Layered Architecture** ซึ่งช่วยแยกความรับผิดชอบของแต่ละส่วนออกจากกันอย่างชัดเจน ทำให้โค้ดมีโครงสร้างชัดเจน และดูแลต่อยอดในอนาคตได้ง่าย

เนื้อหาในส่วนนี้ประกอบด้วย

- Layered Architecture คืออะไร
- Repository Layer – ตัวกลางเชื่อมต่อกับฐานข้อมูล
- Service Layer – ประมวลผลลอจิกทางธุรกิจ
- Presentation Layer – จัดการ HTTP Request/Response
- ประกอบร่างระบบให้สมบูรณ์

### Layered Architecture คืออะไร

**Layered Architecture** คือรูปแบบการออกแบบซอฟต์แวร์ที่แยกโค้ดออกเป็น “ชั้น” หรือ “เลเยอร์” ตามหน้าที่ของมัน เช่น การเข้าถึงฐานข้อมูล, business logic และการจัดการ HTTP Request/Response โดยแต่ละเลเยอร์จะรับผิดชอบหน้าที่ของตัวเองอย่างชัดเจน

**โครงสร้าง Layered Architecture โดยทั่วไป**

```
Client/UI Layer        ← ผู้ใช้โต้ตอบกับระบบ
↓
Presentation Layer     ← Controller, API (จัดการคำขอจากผู้ใช้)
↓ ← DTO
Service Layer          ← Business Logic (กฎทางธุรกิจ)
↓ ← Model
Repository/Data Layer  ← จัดการกับฐานข้อมูล, external APIs
↓
Database/External APIs
```

**เมื่อนำมา implement ในโค้ดของเรา**

```
project/
│
├── handler/          ← Presentation Layer (HTTP handlers)
├── dto/              ← รับ/ส่งข้อมูลระหว่าง handler ↔ service
├── service/          ← Business Logic (core logic)
├── model/            ← ใช้จัดการข้อมูลภายในระบบ service ↔ repository
├── repository/       ← Data Access (DB queries)
└── main.go           ← Entry point (setup DI, server, etc)
```

### Repository Layer – ตัวกลางเชื่อมต่อกับฐานข้อมูล

**Repository Layer** มีหน้าที่หลักในการ ติดต่อกับฐานข้อมูลหรือแหล่งเก็บข้อมูล โดยรับคำสั่งจาก Service Layer แล้วทำหน้าที่ CRUD (Create, Read, Update, Delete) โดยไม่ให้เลเยอร์อื่นรู้ว่าข้อมูลมาจากที่ใด (Postgres, MySQL, Redis หรือแม้แต่ API)

- เริ่มจากการสร้าง Model โดยให้ไฟล์ชื่อ `model/customer.go` เพื่อใช้แทนโครงสร้างข้อมูลที่ตรงกับฐานข้อมูล

    ```go
    package model
    
    import (
     "go-mma/util/idgen"
     "time"
    )
    
    type Customer struct {
     ID          int64     `db:"id"` // tag db ใช้สำหรับ StructScan() ของ sqlx
     Email       string    `db:"email"`
     Credit      int       `db:"credit"`
     CreatedAt   time.Time `db:"created_at"`
     UpdatedAt   time.Time `db:"updated_at"`
    }
    
    func NewCustomer(email string, credit int) *Customer {
     return &Customer{
      ID:     idgen.GenerateTimeRandomID(),
      Email:  email,
      Credit: credit,
     }
    }
    ```

    <aside>
    💡

    ในบทความนี้จะสร้างเป็น [Rich Model](https://somprasongd.work/blog/architecture/anemic-vs-rich-model-ddd)

    </aside>

- สร้าง Repository สำหรับ บันทึกลูกค้าใหม่ลงฐานข้อมูล ไว้ที่ไฟล์ `repository/customer.go`

    ```go
    package repository
    
    import (
     "context"
     "database/sql"
     "fmt"
     "go-mma/model"
     "go-mma/util/storage/sqldb"
     "time"
    )
    
    type CustomerRepository struct {
     dbCtx sqldb.DBContext // ใช้งาน database ผ่าน DBContext interface
    }
    
    func NewCustomerRepository(dbCtx sqldb.DBContext) *CustomerRepository {
     return &CustomerRepository{
      dbCtx: dbCtx, // inject DBContext instance into CustomerRepository
     }
    }
    
    func (r *CustomerRepository) Create(ctx context.Context, customer *model.Customer) error {
     query := `
     INSERT INTO public.customers (id, email, credit)
     VALUES ($1, $2, $3)
     RETURNING *
     `
    
     // กำหนด timeout ของ query
     ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
     defer cancel()
    
     err := r.dbCtx.DB().
      QueryRowxContext(ctx, query, customer.ID, customer.Email, customer.Credit).
      StructScan(customer) // นำค่า created_at, updated_at ใส่ใน struct customer
     if err != nil {
      return fmt.Errorf("an error occurred while inserting customer: %w", err)
     }
     return nil
    }
    
    func (r *CustomerRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
     query := `SELECT 1 FROM public.customers WHERE email = $1 LIMIT 1`
    
     ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
     defer cancel()
    
     var exists int
     err := r.dbCtx.DB().
      QueryRowxContext(ctx, query, email).
      Scan(&exists)
     if err != nil {
      if err == sql.ErrNoRows { // หาไม่เจอแสดงว่ายังไม่มี email ในระบบแล้ว
       return false, nil
      }
      return false, fmt.Errorf("an error occurred while checking email: %w", err)
     }
     return true, nil // ถ้าไม่ error แสดงว่ามี email ในระบบแล้ว
    }
    ```

    <aside>
    💡

    การใช้ `context.WithTimeout` เป็นแนวปฏิบัติมาตรฐานสำหรับระบบงานที่เกี่ยวข้องกับฐานข้อมูลหรือ external service

    </aside>

### Service Layer – ประมวลผลลอจิกทางธุรกิจ

**Service Layer** คือเลเยอร์ที่อยู่ตรงกลางระหว่าง Controller (หรือ Handler) กับ Repository

หน้าที่หลักของ Service Layer คือ รวมและควบคุม Business Logic ของแอปพลิเคชันไว้ในที่เดียว ดังนี้

- **รับ DTO**: รับ DTO จาก Handler เข้ามาเพื่อประมวลผล
- **ตรวจสอบ**: ตรวจสอบความถูกต้องตาม business logic rule
- **แปลงข้อมูล**: แปลง DTO → Model
- **เรียก Repository**: เพื่อทำ CRUD (Create, Read, Update, Delete) ตามเงื่อนไข
- **ส่งผลลัพธ์**: รับผลลัพธ์จาก Repository แล้วแปลงกลับเป็น DTO Response
- **จัดการ error**: แสดง error log แล้วส่งกลับไปให้ Controller (หรือ Handler) จัดการต่อ

ขั้นตอนการสร้าง Service Layer

- สร้าง DTO (Data Transfer Object) ไว้เป็นตัวกลางสำหรับรับ–ส่งข้อมูล ระหว่างชั้น Handler ↔ Service

    สร้างไฟล์ `dto/customer_request.go`

    ```go
    package dto
    
    type CreateCustomerRequest struct {
     Email  string `json:"email"`
     Credit int    `json:"credit"`
    }
    ```

    สร้างไฟล์ `dto/customer_response.go`

    ```go
    package dto
    
    type CreateCustomerResponse struct {
     ID int64 `json:"id"`
    }
    
    func NewCreateCustomerResponse(id int64) *CreateCustomerResponse {
     return &CreateCustomerResponse{ID: id}
    }
    ```

- สร้าง Service สำหรับควบคุม Business Logic ทั้งหมดในการสร้างลูกค้าใหม่

    สร้างไฟล์ `service/customer.go`

    ```go
    package service
    
    import (
     "context"
     "errors"
     "go-mma/dto"
     "go-mma/model"
     "go-mma/repository"
     "go-mma/util/logger"
    )
    
    var (
     ErrCreditValue = errors.New("credit must be greater than 0")
     ErrEmailExists = errors.New("email already exists")
    )
    
    type CustomerService struct {
     custRepo *repository.CustomerRepository
    }
    
    func NewCustomerService(custRepo *repository.CustomerRepository) *CustomerService {
     return &CustomerService{
      custRepo: custRepo,
     }
    }
    
    func (s *CustomerService) CreateCustomer(ctx context.Context, req *dto.CreateCustomerRequest) (*dto.CreateCustomerResponse, error) {
     // ตรวจสอบเงื่อนไขของ business rules
     // Rule 1: credit ต้องมากกว่า 0
     if req.Credit <= 0 {
      return nil, ErrCreditValue
     }
    
     // Rule 2: ตรวจสอบ email ต้องไม่ซ้ำ
     exists, err := s.custRepo.ExistsByEmail(ctx, req.Email)
     if err != nil {
      // error logging
      logger.Log.Error(err.Error())
      return nil, err
     }
    
     if exists {
      return nil, ErrEmailExists
     }
    
     // แปลง DTO → Model
     customer := model.NewCustomer(req.Email, req.Credit)
    
     // ส่งไปที่ Repository Layer เพื่อบันทึกข้อมูลลงฐานข้อมูล
     if err := s.custRepo.Create(ctx, customer); err != nil {
      // error logging
      logger.Log.Error(err.Error())
      return nil, err
     }
    
     // สร้าง DTO Response
     resp := dto.NewCreateCustomerResponse(customer.ID)
    
     return resp, nil
    }
    ```

### Presentation Layer – จัดการ HTTP Request/Response

**Presentation Layer (HTTP Handlers)** คือชั้นที่อยู่บนสุดของระบบในสถาปัตยกรรมแบบ Layered Architecture โดยทำหน้าที่เป็น “จุดเชื่อมต่อระหว่างผู้ใช้ (Client) กับระบบ” ผ่านโปรโตคอล เช่น HTTP หรือ WebSocket

หน้าที่หลักของ Presentation Layer (หรือ HTTP Handler)

- รับคำขอ: รับ HTTP Request จาก Client
- แปลงข้อมูล: แปลง JSON → DTO (ใช้ `BodyParser`, `Bind`, หรือ Unmarshal)
- ตรวจสอบ: ตรวจสอบความถูกต้องของข้อมูล (validation)
- เรียก Service: ส่ง DTO เข้า Service Layer เพื่อประมวลผล
- ส่งผลลัพธ์: รับผลลัพธ์จาก Service แล้วแปลงกลับเป็น JSON Response
- จัดการ error: แปลง error จากชั้นล่างเป็น HTTP response code เช่น 400, 500

ขั้นตอนการสร้าง Presentation Layer (HTTP Handlers)

- แก้ไขไฟล์ `dto/customer_request.go` เพื่อเพิ่ม validation เช่น credit ต้อง ≥ 0 ก่อนส่งให้ Service

    ```go
    package dto
    
    import (
     "errors"
     "net/mail"
    )
    
    // struct
    
    func (r *CreateCustomerRequest) Validate() error {
     var errs error
     if r.Email == "" {
      errs = errors.Join(errs, errors.New("email is required"))
     }
     if _, err := mail.ParseAddress(r.Email); err != nil {
      errs = errors.Join(errs, errors.New("email is invalid"))
     }
     if r.Credit <= 0 {
      errs = errors.Join(errs, errors.New("credit must be greater than 0"))
     }
     return errs
    }
    ```

- แก้ไขไฟล์ `handler/customer.go` เพื่อให้ทำงานตามหน้าที่ของ Presentation Layer

    ```go
    package handler
    
    import (
     "fmt"
     "go-mma/dto"
     "go-mma/service"
     "go-mma/util/logger"
    
     "github.com/gofiber/fiber/v3"
    )
    
    type CustomerHandler struct {
     custService *service.CustomerService
    }
    
    func NewCustomerHandler(custService *service.CustomerService) *CustomerHandler {
     return &CustomerHandler{
      custService: custService,
     }
    }
    
    func (h *CustomerHandler) CreateCustomer(c fiber.Ctx) error {
     // แปลง request body -> dto
     var req dto.CreateCustomerRequest
     if err := c.Bind().Body(&req); err != nil {
      // แปลงไม่ได้ให้ส่ง error 400
      return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
     }
    
     logger.Log.Info(fmt.Sprintf("Received customer: %v", req))
    
     // ตรวจสอบ input fields (e.g., value, format, etc.)
     if err := req.Validate(); err != nil {
      return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
     }
    
     // ส่งไปที่ Service Layer
     resp, err := h.custService.CreateCustomer(c.Context(), &req)
    
     // จัดการ error จาก Service Layer หากเกิดขึ้น
     if err != nil {
      return c.Status(500).JSON(fiber.Map{"error": err.Error()})
     }
    
     // ตอบกลับด้วย status code 201 (created) และข้อมูลแบบ JSON
     return c.Status(fiber.StatusCreated).JSON(resp)
    }
    ```

### ประกอบร่างระบบให้สมบูรณ์

เมื่อเราแยกเลเยอร์ต่าง ๆ ออกมาแล้ว ขั้นตอนสุดท้ายคือการ “ประกอบ” หรือ “wire” ส่วนต่าง ๆ เข้าด้วยกันในไฟล์ `application/http.go` โดยจะใช้การทำ Dependency Injection แต่ละเลเยอร์เข้าไป

```go
func (s *httpServer) RegisterRoutes(db sqldb.DBContext) {
 v1 := s.app.Group("/api/v1")

 customers := v1.Group("/customers")
 {
   // กำหนด dependency ระหว่างเลเยอร์
  repo := repository.NewCustomerRepository(dbCtx)
  // ส่ง instance ของ repository เข้า service
  svc := service.NewCustomerService(repo)
  // ส่ง service เข้า handler
  hdlr := handler.NewCustomerHandler(svc)
  // Register routes เข้ากับ HTTP server
  customers.Post("", hdlr.CreateCustomer)
 }

 // orders
}
```

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
| Input Validation | 400 Bad Request | ข้อมูลไม่ครบ, รูปแบบผิด | ตรวจจับได้ที่ Handler / DTO |
| Authorization | 401 Unauthorized | ยังไม่ login / token ผิด | ตรวจจับได้ที่ Middleware |
|  | 403 Forbidden | login แล้ว แต่ไม่มีสิทธิ์ | ตรวจจับได้ที่ Middleware |
| Business Rule | 404 Not Found | ไม่พบข้อมูล | ตรวจจับได้ที่ Service |
|  | 409 Conflict | ข้อมูลซ้ำกัน, ขัดแย้ง เช่น email ซ้ำ, order ถูก cancel ไปแล้ว | ตรวจจับได้ที่ Service |
|  | 422 Unprocessable Entity |  ข้อมูลมีรูปแบบถูก แต่ logic ผิด เช่น เครดิตไม่พอ, วันที่ย้อนหลัง | ตรวจจับได้ที่ Service |
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

- 23502: Not null violation → **ErrConflict**
- 23503: Foreign key violation → **ErrDataIntegrity**
- 23505: Unique constraint violation → **ErrDataIntegrity**
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

ในไฟล์ `service/customer.go` มีแค่ error การตรวจ business rules เท่านั้น

```go
var (
 ErrCreditValue = errs.BusinessRuleError("credit must be greater than 0")
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
      errResp := errs.InputValidationError(err.Error()) / <-- ปรับมาเป็น AppError
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

## สร้างระบบส่งอีเมลแบบ Reusable ด้วย Notification Service

ในขั้นตอนการสร้างลูกค้าใหม่ เราต้องการเพิ่มฟีเจอร์ **ส่งอีเมลต้อนรับ** หลังจากสร้างลูกค้าสำเร็จ และในอนาคตเราจะใช้ฟีเจอร์ส่งอีเมลในหลายจุด เช่น การยืนยันคำสั่งซื้อ

เพื่อให้สามารถใช้งานซ้ำได้อย่างยืดหยุ่นและง่ายต่อการดูแล เราจะ **แยกการส่งอีเมลออกเป็น Service เฉพาะ** ชื่อว่า `NotificationService` โดยมีขั้นตอนดังนี้

- สร้าง NotificationService แยกออกมา
เราจะสร้าง service ใหม่ในไฟล์ `service/notification.go` เพื่อจัดการการส่งอีเมลทั้งหมดของระบบไว้ในจุดเดียว

    ```go
    package service
    
    import (
     "fmt"
     "go-mma/util/logger"
    )
    
    type NotificationService struct {
    }
    
    func NewNotificationService() *NotificationService {
     return &NotificationService{}
    }
    
    func (s *NotificationService) SendEmail(to string, subject string, payload map[string]any) error {
     // implement email sending logic here
     logger.Log.Info(fmt.Sprintf("Sending email to %s with subject: %s and payload: %v", to, subject, payload))
     return nil
    }
    
    ```

- Inject NotificationService เข้าไปใน CustomerService
แก้ไข `CustomerService` ที่ไฟล์ `service/customer.go` ให้สามารถเรียกใช้ `NotificationService` ได้ เพื่อทำการส่งอีเมลต้อนรับหลังจากสร้างลูกค้าใหม่

    ```go
    package service
    
    // ...
    
    type CustomerService struct {
     custRepo *repository.CustomerRepository
     notiSvc  *NotificationService // <-- เพิ่มตรงนี้
    }
    
    func NewCustomerService(custRepo *repository.CustomerRepository, 
     notiSvc *NotificationService, // <-- เพิ่มตรงนี้
     ) *CustomerService {
     return &CustomerService{
      custRepo: custRepo,
      notiSvc:  notiSvc, // <-- เพิ่มตรงนี้
     }
    }
    
    func (s *CustomerService) CreateCustomer(ctx context.Context, req *dto.CreateCustomerRequest) (*dto.CreateCustomerResponse, error) {
     // Business Logic Rule: ตรวจสอบ email ซ้ำ
     // แปลง DTO → Model
     // ส่งไปที่ Repository Layer เพื่อบันทึกข้อมูลลงฐานข้อมูล
    
     // ส่งอีเมลต้อนรับ // <-- เพิ่มตรงนี้
     if err := s.notiSvc.SendEmail(customer.Email, "Welcome to our service!", map[string]any{
      "message": "Thank you for joining us! We are excited to have you as a member.",
     }); err != nil {
      // error logging
      logger.Log.Error(err.Error())
      return nil, err
     }
    
     // สร้าง DTO Response
    }
    ```

- Wiring Dependencies ใน Layer ของ Application
สุดท้าย ต้องแก้ไฟล์ `application/http.go` เพื่อทำ dependency injection ให้กับ `NotificationService` และ `CustomerService` อย่างถูกต้องในจุดเริ่มต้นของแอป

    ```go
    func (s *httpServer) RegisterRoutes(db sqldb.DBContext) {
     v1 := s.app.Group("/api/v1")
    
     customers := v1.Group("/customers")
     {
      repo := repository.NewCustomerRepository(db)
      svcNoti := service.NewNotificationService()      // <-- เพิ่มตรงนี้
      svc := service.NewCustomerService(repo, svcNoti) // <-- เพิ่มตรงนี้
      hdlr := handler.NewCustomerHandler(svc)
      customers.Post("", hdlr.CreateCustomer)
     }
     // orders
    }
    ```

เมื่อแยก notification logic ออกมาเป็น service จะช่วยให้

- Code **แยกความรับผิดชอบ (separation of concerns)** ชัดเจน
- สามารถนำไปใช้ซ้ำใน module อื่น ๆ ได้ทันที เช่น order
- รองรับการเปลี่ยนแปลง เช่น เปลี่ยนจาก log → SMTP → SendGrid โดยไม่กระทบ business logic

## สร้างระบบจัดการออเดอร์ด้วย Layered Architecture

ในส่วนนี้เราจะพัฒนาระบบสำหรับจัดการออเดอร์โดยใช้สถาปัตยกรรมแบบ Layered Architecture ซึ่งประกอบด้วยขั้นตอนหลักดังนี้

### Repository Layer

- สร้างโมเดล

    สร้างไฟล์ `model/order.go` เพื่อกำหนดโครงสร้างของออเดอร์ และฟังก์ชันสำหรับสร้างออเดอร์ใหม่

    ```go
    package model
    
    import (
     "go-mma/util/idgen"
     "time"
    )
    
    type Order struct {
     ID         int64      `db:"id"`
     CustomerID int64      `db:"customer_id"`
     OrderTotal int        `db:"order_total"`
     CreatedAt  time.Time  `db:"created_at"`
     CanceledAt *time.Time `db:"canceled_at"` // nullable
    }
    
    func NewOrder(customerID int64, orderTotal int) *Order {
     return &Order{
      ID:     idgen.GenerateTimeRandomID(),
      CustomerID: customerID,
      OrderTotal: orderTotal,
     }
    }
    ```

    สำหรับโมเดล `Customer` ในไฟล์ `model/customer.go` ให้เพิ่มเมธอดเพื่อจัดการ credit ได้แก่ ตัดยอด (`ReserveCredit`) และคืนยอด (`ReleaseCredit`) โดยใช้แนวทางของ ([Rich Model](https://somprasongd.work/blog/architecture/anemic-vs-rich-model))

    ```go
    func (c *Customer) ReserveCredit(v int) error {
     newCredit := c.Credit - v
     if newCredit < 0 { // เมื่อตัดยอดติดลบแสดงว่า credit ไม่พอ
      return errs.BusinessRuleError("insufficient credit limit")
     }
     c.Credit = newCredit
     return nil
    }
    
    func (c *Customer) ReleaseCredit(v int) {
     if c.Credit <= 0 { // reset ยอดก่อนถ้าติดลบ
      c.Credit = 0
     }
     c.Credit = c.Credit + v
    }
    ```

- สร้าง Repository

    สร้างไฟล์ `repository/order.go` สำหรับจัดการกับคำสั่ง SQL ที่เกี่ยวข้องกับออเดอร์ เช่น การสร้าง ค้นหา และยกเลิกออเดอร์

    ```go
    package repository
    
    import (
     "context"
     "database/sql"
     "fmt"
     "go-mma/model"
     "go-mma/util/errs"
     "go-mma/util/storage/sqldb"
     "time"
    )
    
    type OrderRepository struct {
     dbCtx sqldb.DBContext
    }
    
    func NewOrderRepository(dbCtx sqldb.DBContext) *OrderRepository {
     return &OrderRepository{
      dbCtx: dbCtx,
     }
    }
    
    func (r *OrderRepository) Create(ctx context.Context, m *model.Order) error {
     query := `
     INSERT INTO public.orders (
       id, customer_id, order_total
     )
     VALUES ($1, $2, $3)
     RETURNING *
     `
    
     ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
     defer cancel()
    
     err := r.dbCtx.DB().QueryRowxContext(ctx, query, m.ID, m.CustomerID, m.OrderTotal).StructScan(m)
     if err != nil {
      return errs.HandleDBError(fmt.Errorf("an error occurred while inserting an order: %w", err))
     }
     return nil
    }
    
    func (r *OrderRepository) FindByID(ctx context.Context, id int64) (*model.Order, error) {
     query := `
     SELECT *
     FROM public.orders
     WHERE id = $1
     AND canceled_at IS NULL -- รายออเดอร์ต้องยังไม่ถูกยกเลิก
    `
     ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
     defer cancel()
    
     var order model.Order
     err := r.dbCtx.DB().QueryRowxContext(ctx, query, id).StructScan(&order)
     if err != nil {
      if err == sql.ErrNoRows {
       return nil, nil
      }
      return nil, errs.HandleDBError(fmt.Errorf("an error occurred while finding a order by id: %w", err))
     }
     return &order, nil
    }
    
    func (r *OrderRepository) Cancel(ctx context.Context, id int64) error {
     query := `
     UPDATE public.orders
     SET canceled_at = current_timestamp -- soft delete record
     WHERE id = $1
    `
     ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
     defer cancel()
     _, err := r.dbCtx.DB().ExecContext(ctx, query, id)
     if err != nil {
      return errs.HandleDBError(fmt.Errorf("failed to cancel order: %w", err))
     }
     return nil
    }
    ```

    ในไฟล์ `repository/customer.go` ให้เพิ่มฟังก์ชัน `FindByID` และ `UpdateCredit` เพื่อค้นหาข้อมูลลูกค้าและอัปเดตยอดเครดิต

    ```go
    func (r *CustomerRepository) FindByID(ctx context.Context, id int64) (*model.Customer, error) {
     query := `
     SELECT *
     FROM public.customers
     WHERE id = $1
    `
     ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
     defer cancel()
    
     var customer model.Customer
     err := r.dbCtx.DB().QueryRowxContext(ctx, query, id).StructScan(&customer)
     if err != nil {
      if err == sql.ErrNoRows {
       return nil, nil
      }
      return nil, errs.HandleDBError(fmt.Errorf("an error occurred while finding customer by id: %w", err))
     }
    
     return &customer, nil
    }
    
    func (r *CustomerRepository) UpdateCredit(ctx context.Context, m *model.Customer) error {
     query := `
     UPDATE public.customers
     SET credit = $2
     WHERE id = $1
     RETURNING *
    `
     ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
     defer cancel()
    
     err := r.dbCtx.DB().QueryRowxContext(ctx, query, m.ID, m.Credit).StructScan(m)
     if err != nil {
      return errs.HandleDBError(fmt.Errorf("an error occurred while updating customer credit: %w", err))
     }
     return nil
    }
    ```

### Service Layer

- สร้าง DTO สำหรับไว้ รับ-ส่งข้อมูลระหว่าง Handler ↔ Service

    สร้างไฟล์ `dto/order_request.go` สำหรับรับข้อมูลการสร้างออเดอร์จากฝั่งผู้ใช้งาน และเพิ่มฟังก์ชัน `Validate` สำหรับตรวจสอบความถูกต้องของข้อมูล

    ```go
    package dto
    
    import "fmt"
    
    type CreateOrderRequest struct {
     CustomerID int64 `json:"customer_id"`
     OrderTotal int   `json:"order_total"`
    }
    
    func (r *CreateOrderRequest) Validate() error {
     if r.CustomerID <= 0 {
      return fmt.Errorf("customer_id is required")
     }
     return nil
    }
    ```

    สร้างไฟล์ `dto/order_response.go` สำหรับสร้าง response ที่จะส่งกลับไปยังผู้ใช้งานหลังจากสร้างออเดอร์สำเร็จ

    ```go
    package dto
    
    type CreateOrderResponse struct {
     ID int64 `json:"id"`
    }
    
    func NewCreateOrderResponse(id int64) *CreateOrderResponse {
     return &CreateOrderResponse{ID: id}
    }
    ```

- สร้าง Service

    สร้างไฟล์ `service/order.go` เพื่อรวม business logic สำหรับการสร้างและยกเลิกออเดอร์ โดยมีขั้นตอนสำคัญดังนี้

  - ตรวจสอบยอดรวมต้องมากกว่า 0
  - ตรวจสอบว่าลูกค้ามีอยู่จริง
  - ตรวจสอบ credit และตัดยอด
  - บันทึกออเดอร์
  - ส่งอีเมลแจ้งเตือน
  - หากยกเลิกออเดอร์ ให้ทำการคืนยอดเครดิตให้ลูกค้า และบันทึกลงฐานข้อมูล

    ```go
    package service
    
    import (
     "context"
     "go-mma/dto"
     "go-mma/model"
     "go-mma/repository"
     "go-mma/util/errs"
     "go-mma/util/logger"
    )
    
    var (
     ErrTotalOrderValue = errs.BusinessRuleError("order_total must be greater than 0")
     ErrNoCustomerID    = errs.ResourceNotFoundError("the customer with given id was not found")
     ErrNoOrderID       = errs.ResourceNotFoundError("the order with given id was not found")
    )
    
    type OrderService struct {
     custRepo  *repository.CustomerRepository
     orderRepo *repository.OrderRepository
     notiSvc   *NotificationService
    }
    
    func NewOrderService(custRepo *repository.CustomerRepository, orderRepo *repository.OrderRepository, notiSvc *NotificationService) *OrderService {
     return &OrderService{
      custRepo:  custRepo,
      orderRepo: orderRepo,
      notiSvc:   notiSvc,
     }
    }
    
    func (s *OrderService) CreateOrder(ctx context.Context, req *dto.CreateOrderRequest) (*dto.CreateOrderResponse, error) {
     // Business Logic Rule: ตรวจสอบ ยอดรวมต้องมากกว่า 0
     if req.OrderTotal <= 0 {
      return nil, ErrTotalOrderValue
     }
    
     // Business Logic Rule: ตรวจสอบ customer id
     customer, err := s.custRepo.FindByID(ctx, req.CustomerID)
     if err != nil {
      logger.Log.Error(err.Error())
      return nil, err
     }
    
     if customer == nil {
      return nil, ErrNoCustomerID
     }
    
     // Business Logic Rule: ตัดยอด credit ถ้าไม่พอให้ error
     if err := customer.ReserveCredit(req.OrderTotal); err != nil {
      return nil, err
     }
    
     // ตัดยอด credit ในตาราง customer
     if err := s.custRepo.UpdateCredit(ctx, customer); err != nil {
      logger.Log.Error(err.Error())
      return nil, err
     }
    
     // สร้าง order ใหม่ DTO -> Model
     order := model.NewOrder(req.CustomerID, req.OrderTotal)
     // บันทึกลงฐานข้อมูล
     err = s.orderRepo.Create(ctx, order)
     if err != nil {
      logger.Log.Error(err.Error())
      return nil, err
     }
    
     // ส่งอีเมลยืนยัน
     err = s.notiSvc.SendEmail(customer.Email, "Order Created", map[string]any{
      "order_id": order.ID,
      "total":    order.OrderTotal,
     })
     if err != nil {
      logger.Log.Error(err.Error())
      return nil, err
     }
    
     // สร้าง DTO Response
     resp := dto.NewCreateOrderResponse(order.ID)
     return resp, nil
    }
    
    func (s *OrderService) CancelOrder(ctx context.Context, id int64) error {
     // Business Logic Rule: ตรวจสอบ order id
     order, err := s.orderRepo.FindByID(ctx, id)
     if err != nil {
      logger.Log.Error(err.Error())
      return err
     }
    
     if order == nil {
      return ErrNoOrderID
     }
    
     // ยกเลิก order
     if err := s.orderRepo.Cancel(ctx, order.ID); err != nil {
      logger.Log.Error(err.Error())
      return err
     }
    
     // Business Logic Rule: ตรวจสอบ customer id
     customer, err := s.custRepo.FindByID(ctx, order.CustomerID)
     if err != nil {
      logger.Log.Error(err.Error())
      return err
     }
    
     if customer == nil {
      return ErrNoCustomerID
     }
    
     // Business Logic: คืนยอด credit
     customer.ReleaseCredit(order.OrderTotal)
    
     // บันทึกการคืนยอด credit
     if err := s.custRepo.UpdateCredit(ctx, customer); err != nil {
      logger.Log.Error(err.Error())
      return err
     }
    
     return nil
    }
    ```

### Presentation Layer

- สร้าง Handler
สร้างไฟล์ `handler/order.go` เพื่อจัดการคำขอ HTTP ที่เกี่ยวข้องกับออเดอร์ เช่น การสร้างและยกเลิกออเดอร์ โดยเรียกใช้ Service Layer และจัดการ response/error ให้เหมาะสม

    ```go
    package handler
    
    import (
     "fmt"
     "go-mma/dto"
     "go-mma/service"
     "go-mma/util/errs"
     "go-mma/util/logger"
     "strconv"
    
     "github.com/gofiber/fiber/v3"
    )
    
    type OrderHandler struct {
     orderSvc *service.OrderService
    }
    
    func NewOrderHandler(orderSvc *service.OrderService) *OrderHandler {
     return &OrderHandler{orderSvc: orderSvc}
    }
    
    func (h *OrderHandler) CreateOrder(c fiber.Ctx) error {
     // แปลง request body -> struct
     var req dto.CreateOrderRequest
     if err := c.Bind().Body(&req); err != nil {
      // จัดการ error response ที่ middleware
      return errs.InputValidationError(err.Error())
     }
    
     logger.Log.Info(fmt.Sprintf("Received Order: %v", req))
    
     // ตรวจสอบ input fields (e.g., value, format, etc.)
     if err := req.Validate(); err != nil {
      // จัดการ error response ที่ middleware
      return errs.InputValidationError(err.Error())
     }
    
     // ส่งไปที่ Service Layer
     resp, err := h.orderSvc.CreateOrder(c.Context(), &req)
    
     // จัดการ error จาก Service Layer หากเกิดขึ้น
     if err != nil {
      // จัดการ error response ที่ middleware
      return err
     }
    
     // ตอบกลับด้วย status code 201 (created) และข้อมูลแบบ JSON
     return c.Status(fiber.StatusCreated).JSON(resp)
    }
    
    func (h *OrderHandler) CancelOrder(c fiber.Ctx) error {
     // ตรวจสอบรูปแบบ orderID
     orderID, err := strconv.Atoi(c.Params("orderID"))
     if err != nil {
      // จัดการ error response ที่ middleware
      return errs.InputValidationError("invalid order id")
     }
    
     logger.Log.Info(fmt.Sprintf("Cancelling order: %v", orderID))
    
     // ส่งไปที่ Service Layer
     err = h.orderSvc.CancelOrder(c.Context(), int64(orderID))
    
     // จัดการ error จาก Service Layer หากเกิดขึ้น
     if err != nil {
      // จัดการ error response ที่ middleware
      return err
     }
    
     // ตอบกลับด้วย status code 204 (no content)
     return c.SendStatus(fiber.StatusNoContent)
    }
    ```

### ประกอบร่างด้วย Dependency Injection

- แก้ไขไฟล์ `application/http.go`

  ```go
  func (s *httpServer) RegisterRoutes(db sqldb.DBContext) {
  v1 := s.app.Group("/api/v1")

  // customers

  orders := v1.Group("/orders")
  {
    repoCust := repository.NewCustomerRepository(dbCtx)
    repoOrder := repository.NewOrderRepository(dbCtx)
    svcNoti := service.NewNotificationService()
    svcOrder := service.NewOrderService(repoCust, repoOrder, svcNoti)
    hdlr := handler.NewOrderHandler(svcOrder)
    orders.Post("", hdlr.CreateOrder)
    orders.Delete("/:orderID", hdlr.CancelOrder)
  }
  }

  ```
