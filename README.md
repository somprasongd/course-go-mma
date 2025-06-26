# แนวทางการสร้างโปรเจกต์แบบ Modular Monolith

> หากคุณเคยเริ่มพัฒนาแอปพลิเคชันด้วยโครงสร้างแบบ Monolith แล้วพบว่าระบบเริ่มยุ่งเหยิงเมื่อทีมเติบโตขึ้น หรือเมื่อฟีเจอร์ใหม่ๆ เริ่มทับซ้อนกัน — Modular Monolith อาจเป็นคำตอบที่เหมาะสม
>

บทความนี้จะพาคุณไล่เรียงตั้งแต่พื้นฐานของการพัฒนาเว็บแอปในรูปแบบ Monolith ไปสู่การออกแบบระบบแบบ **Modular Monolith** อย่างเป็นระบบ พร้อมแนวคิดและแนวปฏิบัติที่นำไปใช้ได้จริง โดยไม่ต้องกระโดดไปสู่ Microservices ตั้งแต่แรก

## สารบัญ

- [โปรเจกต์นี้มีเป้าหมายอะไร](#โปรเจกต์นี้มีเป้าหมายอะไร)
- [ออกแบบ Web Server ให้ดีตั้งแต่แรก](#ออกแบบ-web-server-ให้ดีตั้งแต่แรก)
- [การแยก logic ออกจาก routing ด้วย Handlers](#การแยก-logic-ออกจาก-routing-ด้วย-handlers)
- [เชื่อมต่อฐานข้อมูลอย่างปลอดภัยและยืดหยุ่น](#เชื่อมต่อฐานข้อมูลอย่างปลอดภัยและยืดหยุ่น)
- [เริ่มต้นด้วย Layered Architecture ที่เข้าใจง่าย](#เริ่มต้นด้วย-layered-architecture-ที่เข้าใจง่าย)
- [ออกแบบระบบจัดการ Error ให้ตรวจสอบและแก้ไขง่าย](#ออกแบบระบบจัดการ-error-ให้ตรวจสอบและแก้ไขง่าย)
- [สร้างระบบส่งอีเมลแบบ Reusable ด้วย Notification Service](#สร้างระบบส่งอีเมลแบบ-reusable-ด้วย-notification-service)
- [สร้างระบบจัดการออเดอร์ด้วย Layered Architecture](#สร้างระบบจัดการออเดอร์ด้วย-layered-architecture)
- [ใช้งาน Database Transaction อย่างไรให้ถูกต้อง](#ใช้งาน-database-transaction-อย่างไรให้ถูกต้อง)
- [นำหลักการ Dependency Inversion มาใช้ในระบบจริง](#นำหลักการ-dependency-inversion-มาใช้ในระบบจริง)
- [แปลงโครงสร้างไปสู่ Modular Architecture อย่างเป็นขั้นตอน](#แปลงโครงสร้างไปสู่-modular-architecture-อย่างเป็นขั้นตอน)
- [แยกความรับผิดชอบด้วยการซ่อนรายละเอียดของ Subdomain](#แยกความรับผิดชอบด้วยการซ่อนรายละเอียดของ-subdomain)
- [จัดการ Service ใน Monolith ด้วย Service Registry](#จัดการ-service-ใน-monolith-ด้วย-service-registry)
- [ป้องกันการเข้าถึงข้ามโมดูลด้วยโฟลเดอร์ `internal`](#ป้องกันการเข้าถึงข้ามโมดูลด้วยโฟลเดอร์-internal)
- [รวมโค้ดทั้งหมดไว้ใน Mono-Repository อย่างเป็นระบบ](#รวมโค้ดทั้งหมดไว้ใน-mono-repository-อย่างเป็นระบบ)
- [กำหนด Public API Contract ระหว่างโมดูล](#กำหนด-public-api-contract-ระหว่างโมดูล)
- [การแยกข้อมูลระหว่างโมดูล (Data Isolation)](#การแยกข้อมูลระหว่างโมดูล-data-isolation)
- [การจัดการโมดูลด้วย Feature-Based Structure และ CQRS](#การจัดการโมดูลด้วย-feature-based-structure-และ-cqrs)
- [เพิ่มความยืดหยุ่นด้วยแนวคิด Event-Driven Architecture](#เพิ่มความยืดหยุ่นด้วยแนวคิด-event-driven-architecture)
- บทเสริม
  - [สร้าง API Document ด้วย Swagger](#สร้าง-api-document-ด้วย-swagger)

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
     "-s -w \
     -X 'go-mma/build.Version=${BUILD_VERSION}' \
     -X 'go-mma/build.Time=${BUILD_TIME}'" \
     -o app cmd/api/main.go
    ```

    > `-ldflags="-s -w"` คือ flag ที่ใช้กับ `go build` เพื่อ **strip debugging information** ออกจาก binary ทำให้ไฟล์เล็กลง

- รันคำสั่ง build พร้อมกำหนด build version

    ```bash
    BUILD_VERSION=0.0.1 make build
    ```

อีกทางเลือกหนึ่งคือการ package แอปเป็น Docker image เพื่อให้รันได้แบบ isolated และพร้อม deploy

- สร้าง `Dockerfile`

    ```docker
    # Builder stage
    FROM golang:1.24-alpine AS builder
    WORKDIR /app
    COPY go.mod go.sum ./
    RUN go mod download
    COPY . .

    # ปิด cgo (CGO_ENABLED=0)
    ENV CGO_ENABLED=0 \
        GOOS=linux \
        GOARCH=amd64

    # ตั้งค่า default สำหรับ VERSION
    ARG VERSION=latest
    ENV IMAGE_VERSION=${VERSION}
    RUN echo "Build version: $IMAGE_VERSION"
    RUN go build -ldflags \
    # strip debugging information ออกจาก binary ทำให้ไฟล์เล็กลง
        "-s -w \
        -X 'go-mma/build.Version=${IMAGE_VERSION}' \
        -X 'go-mma/build.Time=$(date +"%Y-%m-%dT%H:%M:%S%z")'" \
        -o /app/app cmd/api/main.go

    # Final stage
    FROM alpine:3.22
    RUN apk add --no-cache ca-certificates tzdata && \
            # สร้าง non-root user
        addgroup -S appgroup && adduser -S appuser -G appgroup 
        
    WORKDIR /app
    COPY --from=builder /app/app .
    USER appuser
    EXPOSE 8090
    ENV TZ=Asia/Bangkok

    # ใช้ ENTRYPOINT เพื่อล็อกว่าต้องรันไฟล์ไหน
    ENTRYPOINT ["./app"]
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
      // แปลงไม่ได้ แสดงว่าตรวจสอบโครงสร้างไม่ผ่าน ให้ส่ง error 400
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
     if req.Credit <= 0 {
      return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "credit must be greater than 0"})
     }
    
     // ตรวจสอบความถูกต้องตาม "กฎทางธุรกิจ" (Business Logic/Semantic Validation)
     // TODO: ตรวจสอบ email ต้องไม่ซ้ำในฐานข้อมูล
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
      // แปลงไม่ได้ แสดงว่าตรวจสอบโครงสร้างไม่ผ่าน ให้ส่ง error 400
      return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
     }
    
     logger.Log.Info(fmt.Sprintf("Received Order: %v", req))
    
     // ตรวจสอบ input fields (e.g., value, format, etc.)
     if req.CustomerID == "" {
      return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "customer_id is required"})
     }
     if req.OrderTotal <= 0 {
      return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "order_total must be greater than 0"})
     }
    
      // ตรวจสอบความถูกต้องตาม "กฎทางธุรกิจ" (Business Logic/Semantic Validation)
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
     // ตรวจสอบ input: ประเภทข้อมูลของ orderID
     orderID, err := strconv.Atoi(c.Params("orderID"))
     if err != nil {
      // ถ้าไม่ถูกต้อง error 400
      return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid order id"})
     }
    
     logger.Log.Info(fmt.Sprintf("Cancelling order: %v", orderID))
    
      // ตรวจสอบความถูกต้องตาม "กฎทางธุรกิจ" (Business Logic/Semantic Validation)
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
     // ตรวจสอบความถูกต้องตาม "กฎทางธุรกิจ" (Business Logic/Semantic Validation)
     // Rule: email ต้องไม่ซ้ำในฐานข้อมูล
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

    สร้าง Model แบบ [Rich Model](https://somprasongd.work/blog/architecture/anemic-vs-rich-model-ddd)

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
- **ตรวจสอบ**: ตรวจสอบความถูกต้องตาม **"กฎทางธุรกิจ"** (Business Logic/Semantic Validation) ซึ่งมักจะต้องมีการประมวลผลหรือตรวจสอบกับข้อมูลส่วนอื่นๆ ในระบบ เช่น การตรวจสอบข้อมูลซ้ำในฐานข้อมูล
- **แปลงข้อมูล**: แปลง DTO → Model
- **เรียก Repository**: เพื่อทำ CRUD (Create, Read, Update, Delete) ตามเงื่อนไข
- **ส่งผลลัพธ์**: รับผลลัพธ์จาก Repository แล้วแปลงกลับเป็น DTO Response
- **จัดการ error**: แสดง error log แล้วส่งกลับไปให้ Controller (หรือ Handler) จัดการต่อ

**ขั้นตอนการสร้าง Service Layer**

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
     // ตรวจสอบความถูกต้องตาม "กฎทางธุรกิจ" (Business Logic/Semantic Validation)
     // Rule: email ต้องไม่ซ้ำในฐานข้อมูล
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

- **รับคำขอ:** รับ HTTP Request จาก Client
- **แปลงข้อมูล:** แปลง JSON → DTO (ใช้ `BodyParser`, `Bind`, หรือ Unmarshal)
- **ตรวจสอบ:** ตรวจสอบความถูกต้องของ **"รูปแบบ"** และ **"โครงสร้าง"** ของข้อมูลที่ส่งเข้ามา (Input/Syntax Validation) เช่น ค่าว่าง, รูปแบบอีเมล, หรือประเภทข้อมูลที่ถูกต้อง
- **เรียก Service:** ส่ง DTO เข้า Service Layer เพื่อประมวลผล
- **ส่งผลลัพธ์:** รับผลลัพธ์จาก Service แล้วแปลงกลับเป็น JSON Response ส่งกลับไปให้ client
- **จัดการ error**: แปลง error จาก Service ให้เป็น HTTP response code เช่น `400`, `500`

**ขั้นตอนการสร้าง Presentation Layer (HTTP Handlers)**

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
     // Rule: email ต้องไม่ซ้ำในฐานข้อมูล
     // แปลง DTO → Model
     // ส่งไปที่ Repository Layer เพื่อบันทึกข้อมูลลงฐานข้อมูล
    
      // <-- เพิ่มตรงนี้
     // ส่งอีเมลต้อนรับ 
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

- สร้างโมเดล `Order` เพื่อกำหนดโครงสร้างของออเดอร์ และฟังก์ชันสำหรับสร้างออเดอร์ใหม่

    > สร้างไฟล์ `model/order.go`
    >

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

- สำหรับโมเดล `Customer` ให้เพิ่มฟังก์ชันเพื่อจัดการ credit ได้แก่ ตัดยอด (`ReserveCredit`) และคืนยอด (`ReleaseCredit`) โดยใช้แนวทางของ ([Rich Model](https://somprasongd.work/blog/architecture/anemic-vs-rich-model))

    > แก้ไขไฟล์ `model/customer.go`
    >

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

- สร้าง `OrderRepository` สำหรับจัดการกับคำสั่ง SQL ที่เกี่ยวข้องกับออเดอร์ เช่น การสร้าง ค้นหา และยกเลิกออเดอร์

    > สร้างไฟล์ `repository/order.go`
    >

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

- สำหรับ `CustomerRepository` ให้เพิ่มฟังก์ชัน `FindByID` และ `UpdateCredit` เพื่อค้นหาข้อมูลลูกค้าและอัปเดตยอดเครดิต

    > แก้ไขไฟล์ `repository/customer.go`
    >

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

- สร้าง DTO สำหรับรับข้อมูลการสร้างออเดอร์จากฝั่งผู้ใช้งาน และเพิ่มฟังก์ชัน `Validate` สำหรับตรวจสอบความถูกต้องของข้อมูล

    > สร้างไฟล์ `dto/order_request.go`
    >

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
     if r.OrderTotal <= 0 {
      return fmt.Errorf("order_total must be greater than 0")
     }
     return nil
    }
    ```

- สร้าง DTO สำหรับสร้าง response ที่จะส่งกลับไปยังผู้ใช้งานหลังจากสร้างออเดอร์สำเร็จใช้งาน

    > สร้างไฟล์ `dto/order_response.go`
    >

    ```go
    package dto
    
    type CreateOrderResponse struct {
     ID int64 `json:"id"`
    }
    
    func NewCreateOrderResponse(id int64) *CreateOrderResponse {
     return &CreateOrderResponse{ID: id}
    }
    ```

- สร้าง `OrderService` เพื่อรวม business logic สำหรับการสร้างและยกเลิกออเดอร์ โดยมีขั้นตอนสำคัญดังนี้
  - ตรวจสอบว่าลูกค้ามีอยู่จริง
  - ตรวจสอบ credit เพียงพอหรือไม่ เพื่อตัดยอด
  - บันทึกการแก้ไขยอดเครดิตลงฐานข้อมูล
  - บันทึกรายออเดอร์ใหม่ลงฐานข้อมูล
  - ส่งอีเมลแจ้งเตือน
  - หากยกเลิกออเดอร์ ให้ทำการคืนยอดเครดิตให้ลูกค้า และบันทึกการแก้ไขลงฐานข้อมูล

    > สร้างไฟล์ `service/order.go`
    >

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
     // ตรวจสอบความถูกต้องตาม "กฎทางธุรกิจ" (Business Logic/Semantic Validation)
     // Business Logic: ตรวจสอบ customer id ในฐานข้อมูล
     customer, err := s.custRepo.FindByID(ctx, req.CustomerID)
     if err != nil {
      logger.Log.Error(err.Error())
      return nil, err
     }
    
     if customer == nil {
      return nil, ErrNoCustomerID
     }
    
     // Business Logic: ตัดยอด credit ถ้าไม่พอให้ error
     if err := customer.ReserveCredit(req.OrderTotal); err != nil {
      return nil, err
     }
    
     // บันทึกการตัดยอด credit ในตาราง customer
     if err := s.custRepo.UpdateCredit(ctx, customer); err != nil {
      logger.Log.Error(err.Error())
      return nil, err
     }
    
     // สร้าง order ใหม่ DTO -> Model
     order := model.NewOrder(req.CustomerID, req.OrderTotal)
     
     // บันทึกรายออเดอร์ใหม่ลงฐานข้อมูล
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
     // Business Logic: ตรวจสอบ order id ในฐานข้อมูล
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
    
     // Business Logic: ตรวจสอบ customer id
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

## ใช้งาน Database Transaction อย่างไรให้ถูกต้อง

ในตัวอย่างโค้ดล่าสุด การสร้างออเดอร์ใหม่มีการเรียกใช้งาน repository หลายครั้ง เช่น

1. หักเครดิตจากลูกค้า
2. บันทึกคำสั่งซื้อ (order)

หากคำสั่งแรกสำเร็จ แต่คำสั่งที่สองล้มเหลว ข้อมูลในระบบจะไม่สมบูรณ์และอาจก่อให้เกิดปัญหาตามมา เช่น เครดิตถูกหักไปแล้วแต่ไม่มีคำสั่งซื้อเกิดขึ้น

```go
+------------------+        +-------------------+        +------------------+
| OrderService     |        | CustomerRepo      |        | OrderRepo        |
|------------------|        |-------------------|        |------------------|
| CreateOrder()    |        | FindCustomerByID()|        | SaveOrder()      |
|                  |        | ReserveCredit()   |        |                  |
+--------+---------+        +--------+----------+        +--------+---------+
         |                           |                            |
         | Find Customer By ID       |                            |
         +-------------------------> |                            |
         |                           |                            |
         |     Reserve Credit        |                            |
         +-------------------------> |                            |
         |                           |                            |
         |   Save Order              |                            |
         +------------------------------------------------------->|
         |                           |                            |
         |   ❌ FAIL (DB Error)      |                            |
         |<-------------------------------------------------------+
         |                           |                            |
         |                           |                            |
         |                           |                            |
         |                           |      ⚠️ CREDIT ALREADY DEDUCTED
         |                           |      ❌ ORDER NOT CREATED
         |                           |
         |  ❌ Inconsistent state!
```

### Database Transaction คืออะไร?

**Database Transaction** คือกลุ่มของคำสั่ง (เช่น `INSERT`, `UPDATE`, `DELETE`) ที่ทำงานกับฐานข้อมูล ซึ่งจะถูกมองว่าเป็น **"หน่วยการทำงานเดียวที่แบ่งแยกไม่ได้"**

หัวใจสำคัญของมันคือคุณสมบัติ **ACID** ที่รับประกันความน่าเชื่อถือของการเปลี่ยนแปลงข้อมูล:

- **Atomicity (ทำงานพร้อมกันทั้งหมด หรือไม่เลย):** การดำเนินการทั้งหมดใน Transaction จะต้องสำเร็จทั้งหมด (เรียกว่า **Commit**) หรือถ้ามีข้อผิดพลาดแม้แต่อันเดียว ทุกอย่างที่ทำไปก่อนหน้าจะถูกยกเลิกทั้งหมด (เรียกว่า **Rollback**) กลับสู่สถานะเริ่มต้น
- **Consistency (ความสอดคล้อง):** Transaction จะต้องทำให้ข้อมูลเปลี่ยนจากสถานะที่ถูกต้องหนึ่ง ไปยังอีกสถานะที่ถูกต้องหนึ่งเสมอ จะไม่มีการทิ้งข้อมูลให้อยู่ในสถานะครึ่งๆ กลางๆ
- **Isolation (การแยกตัว):** Transaction ที่ทำงานพร้อมกันหลายๆ อัน จะต้องไม่กวนกัน ผลลัพธ์จะต้องเหมือนกับว่า Transaction เหล่านั้นทำงานเรียงกันทีละอัน
- **Durability (ความคงทน):** เมื่อ Transaction ถูก Commit แล้ว ข้อมูลนั้นจะถูกบันทึกอย่างถาวรและจะไม่สูญหายไป แม้ว่าจะเกิดไฟดับหรือระบบล่มก็ตาม

**สรุป:** Database Transaction เป็นกลไกใน **ระดับฐานข้อมูล** ที่รับประกันว่าการเขียน/ลบ/แก้ไขข้อมูลจะเสร็จสมบูรณ์หรือล้มเหลวไปพร้อมกันทั้งหมด

### Unit of Work คืออะไร?

**Unit of Work (UoW)** เป็น **Design Pattern** ที่ใช้ใน **Service Layer** ใช้จัดกลุ่มของหลายๆ operation (insert, update, delete) ให้อยู่ใน **transaction เดียวกัน** โดยมุ่งเน้นที่ **ความถูกต้องของข้อมูล (Consistency)** และ **การ rollback อัตโนมัติเมื่อเกิดข้อผิดพลาด**

**องค์ประกอบของ Unit of Work**

1. Start / Begin: เริ่มต้น transaction
2. Register Changes: เก็บรายการ operation ที่จะทำ (insert, update, delete)
3. Commit: ถ้าทุกอย่างผ่าน → commit DB
4. Rollback: ถ้า error เกิดขึ้น → ยกเลิกทั้งหมด (rollback)
5. Post-Commit Hook: รัน side effects (เช่น send email) **หลังจาก** commit สำเร็จ

### สร้าง Transactor(Unit of Work) สำหรับจัดการ Database Transaction

เพื่อให้การทำงานกับหลาย repository ในหนึ่ง business flow (เช่น การหักเครดิตลูกค้าและสร้างออเดอร์) มีความ atomic มากขึ้น เราสามารถใช้ **Transactor** เพื่อควบคุม transaction และทำ rollback อัตโนมัติเมื่อเกิด error

<aside>
💡

โค้ดส่วนนี้จะถูกดัดแปลงมาจาก <https://github.com/Thiht/transactor>

</aside>

- สร้าง Interface กลาง `DBTX` เพื่อให้ repository ทั้งหมดสามารถทำงานได้กับทั้ง `*sqlx.DB` และ `*sqlx.Tx`

    > สร้างไฟล์ `util/storage/sqldb/transactor/types.go`
    >

    ```go
    package transactor
    
    import (
     "context"
     "database/sql"
    
     "github.com/jmoiron/sqlx"
    )
    
    // DBTX คือ interface กลางระหว่าง *sqlx.DB และ *sqlx.Tx
    // รวม method ของทั้ง database/sql และ sqlx เพื่อให้สามารถใช้ interchangeably ได้
    type DBTX interface {
     // database/sql methods
     ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
     PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
     QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
     QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
    
     Exec(query string, args ...any) (sql.Result, error)
     Prepare(query string) (*sql.Stmt, error)
     Query(query string, args ...any) (*sql.Rows, error)
     QueryRow(query string, args ...any) *sql.Row
    
     // sqlx methods
     GetContext(ctx context.Context, dest any, query string, args ...any) error
     MustExecContext(ctx context.Context, query string, args ...any) sql.Result
     NamedExecContext(ctx context.Context, query string, arg any) (sql.Result, error)
     PrepareNamedContext(ctx context.Context, query string) (*sqlx.NamedStmt, error)
     PreparexContext(ctx context.Context, query string) (*sqlx.Stmt, error)
     QueryRowxContext(ctx context.Context, query string, args ...any) *sqlx.Row
     QueryxContext(ctx context.Context, query string, args ...any) (*sqlx.Rows, error)
     SelectContext(ctx context.Context, dest any, query string, args ...any) error
    
     Get(dest any, query string, args ...any) error
     MustExec(query string, args ...any) sql.Result
     NamedExec(query string, arg any) (sql.Result, error)
     NamedQuery(query string, arg any) (*sqlx.Rows, error)
     PrepareNamed(query string) (*sqlx.NamedStmt, error)
     Preparex(query string) (*sqlx.Stmt, error)
     QueryRowx(query string, args ...any) *sqlx.Row
     Queryx(query string, args ...any) (*sqlx.Rows, error)
     Select(dest any, query string, args ...any) error
    
     Rebind(query string) string
     BindNamed(query string, arg any) (string, []any, error)
     DriverName() string
    }
    
    // ขยาย interface DBTX โดยเพิ่ม method สำหรับเริ่ม transaction
    type sqlxDB interface {
     DBTX
     BeginTxx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error)
    }
    
    // interface สำหรับจัดการ transaction lifecycle
    type sqlxTx interface {
     Commit() error
     Rollback() error
    }
    
    // compile-time check เพื่อให้แน่ใจว่า type เหล่านี้ implement interface ที่กำหนด
    var (
     _ DBTX   = &sqlx.DB{}
     _ DBTX   = &sqlx.Tx{}
     _ sqlxDB = &sqlx.DB{}
     _ sqlxTx = &sqlx.Tx{}
    )
    
    type (
     // ใช้เป็น key สำหรับ context value ของ transactor
     transactorKey struct{}
    
     // DBTXContext ใช้ดึง DBTX ปัจจุบันจาก context (อาจเป็น transaction หรือ db ปกติ)
     DBTXContext func(context.Context) DBTX
    )
    
    // ฝัง sqlxDB (เช่น tx) ลงใน context เพื่อใช้ใน logic ต่อไป
    func txToContext(ctx context.Context, tx sqlxDB) context.Context {
     return context.WithValue(ctx, transactorKey{}, tx)
    }
    
    // ดึง sqlxDB ออกมาจาก context ถ้ามี
    func txFromContext(ctx context.Context) sqlxDB {
     if tx, ok := ctx.Value(transactorKey{}).(sqlxDB); ok {
      return tx
     }
     return nil
    }
    ```

    `DBTX` จะถูก inject ผ่าน context เพื่อให้ repository ไม่ต้องรู้ว่าใช้ DB หรือ Tx อยู่

- เพิ่มการรองรับ Nested Transactions
เราสามารถเลือกกลยุทธ์การจัดการ nested transactions ได้ 2 แบบ
  - **ไม่รองรับ** (เหมาะกับระบบที่ควบคุม transaction flow เอง)

        > สร้างไฟล์ `util/storage/sqldb/transactor/nested_transactions_none.go`
        >

        ```go
        package transactor
        
        import (
         "context"
         "database/sql"
         "errors"
        
         "github.com/jmoiron/sqlx"
        )
        
        // NestedTransactionsNone ป้องกันไม่ให้เกิด nested transaction
        // เหมาะสำหรับระบบที่ไม่รองรับหรือไม่ต้องการ nested tx หรือเพื่อป้องกัน logic ซ้อนผิดพลาด
        func NestedTransactionsNone(db sqlxDB, tx *sqlx.Tx) (sqlxDB, sqlxTx) {
         switch typedDB := db.(type) {
         case *sqlx.DB:
          // ถ้าเป็น root DB -> wrap tx ด้วย nestedTransactionNone เพื่อป้องกัน nested
          return &nestedTransactionNone{tx}, tx
        
         case *nestedTransactionNone:
          // ถ้า tx ถูก wrap แล้ว -> คืนอันเดิมไป (กัน nested)
          return typedDB, typedDB
        
         default:
          panic("unsupported type") // ไม่รองรับ type อื่น
         }
        }
        
        // nestedTransactionNone เป็น struct ที่ wrap *sqlx.Tx
        // และ override method ที่เกี่ยวกับการเริ่ม/commit/rollback transaction
        // ให้ return error เสมอ เพื่อ block การทำ nested transaction
        type nestedTransactionNone struct {
         *sqlx.Tx
        }
        
        // Override BeginTxx เพื่อ block nested tx
        func (t *nestedTransactionNone) BeginTxx(_ context.Context, _ *sql.TxOptions) (*sqlx.Tx, error) {
         return nil, errors.New("nested transactions are not supported")
        }
        
        // Override Commit เพื่อป้องกันการ commit nested tx
        func (t *nestedTransactionNone) Commit() error {
         return errors.New("nested transactions are not supported")
        }
        
        // Override Rollback เพื่อป้องกันการ rollback nested tx
        func (t *nestedTransactionNone) Rollback() error {
         return errors.New("nested transactions are not supported")
        }
        ```

  - **ใช้ Savepoints** (เหมาะกับระบบที่อาจซ้อน transaction ได้)

        > สร้างไฟล์ `util/storage/sqldb/transactor/nested_transactions_savepoint.go`
        >

        ```go
        package transactor
        
        import (
         "context"
         "database/sql"
         "fmt"
         "strconv"
         "sync/atomic"
        
         "github.com/jmoiron/sqlx"
        )
        
        // NestedTransactionsSavepoints ใช้ savepoint เพื่อจำลอง nested transaction
        // รองรับ DB ที่สนับสนุน savepoint เช่น PostgreSQL, MySQL, MariaDB, SQLite
        func NestedTransactionsSavepoints(db sqlxDB, tx *sqlx.Tx) (sqlxDB, sqlxTx) {
         switch typedDB := db.(type) {
         case *sqlx.DB:
          // เริ่ม nested transaction จาก root db
          return &nestedTransactionSavepoints{Tx: tx}, tx
        
         case *nestedTransactionSavepoints:
          // ซ้อน nested อีกชั้น (เพิ่ม depth)
          nestedTransaction := &nestedTransactionSavepoints{
           Tx:    tx,
           depth: typedDB.depth + 1,
          }
          return nestedTransaction, nestedTransaction
        
         default:
          panic("unsupported type") // ไม่รองรับ type อื่น
         }
        }
        
        // struct สำหรับจัดการ nested transaction ด้วย savepoint
        type nestedTransactionSavepoints struct {
         *sqlx.Tx
         depth int64       // ลำดับของ nested level (ใช้ในการตั้งชื่อ savepoint)
         done  atomic.Bool // เพื่อป้องกันไม่ให้ Commit() หรือ Rollback() ถูกเรียกซ้ำ
        }
        
        // BeginTxx สร้าง savepoint ใหม่ตามลำดับ depth
        func (t *nestedTransactionSavepoints) BeginTxx(ctx context.Context, _ *sql.TxOptions) (*sqlx.Tx, error) {
         if _, err := t.ExecContext(ctx, "SAVEPOINT sp_"+strconv.FormatInt(t.depth+1, 10)); err != nil {
          return nil, fmt.Errorf("failed to create savepoint: %w", err)
         }
         return t.Tx, nil
        }
        
        // Commit จะ release savepoint ที่เกี่ยวข้องกับ level นี้
        // ใช้ CompareAndSwap เพื่อกันการ commit ซ้ำ
        func (t *nestedTransactionSavepoints) Commit() error {
         if !t.done.CompareAndSwap(false, true) {
          return sql.ErrTxDone // ป้องกันการ commit ซ้ำ
         }
         if _, err := t.Exec("RELEASE SAVEPOINT sp_" + strconv.FormatInt(t.depth, 10)); err != nil {
          return fmt.Errorf("failed to release savepoint: %w", err)
         }
         return nil
        }
        
        // Rollback จะ rollback ไปยัง savepoint ของ level นี้
        // และ mark ว่า transaction นี้จบแล้ว
        func (t *nestedTransactionSavepoints) Rollback() error {
         if !t.done.CompareAndSwap(false, true) {
          return sql.ErrTxDone // ป้องกัน rollback ซ้ำ
         }
         if _, err := t.Exec("ROLLBACK TO SAVEPOINT sp_" + strconv.FormatInt(t.depth, 10)); err != nil {
          return fmt.Errorf("failed to rollback to savepoint: %w", err)
         }
         return nil
        }
        ```

- สร้าง Transactor ซึ่งตัว `Transactor` จะทำหน้าที่เริ่ม transaction, inject context, และ commit/rollback โดยอัตโนมัติ

    > สร้างไฟล์ `util/storage/sqldb/transactor/transactor.go`
    >

    ```go
    // Ref: https://github.com/Thiht/transactor/blob/main/sqlx/transactor.go
    package transactor
    
    import (
     "context"
     "fmt"
     "go-mma/util/logger"
    
     "github.com/jmoiron/sqlx"
    )
    
    // PostCommitHook คือฟังก์ชันที่สามารถลงทะเบียนเพื่อให้ทำงานหลังจาก Commit เสร็จแล้ว
    type PostCommitHook func(ctx context.Context) error
    
    // Transactor คือ interface สำหรับการจัดการ Transaction
    type Transactor interface {
     // WithinTransaction ใช้สำหรับรันโค้ดใน Transaction
     // และสามารถลงทะเบียน PostCommitHook เพื่อรันหลังจาก Commit ได้
     WithinTransaction(
      ctx context.Context,
      txFunc func(ctxWithTx context.Context, registerPostCommitHook func(PostCommitHook)) error,
     ) error
    }
    
    // กำหนด alias สำหรับประเภทที่ใช้เพื่อ inject dependencies
    type (
     sqlxDBGetter               func(context.Context) sqlxDB // ดึง *sqlx.DB หรือ *sqlx.Tx จาก context
     nestedTransactionsStrategy func(sqlxDB, *sqlx.Tx) (sqlxDB, sqlxTx) // กลยุทธ์สำหรับจัดการ nested transaction
    )
    
    // sqlxTransactor คือ implementation ของ Transactor
    type sqlxTransactor struct {
     sqlxDBGetter               // วิธีดึง DB/Tx จาก context
     nestedTransactionsStrategy // กลยุทธ์ในการจัดการ nested transaction
    }
    
    // Option ใช้สำหรับกำหนดค่าเพิ่มเติมให้ sqlxTransactor เช่น กลยุทธ์ nested transaction
    type Option func(*sqlxTransactor)
    
    // New สร้าง instance ของ Transactor และ DBTXContext
    func New(db *sqlx.DB, opts ...Option) (Transactor, DBTXContext) {
     t := &sqlxTransactor{
      // default: ถ้า context มี tx ให้ใช้ tx, ไม่งั้นใช้ db
      sqlxDBGetter: func(ctx context.Context) sqlxDB {
       if tx := txFromContext(ctx); tx != nil {
        return tx
       }
       return db
      },
      nestedTransactionsStrategy: NestedTransactionsNone, // default strategy: ไม่รองรับ nested transaction
     }
    
     // apply ตัวเลือกเพิ่มเติม (ถ้ามี)
     for _, opt := range opts {
      opt(t)
     }
    
     // สร้างฟังก์ชันสำหรับดึง DBTX จาก context
     dbGetter := func(ctx context.Context) DBTX {
      if tx := txFromContext(ctx); tx != nil {
       return tx
      }
      return db
     }
    
     return t, dbGetter
    }
    
    // WithNestedTransactionStrategy ใช้เปลี่ยนกลยุทธ์ nested transaction (เช่น Savepoints)
    func WithNestedTransactionStrategy(strategy nestedTransactionsStrategy) Option {
     return func(t *sqlxTransactor) {
      t.nestedTransactionsStrategy = strategy
     }
    }
    
    // WithinTransaction ใช้สำหรับรันฟังก์ชันหนึ่งใน transaction context
    func (t *sqlxTransactor) WithinTransaction(
     ctx context.Context,
     txFunc func(ctxWithTx context.Context, registerPostCommitHook func(PostCommitHook)) error,
    ) error {
     // ดึง database หรือ transaction object จาก context
     currentDB := t.sqlxDBGetter(ctx)
    
     // เริ่มต้น transaction ใหม่
     tx, err := currentDB.BeginTxx(ctx, nil)
     if err != nil {
      return fmt.Errorf("failed to begin transaction: %w", err)
     }
    
     var hooks []PostCommitHook // เก็บ hook ที่จะเรียกหลังจาก commit
    
     // ฟังก์ชันสำหรับให้ผู้ใช้ลงทะเบียน post-commit hook
     registerPostCommitHook := func(hook PostCommitHook) {
      hooks = append(hooks, hook)
     }
    
     // สร้าง nested transaction context ใหม่ (ใช้กลยุทธ์ที่กำหนดไว้)
     newDB, currentTX := t.nestedTransactionsStrategy(currentDB, tx)
    
     // เผื่อกรณี panic หรือ error แล้วไม่ได้ commit — จะ rollback ให้อัตโนมัติ
     defer func() {
      _ = currentTX.Rollback()
     }()
    
     // inject transaction object เข้า context ใหม่
     ctxWithTx := txToContext(ctx, newDB)
    
     // เรียกใช้ฟังก์ชันที่รับ context + hook registration
     if err := txFunc(ctxWithTx, registerPostCommitHook); err != nil {
      return err // ถ้า error, transaction จะถูก rollback ใน defer
     }
    
     // พยายาม commit
     if err := currentTX.Commit(); err != nil {
      return fmt.Errorf("failed to commit transaction: %w", err)
     }
    
     // หลังจาก commit เสร็จ ให้ run post-commit hooks แบบ async
     go func() {
      for _, hook := range hooks {
       func(h PostCommitHook) {
        defer func() {
         // ดัก panic เพื่อไม่ให้ crash และ log ไว้
         if r := recover(); r != nil {
          logger.Log.Error(fmt.Sprintf("post-commit hook panic: %v", r))
         }
        }()
        // ถ้า error จาก hook ก็ log ไว้
        if err := h(ctx); err != nil {
         logger.Log.Error(fmt.Sprintf("post-commit hook error: %v", err))
        }
       }(hook)
      }
     }()
    
     return nil
    }
    
    // IsWithinTransaction ใช้ตรวจสอบว่า context นี้อยู่ใน transaction หรือไม่
    func IsWithinTransaction(ctx context.Context) bool {
     return ctx.Value(transactorKey{}) != nil
    }
    
    ```

    **สรุป**

  - ใช้ `Transactor` เพื่อควบคุมหลาย DB operation ให้เป็น atomic unit
  - Inject `DBTX` ผ่าน context ทำให้ repository ไม่ต้องรู้ว่าอยู่ใน transaction หรือไม่
  - รองรับ nested transactions ด้วย savepoint หากจำเป็น

### ปรับปรุง Repository Layer

ปรับปรุงโค้ดใน layer ของ repository เพื่อรองรับการใช้งาน transaction ร่วมกันข้ามหลาย repository โดยเปลี่ยนจากการ inject ตัว database connection (`*sql.DB` หรือ `sqldb.DBContext`) มาเป็นการใช้ interface ใหม่ที่ชื่อว่า `transactor.DBContext`

**สิ่งที่เปลี่ยนแปลง**

- เดิมใช้ `dbCtx sqldb.DBContext`
- เปลี่ยนเป็น `dbCtx transactor.DBContext`
- เวลาจะเรียกใช้ `db.QueryContext(...)`, `db.ExecContext(...)` หรืออื่น ๆ ให้ดึง `sql.DB` หรือ `sql.Tx` จาก `context.Context` โดยใช้ `r.dbCtx(ctx)` แทน

**ปรับปรุงโค้ด**

- แก้ไขไฟล์ `repository/customer.go`

    ```go
    package repository
    
    import (
     "context"
     "database/sql"
     "fmt"
     "go-mma/model"
     "go-mma/util/errs"
     "go-mma/util/storage/sqldb/transactor"
     "time"
    )
    
    type CustomerRepository struct {
     dbCtx transactor.DBTXContext // <-- ตรงนี้
    }
    
    func NewCustomerRepository(dbCtx transactor.DBTXContext) // <-- ตรงนี้
    *CustomerRepository {
     // ...
    }
    
    func (r *CustomerRepository) Create(ctx context.Context, customer *model.Customer) error {
     // ...
    
     err := r.dbCtx(ctx). // <-- ตรงนี้ จะเป็น *sqlx.DB หรือ *sqlx.Tx ก็ได้
     
     // ...
    }
    
    func (r *CustomerRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
     // ...
     
     err := r.dbCtx(ctx). // <-- ตรงนี้ จะเป็น *sqlx.DB หรือ *sqlx.Tx ก็ได้
     
     // ...
    }
    
    func (r *CustomerRepository) FindByID(ctx context.Context, id int64) (*model.Customer, error) {
     // ...
     
     err := r.dbCtx(ctx). // <-- ตรงนี้ จะเป็น *sqlx.DB หรือ *sqlx.Tx ก็ได้
     
     // ...
    }
    
    func (r *CustomerRepository) UpdateCredit(ctx context.Context, m *model.Customer) error {
     // ...
    
     err := r.dbCtx(ctx). // <-- ตรงนี้ จะเป็น *sqlx.DB หรือ *sqlx.Tx ก็ได้
     
     // ...
    }
    
    ```

- แก้ไขไฟล์ `repository/order.go`

    ```go
    package repository
    
    import (
     "context"
     "database/sql"
     "fmt"
     "go-mma/model"
     "go-mma/util/errs"
     "go-mma/util/storage/sqldb/transactor"
     "time"
    )
    
    type OrderRepository struct {
     dbCtx transactor.DBTXContext  // <-- ตรงนี้
    }
    
    func NewOrderRepository(dbCtx transactor.DBTXContext) // <-- ตรงนี้
    *OrderRepository {
     // ...
    }
    
    func (r *OrderRepository) Create(ctx context.Context, m *model.Order) error {
     // ...
    
     err := r.dbCtx(ctx). // <-- ตรงนี้ จะเป็น *sqlx.DB หรือ *sqlx.Tx ก็ได้
     
     // ...
    }
    
    func (r *OrderRepository) FindByID(ctx context.Context, id int64) (*model.Order, error) {
     // ...
     
     err := r.dbCtx(ctx). // <-- ตรงนี้ จะเป็น *sqlx.DB หรือ *sqlx.Tx ก็ได้
     
     // ...
    }
    
    func (r *OrderRepository) Cancel(ctx context.Context, id int64) error {
     // ...
     
     _, err := r.dbCtx(ctx). // <-- ตรงนี้ จะเป็น *sqlx.DB หรือ *sqlx.Tx ก็ได้
     
     // ...
    }
    ```

**สรุปแนวทาง**

- ใช้ `transactor.DBTXContext` แทนการเข้าถึง `sql.DB` โดยตรง
- `dbCtx(ctx)` จะ return `sql.Tx` ถ้ามี transaction, หรือ `sql.DB` ถ้าไม่มี
- ทำให้ repository ทุกตัวสามารถทำงานร่วมกันใน transaction เดียวได้ โดยไม่ต้องรู้ว่าใช้ `sql.Tx` หรือ `sql.DB`

### ปรับปรุง Service Layer

ในชั้น `service` เราจะรับ `transactor.Transactor` มาตั้งแต่ตอนสร้าง Service และจะใช้ `WithinTransaction()` เพื่อให้ทุกคำสั่งฐานข้อมูลอยู่ภายใต้ transaction เดียวกัน ซึ่งช่วยให้ rollback ได้หากเกิดข้อผิดพลาดในกระบวนการใดๆ

**ปรับปรุงโค้ด**

- แก้ไขไฟล์ `service/customer.go` ย้ายส่วนการบันทึกลูกค้าใหม่ กับส่งอีเมล มาไว้ใน `WithinTransaction`

    ```go
    package service
    
    // ...
    
    type CustomerService struct {
     transactor transactor.Transactor // <-- ตรงนี้
     custRepo   *repository.CustomerRepository
     notiSvc    *NotificationService
    }
    
    func NewCustomerService(
     transactor transactor.Transactor, // <-- ตรงนี้
     custRepo *repository.CustomerRepository,
     notiSvc *NotificationService,
    ) *CustomerService {
     return &CustomerService{
      transactor: transactor, // <-- ตรงนี้
      custRepo:   custRepo,
      notiSvc:    notiSvc,
     }
    }
    
    func (s *CustomerService) CreateCustomer(ctx context.Context, req *dto.CreateCustomerRequest) (*dto.CreateCustomerResponse, error) {
     // ตรวจสอบความถูกต้องตาม "กฎทางธุรกิจ" (Business Logic/Semantic Validation)
     // Business Logic: email ต้องไม่ซ้ำในฐานข้อมูล
     
     // แปลง DTO → Model
     customer := model.NewCustomer(req.Email, req.Credit)
     
      // <-- แก้ตรงนี้
     // ย้ายส่วนที่ติดต่อฐานข้อมูล กับส่งอีเมลมาทำงานใน WithinTransaction
     err = s.transactor.WithinTransaction(ctx, func(ctx context.Context, registerPostCommitHook func(transactor.PostCommitHook)) error {
         
      // Unit of Work: register change
      if err := s.custRepo.Create(ctx, customer); err != nil {
       // error logging
       logger.Log.Error(err.Error())
       return err // Unit of Work: rollback จะทำงาน
      }
    
      // ส่งอีเมลต้อนรับ
      if err := s.notiSvc.SendEmail(customer.Email, "Welcome to our service!", map[string]any{
       "message": "Thank you for joining us! We are excited to have you as a member.",
      }); err != nil {
       // error logging
       logger.Log.Error(err.Error())
       return err // Unit of Work: rollback จะทำงาน
      }
    
        // Unit of Work: commit จะเกิดขึ้นอัตโนมัติหลัง func สำเร็จ
      return nil
     })
    
     if err != nil {
      return nil, err
     }
    
     // สร้าง DTO Response
     // ...
    }
    ```

- แก้ไขไฟล์ `service/order.go` ย้ายส่วนที่ติดต่อฐานข้อมูล กับส่งอีเมล มาไว้ใน `WithinTransaction`

    ```go
    package service
    
    // ...
    
    type OrderService struct {
     transactor transactor.Transactor // <-- ตรงนี้
     custRepo   *repository.CustomerRepository
     orderRepo  *repository.OrderRepository
     notiSvc    *NotificationService
    }
    
    func NewOrderService(
     transactor transactor.Transactor, // <-- ตรงนี้
     custRepo *repository.CustomerRepository,
     orderRepo *repository.OrderRepository,
     notiSvc *NotificationService) *OrderService {
     return &OrderService{
      transactor: transactor, // <-- ตรงนี้
      custRepo:   custRepo,
      orderRepo:  orderRepo,
      notiSvc:    notiSvc,
     }
    }
    
    func (s *OrderService) CreateOrder(ctx context.Context, req *dto.CreateOrderRequest) (*dto.CreateOrderResponse, error) {
     // Business Logic: ตรวจสอบ customer id ในฐานข้อมูล
     // ...
    
        // <-- แก้ตรงนี้
     // ย้ายส่วนที่ติดต่อฐานข้อมูล กับส่งอีเมลมาทำงานใน WithinTransaction
     var order *model.Order
     err = s.transactor.WithinTransaction(ctx, func(ctx context.Context, registerPostCommitHook func(transactor.PostCommitHook)) error {
       
       // Business Logic Rule: ตัดยอด credit ถ้าไม่พอให้ error
      if err := customer.ReserveCredit(req.OrderTotal); err != nil {
       return err
      }
     
        // Unit of Work: register change
      // ตัดยอด credit ในตาราง customer
      if err := s.custRepo.UpdateCredit(ctx, customer); err != nil {
       logger.Log.Error(err.Error())
       return err // Unit of Work: rollback จะทำงาน
      }
    
      // สร้าง order ใหม่ DTO -> Model
      order = model.NewOrder(req.CustomerID, req.OrderTotal)
      // Unit of Work: register change
      // บันทึกลงฐานข้อมูล
      err = s.orderRepo.Create(ctx, order)
      if err != nil {
       logger.Log.Error(err.Error())
       return err // Unit of Work: rollback จะทำงาน
      }
    
      // ส่งอีเมลยืนยัน
      err = s.notiSvc.SendEmail(customer.Email, "Order Created", map[string]any{
       "order_id": order.ID,
       "total":    order.OrderTotal,
      })
      if err != nil {
       logger.Log.Error(err.Error())
       return err // Unit of Work: rollback จะทำงาน
      }
        
        // Unit of Work: commit จะเกิดขึ้นอัตโนมัติหลัง func สำเร็จ
      return nil
     })
    
     // จัดการ error จากใน transactor
     if err != nil {
      return nil, err
     }
    
     // สร้าง DTO Response
     // ...
    }
    ```

### Post-Commit Hook

เราควรแยกโค้ดส่วนที่ทำให้เกิด side effects เช่น ส่งอีเมล, call external service มาทำงาน**หลังจาก** commit สำเร็จ เพื่อทำให้มั่นใจว่า data persist แล้ว จะไม่ถูก rollback เพราะ external failure

**ก่อนแก้ไข**

```go
// service/customer.go
// func (s *customerService) CreateCustomer(...)
err = s.transactor.WithinTransaction(ctx, func(ctx context.Context, registerPostCommitHook func(PostCommitHook)) error {
     
  // Unit of Work: register change
  if err := s.custRepo.Create(ctx, customer); err != nil {
   // error logging
   logger.Log.Error(err.Error())
   return err // Unit of Work: rollback จะทำงาน
  }

  // ส่งอีเมลต้อนรับ <-- side-effect
  if err := s.notiSvc.SendEmail(customer.Email, "Welcome to our service!", map[string]any{
   "message": "Thank you for joining us! We are excited to have you as a member.",
  }); err != nil {
   // error logging
   logger.Log.Error(err.Error())
   return err // Unit of Work: rollback จะทำงาน
  }

    // Unit of Work: commit จะเกิดขึ้นอัตโนมัติหลัง func สำเร็จ
  return nil
 })
```

**หลังแก้ไข**

```go
// service/customer.go
// func (s *customerService) CreateCustomer(...)
err = s.transactor.WithinTransaction(ctx, func(ctx context.Context, registerPostCommitHook func(PostCommitHook)) error {
     
  // Unit of Work: register change
  if err := s.custRepo.Create(ctx, customer); err != nil {
   // error logging
   logger.Log.Error(err.Error())
   return err // Unit of Work: rollback จะทำงาน
  }

  // เพิ่มส่งอีเมลต้อนรับ เข้าไปใน hook แทนการเรียกใช้งานทันที
  registerPostCommitHook(func(ctx context.Context) error {
   return s.notiSvc.SendEmail(customer.Email, "Welcome to our service!", map[string]any{
    "message": "Thank you for joining us! We are excited to have you as a member."})
  })

    // Unit of Work: commit จะเกิดขึ้นอัตโนมัติหลัง func สำเร็จ
  return nil
 })
```

แก้ไขที่ `func (s *orderService) CreateOrder(…)` ด้วย

### ปรับปรุง Dependency Injection

เพิ่มการ inject `transactor.Transactor` เข้าไปใน Service Layer

- แก้ไขไฟล์ `application/http.go`

    ```go
    func (s *httpServer) RegisterRoutes(dbCtx sqldb.DBContext) {
     v1 := s.app.Group("/api/v1")
    
      // สร้าง transactor กับ dbCtx จาก transactor
     transactor, dbtxCtx := transactor.New(dbCtx.DB()) // <-- ตรงนี้
     customers := v1.Group("/customers")
     {
      repo := repository.NewCustomerRepository(dbtxCtx) // <-- ตรงนี้
      svcNoti := service.NewNotificationService()
      svc := service.NewCustomerService(transactor, repo, svcNoti) // <-- ตรงนี้
      hdl := handler.NewCustomerHandler(svc)
      customers.Post("", hdl.CreateCustomer)
     }
    
     orders := v1.Group("/orders")
     {
      repoCust := repository.NewCustomerRepository(dbtxCtx) // <-- ตรงนี้
      repoOrder := repository.NewOrderRepository(dbtxCtx) // <-- ตรงนี้
      svcNoti := service.NewNotificationService()
      svcOrder := service.NewOrderService(transactor, repoCust, repoOrder, svcNoti) // <-- ตรงนี้
      hdl := handler.NewOrderHandler(svcOrder)
      orders.Post("", hdlr.CreateOrder)
      orders.Delete("/:orderID", hdl.CancelOrder)
     }
    }
    ```

## นำหลักการ Dependency Inversion มาใช้ในระบบจริง

**Dependency Inversion** คือ โค้ดส่วนหลัก (เช่น Handler, Service) ไม่ควรขึ้นกับโค้ดส่วนล่าง (เช่น Repository แบบเฉพาะเจาะจง), แต่ควรขึ้นกับ Interface แทน

มีเป้าหมาย คือ

- ลดการผูกติดกันของโค้ด (loose coupling)
- เปลี่ยน implementation ได้ง่าย เช่นเปลี่ยนจาก PostgreSQL → MongoDB
- ทำ unit test ได้ง่าย เพราะ mock ได้จาก interface

เมื่อใช้ Dependency Inversion

```go
┌────────────┐
│  Handler   │ ← struct: CustomerHandler
└────┬───────┘
     │ depends on interface
     ▼
┌────────────┐
│  Service   │  ← interface: CustomerService
└────┬───────┘
     │ implemented by
     ▼
┌────────────────────┐
│ ServiceImp         │ ← struct: customerService
└────────────────────┘
     │ depends on interface
     ▼
┌────────────┐
│ Repository │  ← interface: CustomerRepository
└────┬───────┘
     │ implemented by
     ▼
┌────────────────────┐
│ PostgresRepository │ ← struct: customerRepository
└────────────────────┘
```

### Repository Layer

- แก้ไขไฟล์ `repository/customer.go`

    ```go
    package repository
    
    // ...
    
    // --> Step 1: สร้าง interface
    type CustomerRepository interface {
     Create(ctx context.Context, customer *model.Customer) error
     ExistsByEmail(ctx context.Context, email string) (bool, error)
     FindByID(ctx context.Context, id int64) (*model.Customer, error)
     UpdateCredit(ctx context.Context, customer *model.Customer) error
    }
    
    type customerRepository struct { // --> Step 2: เปลี่ยนชื่อ struct เป็นตัวพิมพ์เล็ก
     dbCtx transactor.DBTXContext
    }
    
    // --> Step 3: return เป็น interface
    func NewCustomerRepository(dbCtx transactor.DBTXContext) CustomerRepository {
     return &customerRepository{ // --> Step 4: เปลี่ยนชื่อ struct เป็นตัวพิมพ์เล็ก
      dbCtx: dbCtx,
     }
    }
    
    // --> Step 5: เปลี่ยนชื่อ struct เป็นตัวพิมพ์เล็ก
    func (r *customerRepository) Create(ctx context.Context, customer *model.Customer) error {
     // ...
    }
    
    // --> Step 6: เปลี่ยนชื่อ struct เป็นตัวพิมพ์เล็ก
    func (r *customerRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
     // ...
    }
    
    // --> Step 7: เปลี่ยนชื่อ struct เป็นตัวพิมพ์เล็ก
    func (r *customerRepository) FindByID(ctx context.Context, id int64) (*model.Customer, error) {
     // ...
    }
    
    // --> Step 8: เปลี่ยนชื่อ struct เป็นตัวพิมพ์เล็ก
    func (r *customerRepository) UpdateCredit(ctx context.Context, m *model.Customer) error {
     // ...
    }
    ```

- แก้ไขไฟล์ `repository/order.go`

    ```go
    package repository
    
    // ...
    
    // --> Step 1: สร้าง interface
    type OrderRepository interface {
     Create(ctx context.Context, order *model.Order) error
     FindByID(ctx context.Context, id int64) (*model.Order, error)
     Cancel(ctx context.Context, id int64) error
    }
    
    type orderRepository struct { // --> Step 2: เปลี่ยนชื่อ struct เป็นตัวพิมพ์เล็ก
     dbCtx transactor.DBTXContext
    }
    
    // --> Step 3: return เป็น interface
    func NewOrderRepository(dbCtx transactor.DBTXContext) OrderRepository {
     return &orderRepository{ // --> Step 4: เปลี่ยนชื่อ struct เป็นตัวพิมพ์เล็ก
      dbCtx: dbCtx,
     }
    }
    
    // --> Step 5: เปลี่ยนชื่อ struct เป็นตัวพิมพ์เล็ก
    func (r *orderRepository) Create(ctx context.Context, m *model.Order) error {
     // ...
    }
    
    // --> Step 6: เปลี่ยนชื่อ struct เป็นตัวพิมพ์เล็ก
    func (r *orderRepository) FindByID(ctx context.Context, id int64) (*model.Order, error) {
     // ...
    }
    
    // --> Step 7: เปลี่ยนชื่อ struct เป็นตัวพิมพ์เล็ก
    func (r *orderRepository) Cancel(ctx context.Context, id int64) error {
     // ...
    }
    ```

### Service Layer

- แก้ไขไฟล์ `service/notification.go`

    ```go
    package service
    
    import (
     "fmt"
     "go-mma/util/logger"
    )
    
    // --> Step 1: สร้าง interface
    type NotificationService interface {
     SendEmail(to string, subject string, payload map[string]any) error
    }
    
    // --> Step 2: เปลี่ยนชื่อ struct เป็นตัวพิมพ์เล็ก
    type notificationService struct {
    }
    
    // --> Step 3: return เป็น interface
    func NewNotificationService() NotificationService {
     return &notificationService{} // --> Step 4: เปลี่ยนชื่อ struct เป็นตัวพิมพ์เล็ก
    }
    
    // --> Step 5: เปลี่ยนชื่อ struct เป็นตัวพิมพ์เล็ก
    func (s *notificationService) SendEmail(to string, subject string, payload map[string]any) error {
     // ...
    }
    ```

- แก้ไขไฟล์ `service/customer.go`

    ```go
    package service
    
    // ...
    
    // --> Step 1: สร้าง interface
    type CustomerService interface {
     CreateCustomer(ctx context.Context, req *dto.CreateCustomerRequest) (*dto.CreateCustomerResponse, error)
    }
    
    // --> Step 2: เปลี่ยนชื่อ struct เป็นตัวพิมพ์เล็ก
    type customerService struct {
     transactor transactor.Transactor
     custRepo   repository.CustomerRepository // --> step 3: เปลี่ยนจาก pointer เป็น interface
     notiSvc    NotificationService // --> step 4: เปลี่ยนจาก pointer เป็น interface
    }
    
    func NewCustomerService(
     transactor transactor.Transactor,
     custRepo repository.CustomerRepository, // --> step 5: เปลี่ยนจาก pointer เป็น interface
     notiSvc NotificationService, // --> step 6: เปลี่ยนจาก pointer เป็น interface
    ) CustomerService {            // --> Step 7: return เป็น interface
     return &customerService{     // --> Step 8: เปลี่ยนชื่อ struct เป็นตัวพิมพ์เล็ก
      transactor: transactor,
      custRepo:   custRepo,
      notiSvc:    notiSvc,
     }
    }
    
    // --> Step 9: เปลี่ยนชื่อ struct เป็นตัวพิมพ์เล็ก
    func (s *customerService) CreateCustomer(ctx context.Context, req *dto.CreateCustomerRequest) (*dto.CreateCustomerResponse, error) {
     // ...
    }
    ```

- แก้ไขไฟล์ `service/order.go`

    ```go
    package service
    
    // ...
    
    // --> Step 1: สร้าง interface
    type OrderService interface {
     CreateOrder(ctx context.Context, req *dto.CreateOrderRequest) (*dto.CreateOrderResponse, error)
     CancelOrder(ctx context.Context, id int64) error
    }
    
    // --> Step 2: เปลี่ยนชื่อ struct เป็นตัวพิมพ์เล็ก
    type orderService struct {
     transactor transactor.Transactor
     custRepo   repository.CustomerRepository // --> step 3: เปลี่ยนจาก pointer เป็น interface
     orderRepo  repository.OrderRepository // --> step 4: เปลี่ยนจาก pointer เป็น interface
     notiSvc    NotificationService // --> step 5: เปลี่ยนจาก pointer เป็น interface
    }
    
    func NewOrderService(
     transactor transactor.Transactor,
     custRepo repository.CustomerRepository, // --> step 6: เปลี่ยนจาก pointer เป็น interface
     orderRepo repository.OrderRepository, // --> step 7: เปลี่ยนจาก pointer เป็น interface
     notiSvc NotificationService, // --> step 8: เปลี่ยนจาก pointer เป็น interface
     ) OrderService {            // --> Step 9: return เป็น interface
     return &orderService{       // --> Step 10: เปลี่ยนชื่อ struct เป็นตัวพิมพ์เล็ก
      transactor: transactor,
      custRepo:   custRepo,
      orderRepo:  orderRepo,
      notiSvc:    notiSvc,
     }
    }
    
    // --> Step 11: เปลี่ยนชื่อ struct เป็นตัวพิมพ์เล็ก
    func (s *orderService) CreateOrder(ctx context.Context, req *dto.CreateOrderRequest) (*dto.CreateOrderResponse, error) {
     // ...
    }
    
    // --> Step 12: เปลี่ยนชื่อ struct เป็นตัวพิมพ์เล็ก
    func (s *orderService) CancelOrder(ctx context.Context, id int) error {
     // ...
    }
    ```

### Presentation Layer

- แก้ไขไฟล์ `handler/customer.go` แก้ให้รับ service มาเป็น interface

    ```go
    package handler
    
    // ...
    
    type CustomerHandler struct {
     custService service.CustomerService // <-- Step1: เปลี่ยนจาก pointer เป็น interface
    }
    
    func NewCustomerHandler(custService service.CustomerService, // <-- Step2: เปลี่ยนจาก pointer เป็น interface
     ) *CustomerHandler {
     return &CustomerHandler{
      custService: custService,
     }
    }
    ```

- แก้ไขไฟล์ `handler/order.go` แก้ให้รับ service มาเป็น interface

    ```go
    package handler
    
    // ...
    
    type OrderHandler struct {
     orderSvc service.OrderService // <-- Step1: เปลี่ยนจาก pointer เป็น interface
    }
    
    func NewOrderHandler(orderSvc service.OrderService, // <-- Step1: เปลี่ยนจาก pointer เป็น interface
    ) *OrderHandler {
     return &OrderHandler{orderSvc: orderSvc}
    }
    ```

## แปลงโครงสร้างไปสู่ Modular Architecture อย่างเป็นขั้นตอน

ถัดมาเราจะมาเปลี่ยนโครงสร้างจากที่แยกตาม "layer" (เช่น handler, service, repository) ไปเป็นการแยกตาม "feature หรือ use case” โดยใช้หลักการของ [Vertical Slice Architecture](https://somprasongd.work/blog/architecture/vertical-slice) คือ

- แยกตามฟีเจอร์ เช่น `customer`, `order`, `notification`
- ภายในแต่ละฟีเจอร์มีโค้ดของมันเอง: `handler`, `dto`, `service`, `model`, `repository`, `test`
- ทำให้ แยกอิสระ, ลดการพึ่งพาข้าม slice, เพิ่ม modularity

### โครงสร้างใหม่

```bash
.
├── cmd
│   └── api
│       └── main.go         # bootstraps all modules
├── config
│   └── config.go
├── modules                 
│   ├── customer
│   │   ├── handler
│   │   ├── dto
│   │   ├── model
│   │   ├── repository
│   │   ├── service
│   │   └── module.go       # wiring
│   ├── notification
│   │   ├── service
│   │   └── module.go 
│   └── order
│       ├── handler
│       ├── dto
│       ├── model
│       ├── repository
│       ├── service
│       └── module.go
├── application
│   ├── application.go      # register all modules
│   ├── http.go             # remove register all routes
│   └── middleware
│       ├── request_logger.go
│       └── response_error.go
├── migrations
│   └── ...sql
├── util
│   ├── module              # new
│   │   └── module.go       # module interface
│   └── ...
└── go.mod
```

### Notification Module

ทำการย้ายโค้ดทีเกี่ยวกับ notification มาไว้ที่ `modules/notification`

- ย้ายไฟล์ `service/notification.go` มาไว้ที่ `modules/notification/service/notification.go`

### Customer Module

ทำการย้ายโค้ดทีเกี่ยวกับ customer มาไว้ที่ `modules/customer`

- ย้ายไฟล์ `model/customer.go` มาไว้ที่ `modules/customer/model/customer.go`
- ย้ายไฟล์ `dto/customer_*.go` มาไว้ที่ `modules/customer/dto/customer_*.go`
- ย้ายไฟล์ `repository/customer.go` มาไว้ที่ `modules/customer/repository/customer.go`

    ```go
    package repository
    
    import (
     "context"
     "database/sql"
     "fmt"
     "go-mma/modules/customer/model" // <-- แก้ตรงนี้ด้วย
     "go-mma/util/errs"
     "go-mma/util/storage/sqldb/transactor"
     "time"
    )
    ```

- ย้ายไฟล์ `service/customer.go` มาไว้ที่ `modules/customer/service/customer.go`

    ```go
    package service
    
    import (
     "context"
     "go-mma/modules/customer/dto"        // <-- แก้ตรงนี้ด้วย
     "go-mma/modules/customer/model"      // <-- แก้ตรงนี้ด้วย
     "go-mma/modules/customer/repository" // <-- แก้ตรงนี้ด้วย
     "go-mma/util/errs"
     "go-mma/util/logger"
     "go-mma/util/storage/sqldb/transactor"
    
     notiService "go-mma/modules/notification/service" // <-- แก้ตรงนี้ด้วย
    )
    
    // ...
    
    type customerService struct {
     transactor transactor.Transactor
     custRepo   repository.CustomerRepository
     notiSvc    notiService.NotificationService // <-- แก้ตรงนี้ด้วย
    }
    
    func NewCustomerService(
     transactor transactor.Transactor,
     custRepo repository.CustomerRepository,
     notiSvc notiService.NotificationService, // <-- แก้ตรงนี้ด้วย
    ) CustomerService {
     // ...
    }
    ```

- ย้ายไฟล์ `handler/customer.go` มาไว้ที่ `modules/customer/handler/customer.go`

    ```go
    package handler
    
    import (
     "go-mma/modules/customer/dto"     // <-- แก้ตรงนี้ด้วย
     "go-mma/modules/customer/service" // <-- แก้ตรงนี้ด้วย
     "go-mma/util/errs"
     "strings"
    
     "github.com/gofiber/fiber/v3"
    )
    ```

- ย้ายไฟล์ `tests/customer.http` มาไว้ที่ `modules/customer/test/customer.http`

### Order Module

ทำการย้ายโค้ดทีเกี่ยวกับ order มาไว้ที่ `modules/order`

- ย้ายไฟล์ `model/order.go` มาไว้ที่ `modules/customer/model/order.go`
- ย้ายไฟล์ `dto/order*.go` มาไว้ที่ `modules/customer/dto/order*.go`
- ย้ายไฟล์ `repository/order.go` มาไว้ที่ `modules/customer/repository/order.go`

    ```go
    package repository
    
    import (
     "context"
     "database/sql"
     "fmt"
     "go-mma/modules/order/model" // <-- แก้ตรงนี้ด้วย
     "go-mma/util/errs"
     "go-mma/util/storage/sqldb/transactor"
     "time"
    )
    ```

- ย้ายไฟล์ `service/order.go` มาไว้ที่ `modules/customer/service/order.go`

    ```go
    package service
    
    import (
     "context"
     "go-mma/modules/order/dto"        // <-- แก้ตรงนี้ด้วย
     "go-mma/modules/order/model"      // <-- แก้ตรงนี้ด้วย
     "go-mma/modules/order/repository" // <-- แก้ตรงนี้ด้วย
     "go-mma/util/errs"
     "go-mma/util/logger"
     "go-mma/util/storage/sqldb/transactor"
    
     custRepository "go-mma/modules/customer/repository" // <-- แก้ตรงนี้ด้วย
     notiService "go-mma/modules/notification/service"   // <-- แก้ตรงนี้ด้วย
    )
    
    // ...
    
    type orderService struct {
     transactor transactor.Transactor
     custRepo   custRepository.CustomerRepository // <-- แก้ตรงนี้ด้วย
     orderRepo  repository.OrderRepository
     notiSvc    notiService.NotificationService   // <-- แก้ตรงนี้ด้วย
    }
    
    func NewOrderService(
     transactor transactor.Transactor,
     custRepo custRepository.CustomerRepository,   // <-- แก้ตรงนี้ด้วย
     orderRepo repository.OrderRepository,
     notiSvc notiService.NotificationService       // <-- แก้ตรงนี้ด้วย
     ) OrderService {
     // ...
    }
    ```

- ย้ายไฟล์ `handler/order.go` มาไว้ที่ `modules/customer/handler/order.go`

    ```go
    package handler
    
    import (
     "go-mma/modules/order/dto"      // <-- แก้ตรงนี้ด้วย
     "go-mma/modules/order/service"  // <-- แก้ตรงนี้ด้วย
     "go-mma/util/errs"
     "strconv"
     "strings"
    
     "github.com/gofiber/fiber/v3"
    )
    ```

- ย้ายไฟล์ `tests/order.http` มาไว้ที่ `modules/customer/test/order.http`

### Feature-level constructor

คือ แนวคิดที่ใช้ *constructor function* เฉพาะสำหรับ "feature" หรือ "module" หนึ่ง ๆ ในระบบ เพื่อ ประกอบ dependencies ทั้งหมดของ "feature" หรือ "module" นั้นเข้าเป็นหน่วยเดียว และซ่อนไว้เบื้องหลัง interface หรือ struct เพื่อให้ใช้งานได้ง่ายและยืดหยุ่น

**ตัวอย่างการใช้งาน**

```go
// module/customer/module.go
func NewCustomerModule(mCtx *module.ModuleContext) module.Module {
 repo := repository.NewCustomerRepository(mCtx.DBCtx)
 svc := service.NewCustomerService(repo)
 hdl := handler.NewCustomerHandler(svc)

 return &customerModule{
  handler: hdl,
 }
}
```

**ขั้นตอนการสร้าง**

- สร้าง Module Interface

    > สร้างไฟล์ `util/module/module.go`
    >

    ```go
    package module
    
    import (
     "go-mma/util/storage/sqldb/transactor"
    
     "github.com/gofiber/fiber/v3"
    )
    
    type Module interface {
     APIVersion() string
     RegisterRoutes(r fiber.Router)
    }
    
    type ModuleContext struct {
     Transactor transactor.Transactor
     DBCtx      transactor.DBTXContext
    }
    
    func NewModuleContext(transactor transactor.Transactor, dbCtx transactor.DBTXContext) *ModuleContext {
     return &ModuleContext{
      Transactor: transactor,
      DBCtx:      dbCtx,
     }
    }
    ```

- สร้าง Notification Module โดยใช้ Factory pattern

    > สร้างไฟล์ `modules/notification/module.go`
    >

    ```go
    package notification
    
    import (
     "go-mma/util/module"
    
     "github.com/gofiber/fiber/v3"
    )
    
    func NewModule(mCtx *module.ModuleContext) module.Module {
     return &moduleImp{mCtx}
    }
    
    type moduleImp struct {
     mCtx *module.ModuleContext
    }
    
    func (m *moduleImp) APIVersion() string {
     return "v1"
    }
    
    func (m *moduleImp) RegisterRoutes(router fiber.Router) {
     // โมดูลนี้ยังไม่มี routes
    }
    ```

- สร้าง Customer Module โดยใช้ Factory pattern และย้ายการ wiring component ต่าง ๆ (เช่น repository, service, handler) สำหรับ customer จาก `application/http.go` มาใส่ `RegisterRoutes()`

    สร้างไฟล์ `modules/customer/module.go`

    ```go
    package customer
    
    import (
     "go-mma/modules/customer/handler"
     "go-mma/modules/customer/repository"
     "go-mma/modules/customer/service"
     "go-mma/util/module"
    
     notiService "go-mma/modules/notification/service"
    
     "github.com/gofiber/fiber/v3"
    )
    
    func NewModule(mCtx *module.ModuleContext) module.Module {
     return &moduleImp{mCtx}
    }
    
    type moduleImp struct {
     mCtx *module.ModuleContext
    }
    
    func (m *moduleImp) APIVersion() string {
     return "v1"
    }
    
    func (m *moduleImp) RegisterRoutes(router fiber.Router) {
     // wiring dependencies
     repo := repository.NewCustomerRepository(m.mCtx.DBCtx)
     svcNoti := notiService.NewNotificationService()
     svc := service.NewCustomerService(m.mCtx.Transactor, repo, svcNoti)
     hdl := handler.NewCustomerHandler(svc)
    
     customers := router.Group("/customers")
     customers.Post("", hdl.CreateCustomer)
    }
    ```

- สร้าง Order Module โดยใช้ Factory pattern และย้ายการ wiring component ต่าง ๆ (เช่น repository, service, handler) สำหรับ order จาก `application/http.go` มาใส่ `RegisterRoutes()`

    สร้างไฟล์ `modules/order/module.go`

    ```go
    package order
    
    import (
     "go-mma/modules/order/handler"
     "go-mma/modules/order/repository"
     "go-mma/modules/order/service"
     "go-mma/util/module"
    
     custRepository "go-mma/modules/customer/repository"
     notiService "go-mma/modules/notification/service"
    
     "github.com/gofiber/fiber/v3"
    )
    
    func NewModule(mCtx *module.ModuleContext) module.Module {
     return &moduleImp{mCtx}
    }
    
    type moduleImp struct {
     mCtx *module.ModuleContext
    }
    
    func (m *moduleImp) APIVersion() string {
     return "v1"
    }
    
    func (m *moduleImp) RegisterRoutes(router fiber.Router) {
     // wiring dependencies
     repoCust := custRepository.NewCustomerRepository(m.mCtx.DBCtx)
     repoOrder := repository.NewOrderRepository(m.mCtx.DBCtx)
     svcNoti := notiService.NewNotificationService()
     svc := service.NewOrderService(m.mCtx.Transactor, repoCust, repoOrder, svcNoti)
     hdl := handler.NewOrderHandler(svc)
    
     orders := router.Group("/orders")
     orders.Post("", hdl.CreateOrder)
     orders.Delete("/:orderID", hdl.CancelOrder)
    }
    ```

- ลบ `RegisterRoutes()` ใน `application/http.go` และเพิ่มเติมโค้ดตามนี้

    ```go
    type HTTPServer interface {
     Start()
     Shutdown() error
     Group(prefix string) fiber.Router  // <-- ตรงนี้
    }
    
    // ใช้สำหรับสร้าง base url router เช่น /api/v1
    func (s *httpServer) Group(prefix string) fiber.Router {
     return s.app.Group(prefix)
    }
    ```

- ลบ `RegisterRoutes()` ใน `application/application.go` และเพิ่ม `RegisterModules()` เข้าไปแทน

    ```go
    func (app *Application) RegisterModules(modules ...module.Module) error {
     for _, m := range modules {
      app.registerModuleRoutes(m)
     }
    
     return nil
    }
    
    // แยกเป็นฟังก์ชันตาม single-responsibility principle (SRP)
    func (app *Application) registerModuleRoutes(m module.Module) {
     prefix := app.buildGroupPrefix(m)
     group := app.httpServer.Group(prefix)
     m.RegisterRoutes(group)
    }
    
    func (app *Application) buildGroupPrefix(m module.Module) string {
     apiBase := "/api"
     version := m.APIVersion()
     if version != "" {
      return fmt.Sprintf("%s/%s", apiBase, version)
     }
     return apiBase
    }
    ```

- ลบ `app.RegisterRoutes()` ใน `cmd/api/main.go` และเพิ่มโค้ดเพื่อสร้างโมดูล

    ```go
    package main
    
    import (
     "fmt"
     "go-mma/application"
     "go-mma/config"
     "go-mma/data/sqldb"
     "go-mma/modules/customer"
     "go-mma/modules/notification"
     "go-mma/modules/order"
     "go-mma/util/logger"
     "go-mma/util/module"
     "go-mma/util/transactor"
     "os"
     "os/signal"
     "syscall"
    )
    
    func main() {
     // log
     // config
     // db
    
     app := application.New(*config, dbCtx)
    
     transactor, dbtxCtx := transactor.New(dbCtx.DB())
     mCtx := module.NewModuleContext(transactor, dbtxCtx)
     app.RegisterModules(
      notification.NewModule(mCtx),
      customer.NewModule(mCtx),
      order.NewModule(mCtx),
     )
    
     app.Run()
    
     // ...
    }
    ```

## แยกความรับผิดชอบด้วยการซ่อนรายละเอียดของ Subdomain

เพื่อให้ระบบของเราชัดเจนและจัดการง่าย เราควร **ซ่อนรายละเอียดภายในของแต่ละ subdomain** (เช่น `Customer`, `Order`, `Notification`) ไว้เบื้องหลัง **จุดเชื่อมต่อเดียว (Facade)**

**Facade** ทำหน้าที่เหมือน "ประตูทางเข้า" ให้ module อื่นสามารถใช้งาน subdomain ได้ โดยไม่ต้องรู้ว่าเบื้องหลังมีโค้ดหรือโครงสร้างอะไรซับซ้อนบ้าง

### ทำไมต้องซ่อน?

- เพื่อให้ **แต่ละ module แยกกันอย่างชัดเจน** (Bounded Context)
- ลดการ **พึ่งพาซึ่งกันและกันโดยตรง** ระหว่าง module
- ทำให้การสื่อสารระหว่าง subdomain เป็นไปได้ง่าย และ **ไม่ต้องรู้โครงสร้างภายใน**

<aside>
💡

**สรุปสั้น ๆ:** **Facade = ประตูทางเข้าเพียงจุดเดียว** ที่ให้ module อื่นใช้เรียก functionality ของ subdomain ได้ โดย **ไม่ต้องรู้รายละเอียดภายใน**

</aside>

### ตัวอย่างการใช้งาน

ก่อน: ระบบเข้าถึงหลายชั้นโดยตรง

```
Order Handler
     │
     ▼
Order Service
     │
     ├──────────────▶ Order Repository
     │
     └──────────────▶ Customer Repository
```

ตัวอย่าง

```go
// OrderService เรียก CustomerRepository ตรง ๆ
customer, err := customerRepo.FindByID(ctx, order.CustomerID)
if customer.Credit < order.Total {
    return errors.New("insufficient credit")
}
```

- `Order Service` เรียกทั้ง `OrderRepo` และ `CustomerRepo` โดยตรง

หลัง: ใช้ Encapsulation

```
Order Handler
     │
     ▼
Order Service
     │
     ├──────────────▶ Order Repository
     │
     └──────────────▶ Customer Service
                             │
                             └────────▶ Customer Repository

```

ตัวอย่าง

```go
// OrderService ใช้ CustomerFacade แทน
ok, err := customerService.HasSufficientCredit(ctx, order.CustomerID, order.Total)
if !ok {
    return errors.New("insufficient credit")
}
```

- `CustomerService`  เป็นจุดเดียวที่เปิดเผย logic ภายใน subdomain customer
- ภายใน facade จะจัดการ repository, validation, business rule ทั้งหมดเอง

### ซ่อนรายละเอียดภายในของ Customer

ตอนนี้ใน OrderService มีการเรียกใช้ `model` และ `repository` ของโมดูล customer โดยตรง ถ้าจะซ่อนรายละเอียดภายใน ทำได้ ดังนี้

- สร้าง DTO สำหรับส่งค่า customer กลับออกไปจาก CustomerService สำหรับการค้นหาลูกค้าจาก id

    > สร้างไฟล์ `module/customer/dto/customer_info.go`
    >

    ```go
    package dto
    
    type CustomerInfo struct {
     ID     int64  `json:"id"`
     Email  string `json:"email"`
     Credit int    `json:"credit"`
    }
    
    func NewCustomerInfo(id int64, email string, credit int) *CustomerInfo {
     return &CustomerInfo{ID: id, Email: email, Credit: credit}
    }
    ```

- เพื่อเพิ่มฟังก์ชันสำหรับการค้นหาจาก id

    > แก้ไข `module/customer/service/customer.go`
    >

    ```go
    var (
     ErrEmailExists      = errs.ConflictError("email already exists")
     ErrCustomerNotFound = errs.ResourceNotFoundError("the customer with given id was not found") // <-- ตรงนี้
    )
    
    type CustomerService interface {
     CreateCustomer(ctx context.Context, req *dto.CreateCustomerRequest) (*dto.CreateCustomerResponse, error)
     GetCustomerByID(ctx context.Context, id int64) (*dto.CustomerInfo, error) // <-- ตรงนี้
    }
    
    // ...
    
    // <-- ตรงนี้
    func (s *customerService) GetCustomerByID(ctx context.Context, id int64) (*dto.CustomerInfo, error) {
     customer, err := s.custRepo.FindByID(ctx, id)
     if err != nil {
      // error logging
      logger.Log.Error(err.Error())
      return nil, err
     }
    
     if customer == nil {
      return nil, ErrCustomerNotFound
     }
    
     // สร้าง DTO Response
     return dto.NewCustomerInfo(
      customer.ID, 
      customer.Email, 
      customer.Credit), nil
    }
    ```

- เพื่อเพิ่มฟังก์ชันสำหรับการตัดยอด และคืนยอด credit

    > แก้ไข `module/customer/service/customer.go`
    >

    ```go
    var (
     ErrEmailExists                  = errs.ConflictError("email already exists")
     ErrCustomerNotFound             = errs.ResourceNotFoundError("the customer with given id was not found")
     ErrOrderTotalExceedsCreditLimit = errs.BusinessRuleError("order total exceeds credit limit") // <-- ตรงนี้
    )
    
    type CustomerService interface {
     // ...
     // <-- ตรงนี้
     ReserveCredit(ctx context.Context, id int64, amount int) error
     ReleaseCredit(ctx context.Context, id int64, amount int) error
    }
    
    // ...
    
    // <-- ตรงนี้
    func (s *customerService) ReserveCredit(ctx context.Context, id int64, amount int) error {
     err := s.transactor.WithinTransaction(ctx, func(ctx context.Context, registerPostCommitHook func(transactor.PostCommitHook)) error {
      customer, err := s.custRepo.FindByID(ctx, id)
      if err != nil {
       logger.Log.Error(err.Error())
       return err
      }
    
      if customer == nil {
       return ErrCustomerNotFound
      }
    
      if err := customer.ReserveCredit(amount); err != nil {
       return ErrOrderTotalExceedsCreditLimit
      }
    
      if err := s.custRepo.UpdateCredit(ctx, customer); err != nil {
       logger.Log.Error(err.Error())
       return err
      }
    
      return nil
     })
     return err
    }
    
    func (s *customerService) ReleaseCredit(ctx context.Context, id int64, amount int) error {
     err := s.transactor.WithinTransaction(ctx, func(ctx context.Context, registerPostCommitHook func(transactor.PostCommitHook)) error {
      customer, err := s.custRepo.FindByID(ctx, id)
      if err != nil {
       logger.Log.Error(err.Error())
       return err
      }
    
      if customer == nil {
       return ErrCustomerNotFound
      }
    
      customer.ReleaseCredit(amount)
    
      if err := s.custRepo.UpdateCredit(ctx, customer); err != nil {
       logger.Log.Error(err.Error())
       return err
      }
    
      return nil
     })
    
     return err
    }
    ```

### เรียกใช้ CustomerService ในโมดูล Order

- ทำให้ `OrderService` เรียกใช้งาน `CustomerService` แทน `CustomerRepository`

    > แก้ไข `module/order/service/order.go`
    >

    ```go
    package service
    
    import (
     "context"
     "go-mma/modules/order/dto"
     "go-mma/modules/order/model"
     "go-mma/modules/order/repository"
     "go-mma/util/errs"
     "go-mma/util/logger"
     "go-mma/util/transactor"
    
     custService "go-mma/modules/customer/service" // <-- ตรงนี้
     notiService "go-mma/modules/notification/service"
    )
    
    var (
     ErrNoOrderID = errs.ResourceNotFoundError("the order with given id was not found") // <-- ตรงนี้ เหลือแค่ตัวเดียว
    )
    
    // ...
    
    type orderService struct {
     transactor transactor.Transactor
     custSvc    custService.CustomerService // <-- ตรงนี้
     orderRepo  repository.OrderRepository
     notiSvc    notiService.NotificationService
    }
    
    func NewOrderService(
     transactor transactor.Transactor,
     custSvc custService.CustomerService, // <-- ตรงนี้
     orderRepo repository.OrderRepository,
     notiSvc notiService.NotificationService) OrderService {
     // ...
    }
    
    func (s *orderService) CreateOrder(ctx context.Context, req *dto.CreateOrderRequest) (*dto.CreateOrderResponse, error) {
     // Business Logic Rule: ตรวจสอบ customer id ในฐานข้อมูล
     // <-- ตรงนี้
     customer, err := s.custSvc.GetCustomerByID(ctx, req.CustomerID)
     if err != nil {
      return nil, err
     }
     // ...
     // ย้ายส่วนที่ติดต่อฐานข้อมูล กับส่งอีเมลมาทำงานใน WithinTransaction
     var order *model.Order
     err = s.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
    
      // ตัดยอด credit ในตาราง customer
      // <-- ตรงนี้
      if err := s.custSvc.ReserveCredit(ctx, customer.ID, req.OrderTotal); err != nil { 
       return err
      }
    
      // ...
     })
    
     // ...
    }
    
    func (s *orderService) CancelOrder(ctx context.Context, id int64) error {
     // Business Logic Rule: ตรวจสอบ order id ในฐานข้อมูล
     order, err := s.orderRepo.FindByID(ctx, id)
     if err != nil {
      logger.Log.Error(err.Error())
      return err
     }
    
     if order == nil {
      return ErrNoOrderID
     }
    
     err = s.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
    
      // ยกเลิก order
      if err := s.orderRepo.Cancel(ctx, order.ID); err != nil {
       logger.Log.Error(err.Error())
       return err
      }
    
      // Business Logic: คืนยอด credit
      // <-- ตรงนี้
      err = s.custSvc.ReleaseCredit(ctx, order.CustomerID, order.OrderTotal)
      if err != nil {
       return err
      }
    
      return nil
     })
    
     return err
    }
    ```

- ปรับใน `RegisterRoutes` ส่ง `CustomerService` ไปแทน `CustomerRepository`

    > แก้ไข `module/order/module.go`
    >

    ```go
    package order
    
    import (
     "go-mma/modules/order/handler"
     "go-mma/modules/order/repository"
     "go-mma/modules/order/service"
     "go-mma/util/module"
    
     custRepository "go-mma/modules/customer/repository"
     custService "go-mma/modules/customer/service" // เพิ่มตรงนี้
     notiService "go-mma/modules/notification/service"
    
     "github.com/gofiber/fiber/v3"
    )
    
    // ...
    
    func (m *moduleImp) RegisterRoutes(router fiber.Router) {
     // wiring dependencies
     svcNoti := notiService.NewNotificationService()
    
     repoCust := custRepository.NewCustomerRepository(m.mCtx.DBCtx)
     svcCust := custService.NewCustomerService(m.mCtx.Transactor, repoCust, svcNoti) // <-- เพิ่มตรงนี้
    
     repoOrder := repository.NewOrderRepository(m.mCtx.DBCtx)
    
      // ส่ง svcCust แทน repoCust
     svc := service.NewOrderService(m.mCtx.Transactor, svcCust, repoOrder, svcNoti)
     hdl := handler.NewOrderHandler(svc)
    
     orders := router.Group("/orders")
     orders.Post("", hdl.CreateOrder)
     orders.Delete("/:orderID", hdl.CancelOrder)
    }
    
    ```

### Nested Transactions

เมื่อทดสอบรันโปรแกรมใหม่อีกครั้ง แล้วลองสร้างออเดอร์ใหม่ จะได้รับ error ว่า

```go
HTTP/1.1 500 Internal Server Error
Date: Fri, 06 Jun 2025 03:50:32 GMT
Content-Type: application/json
Content-Length: 106
X-Request-Id: 771c27c3-7528-4b93-bc6d-e1696c4727ae
Connection: close

{
  "type": "operation_failed",
  "message": "failed to begin transaction: nested transactions are not supported"
}
```

เนื่องจากในฟังก์ชันการตัดยอด และคืนยอด credit ใน `CustomerService` นั้น มีการเปิดใช้งาน transaction ขึ้นมาใหม่ ซึ่งที่ถูกต้องจะต้องเป็น transaction เดียวกันที่ได้มาจาก `OrderService`

ดังนั้น ตอนสร้าง `transactor` ใน `main.go` ต้องระบุว่าด้วยว่าให้มีการใช้งาน nested transactions

```go
// src/app/cmd/api/main.go

func main() {
 // ...

 app := application.New(*config)

 transactor, dbCtx := transactor.New(
  db.DB(),
  // <-- เพิ่มใช้งาน nested transaction strategy ที่ใช้ Savepoints
  transactor.WithNestedTransactionStrategy(transactor.NestedTransactionsSavepoints))
 mCtx := module.NewModuleContext(transactor, dbCtx)
 
 // ...
}
```

## จัดการ Service ใน Monolith ด้วย Service Registry

ในระบบปัจจุบัน เรามักจะ **สร้าง service เดิมซ้ำ ๆ** ในแต่ละโมดูล เช่น สร้าง `NotificationService` ใหม่ในทั้ง `Customer` และ `Order` ทั้งที่ความจริงแล้วสามารถใช้ตัวเดียวกันได้

เพื่อแก้ปัญหานี้ เราจะใช้แนวคิด **Service Registry** หรือ “คลังเก็บ service” ไว้รวม service ทั้งหมดในระบบไว้ที่เดียว

### **ข้อดีของการใช้ Service Registry**

- ไม่ต้องสร้าง service ซ้ำในหลายที่
- ลดความซับซ้อนของการ inject dependency
- ทำให้การจัดการ service เป็นระบบระเบียบมากขึ้น
- รองรับการใช้ร่วมกันในหลายโมดูลง่ายขึ้น

### สร้าง Service Registry

Service Registry คือ struct หรือ container ที่ทำหน้าที่ **เก็บ instance ของ service ต่าง ๆ** ไว้ให้พร้อมใช้งานตลอดเวลา โดยไม่ต้อง new ซ้ำ

- สร้าง Service Registry

    > สร้างไฟล์ `util/registry/service_registry.go`
    >

    ```go
    package registry
    
    import "fmt"
    
    // สำหรับกำหนด key ของ service ที่จะ export
    type ServiceKey string
    
    // สำหรับ map key กับ service ที่จะ export
    type ProvidedService struct {
     Key   ServiceKey
     Value any
    }
    
    type ServiceRegistry interface {
     Register(key ServiceKey, svc any)
     Resolve(key ServiceKey) (any, error)
    }
    
    type serviceRegistry struct {
     services map[ServiceKey]any
    }
    
    func NewServiceRegistry() ServiceRegistry {
     return &serviceRegistry{
      services: make(map[ServiceKey]any),
     }
    }
    
    func (r *serviceRegistry) Register(key ServiceKey, svc any) {
     r.services[key] = svc
    }
    
    func (r *serviceRegistry) Resolve(key ServiceKey) (any, error) {
     svc, ok := r.services[key]
     if !ok {
      return nil, fmt.Errorf("service not found: %s", key)
     }
     return svc, nil
    }
    
    ```

- สร้างฟังก์ชันสำหรับช่วยแปลง Service กลับมาให้ถูกต้อง

    > สร้างไฟล์ `util/registry/helper.go`
    >

    ```go
    package registry
    
    import "fmt"
    
    func ResolveAs[T any](r ServiceRegistry, key ServiceKey) (T, error) {
     var zero T
     svc, err := r.Resolve(key)
     if err != nil {
      return zero, err
     }
     typedSvc, ok := svc.(T)
     if !ok {
      return zero, fmt.Errorf("service registered under key %s does not implement the expected type", key)
     }
     return typedSvc, nil
    }
    ```

### แก้ไข Module Interface

แก้ไขให้ Module มีฟังก์ชันสำหรับ เพิ่ม service ของตัวเองเข้า Registry

> แก้ไขไฟล์ `util/module/module.go`
>

```go
package module

import (
 "go-mma/util/registry"    // <-- ตรงนี้
 "go-mma/util/transactor"

 "github.com/gofiber/fiber/v3"
)

type Module interface {
 APIVersion() string
 Init(reg registry.ServiceRegistry) error // <-- ตรงนี้
 RegisterRoutes(r fiber.Router)
}

// <-- ตรงนี้
// แยกออกมาเพราะว่า บางโมดูลอาจไม่ต้อง export service
type ServiceProvider interface {
 Services() []registry.ProvidedService
}
```

### แก้ไข Application

แก้ไข Application ให้เป็นที่เก็บ service registry

> แก้ไขไฟล์ `application/application.go`
>

```go
package application

import (
 "fmt"
 "go-mma/config"
 "go-mma/data/sqldb"
 "go-mma/util/logger"
 "go-mma/util/module"
 "go-mma/util/registry"  // <-- ตรงนี้
)

type Application struct {
 config          config.Config
 httpServer      HTTPServer
 serviceRegistry registry.ServiceRegistry // <-- ตรงนี้
}

func New(config config.Config, db sqldb.DBContext) *Application {
 return &Application{
  config:          config,
  httpServer:      newHTTPServer(config),
  serviceRegistry: registry.NewServiceRegistry(), // <-- ตรงนี้
 }
}

// ...

func (app *Application) RegisterModules(modules ...module.Module) error {
 for _, m := range modules {
  // Initialize each module
  if err := app.initModule(m); err != nil {
   return fmt.Errorf("failed to init module [%T]: %w", m, err)
  }

  // ถ้าโมดูลเป็น ServiceProvider ให้เอา service มาลง registry
  if sp, ok := m.(module.ServiceProvider); ok {
   for _, p := range sp.Services() {
    app.serviceRegistry.Register(p.Key, p.Value)
   }
  }

  // Register routes for each module
  app.registerModuleRoutes(m)
 }

 return nil
}

func (app *Application) initModule(m module.Module) error {
 return m.Init(app.serviceRegistry)
}

// ...
```

### เพิ่มการ Initialize แต่ละโมดูล

ปรับให้แต่ละโมดูลเพิ่ม `Init()` เพื่อสร้าง service ของตัวเอง

- แก้ไขไฟล์ `modules/notification/module.go`

    ```go
    package notification
    
    import (
     "go-mma/modules/notification/service"
     "go-mma/util/module"
     "go-mma/util/registry"
    
     "github.com/gofiber/fiber/v3"
    )
    
    const (
     NotificationServiceKey registry.ServiceKey = "NotificationService"
    )
    
    func NewModule(mCtx *module.ModuleContext) module.Module {
     return &moduleImp{mCtx: mCtx}
    }
    
    type moduleImp struct {
     mCtx    *module.ModuleContext
     notiSvc service.NotificationService
    }
    
    func (m *moduleImp) APIVersion() string {
     return "v1"
    }
    
    func (m *moduleImp) Init(reg registry.ServiceRegistry) error {
     m.notiSvc = service.NewNotificationService()
    
     return nil
    }
    
    func (m *moduleImp) Services() []registry.ProvidedService {
     return []registry.ProvidedService{
      {Key: NotificationServiceKey, Value: m.notiSvc},
     }
    }
    
    func (m *moduleImp) RegisterRoutes(router fiber.Router) {
     // โมดูลนี้ยังไม่มี routes
    }
    ```

- แก้ไขไฟล์ `modules/customer/module.go`

    ```go
    package customer
    
    import (
     "go-mma/modules/customer/handler"
     "go-mma/modules/customer/repository"
     "go-mma/modules/customer/service"
     "go-mma/util/module"
     "go-mma/util/registry"
    
     notiModule "go-mma/modules/notification"
     notiService "go-mma/modules/notification/service"
    
     "github.com/gofiber/fiber/v3"
    )
    
    const (
     CustomerServiceKey registry.ServiceKey = "CustomerService"
    )
    
    func NewModule(mCtx *module.ModuleContext) module.Module {
     return &moduleImp{mCtx: mCtx}
    }
    
    type moduleImp struct {
     mCtx    *module.ModuleContext
     custSvc service.CustomerService
    }
    
    func (m *moduleImp) APIVersion() string {
     return "v1"
    }
    
    func (m *moduleImp) Init(reg registry.ServiceRegistry) error {
     // Resolve NotificationService from the registry
     notiSvc, err := registry.ResolveAs[notiService.NotificationService](reg, notiModule.NotificationServiceKey)
     if err != nil {
      return err
     }
    
     repo := repository.NewCustomerRepository(m.mCtx.DBCtx)
     m.custSvc = service.NewCustomerService(m.mCtx.Transactor, repo, notiSvc)
    
     return nil
    }
    
    func (m *moduleImp) Services() []registry.ProvidedService {
     return []registry.ProvidedService{
      {Key: CustomerServiceKey, Value: m.custSvc},
     }
    }
    
    func (m *moduleImp) RegisterRoutes(router fiber.Router) {
     // wiring dependencies
     hdl := handler.NewCustomerHandler(m.custSvc)
    
     customers := router.Group("/customers")
     customers.Post("", hdl.CreateCustomer)
    }
    ```

    <aside>
    💡

    ทำไมถึงสร้าง handler ใน `RegisterRoutes`

  - แยก concern ชัด: `RegisterRoutes` ดูแล “transport layer” ทั้งหมดในฟังก์ชันเดียว
  - อ่านง่าย: เห็นเส้นทางและ handler คู่กันทันที
  - ใช้ที่เดียว: ไม่มี state เพิ่มใน `moduleImp`
    </aside>

- แก้ไขไฟล์ `modules/order/module.go`

    ```go
    package order
    
    import (
     "go-mma/modules/order/handler"
     "go-mma/modules/order/repository"
     "go-mma/modules/order/service"
     "go-mma/util/module"
     "go-mma/util/registry"
    
     custModule "go-mma/modules/customer"
     custService "go-mma/modules/customer/service"
     notiModule "go-mma/modules/notification"
     notiService "go-mma/modules/notification/service"
    
     "github.com/gofiber/fiber/v3"
    )
    
    func NewModule(mCtx *module.ModuleContext) module.Module {
     return &moduleImp{mCtx: mCtx}
    }
    
    type moduleImp struct {
     mCtx     *module.ModuleContext
     orderSvc service.OrderService
    }
    
    func (m *moduleImp) APIVersion() string {
     return "v1"
    }
    
    func (m *moduleImp) Init(reg registry.ServiceRegistry) error {
     // Resolve CustomerService from the registry
     custSvc, err := registry.ResolveAs[custService.CustomerService](reg, custModule.CustomerServiceKey)
     if err != nil {
      return err
     }
    
     // Resolve NotificationService from the registry
     notiSvc, err := registry.ResolveAs[notiService.NotificationService](reg, notiModule.NotificationServiceKey)
     if err != nil {
      return err
     }
    
     repo := repository.NewOrderRepository(m.mCtx.DBCtx)
     m.orderSvc = service.NewOrderService(m.mCtx.Transactor, custSvc, repo, notiSvc)
    
     return nil
    }
    
    func (m *moduleImp) RegisterRoutes(router fiber.Router) {
     // wiring dependencies
     hdl := handler.NewOrderHandler(m.orderSvc)
    
     orders := router.Group("/orders")
     orders.Post("", hdl.CreateOrder)
     orders.Delete("/:orderID", hdl.CancelOrder)
    }
    ```

## ป้องกันการเข้าถึงข้ามโมดูลด้วยโฟลเดอร์ `internal`

หลังจากที่เรา แยกขอบเขตของแต่ละ sub-domain (Encapsulation) แล้ว ปัญหาที่ยังเหลือคือ โค้ดในโมดูล order ยังสามารถ `import` `model` หรือ `repository` ของโมดูล customer ได้โดยตรง นั่นทำให้ละเมิดขอบเขต (boundary) ของโดเมนและสร้างความพึ่งพา (coupling) ที่ไม่พึงประสงค์

ในภาษา Go สามารถย้ายไฟล์ที่ “ห้ามภายนอกใช้” เข้าไปไว้ภายใต้โฟลเดอร์ **`internal`** ได้ ตัวคอมไพเลอร์จะบังคับไม่ให้ path นอกโฟลเดอร์แม่ (root) ของ `internal` ทำ `import` ได้เลย

```go
customer/
├── internal/
│   ├── model/        // โครงสร้างข้อมูลเฉพาะ customer
│   └── repository/   // DB logic ของ customer
└── service/          // business logic (export)
```

ถ้าโมดูลอื่น เช่น `order` พยายาม `import "go-mma/modules/customer/internal/repository"` จะขึ้นข้อความ error แบบนี้

```go
could not import go-mma/modules/customer/internal/repository (invalid use of internal package "go-mma/modules/customer/internal/repository")
```

## รวมโค้ดทั้งหมดไว้ใน Mono-Repository อย่างเป็นระบบ

เพื่อให้การพัฒนาเป็นระบบมากขึ้น และลดความซับซ้อนเมื่อโปรเจกต์เติบโต เราแนะนำให้แยกแต่ละโมดูลหลัก เช่น `customer`, `order`, และ `notification` ออกเป็น **โปรเจกต์ย่อย (submodule)** โดยแต่ละโปรเจกต์จะมี `go.mod` ของตัวเอง

แต่ยังคงเก็บทุกอย่างไว้ใน Git repository เดียวกัน เรียกว่า **Mono-Repository**

**โดยจะแบ่งโค้ดออกเป็น 3 ส่วน หลักๆ คือ**

- **app:** สำหรับโหลดโมดูล และรันโปรแกรม
- **modules:** สำหรับสร้างโมดูลต่างๆ
- **shared:** สำหรับโค้ดที่ใช้งานร่วมกัน

### โครงสร้างใหม่

```bash
.
├── docker-compose.dev.yml
├── docker-compose.yml
├── go-mma.code-workspace
├── Makefile
├── migrations
│   ├── 20250529103238_create_customer.down.sql
│   ├── 20250529103238_create_customer.up.sql
│   ├── 20250529103715_create_order.down.sql
│   └── 20250529103715_create_order.up.sql
└── src
    ├── app
    │   ├── application
    │   │   ├── application.go
    │   │   ├── http.go
    │   │   └── middleware.go
    │   │   │   ├── request_logger.go
    │   │   │   └── response_error.go
    │   ├── cmd
    │   │   └── api
    │   │       └── main.go
    │   ├── config
    │   │   └── config.go
    │   ├── go.mod
    │   ├── go.sum
    │   └── util
    │       └── env
    │           └── env.go
    ├── modules
    │   ├── customers
    │   │   ├── dtos
    │   │   │   ├── customer_request.go
    │   │   │   ├── customer_response.go
    │   │   │   └── customer.go
    │   │   ├── handler
    │   │   │   └── customer.go
    │   │   ├── internal
    │   │   │   ├── model
    │   │   │   │   └── customer.go
    │   │   │   └── repository
    │   │   │       └── customer.go
    │   │   ├── module.go
    │   │   ├── service
    │   │   │   └── customer.go
    │   │   ├── test
    │   │   │   └── customer.http
    │   │   ├── go.mod
    │   │   └── go.sum
    │   ├── notifications
    │   │   ├── module.go
    │   │   ├── service
    │   │   │   └── notification.go
    │   │   ├── go.mod
    │   │   └── go.sum
    │   └── orders
    │   │   ├── dtos
    │   │   │   ├── order_request.go
    │   │   │   └── order_response.go
    │   │   ├── handler
    │   │   │   └── order.go
    │   │   ├── internal
    │   │   │   ├── model
    │   │   │   │   └── order.go
    │   │   │   └── repository
    │   │   │       └── order.go
    │   │   ├── module.go
    │   │   ├── service
    │   │   │   └── order.go
    │   │   ├── test
    │   │   │   └── order.http
    │   │   ├── go.mod
    │   │   └── go.sum
    └── shared
        └──common
            ├── errs
            │   ├── errs.go
            │   ├── helper.go
            │   └── types.go
            ├── logger
            │   └── logger.go
            ├── module
            │   └── module.go
            ├── registry
            │   ├── helper.go
            │   └── service_registry.go
            ├── storage
            │   └── db
            │       ├── db.go
            │       └── transactor
            │           ├── nested_transactions_none.go
            │           ├── nested_transactions_savepoints.go
            │           ├── transactor.go
            │           └── types.go
            ├── go.mod
            └── go.sum
```

**ข้อดีของการจัดแบบนี้**

- **แยกขอบเขตชัดเจน**: แต่ละโมดูลพัฒนาและทดสอบได้แบบอิสระ
- **ควบคุมเวอร์ชันได้ง่าย**: ใช้ `go.mod` จัดการ dependency ภายในแต่ละโมดูล
- **รวมศูนย์การจัดการ**: ยังใช้ Git ร่วมกันใน repo เดียว ไม่ต้องแยกหลาย repo
- **พร้อมสำหรับการแยกเป็น microservice ในอนาคต**: โครงสร้างรองรับการแยก deploy ได้หากจำเป็น

### สร้างโปรเจกต์ใหม่

- สร้าง Folder ใหม่ ดังนี้

    ```bash
    # อยู่ที่ root project
    mkdir -p src/app
    mkdir -p src/modules/customer
    mkdir -p src/modules/notification
    mkdir -p src/modules/order
    mkdir -p src/shared/common
    ```

- สร้าง app โปรเจกต์

    ```bash
    # อยู่ที่ root project
    cd src/app
    go mod init go-mma
    ```

- สร้าง customer โปรเจกต์

    ```bash
    cd ../..
    # อยู่ที่ root project
    cd src/modules/customer
    go mod init go-mma/modules/customer
    ```

- สร้าง notification โปรเจกต์

    ```bash
    cd ../../..
    # อยู่ที่ root project
    cd src/modules/notification
    go mod init go-mma/modules/notification
    ```

- สร้าง order โปรเจกต์

    ```bash
    cd ../../..
    # อยู่ที่ root project
    cd src/modules/order
    go mod init go-mma/modules/order
    ```

- สร้าง common โปรเจค

    ```bash
    cd ../../..
    # อยู่ที่ root project
    cd src/shared/common
    go mod init go-mma/shared/common
    ```

### ทำ L**ocal module replacement**

เมื่อพัฒนาแบบ Monorepo เพื่อให้แต่ละโมดูลสามารถอ้างถึงกันได้โดยตรงจากไฟล์ในเครื่อง โดยไม่ต้อง publish ไปที่ remote repo ใน Go ทำได้โดย การใช้คำสั่ง `replace` ใน `go.mod`

- โปรเจกต์ notification มีการใช้งาน common

    > แก้ไขไฟล์ `notification/go.mod`
    >

    ```bash
    module go-mma/modules/notification
    
    go 1.24.1
    
    replace go-mma/shared/common v0.0.0 => ../../shared/common
    ```

- โปรเจกต์ customer มีการใช้งาน common, notification

    > แก้ไขไฟล์ `customer/go.mod`
    >

    ```bash
    module go-mma/modules/customer
    
    go 1.24.1
    
    replace go-mma/shared/common v0.0.0 => ../../shared/common
    
    replace go-mma/modules/notification v0.0.0 => ../../modules/notification
    ```

- โปรเจกต์ order มีการใช้งาน common, notification, customer

    > แก้ไขไฟล์ `order/go.mod`
    >

    ```bash
    module go-mma/modules/order
    
    go 1.24.1
    
    replace go-mma/shared/common v0.0.0 => ../../shared/common
    
    replace go-mma/modules/notification v0.0.0 => ../../modules/notification
    
    replace go-mma/modules/customer v0.0.0 => ../../modules/customer
    ```

- โปรเจกต์ app มีการใช้งาน common, notification, customer, order

    แก้ไขไฟล์ `app/go.mod`

    ```bash
    module go-mma
    
    go 1.24.1
    
    replace go-mma/shared/common v0.0.0 => ../shared/common
    
    replace go-mma/modules/notification v0.0.0 => ../modules/notification
    
    replace go-mma/modules/customer v0.0.0 => ../modules/customer
    
    replace go-mma/modules/order v0.0.0 => ../modules/order
    ```

### สร้าง VS Code Workspace

สำหรับการทำ Mono-Repo ใน VS Code ต้องเปิดแบบ Workspace ถึงจะสามารถทำงานได้ถูกต้อง

- สร้างไฟล์ `go-mma.code-workspace`

    ```bash
    {
     "folders": [
      {
       "path": "."
      },
      {
       "path": "src/app"
      },
      {
       "path": "src/modules/customer"
      },
      {
       "path": "src/modules/order"
      },
      {
       "path": "src/modules/notification"
      },
      {
       "path": "src/shared/common"
      }
     ],
     "settings": {}
    }
    ```

- เลือกที่เมนู File เลือก Open Workspace from file…
- เลือกที่ไฟล์ `go-mma.code-workspace`
- กด Open
- ใน Explorer จะแสดง แบบนี้

    ```bash
    go-mma
    app
    customer
    order
    notification
    common
    ```

### โปรเจกต์ common

- ให้ทำการย้ายโค้ดใน `util` ทั้งหมด ยกเว้น `env` มาไว้ในโปรเจกต์ `common`

    ```bash
    common
    ├── go.mod
    ├── errs
    │   ├── errs.go
    │   ├── helpers.go
    │   └── types.go
    ├── idgen
    │   └── idgen.go
    ├── logger
    │   └── logger.go
    ├── module
    │   └── module.go
    ├── registry
    │   ├── helper.go
    │   └── service_registry.go
    └── storage
        └── sqldb
            ├── sqldb.go
            └── transactor
                ├── nested_transactions_none.go
                ├── nested_transactions_savepoints.go
                ├── transactor.go
                └── types.go
    ```

- ติดตั้ง dependencies ด้วย `go mod tidy`

### โปรเจกต์ notification

- ให้ทำการย้ายโค้ดใน `modules/notification` ทั้งหมด  มาไว้ในโปรเจกต์ `notification`

    ```bash
    notification
    ├── go.mod
    ├── go.sum
    ├── module.go
    └── service
        └── notification.go
    ```

- แก้ไข path ของการ `import` ดังนี้
  - `go-mma/util/logger` → `go-mma/shared/common/logger`
- ติดตั้ง dependencies ด้วย `go mod tidy`

### โปรเจกต์ customer

- ให้ทำการย้ายโค้ดใน `modules/customer` ทั้งหมด  มาไว้ในโปรเจกต์ `customer`

    ```bash
    customer
    ├── dto
    │   ├── customer_request.go
    │   ├── customer_response.go
    │   └── customer.go
    ├── go.mod
    ├── go.sum
    ├── handler
    │   └── customer.go
    ├── internal
    │   ├── model
    │   │   └── customer.go
    │   └── repository
    │       └── customer.go
    ├── module.go
    ├── service
    │   └── customer.go
    └── test
        └── customers.http
    ```

- แก้ไข path ของการ `import` ดังนี้
  - `go-mma/util/errs` → `go-mma/shared/common/errs`
  - `go-mma/util/logger` → `go-mma/shared/common/logger`
  - `go-mma/util/module` → `go-mma/shared/common/module`
  - `go-mma/util/registry` → `go-mma/shared/common/registry`
  - `go-mma/util/storage/sqldb/transactor` → `go-mma/shared/common/storage/sqldb/transactor`
- ติดตั้ง dependencies ด้วย `go mod tidy`

### โปรเจกต์ order

- ให้ทำการย้ายโค้ดใน `modules/order` ทั้งหมด  มาไว้ในโปรเจกต์ `order`

    ```bash
    order
    ├── dto
    │   ├── customer_request.go
    │   ├── customer_response.go
    │   └── customer.go
    ├── go.mod
    ├── go.sum
    ├── handler
    │   └── customer.go
    ├── internal
    │   ├── model
    │   │   └── customer.go
    │   └── repository
    │       └── customer.go
    ├── module.go
    ├── service
    │   └── customer.go
    └── test
        └── customers.http
    ```

- แก้ไข path ของการ `import` ดังนี้
  - `go-mma/util/errs` → `go-mma/shared/common/errs`
  - `go-mma/util/logger` → `go-mma/shared/common/logger`
  - `go-mma/util/module` → `go-mma/shared/common/module`
  - `go-mma/util/registry` → `go-mma/shared/common/registry`
  - `go-mma/util/storage/sqldb/transactor` → `go-mma/shared/common/storage/sqldb/transactor`
- ติดตั้ง dependencies ด้วย `go mod tidy`

### โปรเจกต์ app

- ให้ทำการย้ายโค้ดใน `application`, `cmd`, `config` และ `util`   มาไว้ในโปรเจกต์ `app`

    ```bash
    app
    ├── application
    │   ├── application.go
    │   ├── http.go
    │   └── middleware
    │       ├── request_logger.go
    │       └── response_error.go
    ├── cmd
    │   └── api
    │       └── main.go
    ├── config
    │   └── config.go
    ├── go.mod
    ├── go.sum
    └── util
        └── env
            └── env.go
    ```

- แก้ไข path ของการ `import` ดังนี้
  - `go-mma/util/errs` → `go-mma/shared/common/errs`
  - `go-mma/util/logger` → `go-mma/shared/common/logger`
  - `go-mma/util/module` → `go-mma/shared/common/module`
  - `go-mma/util/registry` → `go-mma/shared/common/registry`
  - `go-mma/util/storage/sqldb` → `go-mma/shared/common/storage/sqldb`
  - `go-mma/util/storage/sqldb/transactor` → `go-mma/shared/common/storage/sqldb/transactor`
- ติดตั้ง dependencies ด้วย `go mod tidy`

### รันโปรแกรม

- แก้ไฟล์ `Makefile` เพื่อแก้ path ในการรัน

    > แก้ไขไฟล์ `main.go`
    >

    ```bash
    .PHONY: run
    run:
     cd src/app && \
     go run cmd/api/main.go
    ```

- ทดลองรันโปรแกรม

    ```bash
    make run
    ```

### Build โปรแกรม

- แก้ไฟล์ `Makefile` เพื่อแก้ path ในการ build

    > แก้ไขไฟล์ `main.go`

    ```bash
    .PHONY: build
    build:
        cd src/app && \
        go build -ldflags \
        "-s -w \
        -X 'go-mma/build.Version=${BUILD_VERSION}' \
        -X 'go-mma/build.Time=${BUILD_TIME}'" \
        -o ../../app cmd/api/main.go
    ```

- ทดลองรัน build

    ```bash
    BUILD_VERSION=0.0.2 make build
    ```

### Build โปรแกรมเป็น Docker image

- แก้ไฟล์ `Dockerfile` เพื่อแก้ path ในการ build

    > แก้ไขไฟล์ `Dockerfile`

    ```docker
    # Builder stage
    FROM golang:1.24-alpine AS builder
    WORKDIR /app
    COPY . .
    RUN cd src/app && go mod download

    # ปิด cgo (CGO_ENABLED=0)

    ENV CGO_ENABLED=0 \
        GOOS=linux \
        GOARCH=amd64

    # ตั้งค่า default สำหรับ VERSION

    ARG VERSION=latest
    ENV IMAGE_VERSION=${VERSION}
    RUN echo "Build version: $IMAGE_VERSION"
    RUN cd src/app && \
    go build -ldflags \
    # strip debugging information ออกจาก binary ทำให้ไฟล์เล็กลง
    "-s -w \
    -X 'go-mma/build.Version=${IMAGE_VERSION}' \
    -X 'go-mma/build.Time=$(date +"%Y-%m-%dT%H:%M:%S%z")'" \
    -o /app/app cmd/api/main.go

    # Final stage

    FROM alpine:3.22
    RUN apk add --no-cache ca-certificates tzdata && \
    # สร้าง non-root user
        addgroup -S appgroup && adduser -S appuser -G appgroup
    
    WORKDIR /app
    COPY --from=builder /app/app .
    USER appuser
    EXPOSE 8090
    ENV TZ=Asia/Bangkok

    # ใช้ ENTRYPOINT เพื่อล็อกว่าต้องรันไฟล์ไหน
    ENTRYPOINT ["./app"]
    ```

- ทดลองรัน build

    ```bash
    BUILD_VERSION=0.0.2 make image
    ```

## กำหนด Public API Contract ระหว่างโมดูล

ในโค้ดปัจจุบัน โมดูล `order` สามารถเรียกใช้ `CustomerService` ของโมดูล `customer` ได้โดยตรง ซึ่ง **ไม่เหมาะสม** เพราะเท่ากับว่าโมดูล `order` รู้รายละเอียดภายในทั้งหมดของ `customer` ซึ่งจะทำให้ระบบขาดความยืดหยุ่นและยากต่อการบำรุงรักษา

แนวทางที่ถูกต้องคือโมดูลแต่ละตัวควรเปิดเผยเฉพาะ interface ที่จำเป็นต่อการใช้งานจากภายนอกเท่านั้น ซึ่งเรียกว่า Public API Contract ซึ่งมีลักษณะ ดังนี้

- ระบุว่า **โมดูลนี้ให้บริการอะไรบ้าง**
- กำหนด **รูปแบบการเรียกใช้** (method, input, output)
- **ซ่อนการทำงานภายใน** (encapsulation)

### ตัวอย่างแนวทางการออกแบบ

```bash
                        ┌────────────────────────────┐
                        │     customercontract       │
                        │ ┌────────────────────────┐ │
                        │ │  CreditManager         │ │
                        │ │                        │ │
                        │ │ + ReserveCredit()      │ │
                        │ │ + ReleaseCredit()      │ │
                        │ └────────────────────────┘ │
                        └────────────▲───────────────┘
                                     │
        implements                   │  depends on
                                     │
┌────────────────────┐     uses      │   ┌────────────────────┐
│     customer       │───────────────┘   │       order        │
│ ┌────────────────┐ │                   │ ┌─────────────────┐│
│ │ CustomerService│◄────────────────────┤ │ OrderService    ││
│ │ (implements    │ │                   │ │ (depends on     ││
│ │  CreditManager)│ │                   │ │  CreditManager) ││
│ └────────────────┘ │                   │ └─────────────────┘│
└────────────────────┘                   └────────────────────┘
```

### ประโยชน์ของการใช้ Public API Contract

- ลดการพึ่งพาภายใน (Loose Coupling)
- เปลี่ยนแปลงภายในได้อิสระ โดยไม่กระทบโมดูลอื่น
- รองรับการทดสอบง่ายขึ้น (mock ได้)
- เตรียมพร้อมสำหรับการแยกเป็น microservice หากจำเป็น

### สร้าง Customer Contract

`customercontract` เป็น โปรเจกต์กลาง ที่เก็บ public interfaces เช่น `CreditManager` ในการสร้างนั้นใช้ 2 หลักการนี้

1. **Interface Segregation Principle (ISP)** ใช้เพื่อแยก interface ของ `CustomerService` ให้เป็น interface ย่อยๆ
2. เนื่องจากเราทำเป็น mono-repo ดังนั้น จะ**สร้าง contract เป็นโปรเจกต์แยกออกมาจากโปรเจกต์โมดูล customer** เพราะว่า
    - Low Coupling: `order` ไม่ต้อง import logic หรือ dependency ของ `customer` โดยตรง
    - เปลี่ยน implementation ได้อิสระ: เปลี่ยน logic ภายใน `customer` โดยไม่กระทบ `order`
    - Encapsulation: ป้องกันการ import โค้ดภายใน customer ที่ไม่ได้ตั้งใจเปิดเผย

ขั้นตอนการสร้าง customer contract

- สร้างโปรเจกต์ใหม่

    ```bash
    mkdir -p src/shared/contract/customercontract
    cd src/shared/contract/customercontract
    go mod init go-mma/shared/contract/customercontract
    ```

- เพิ่มโปรเจกต์เข้า workspace

    > แก้ไขไฟล์ `go-mma.code-workspace`
    >

    ```bash
    {
      "folders": [
        {
          "path": "."
        },
        {
          "path": "src/app"
        },
        {
          "path": "src/modules/customer"
        },
        {
          "path": "src/shared/contract/customercontract"
        },
        {
          "path": "src/modules/order"
        },
        {
          "path": "src/modules/notification"
        },
        {
          "path": "src/shared/common"
        }
      ],
      "settings": {}
    }
    ```

- แก้ไขไฟล์ `go.mod`

    ```go
    module go-mma/shared/contract/customercontract
    
    go 1.24.1
    
    replace go-mma/shared/common v0.0.0 => ../../common
    ```

- สร้าง customer contract

    > สร้างไฟล์ `contract/customercontract/contract.go`
    >

    ```go
    package customercontract
    
    import (
     "context"
     "go-mma/shared/common/registry"
    )
    
    const (
     CreditManagerKey registry.ServiceKey = "customer:contract:credit"
    )
    
    type CustomerInfo struct {
     ID     int64    `json:"id"`
     Email  string   `json:"email"`
     Credit int      `json:"credit"`
    }
    
    func NewCustomerInfo(id int64, email string, credit int) *CustomerInfo {
     return &CustomerInfo{ID: id, Email: email, Credit: credit}
    }
    
    type CustomerReader interface {
     GetCustomerByID(ctx context.Context, id int64) (*CustomerInfo, error)
    }
    
    type CreditManager interface {
     CustomerReader // embed เพื่อ reuse
     ReserveCredit(ctx context.Context, id int64, amount int) error
     ReleaseCredit(ctx context.Context, id int64, amount int) error
    }
    ```

- ติดตั้ง dependencies ด้วย `go mod tidy`

### โมดูล Customer

ต้องปรับให้ `CustomerService` มา implement `customercontract`

- ทำ module replacement

    > แก้ไขไฟล์ `customer/go.mod`
    >

    ```go
    module go-mma/modules/customer
    
    go 1.24.1
    
    replace go-mma/shared/common v0.0.0 => ../../shared/common
    
    replace go-mma/modules/notification v0.0.0 => ../../modules/notification
    
    replace go-mma/shared/contract/customercontract v0.0.0 => ../../shared/contract/customercontract
    
    require (
     github.com/gofiber/fiber/v3 v3.0.0-beta.4
     go-mma/modules/notification v0.0.0
     go-mma/shared/common v0.0.0
     go-mma/shared/contract/customercontract v0.0.0
    )
    
    // ...
    ```

- ปรับให้ `CustomerService` มา implement `customercontract`

    > แก้ไขไฟล์ `src/modules/customer/service/customer.go`
    >

    ```go
    package service
    
    import (
     "context"
    
      "go-mma/modules/customer/dto"
     "go-mma/modules/customer/internal/model"
     "go-mma/modules/customer/internal/repository"
     "go-mma/shared/common/errs"
     "go-mma/shared/common/logger"
     "go-mma/shared/common/storage/sqldb/transactor"
     "go-mma/shared/contract/customercontract"  // <-- ตรงนี้
    
     notiService "go-mma/modules/notification/service"
    )
    
    // ...
    
    type CustomerService interface {
     CreateCustomer(ctx context.Context, req *dto.CreateCustomerRequest) (*dto.CreateCustomerResponse, error)
     customercontract.CreditManager   // <-- implement customercontract
    }
    
    // ...
    
    func (s *customerService) CreateCustomer(ctx context.Context, req *customercontract.CreateCustomerRequest) (*customercontract.CreateCustomerResponse, error) { // <-- แก้ให้ return เป็น contract แทน dto
     // ...
     
     // <-- เปลี่ยนมาสร้าง Response จาก contract
     return customercontract.NewCustomerInfo(customer.ID, customer.Email, customer.Credit), nil
    }
    ```

- เปลี่ยนส่งออก service ด้วย key `customercontract.CreditManagerKey`

    > แก้ไขไฟล์ `customer/module.go`
    >

    ```go
    // ยกเลิกการใช้งาน
    // const (
    //  CustomerServiceKey registry.ServiceKey = "CustomerService"
    // )
    
    func (m *moduleImp) Services() []registry.ProvidedService {
     return []registry.ProvidedService{
       // เปลี่ยน key มาจาก Contract แทน
      {Key: customercontract.CreditManagerKey, Value: m.custSvc},
     }
    }
    ```

### โมดูล Order

รู้จักแค่ interface `CreditManager` ที่มาจาก `customercontract`

- ทำ module replacement

    > แก้ไขไฟล์  `order/go.mod`
    >

    ```go
    module go-mma/modules/order
    
    go 1.24.4
    
    replace go-mma/shared/common v0.0.0 => ../../shared/common
    
    replace go-mma/modules/notification v0.0.0 => ../../modules/notification
    
    replace go-mma/modules/customer v0.0.0 => ../../modules/customer
    
    replace go-mma/shared/contract/customercontract v0.0.0 => ../../shared/contract/customercontract
    
    // ...
    ```

- เปลี่ยนมาเรียกใช้ `customercontract` แทนการเรียกใช้ `CustomerService` ตรงๆ

    > แก้ไขไฟล์ `order/service/order.go`
    >

    ```go
    package service
    
    import (
     "context"
     "go-mma/modules/order/dto"
     "go-mma/modules/order/internal/model"
     "go-mma/modules/order/internal/repository"
     "go-mma/shared/common/errs"
     "go-mma/shared/common/logger"
     "go-mma/shared/common/storage/sqldb/transactor"
     "go-mma/shared/contract/customercontract" // <-- ตรงนี้
    
     notiService "go-mma/modules/notification/service"
    )
    
    var (
     ErrNoOrderID = errs.ResourceNotFoundError("the order with given id was not found")
    )
    
    type OrderService interface {
     CreateOrder(ctx context.Context, req *dto.CreateOrderRequest) (*dto.CreateOrderResponse, error)
     CancelOrder(ctx context.Context, id int64) error
    }
    
    type orderService struct {
     transactor transactor.Transactor
     custSvc    customercontract.CreditManager // <-- ตรงนี้
     orderRepo  repository.OrderRepository
     notiSvc    notiService.NotificationService
    }
    
    func NewOrderService(
     transactor transactor.Transactor,
     custSvc customercontract.CreditManager, // <-- ตรงนี้
     orderRepo repository.OrderRepository,
     notiSvc notiService.NotificationService) OrderService {
     return &orderService{
      transactor: transactor,
      custSvc:    custSvc,
      orderRepo:  orderRepo,
      notiSvc:    notiSvc,
     }
    }
    
    // ...
    ```

    <aside>
    💡

    ถ้าลองดูที่ `custSvc` จะเห็นว่าสามารถเรียกใช้ได้แค่ 3 methods เท่าที่ contract ระบุไว้เท่านั้น

    </aside>

- เปลี่ยนการดึง Service จาก registry มาเป็น `customercontract.CreditManager` แทน `CustomerService`

    > แก้ไขไฟล์ `src/modules/order/module.go`
    >

    ```go
    package order
    
    import (
     "go-mma/modules/order/handler"
     "go-mma/modules/order/internal/repository"
     "go-mma/modules/order/service"
     "go-mma/shared/common/module"
     "go-mma/shared/common/registry"
     "go-mma/shared/contract/customercontract" // <-- ตรงนี้
    
     notiModule "go-mma/modules/notification"
     notiService "go-mma/modules/notification/service"
    
     "github.com/gofiber/fiber/v3"
    )
    
    // ...
    
    func (m *moduleImp) Init(reg registry.ServiceRegistry) error {
     // <-- ตรงนี้
     // Resolve CustomerService as CreditManagerKey from the registry
     custSvc, err := registry.ResolveAs[customercontract.CreditManager](reg, customercontract.CreditManagerKey)
     if err != nil {
      return err
     }
      
      // ...
    }
    ```

### รันโปรแกรม

ก่อนจะรันโปรแกรมต้องทำให้โปรเจกต์ `app` รู้จัก `customercontract`  ด้วย

- ทำ module replacement

    > แก้ไขไฟล์ `src/app/go.mod`
    >

    ```go
    module go-mma
    
    go 1.24.1
    
    replace go-mma/shared/common v0.0.0 => ../shared/common
    
    replace go-mma/modules/notification v0.0.0 => ../modules/notification
    
    replace go-mma/modules/customer v0.0.0 => ../modules/customer
    
    replace go-mma/modules/order v0.0.0 => ../modules/order
    
    replace go-mma/shared/contract/customercontract v0.0.0 => ../shared/contract/customer-contract
    
    // ...
    ```

- ติดตั้ง dependencies ด้วย `go mod tidy`
- รันโปรแกรม `make run`

## การแยกข้อมูลระหว่างโมดูล (Data Isolation)

หลังจากที่เราการกำหนดขอบเขตโมดูลและรูปแบบการสื่อสารที่ชัดเจนแล้ว สิ่งหนึ่งที่มองข้ามไม่ได้คือ **การแยกข้อมูลระหว่างโมดูล ([Data Isolation](https://somprasongd.work/blog/architecture/modular-monolith-data-isolation))** ซึ่งช่วยให้โมดูลมีความเป็นอิสระและมีการผูกแน่นที่หลวม (loosely coupled) สถาปัตยกรรมแบบ Modular Monolith มีกฎที่เข้มงวดสำหรับการรักษาความสมบูรณ์ของข้อมูล:

- แต่ละโมดูลสามารถเข้าถึงได้เฉพาะตารางของตนเองเท่านั้น
- ไม่มีการแชร์ตารางหรืออ็อบเจกต์ระหว่างโมดูล
- การ Join สามารถทำได้เฉพาะตารางภายในโมดูลเดียวกันเท่านั้น

ประโยชน์ของการออกแบบนี้คือการส่งเสริมความเป็นโมดูลาร์และการผูกแน่นที่หลวม ทำให้ง่ายต่อการเปลี่ยนแปลงระบบและลดผลข้างเคียงที่ไม่พึงประสงค์

### ระดับการแยกข้อมูล (Data Isolation Levels)

นี่คือสี่แนวทางในการแยกข้อมูลสำหรับ Modular Monoliths

- **Level 1 - Separate Table (ไม่แนะนำ)**

    เป็นวิธีการที่ง่ายที่สุดคือไม่มีการแยกข้อมูลในระดับฐานข้อมูล ทุกตารางของทุกโมดูลจะอยู่ในฐานข้อมูลเดียวกัน ทำให้ยากที่จะระบุว่าตารางใดเป็นของโมดูลใด

    ```bash
    ┌─────────────────────────────────┐
    │         Single Database         │
    │ ┌─────────────────────────────┐ │
    │ │   Module A - Schema public  │ │
    │ │  Table A1  Table A2  ...    │ │
    │ └─────────────────────────────┘ │
    │ ┌─────────────────────────────┐ │
    │ │   Module B - Schema public  │ │
    │ │  Table B1  Table B2  ...    │ │
    │ └─────────────────────────────┘ │
    │ ┌─────────────────────────────┐ │
    │ │   Module B - Schema public  │ │
    │ │  Table C1  Table C2  ...    │ │
    │ └─────────────────────────────┘ │
    └─────────────────────────────────┘
    ```

- **Level 2 - Separate Schema (Logical Isolation)**

    วิธีนี้เป็นการจัดกลุ่มตารางที่เกี่ยวข้องเข้าด้วยกันในฐานข้อมูลโดยใช้ **Database Schemas** แต่ละโมดูลมี Schema เฉพาะของตนเองที่มีตารางของโมดูลนั้นๆ ทำให้ง่ายต่อการจำแนกว่าตารางใดเป็นของโมดูลใด

    ```bash
    ┌─────────────────────────────────┐
    │         Single Database         │
    │ ┌─────────────────────────────┐ │
    │ │     Module A - Schema A     │ │
    │ │  Table A1  Table A2  ...    │ │
    │ └─────────────────────────────┘ │
    │ ┌─────────────────────────────┐ │
    │ │     Module B - Schema B     │ │
    │ │  Table B1  Table B2  ...    │ │
    │ └─────────────────────────────┘ │
    │ ┌─────────────────────────────┐ │
    │ │     Module C - Schema C     │ │
    │ │  Table C1  Table C2  ...    │ │
    │ └─────────────────────────────┘ │
    └─────────────────────────────────┘
    ```

- **Level 3 - Separate Database (Physical Isolation)**

    ระดับถัดไปคือการย้ายข้อมูลของแต่ละโมดูลไปยัง **ฐานข้อมูลที่แยกจากกัน** วิธีนี้มีการจำกัดมากกว่าการแยกข้อมูลด้วย Schema และเป็นทางเลือกที่ดีหากคุณต้องการกฎการแยกข้อมูลที่เข้มงวดระหว่างโมดูล

    ```bash
    ┌─────────────────┐   ┌─────────────────┐   ┌─────────────────┐
    │  Database A     │   │  Database B     │   │  Database C     │
    │ ┌─────────────┐ │   │ ┌─────────────┐ │   │ ┌─────────────┐ │
    │ │ Module A    │ │   │ │ Module B    │ │   │ │ Module C    │ │
    │ │ Table A1    │ │   │ │ Table B1    │ │   │ │ Table C1    │ │
    │ │ Table A2    │ │   │ │ Table B2    │ │   │ │ Table C2    │ │
    │ │     ...     │ │   │ │     ...     │ │   │ │     ...     │ │
    │ └─────────────┘ │   │ └─────────────┘ │   │ └─────────────┘ │
    └─────────────────┘   └─────────────────┘   └─────────────────┘
    ```

- **Level 4 - Different Persistence (Polyglot Persistence)**

    วิธีนี้ไปไกลกว่านั้นโดยการใช้ **ฐานข้อมูลประเภทที่แตกต่างกัน** สำหรับแต่ละโมดูล เพื่อแก้ปัญหาเฉพาะทางของแต่ละเรื่อง

    ```bash
    ┌─────────────────┐   ┌─────────────────┐   ┌─────────────────┐
    │  Database A     │   │  Database B     │   │  Database C     │
    │ (Relational DB) │   │ (Document DB)   │   │ (Graph DB)      │
    │ ┌─────────────┐ │   │ ┌─────────────┐ │   │ ┌─────────────┐ │
    │ │ Module A    │ │   │ │ Module B    │ │   │ │ Module C    │ │
    │ │ Table A1    │ │   │ │ Doc B1      │ │   │ │ Node C1     │ │
    │ │ Table A2    │ │   │ │ Doc B2      │ │   │ │ Edge C1     │ │
    │ │     ...     │ │   │ │     ...     │ │   │ │     ...     │ │
    │ └─────────────┘ │   │ └─────────────┘ │   │ └─────────────┘ │
    └─────────────────┘   └─────────────────┘   └─────────────────┘
    ```

### **Level 2 - Separate Schema (Logical Isolation)**

โดยทั่วไปแล้ว การเริ่มต้นด้วยการแยกข้อมูลแบบ Logical Isolation โดยใช้ Schemas เป็นวิธีที่ง่ายต่อการนำไปใช้และช่วยให้เข้าใจขอบเขตของระบบได้ดีขึ้น และคุณสามารถพิจารณาใช้ฐานข้อมูลแยกกัน (Separate Database) ในภายหลังได้ตามความต้องการของระบบที่เปลี่ยนแปลงไป

**ขั้นตอนการแยก schema**

- สร้างไฟล์ migration สำหรับการแยก schema

    ```bash
    make mgc filename=separate_schema
    ```

- เพิ่มคำสั่งสำหรับแยก schema

    > แก้ไขไฟล์ `xxx_separate_schema.up.sql`
    >

    ```sql
    -- สร้าง schema ใหม่สำหรับโมดูลลูกค้า
    CREATE SCHEMA customer;
    
    -- ย้ายตาราง 'customers' จาก schema 'public' ไปยัง schema 'customer'
    ALTER TABLE public.customers SET SCHEMA customer;
    
    -- สร้าง schema ใหม่สำหรับโมดูลคำสั่งซื้อ
    CREATE SCHEMA sales;
    
    -- ลบ Foreign Key Constraint เดิมก่อนย้ายตาราง
    ALTER TABLE public.orders DROP CONSTRAINT IF EXISTS fk_customer;
    
    -- ย้ายตาราง 'orders' จาก schema 'public' ไปยัง schema 'order'
    ALTER TABLE public.orders SET SCHEMA sales;
    
    -- หมายเหตุ: ตามหลัก Modular Monolith จะไม่มีการสร้าง Foreign Key ข้ามโมดูล
    -- การตรวจสอบความถูกต้องของ customer_id จะจัดการที่ระดับแอปพลิเคชัน
    ```

- เพิ่มคำสั่งสำหรับการย้อนกลับ

    > แก้ไขไฟล์ `xxx_separate_schema.down.sql`
    >

    ```sql
    -- ย้ายตาราง 'customers' กลับจาก schema 'customer' ไปยัง 'public'
    ALTER TABLE customer.customers SET SCHEMA public;
    
    -- ลบ schema 'customer' (จะสำเร็จเมื่อ schema ว่างเปล่า)
    DROP SCHEMA customer;
    
    -- ลบ Foreign Key Constraint บน 'sales.orders' ถ้ามีอยู่ (ป้องกันข้อผิดพลาด)
    ALTER TABLE sales.orders DROP CONSTRAINT IF EXISTS fk_customer;
    
    -- ย้ายตาราง 'orders' กลับจาก schema 'sales' ไปยัง 'public'
    ALTER TABLE sales.orders SET SCHEMA public;
    
    -- ลบ schema 'sales' (จะสำเร็จเมื่อ schema ว่างเปล่า)
    DROP SCHEMA sales;
    
    -- เพิ่ม Foreign Key Constraint กลับคืนที่ 'public.sales' โดยอ้างอิง 'public.customers'
    ALTER TABLE public.orders
    ADD CONSTRAINT fk_customer FOREIGN KEY (customer_id) REFERENCES public.customers(id);
    ```

- รันคำสั่ง `make mgu` เพิ่มสั่งสร้าง schema และย้ายตาราง
- ปรับปรุง `CustomerRepository` ให้ใช้ schema ที่ถูกต้อง `public` → `customer`

    > แก้ไขไฟล์ `customer/internal/repository/customer.go`
    >

    ```go
    package repository
    
    // ...
    
    func (r *customerRepository) Create(ctx context.Context, customer *model.Customer) error {
     query := `
     INSERT INTO customer.customers (id, email, credit)
     VALUES ($1, $2, $3)
     RETURNING *
     `
    
     // ...
     return nil
    }
    
    func (r *customerRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
     query := `SELECT 1 FROM customer.customers WHERE email = $1 LIMIT 1`
    
     // ...
     return true, nil
    }
    
    func (r *customerRepository) FindByID(ctx context.Context, id int64) (*model.Customer, error) {
     query := `
     SELECT *
     FROM customer.customers
     WHERE id = $1
    `
     // ...
     return &customer, nil
    }
    
    func (r *customerRepository) UpdateCredit(ctx context.Context, m *model.Customer) error {
     query := `
     UPDATE customer.customers
     SET credit = $2
     WHERE id = $1
     RETURNING *
    `
      // ...
     return nil
    }
    
    ```

- ปรับปรุง `OrderRepository` ให้ใช้ schema ที่ถูกต้อง `public` → `sales`

    > แก้ไขไฟล์ `order/internal/repository/order.go`
    >

    ```go
    package repository
    
    // ...
    
    func (r *orderRepository) FindByID(ctx context.Context, id int64) (*model.Order, error) {
     query := `
     SELECT *
     FROM sales.orders
     WHERE id = $1
     AND canceled_at IS NULL -- รายออเดอร์ต้องยังไม่ถูกยกเลิก
    `
     // ...
     return &order, nil
    }
    
    func (r *orderRepository) Cancel(ctx context.Context, id int64) error {
     query := `
     UPDATE sales.orders
     SET canceled_at = current_timestamp -- soft delete record
     WHERE id = $1
    `
     // ...
     return nil
    }
    
    ```

## การจัดการโมดูลด้วย Feature-Based Structure และ CQRS

ในโครงสร้างเดิม โมดูล `customer` มักรวมทุกฟังก์ชันไว้ภายใน interface เดียวคือ `CustomerService` ซึ่งเมื่อระบบโตขึ้น จะทำให้โค้ดเริ่มยุ่งเหยิงและยากต่อการดูแล

เพื่อให้โมดูลมีความ *แยกส่วน* (modular) และ *ปรับขยายง่าย* มากขึ้น เราจะเปลี่ยนไปใช้แนวทางใหม่ดังนี้

1. **CQRS (Command Query Responsibility Segregation)**

    แยกโค้ดที่ "เขียนข้อมูล" (Command) ออกจาก "อ่านข้อมูล" (Query) อย่างชัดเจน

    - เพิ่มความชัดเจนของ intent (เรากำลังอ่าน หรือเขียน?)
    - ปรับปรุง performance ฝั่ง read โดยไม่กระทบฝั่ง write
    - ให้แต่ละส่วนสามารถเปลี่ยนแปลง / ขยาย / ทดสอบ ได้อิสระ
2. **Mediator Pattern**

    ใช้ Mediator (หรือ Message Bus) เป็น **ตัวกลาง** ในการส่ง Command/Query แทนการเรียก Service ตรงๆ ช่วยลดการผูกติดกันระหว่าง component (decoupling) และรองรับ cross-cutting concerns เช่น logging, transaction, validation

    **ตัวอย่างการใช้งาน**

    ```go
    // customer/module.go
    
    // ลงทะเบียน feature handler สำหรับ query `GetCustomerByID`
    // โดยผูก handler กับ request type เพื่อให้ mediator เรียกใช้ได้ในภายหลัง
    mediator.Register(getbyid.NewGetCustomerByIDQueryHandler(repo))
    ```

    ```go
    // order/internal/feature/create/command_handler.go
    
    // ใช้ mediator เรียก query `GetCustomerByIDQuery` พร้อมกำหนด type ของ request และ response
    customer, err := mediator.Send[*customercontract.GetCustomerByIDQuery, *customercontract.GetCustomerByIDQueryResult](
     ctx,
     &customercontract.GetCustomerByIDQuery{ID: cmd.CustomerID}, // ส่ง request พร้อมข้อมูล customer ID
    )
    ```

### **สร้าง Mediator**

สร้างตัวจัดการ `Request/Response` ของแต่ละการเขียนข้อมูล (Command) และ อ่านข้อมูล (Query)

> สร้างไฟล์ `common/mediator/mediator.go`
>

```go
package mediator

import (
 "context"
 "errors"
 "fmt"
 "reflect"
)

// ใช้แทนกรณีไม่ต้องการ response ใด ๆ
type NoResponse struct{}

// Interface สำหรับ handler ที่รับ request และ return response
type RequestHandler[TRequest any, TResponse any] interface {
 Handle(ctx context.Context, request TRequest) (TResponse, error)
}

// registry สำหรับเก็บ handler ตาม type ของ request
var handlers = map[reflect.Type]func(ctx context.Context, req interface{}) (interface{}, error){}

// Register: ผูก handler กับ type ของ request ที่รองรับ
func Register[TRequest any, TResponse any](handler RequestHandler[TRequest, TResponse]) {
 var req TRequest // สร้าง zero value เพื่อใช้หา type
 reqType := reflect.TypeOf(req)

 // wrap handler ให้รองรับ interface{}
 handlers[reqType] = func(ctx context.Context, request interface{}) (interface{}, error) {
  typedReq, ok := request.(TRequest)
  if !ok {
   return nil, errors.New("invalid request type")
  }
  return handler.Handle(ctx, typedReq)
 }
}

// Send: dispatch request ไปยัง handler ที่ match กับ type ของ request
func Send[TRequest any, TResponse any](ctx context.Context, req TRequest) (TResponse, error) {
 reqType := reflect.TypeOf(req)
 handler, ok := handlers[reqType]
 if !ok {
  var empty TResponse
  return empty, fmt.Errorf("no handler for request %T", req)
 }

 result, err := handler(ctx, req)
 if err != nil {
  var empty TResponse
  return empty, err
 }

 // ตรวจสอบ type ของ response ก่อน return
 typedRes, ok := result.(TResponse)
 if !ok {
  var empty TResponse
  return empty, errors.New("invalid response type")
 }

 return typedRes, nil
}
```

### วิเคาระห์ Customer Features

นำฟังก์ชันทั้งหมดใน interface `CustomerService` เดิม มาแยกออกเป็น 1 ฟังก์ชัน 1 ฟีเจอร์ ได้ดังนี้

1. **create**: สร้างลูกค้าใหม่
2. **get-by-id**: ค้นหาลูกค้าจาก ID
3. **reserve-credit**: ตัดยอด credit
4. **release-credit**: คืนยอด credit

**โครงสร้างใหม่**

```bash
customer
├── domainerrors
│   └── domainerrors.go             # ไว้รวบรวม error ทั้่งหมด ของ customer
├── internal
│   ├── feature                     # สร้างใน internal ป้องกันไม่ให้ import
│   │   ├── create
│   │   │   ├── dto.go              # ย้าย dto มาที่นี่
│   │   │   ├── endpoint.go         # ย้าย http handler มาที่นี่
│   │   │   ├── command.go          # กำหนดรูปแบบของ Request/Response ของ command
│   │   │   └── command_handler.go  # จัดการ command handler
│   │   ├── get-by-id
│   │   │   └── query_handler.go    # จัดการ query handler
│   │   ├── release-credit
│   │   │   └── command_handler.go
│   │   └── reserve-credit
│   │       └── command_handler.go
│   ├── model
│   │   └── customer.go
│   └── repository
│       └── customer.go
├── test
│   └── customers.http
├── module.go          # เปลี่ยนจาก register service เป็น command/query handler แทน
├── go.mod
└── go.sum
```

### สร้าง Customer Domain Error

เริ่มจากรวบรวม error ที่จะเกิดขึ้นทั้งหมดจาก command handler, query handler และ rich model มาไว้ที่เดียวเพื่อสร้าง Domain error และใช้งานร่วมกันทุกฟีเจอร์

> สร้างไฟล์ `customer/domainerrors/domainerrors.go`
>

```go
package domainerrors

import "go-mma/shared/common/errs"

var (
 ErrEmailExists        = errs.ConflictError("email already exists")
 ErrCustomerNotFound   = errs.ResourceNotFoundError("the customer with given id was not found")
 ErrInsufficientCredit = errs.BusinessRuleError("insufficient credit")
)
```

### สร้างฟีเจอร์ **get-by-id**: ค้นหาลูกค้าจาก ID

ฟีเจอร์นี้ เป็นการค้นหาข้อมูลลูกค้า จัดเป็น Query ตาม CQRS และมีการเรียกใช้ในโมดูล order ด้วย ให้เริ่มจากสร้าง contract ขึ้นมาก่อน

<aside>
💡

ให้ลบไฟล์ `customercontract/contract.go` เพราะจะเปลี่ยนจาก interface ของ public api contract เป็น struct ของ command กับ query แทน

</aside>

- สร้างไฟล์ `customercontract/query_customer_by_id.go`

    ```go
    package customercontract
    
    type GetCustomerByIDQuery struct {
     ID int64 `json:"id"`
    }
    
    type GetCustomerByIDQueryResult struct {
     ID     int64    `json:"id"`
     Email  string   `json:"email"`
     Credit int      `json:"credit"`
    }
    ```

- สร้างฟีเจอร์ get-by-id โดยการย้าย logic จาก `CustomerService.GetCustomerByID(…)` ออกมาไว้ใน `Handle(…)`

    > สร้างไฟล์ `customer/internal/feature/get-by-id/query_handler.go`
    >

    ```go
    package getbyid
    
    import (
     "context"
     "go-mma/modules/customer/domainerrors"
     "go-mma/modules/customer/internal/model"
     "go-mma/modules/customer/internal/repository"
     "go-mma/shared/contract/customercontract"
    )
    
    type getCustomerByIDQueryHandler struct {
     custRepo repository.CustomerRepository
    }
    
    func NewGetCustomerByIDQueryHandler(custRepo repository.CustomerRepository) *getCustomerByIDQueryHandler {
     return &getCustomerByIDQueryHandler{
      custRepo: custRepo,
     }
    }
    
    func (h *getCustomerByIDQueryHandler) Handle(ctx context.Context, query *customercontract.GetCustomerByIDQuery) (*customercontract.GetCustomerByIDQueryResult, error) {
     customer, err := h.custRepo.FindByID(ctx, query.ID)
     if err != nil {
      return nil, err
     }
     if customer == nil {
      return nil, domainerrors.ErrCustomerNotFound
     }
     return h.newGetCustomerByIDQueryResult(customer), nil
    }
    
    func (h *getCustomerByIDQueryHandler) newGetCustomerByIDQueryResult(customer *model.Customer) *customercontract.GetCustomerByIDQueryResult {
     return &customercontract.GetCustomerByIDQueryResult{
      ID:     customer.ID,
      Email:  customer.Email,
      Credit: customer.Credit,
     }
    }
    ```

### สร้างฟีเจอร์ **reserve-credit**: ตัดยอด credit

ฟีเจอร์์นี้ เป็นการตัดยอด credit ซึ่งเป็นการอัพเดทค่าในฐานข้อมูล จัดเป็น Command ตาม CQRS และมีการเรียกใช้ในโมดูล order ด้วย ให้เริ่มจากสร้าง contract ขึ้นมาก่อน

- สร้างไฟล์ `customer-contract/command_reserve_credit.go`

    ```go
    package customercontract
    
    type ReserveCreditCommand struct {
     CustomerID   int64 `json:"customer_id"`
     CreditAmount int   `json:"credit_amount"`
    }
    ```

- สร้างฟีเจอร์ reserve-credit โดยการย้าย logic จาก `CustomerService.ReserveCredit(…)` ออกมาไว้ใน `Handle(…)`

    > สร้างไฟล์ `customer/internal/feature/reserve-credit/command_handler.go`
    >

    ```go
    package reservecredit
    
    import (
     "context"
     "go-mma/modules/customer/domainerrors"
     "go-mma/modules/customer/internal/repository"
     "go-mma/shared/common/errs"
     "go-mma/shared/common/logger"
     "go-mma/shared/common/mediator"
     "go-mma/shared/common/storage/sqldb/transactor"
     "go-mma/shared/contract/customercontract"
    )
    
    type reserveCreditCommandHandler struct {
     transactor transactor.Transactor
     custRepo   repository.CustomerRepository
    }
    
    func NewReserveCreditCommandHandler(
     transactor transactor.Transactor,
     repo repository.CustomerRepository) *reserveCreditCommandHandler {
     return &reserveCreditCommandHandler{
      transactor: transactor,
      custRepo:   repo,
     }
    }
    
    func (h *reserveCreditCommandHandler) Handle(ctx context.Context, cmd *customercontract.ReserveCreditCommand) (*mediator.NoResponse, error) {
     err := h.transactor.WithinTransaction(ctx, func(ctx context.Context, registerPostCommitHook func(transactor.PostCommitHook)) error {
      customer, err := h.custRepo.FindByID(ctx, cmd.CustomerID)
      if err != nil {
       logger.Log.Error(err.Error())
       return err
      }
    
      if customer == nil {
       return domainerrors.ErrCustomerNotFound
      }
    
      if err := customer.ReserveCredit(cmd.CreditAmount); err != nil {
       return err
      }
    
      if err := h.custRepo.UpdateCredit(ctx, customer); err != nil {
       logger.Log.Error(err.Error())
       return errs.DatabaseFailureError(err.Error())
      }
    
      return nil
     })
    
     return nil, err
    }
    
    ```

### สร้างฟีเจอร์ **release-credit**: คืนยอด credit

ฟีเจอร์์นี้ เป็นการคืนยอด credit ซึ่งเป็นการอัพเดทค่าในฐานข้อมูล จัดเป็น Command ตาม CQRS และมีการเรียกใช้ในโมดูล order ด้วย ให้เริ่มจากสร้าง contract ขึ้นมาก่อน

- สร้างไฟล์ `customer-contract/command_release_credit.go`

    ```go
    package customercontract
    
    type ReleaseCreditCommand struct {
     CustomerID   int64 `json:"customer_id"`
     CreditAmount int   `json:"credit_amount"`
    }
    ```

- สร้างฟีเจอร์ release-credit โดยการย้าย logic จาก `CustomerService.ReleaseCredit(…)` ออกมาไว้ใน `Handle(…)`

    > สร้างไฟล์ `customer/internal/feature/release-credit/command_handler.go`
    >

    ```go
    package releasecredit
    
    import (
     "context"
     "go-mma/modules/customer/domainerrors"
     "go-mma/modules/customer/internal/repository"
     "go-mma/shared/common/logger"
     "go-mma/shared/common/mediator"
     "go-mma/shared/common/storage/sqldb/transactor"
     "go-mma/shared/contract/customercontract"
    )
    
    type releaseCreditCommandHandler struct {
     transactor transactor.Transactor
     custRepo   repository.CustomerRepository
    }
    
    func NewReleaseCreditCommandHandler(
     transactor transactor.Transactor,
     repo repository.CustomerRepository) *releaseCreditCommandHandler {
     return &releaseCreditCommandHandler{
      transactor: transactor,
      custRepo:   repo,
     }
    }
    
    func (h *releaseCreditCommandHandler) Handle(ctx context.Context, cmd *customercontract.ReleaseCreditCommand) (*mediator.NoResponse, error) {
     err := h.transactor.WithinTransaction(ctx, func(ctx context.Context, registerPostCommitHook func(transactor.PostCommitHook)) error {
      customer, err := h.custRepo.FindByID(ctx, cmd.CustomerID)
      if err != nil {
       logger.Log.Error(err.Error())
       return err
      }
    
      if customer == nil {
       return domainerrors.ErrCustomerNotFound
      }
    
      customer.ReleaseCredit(cmd.CreditAmount)
    
      if err := h.custRepo.UpdateCredit(ctx, customer); err != nil {
       logger.Log.Error(err.Error())
       return err
      }
    
      return nil
     })
    
     return nil, err
    }
    ```

### สร้างฟีเจอร์ **create**: สร้างลูกค้าใหม่

ฟีเจอร์์นี้ เป็นการบันทึกข้อมูลลูกค้าใหม่ลงในฐานข้อมูล จัดเป็น Command ตาม CQRS และไม่มีการเรียกใช้ที่โมดูลอื่น จึงไม่จำเป็นต้องมี contract

- เนื่องฟีเจอร์นี้จะมีการเรียกใช้งานผ่าน REST API จึงต้องมี endpoint สำหรับจัดการ request/response ด้วย เริ่มจากย้าย `dto` มาไว้ที่นี้

    > สร้างไฟล์ `customer/internal/feature/create/dto.go`
    >

    ```go
    package create
    
    import (
     "errors"
     "net/mail"
    )
    
    type CreateCustomerRequest struct {
     Email  string `json:"email"`
     Credit int    `json:"credit"`
    }
    
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
    
    type CreateCustomerResponse struct {
     ID int64 `json:"id"`
    }
    ```

- ออกแบบ Command สำหรับฟีเจอร์ create

    > สร้างไฟล์ `customer/internal/feature/create/command.go`
    >

    ```go
    package create
    
    type CreateCustomerCommand struct {
     CreateCustomerRequest  // embeded type มาเพราะหน้าตาเหมือนกัน
    }
    
    type CreateCustomerCommandResult struct {
     CreateCustomerResponse // embeded type มาเพราะหน้าตาเหมือนกัน
    }
    
    // ฟังก์ชันช่วยสร้าง CreateCustomerCommandResult
    func NewCreateCustomerCommandResult(id int64) *CreateCustomerCommandResult {
     return &CreateCustomerCommandResult{
      CreateCustomerResponse{
       ID: id,
      },
     }
    }
    ```

- สร้างฟีเจอร์ create

    > สร้างไฟล์ `customer/internal/feature/create/command_handler.go`
    >

    ```go
    package create
    
    import (
     "context"
     "go-mma/modules/customer/domainerrors"
     "go-mma/modules/customer/internal/model"
     "go-mma/modules/customer/internal/repository"
     "go-mma/shared/common/logger"
     "go-mma/shared/common/storage/sqldb/transactor"
    
     notiService "go-mma/modules/notification/service"
    )
    
    type createCustomerCommandHandler struct {
     transactor transactor.Transactor
     custRepo   repository.CustomerRepository
     notiSvc    notiService.NotificationService
    }
    
    func NewCreateCustomerCommandHandler(
     transactor transactor.Transactor,
     custRepo repository.CustomerRepository,
     notiSvc notiService.NotificationService) *createCustomerCommandHandler {
     return &createCustomerCommandHandler{
      transactor: transactor,
      custRepo:   custRepo,
      notiSvc:    notiSvc,
     }
    }
    
    func (h *createCustomerCommandHandler) Handle(ctx context.Context, cmd *CreateCustomerCommand) (*CreateCustomerCommandResult, error) {
     // ตรวจสอบ business rule/invariant
     if err := h.validateBusinessInvariant(ctx, cmd); err != nil {
      return nil, err
     }
    
     // แปลง Command → Model
     customer := model.NewCustomer(cmd.Email, cmd.Credit)
    
     // ย้ายส่วนที่ติดต่อฐานข้อมูล กับส่งอีเมลมาทำงานใน WithinTransaction
     err := h.transactor.WithinTransaction(ctx, func(ctx context.Context, registerPostCommitHook func(transactor.PostCommitHook)) error {
    
      // ส่งไปที่ Repository Layer เพื่อบันทึกข้อมูลลงฐานข้อมูล
      if err := h.custRepo.Create(ctx, customer); err != nil {
       // error logging
       logger.Log.Error(err.Error())
       return err
      }
    
      // เพิ่มส่งอีเมลต้อนรับ เข้าไปใน hook แทน การเรียกใช้งานทันที
      registerPostCommitHook(func(ctx context.Context) error {
       return h.notiSvc.SendEmail(customer.Email, "Welcome to our service!", map[string]any{
        "message": "Thank you for joining us! We are excited to have you as a member."})
      })
    
      return nil
     })
    
     if err != nil {
      return nil, err
     }
    
     return NewCreateCustomerCommandResult(customer.ID), nil
    }
    
    func (h *createCustomerCommandHandler) validateBusinessInvariant(ctx context.Context, cmd *CreateCustomerCommand) error {
     // ตรวจสอบ email ซ้ำ
     exists, err := h.custRepo.ExistsByEmail(ctx, cmd.Email)
     if err != nil {
      // error logging
      logger.Log.Error(err.Error())
      return err
     }
    
     if exists {
      return domainerrors.ErrEmailExists
     }
     return nil
    }
    ```

- สร้าง endpoint ของฟีเจอร์นี้

    สร้างไฟล์ `customer/internal/feature/create/endpoint.go`

    ```go
    package create
    
    import (
     "go-mma/shared/common/errs"
     "go-mma/shared/common/mediator"
     "strings"
    
     "github.com/gofiber/fiber/v3"
    )
    
    func NewEndpoint(router fiber.Router, path string) {
     router.Post(path, createCustomerHTTPHandler)
    }
    
    func createCustomerHTTPHandler(c fiber.Ctx) error {
     // แปลง request body -> dto
     var req dto.CreateCustomerRequest
     if err := c.Bind().Body(&req); err != nil {
      // จัดการ error response ที่ middleware
      return errs.InputValidationError(err.Error())
     }
    
     logger.Log.Info(fmt.Sprintf("Received customer: %v", req))
    
     // ตรวจสอบ input fields (e.g., value, format, etc.)
     if err := req.Validate(); err != nil {
      // จัดการ error response ที่ middleware
      return errs.InputValidationError(err.Error())
     }
    
     // *** ส่งไปที่ Command Handler แทน Service ***
     resp, err := mediator.Send[*CreateCustomerCommand, *CreateCustomerCommandResult](
      c.Context(),
      &CreateCustomerCommand{CreateCustomerRequest: req},
     )
    
     // จัดการ error จาก feature หากเกิดขึ้น
     if err != nil {
      // จัดการ error response ที่ middleware
      return err
     }
    
     // ตอบกลับด้วย status code 201 (created) และข้อมูลแบบ JSON
     return c.Status(fiber.StatusCreated).JSON(resp)
    }
    ```

### ปรับแก้การ Init โมดูล Customer

จากเดิมใน `customer/module.go` จะมีการระบุว่าต้องการเปิด service อะไรให้ใช้งานบ้าง เราจะเอาตรงนี้ออกไป(ไม่มี `CustomerService` แล้ว) โดยจะใช้ mediator มาจัดการแทน

```go
package customer

import (
 "go-mma/modules/customer/internal/feature/create"
 getbyid "go-mma/modules/customer/internal/feature/get-by-id"
 releasecredit "go-mma/modules/customer/internal/feature/release-credit"
 reservecredit "go-mma/modules/customer/internal/feature/reserve-credit"
 "go-mma/modules/customer/internal/repository"
 "go-mma/shared/common/mediator"
 "go-mma/shared/common/module"
 "go-mma/shared/common/registry"

 notiModule "go-mma/modules/notification"
 notiService "go-mma/modules/notification/service"

 "github.com/gofiber/fiber/v3"
)

func NewModule(mCtx *module.ModuleContext) module.Module {
 return &moduleImp{mCtx: mCtx}
}

type moduleImp struct {
 mCtx *module.ModuleContext
 // เอา service ออก
}

func (m *moduleImp) APIVersion() string {
 return "v1"
}

func (m *moduleImp) Init(reg registry.ServiceRegistry) error {
 // Resolve NotificationService from the registry
 notiSvc, err := registry.ResolveAs[notiService.NotificationService](reg, notiModule.NotificationServiceKey)
 if err != nil {
  return err
 }

 repo := repository.NewCustomerRepository(m.mCtx.DBCtx)

  // <-- ตรงนี้
  // ให้ทำการ register handler เข้า mediator
 mediator.Register(create.NewCreateCustomerCommandHandler(m.mCtx.Transactor, repo, notiSvc))
 mediator.Register(getbyid.NewGetCustomerByIDQueryHandler(repo))
 mediator.Register(reservecredit.NewReserveCreditCommandHandler(m.mCtx.Transactor, repo))
 mediator.Register(releasecredit.NewReleaseCreditCommandHandler(m.mCtx.Transactor, repo))

 return nil
}

// ลบ Services() []registry.ProvidedService ออก

func (m *moduleImp) RegisterRoutes(router fiber.Router) {
 customers := router.Group("/customers")
 create.NewEndpoint(customers, "")
}
```

### ปรับแก้โมดูล Order

ปรับโมดูล Order ให้เรียกใช้ Command/Query ของโมดูล Customer แทนการเรียกจาก service

เริ่มจากแยก `OrderService` เป็นฟีเจอร์

```bash
order
├── domainerrors
│   └── domainerrors.go             # ไว้รวบรวม error ทั้่งหมด ของ order
├── internal
│   ├── feature                     # สร้างใน internal ป้องกันไม่ให้ import
│   │   ├── create
│   │   │   ├── dto.go              # ย้าย dto มาที่นี่
│   │   │   ├── endpoint.go         # ย้าย http handler มาที่นี่
│   │   │   ├── command.go          # กำหนดรูปแบบของ Request/Response ของ command
│   │   │   └── command_handler.go  # จัดการ command handler
│   │   └── cancel
│   │       ├── dto.go              # ย้าย dto มาที่นี่
│   │       ├── endpoint.go         # ย้าย http handler มาที่นี่
│   │       ├── command.go          # กำหนดรูปแบบของ Request/Response ของ command
│   │       └── command_handler.go  # จัดการ command handler
│   ├── model
│   │   └── order.go
│   └── repository
│       └── order.go
├── test
│   └── orders.http
├── module.go                        # register command/query handler
├── go.mod
└── go.sum
```

- รวบรวม error ทั้งหมดใน feature มาไว้ที่เดียวกัน

    > สร้างไฟล์ `order/domainerrors/domainerrors.go`
    >

    ```go
    package domainerrors
    
    import "go-mma/shared/common/errs"
    
    var (
     ErrNoOrderID = errs.ResourceNotFoundError("the order with given id was not found")
    )
    ```

- สร้างฟีเจอร์ create สำหรับสร้างออเดอร์ใหม่ และมีการเรียกใช้งานผ่าน REST API

    > สร้างไฟล์ `order/internal/feature/create/dto.go`
    >

    ```go
    package create
    
    import "fmt"
    
    type CreateOrderRequest struct {
     CustomerID int64 `json:"customer_id"`
     OrderTotal int   `json:"order_total"`
    }
    
    func (r *CreateOrderRequest) Validate() error {
     if r.CustomerID <= 0 {
      return fmt.Errorf("customer_id is required")
     }
     if r.OrderTotal <= 0 {
      return fmt.Errorf("order_total must be greater than 0")
     }
     return nil
    }
    
    type CreateOrderResponse struct {
     ID int64 `json:"id"`
    }
    ```

    สร้างไฟล์ `order/internal/feature/create/command.go`

    ```go
    package create
    
    type CreateOrderCommand struct {
     CreateOrderRequest
    }
    
    type CreateOrderCommandResult struct {
     CreateOrderResponse
    }
    
    func NewCreateOrderCommandResult(id int64) *CreateOrderCommandResult {
     return &CreateOrderCommandResult{
      CreateOrderResponse{ID: id},
     }
    }
    ```

    สร้างไฟล์ `order/internal/feature/create/command_handler.go`

    ```go
    package create
    
    import (
     "context"
     "go-mma/modules/order/internal/model"
     "go-mma/modules/order/internal/repository"
     "go-mma/shared/common/logger"
     "go-mma/shared/common/mediator"
     "go-mma/shared/common/storage/sqldb/transactor"
     "go-mma/shared/contract/customercontract"
    
     notiService "go-mma/modules/notification/service"
    )
    
    type createOrderCommandHandler struct {
     transactor transactor.Transactor
     orderRepo  repository.OrderRepository
     notiSvc    notiService.NotificationService
    }
    
    func NewCreateOrderCommandHandler(
     transactor transactor.Transactor,
     orderRepo repository.OrderRepository,
     notiSvc notiService.NotificationService) *createOrderCommandHandler {
     return &createOrderCommandHandler{
      transactor: transactor,
      orderRepo:  orderRepo,
      notiSvc:    notiSvc,
     }
    }
    
    func (h *createOrderCommandHandler) Handle(ctx context.Context, cmd *CreateOrderCommand) (*CreateOrderCommandResult, error) {
     // Business Logic Rule: ตรวจสอบ customer id ในฐานข้อมูล
     customer, err := mediator.Send[*customercontract.GetCustomerByIDQuery, *customercontract.GetCustomerByIDQueryResult](
      ctx,
      &customercontract.GetCustomerByIDQuery{ID: cmd.CustomerID},
     )
     if err != nil {
      return nil, err
     }
    
     var order *model.Order
     err = h.transactor.WithinTransaction(ctx, func(ctx context.Context, registerPostCommitHook func(transactor.PostCommitHook)) error {
    
      // Business Logic Rule: ตัดยอด credit ในตาราง customer
      if _, err := mediator.Send[*customercontract.ReserveCreditCommand, *mediator.NoResponse](
       ctx,
       &customercontract.ReserveCreditCommand{CustomerID: cmd.CustomerID, CreditAmount: cmd.OrderTotal},
      ); err != nil {
       return err
      }
    
      // สร้าง order ใหม่ DTO -> Model
      order = model.NewOrder(cmd.CustomerID, cmd.OrderTotal)
      
      // บันทึกลงฐานข้อมูล
      err := h.orderRepo.Create(ctx, order)
      if err != nil {
       logger.Log.Error(err.Error())
       return err
      }
    
      // ส่งอีเมลยืนยันหลัง commit
      registerPostCommitHook(func(ctx context.Context) error {
       return h.notiSvc.SendEmail(customer.Email, "Order Created", map[string]any{
        "order_id": order.ID,
        "total":    order.OrderTotal,
       })
      })
    
      return nil
     })
    
     if err != nil {
      return nil, err
     }
    
     return NewCreateOrderCommandResult(order.ID), nil
    }
    
    ```

    สร้างไฟล์ `order/internal/feature/create/endpoint.go`

    ```go
    package create
    
    import (
     "go-mma/shared/common/errs"
     "go-mma/shared/common/mediator"
     "strings"
    
     "github.com/gofiber/fiber/v3"
    )
    
    func NewEndpoint(router fiber.Router, path string) {
     router.Post(path, createOrderHTTPHandler)
    }
    
    func createOrderHTTPHandler(c fiber.Ctx) error {
     // แปลง request body -> struct
     var req CreateOrderRequest
     if err := c.Bind().Body(&req); err != nil {
       // จัดการ error response ที่ middleware
      return errs.InputValidationError(err.Error())
     }
    
     // ตรวจสอบ input fields (e.g., value, format, etc.)
     if err := req.Validate(); err != nil {
       // จัดการ error response ที่ middleware
      return errs.InputValidationError(err.Error())
     }
    
     // ส่งไปที่ Command Handler
     resp, err := mediator.Send[*CreateOrderCommand, *CreateOrderCommandResult](
      c.Context(),
      &CreateOrderCommand{CreateOrderRequest: req},
     )
    
     // จัดการ error จาก feature หากเกิดขึ้น
     if err != nil {
      // จัดการ error response ที่ middleware
      return err
     }
    
     // ตอบกลับด้วย status code 201 (created) และข้อมูลแบบ JSON
     return c.Status(fiber.StatusCreated).JSON(resp)
    }
    ```

- สร้างฟีเจอร์ cancel สำหรับยกเลิกออเดอร์ และมีการเรียกใช้งานผ่าน REST API

    สร้างไฟล์ `order/internal/feature/cancel/command.go`

    ```go
    package cancel
    
    type CancelOrderCommand struct {
     ID int64 `json:"id"`
    }
    ```

    สร้างไฟล์ `order/internal/feature/cancel/command_handler.go`

    ```go
    package cancel
    
    import (
     "context"
     "go-mma/modules/order/domainerrors"
     "go-mma/modules/order/internal/repository"
     "go-mma/shared/common/logger"
     "go-mma/shared/common/mediator"
     "go-mma/shared/common/storage/sqldb/transactor"
     "go-mma/shared/contract/customercontract"
    )
    
    type cancelOrderCommandHandler struct {
     transactor transactor.Transactor
     orderRepo  repository.OrderRepository
    }
    
    func NewCancelOrderCommandHandler(
     transactor transactor.Transactor,
     orderRepo repository.OrderRepository) *cancelOrderCommandHandler {
     return &cancelOrderCommandHandler{
      transactor: transactor,
      orderRepo:  orderRepo,
     }
    }
    
    func (h *cancelOrderCommandHandler) Handle(ctx context.Context, cmd *CancelOrderCommand) (*mediator.NoResponse, error) {
     err := h.transactor.WithinTransaction(ctx, func(ctx context.Context, registerPostCommitHook func(transactor.PostCommitHook)) error {
      // Business Logic Rule: ตรวจสอบ order id
      order, err := h.orderRepo.FindByID(ctx, cmd.ID)
      if err != nil {
       logger.Log.Error(err.Error())
       return err
      }
    
      if order == nil {
       return domainerrors.ErrNoOrderID
      }
    
      // ยกเลิก order
      if err := h.orderRepo.Cancel(ctx, order.ID); err != nil {
       logger.Log.Error(err.Error())
       return err
      }
    
      // Business Logic Rule: คืน credit ในตาราง customer
      if _, err := mediator.Send[*customercontract.ReleaseCreditCommand, *mediator.NoResponse](
       ctx,
       &customercontract.ReleaseCreditCommand{CustomerID: order.CustomerID, CreditAmount: order.OrderTotal},
      ); err != nil {
       return err
      }
    
      return nil
     })
    
     if err != nil {
      return nil, err
     }
    
     return nil, nil
    }
    ```

    สร้างไฟล์ `order/internal/feature/cancel/endpoint.go`

    ```go
    package cancel
    
    import (
     "go-mma/shared/common/errs"
     "go-mma/shared/common/mediator"
     "strconv"
    
     "github.com/gofiber/fiber/v3"
    )
    
    func NewEndpoint(router fiber.Router, path string) {
     router.Delete(path, cancelOrderHTTPHandler)
    }
    
    func cancelOrderHTTPHandler(c fiber.Ctx) error {
     // ตรวจสอบรูปแบบ orderID
     orderID, err := strconv.Atoi(c.Params("orderID"))
     if err != nil {
      // จัดการ error response ที่ middleware
      return errs.InputValidationError("invalid order id")
     }
    
     logger.Log.Info(fmt.Sprintf("Cancelling order: %v", orderID))
    
     // ส่งไปที่ Command Handler
     _, err = mediator.Send[*CancelOrderCommand, *mediator.NoResponse](
      c.Context(),
      &CancelOrderCommand{ID: int64(orderID)},
     )
    
     // จัดการ error จาก feature หากเกิดขึ้น
     if err != nil {
      // จัดการ error response ที่ middleware
      return err
     }
    
     // ตอบกลับด้วย status code 204 (no content)
     return c.SendStatus(fiber.StatusNoContent)
    }
    ```

- เพิ่มการ register command handlers ทั้งหมด ใน `order/module.go`

    ```go
    package order
    
    import (
     "go-mma/modules/order/internal/feature/cancel"
     "go-mma/modules/order/internal/feature/create"
     "go-mma/modules/order/internal/repository"
     "go-mma/shared/common/mediator"
     "go-mma/shared/common/module"
     "go-mma/shared/common/registry"
    
     notiModule "go-mma/modules/notification"
     notiService "go-mma/modules/notification/service"
    
     "github.com/gofiber/fiber/v3"
    )
    
    func NewModule(mCtx *module.ModuleContext) module.Module {
     return &moduleImp{mCtx: mCtx}
    }
    
    type moduleImp struct {
     mCtx *module.ModuleContext
    }
    
    func (m *moduleImp) APIVersion() string {
     return "v1"
    }
    
    func (m *moduleImp) Init(reg registry.ServiceRegistry) error {
    
     // Resolve NotificationService from the registry
     notiSvc, err := registry.ResolveAs[notiService.NotificationService](reg, notiModule.NotificationServiceKey)
     if err != nil {
      return err
     }
    
     repo := repository.NewOrderRepository(m.mCtx.DBCtx)
    
     mediator.Register(create.NewCreateOrderCommandHandler(m.mCtx.Transactor, repo, notiSvc))
     mediator.Register(cancel.NewCancelOrderCommandHandler(m.mCtx.Transactor, repo))
    
     return nil
    }
    
    func (m *moduleImp) RegisterRoutes(router fiber.Router) {
     orders := router.Group("/orders")
     create.NewEndpoint(orders, "")
     cancel.NewEndpoint(orders, "/:orderID")
    }
    ```

### โมดูล Notification

ขอข้ามการแปลงโมดูล notification ไปก่อน

## เพิ่มความยืดหยุ่นด้วยแนวคิด Event-Driven Architecture

ในโค้ดปัจจุบัน เช่น ภายใน `CreateCustomerCommandHandler` ซึ่งมีหน้าที่สร้างลูกค้าใหม่ เรายังใช้รูปแบบการเขียนแบบ **Imperative Style** — คือการสั่งงานโดยตรง เช่น

```go
h.notiSvc.SendEmail(...)
```

```bash
+-------------------------------+
| CreateCustomerCommandHandler  |
+-------------------------------+
           |
           | creates Customer
           v
+-------------------------------+
|  CustomerRepository           |
+-------------------------------+
           |
           | persists to DB
           v
+-------------------------------+
|  NotificationService          |
+-------------------------------+
           |
           | sends welcome email
           v
       External System
```

แม้จะทำงานได้ดีในระบบขนาดเล็ก แต่แนวทางนี้มีข้อจำกัด

- ตัว Handler ผูกกับ service อื่นโดยตรง (tight coupling)
- หากต้องเพิ่ม action อื่น เช่น ส่ง SMS หรือบันทึก audit log → ต้องแก้โค้ดตรงนี้เพิ่ม
- ขัดกับหลัก Single Responsibility

### แนวทางใหม่: เปลี่ยนเป็น Event-Driven

แทนที่จะ “สั่งให้ทำงาน” โดยตรง → ให้ "ประกาศเหตุการณ์" แล้วให้ส่วนอื่นในระบบมารับฟังและจัดการ

ข้อดีของแนวทางนี้

- **ทำให้ระบบ หลวมตัว (loosely coupled):** Handler ไม่รู้ว่ามีใครฟัง event หรือใครจะทำอะไร
- **ขยายระบบง่าย:** อยากเพิ่ม logic ใหม่ แค่เพิ่ม event handler โดยไม่ต้องแก้ handler เดิม
- **พร้อมสำหรับ scaling:** สามารถแยกบาง handler ไปทำงาน async หรือใช้ message queue ได้ทันที

### Domain Events

- เป็น event ที่เกิดภายใน domain (bounded context)
- ใช้เพื่อแจ้งว่า *"สิ่งนี้เกิดขึ้นแล้ว"* (เช่น `CustomerCreated`)
- ถูก publish และ consume *ในกระบวนการเดียวกัน* (in-process)
- มักใช้กับ business logic ภายใน

### Integration Events

- ถูกใช้เพื่อสื่อสารข้าม bounded context / microservice
- ใช้ messaging system เช่น Kafka, RabbitMQ
- มักเกิดจาก domain event แล้วถูกแปลง (map) เป็น integration event
- ทำงานแบบ async

### โครงสร้างหลังเพิ่ม Domain Events กับ Integration Events

ตัวอย่าง การแยก logic การส่งอีเมลออกจาก Handler และรองรับการทำงานแบบ asynchronous ซึ่งจะมีขั้นตอนการทำงานแบบนี้

```bash
+-------------------------------+
| CreateCustomerCommandHandler  |
+-------------------------------+
           |
           | creates Customer
           v
+-------------------------+
|  Customer Entity        |
|  + AddDomainEvent()     |
+-------------------------+
           |
           | emits domain event
           v
+-------------------------------+
| DomainEventDispatcher         |
| (in-process, synchronous)     |
+-------------------------------+
           |
           | calls domain handler
           v
+------------------------------------------+
| CustomerCreatedDomainEventHandler        |
| - Converts to Integration Event          |
| - Calls EventBus.Publish()               |
+------------------------------------------+
           |
           | emits async message (Kafka/Outbox)
           v
+------------------------------+
|  NotificationService         |
| (another module/microservice)|
+------------------------------+
           |
           | sends welcome email
           v
       External System
```

1. `CreateCustomerHandler` → สร้าง Customer และเพิ่ม `CustomerCreated` domain event
2. `DomainEventDispatcher` → dispatch event นี้ให้ `CustomerCreatedDomainEventHandler`
3. Handler → สร้าง `CustomerCreatedIntegrationEvent` แล้วส่งผ่าน EventBus
4. ระบบภายนอก (เช่น Notification Module) consume แล้วจัดการเรื่อง Email

## Refactor เพิ่ม Domain Event

การทำ Domain Event ให้สมบูรณ์ในระบบที่ใช้ DDD (Domain-Driven Design) และ Event-Driven Architecture มีองค์ประกอบหลัก ดังนี้

1. **Domain Event**
    - เป็น struct ที่บรรยายเหตุการณ์ที่ “เกิดขึ้นแล้ว” ใน domain
    - อยู่ใน layer `domain` หรือ `internal/domain/event`
2. **Aggregate/Entity ที่สร้าง Event**
    - Entity เช่น `Customer` ต้องมีช่องทางในการเก็บ domain events (เช่น slice `[]DomainEvent`)
    - เมื่อเกิดเหตุการณ์ ให้ `append()` ลงไป
3. **DomainEvent Interface**
    - ใช้เป็น abstraction สำหรับ event ทั้งหมด เช่น: มี method `EventName()` หรือ `OccurredAt()`
4. **Event Dispatcher**
    - ดึง events จาก aggregate แล้ว dispatch ไปยังผู้รับ (handler)
5. **Event Handler**
    - โค้ดที่รับ event และทำงานตอบสนอง
    - อยู่ใน layer `domain` หรือ `internal/domain/eventhandler`
6. **Trigger Point**
    - จุดที่ pull domain events เพื่อนำไปส่งผ่าน dispatcher (มักอยู่หลัง transaction สำเร็จ)
7. **Dispatch Events มี 2 แนวทางหลัก**
    - ภายใน transaction (immediate dispatch) เหมาะกับ use case ที่ event handler แค่ปรับ state ภายใน เช่น update model อื่น ซึ่งจะ coupling กับ transaction logic ถ้า event handler fail จะต้อง rollback transaction ด้วย
    - หลังจาก commit แล้ว คือ ดึง domain events → รอ DB commit → dispatch เช่น post-commit hook เหมาะกับ handler ที่มี side-effect เช่น ส่งอีเมล, call external service แต่ต้องมีการจัดการ error และ retry เอง แยกออกมาจาก transaction logic

### DomainEvent Interface

ใช้เป็น abstraction สำหรับ event ทั้งหมด เช่น: มี method `EventName()` หรือ `OccurredAt()`

- สร้างไฟล์ `common/domain/event.go`

    ```go
    package domain
    
    import "time"
    
    // EventName คือ alias ของ string เพื่อใช้แทนชื่อ event เช่น "CustomerCreated", "OrderPlaced" เป็นต้น
    type EventName string
    
    // DomainEvent เป็น interface สำหรับ event ที่เกิดขึ้นใน domain (Domain Event ตาม DDD)
    // ใช้เพื่อให้สามารถบันทึกหรือส่ง event ได้ โดยไม่ต้องรู้โครงสร้างภายใน
    type DomainEvent interface {
     EventName() EventName     // คืนชื่อ event
     OccurredAt() time.Time    // คืนเวลาที่ event เกิด
    }
    
    // BaseDomainEvent เป็น struct พื้นฐานที่ implement DomainEvent
    // ใช้ฝังใน struct อื่นๆ ที่เป็น event เพื่อ reuse method ได้
    type BaseDomainEvent struct {
     Name EventName   // ชื่อของ event เช่น "UserRegistered"
     At   time.Time   // เวลาที่ event นี้เกิดขึ้น
    }
    
    // EventName คืนชื่อของ event นี้
    func (e BaseDomainEvent) EventName() EventName {
     return e.Name
    }
    
    // OccurredAt คืนเวลาที่ event นี้เกิดขึ้น
    // ใช้แบบ value receiver ด้วยเหตุผลเดียวกันกับข้างต้น
    func (e BaseDomainEvent) OccurredAt() time.Time {
     return e.At
    }
    
    ```

    <aside>
    💡

    ใช้ **value receiver** เพราะ struct เล็ก ไม่มี mutation และปลอดภัยในเชิง concurrent

    </aside>

### Aggregate

`Aggregate` เป็น **ฐานแม่แบบ (base struct)** สำหรับ aggregate root ทุกตัว เช่น `Customer`, `Order` ฯลฯ จะทำหน้าที่ **บันทึกเหตุการณ์ที่เกิดขึ้น (Domain Events)** เพื่อให้ layer ภายนอก (เช่น Application หรือ Infrastructure) นำไปประมวลผลต่อ

- สร้างไฟล์ `common/domain/aggregate.go`

    ```go
    package domain
    
    // Aggregate เป็น struct พื้นฐานสำหรับ aggregate root ทั้งหมดใน DDD
    // ใช้เพื่อเก็บรวบรวม domain events ที่เกิดขึ้นภายใน aggregate
    type Aggregate struct {
     domainEvents []DomainEvent // เก็บรายการของ event ที่เกิดขึ้นใน aggregate นี้
    }
    
    // AddDomainEvent ใช้สำหรับเพิ่ม domain event เข้าไปใน aggregate
    // ฟังก์ชันนี้จะถูกเรียกภายใน method อื่น ๆ ของ aggregate เมื่อต้องการประกาศว่า event บางอย่างได้เกิดขึ้นแล้ว
    func (a *Aggregate) AddDomainEvent(dv DomainEvent) {
     // สร้าง slice เปล่าหากยังไม่มี
     if a.domainEvents == nil {
      a.domainEvents = make([]DomainEvent, 0)
     }
    
     // เพิ่ม event ลงใน slice
     a.domainEvents = append(a.domainEvents, dv)
    }
    
    // PullDomainEvents จะดึง domain events ทั้งหมดออกจาก aggregate
    // พร้อมกับเคลียร์ events เหล่านั้นจาก memory (เพราะ events ถูกส่งออกไปแล้ว)
    // เหมาะสำหรับใช้ใน layer ที่ทำการ publish หรือ persist event
    func (a *Aggregate) PullDomainEvents() []DomainEvent {
     events := a.domainEvents    // ดึง event ทั้งหมดที่บันทึกไว้
     a.domainEvents = nil        // เคลียร์ event list เพื่อป้องกันการส่งซ้ำ
     return events
    }
    ```

### Event Dispatcher

สำหรับการ register handler และ dispatch ไปยังผู้รับ (handler)

- สร้างไฟล์ `common/domain/event_dispatcher.go`

    ```go
    package domain
    
    import (
     "context"
     "fmt"
     "sync"
    )
    
    // Error ที่ใช้ตรวจสอบความถูกต้องของ event
    var (
     ErrInvalidEvent = fmt.Errorf("invalid domain event")
    )
    
    // DomainEventHandler คือ interface ที่ทุก handler ของ event ต้อง implement
    // โดยจะมี method เดียวคือ Handle เพื่อรับ event และทำงานตาม logic ที่ต้องการ
    type DomainEventHandler interface {
     Handle(ctx context.Context, event DomainEvent) error
    }
    
    // DomainEventDispatcher คือ interface สำหรับระบบที่ทำหน้าที่กระจาย (dispatch) event
    // โดยสามารถ register handler สำหรับแต่ละ EventName และ dispatch หลาย event พร้อมกันได้
    type DomainEventDispatcher interface {
     Register(eventType EventName, handler DomainEventHandler)
     Dispatch(ctx context.Context, events []DomainEvent) error
    }
    
    // simpleDomainEventDispatcher เป็น implementation ง่าย ๆ ของ DomainEventDispatcher
    // ใช้ map เก็บ handler แยกตาม EventName
    type simpleDomainEventDispatcher struct {
     handlers map[EventName][]DomainEventHandler // แผนที่ของ EventName ไปยัง handler หลายตัว
     mu       sync.RWMutex                       // ใช้ mutex เพื่อป้องกัน concurrent read/write
    }
    
    // NewSimpleDomainEventDispatcher สร้าง instance ใหม่ของ dispatcher
    func NewSimpleDomainEventDispatcher() DomainEventDispatcher {
     return &simpleDomainEventDispatcher{
      handlers: make(map[EventName][]DomainEventHandler),
     }
    }
    
    // Register ใช้สำหรับลงทะเบียน handler กับ EventName
    // Handler จะถูกเรียกเมื่อมี event นั้น ๆ ถูก dispatch
    func (d *simpleDomainEventDispatcher) Register(eventType EventName, handler DomainEventHandler) {
     d.mu.Lock()
     defer d.mu.Unlock()
    
     // เพิ่ม handler ไปยัง slice ของ event นั้น ๆ
     d.handlers[eventType] = append(d.handlers[eventType], handler)
    }
    
    // Dispatch รับ slice ของ event แล้ว dispatch ไปยัง handler ที่ลงทะเบียนไว้
    // ถ้ามี handler มากกว่าหนึ่งตัวสำหรับ event เดียวกัน จะเรียกทุกตัว
    func (d *simpleDomainEventDispatcher) Dispatch(ctx context.Context, events []DomainEvent) error {
     for _, event := range events {
      // อ่าน handler ของ event นี้ (copy slice เพื่อป้องกัน concurrent modification)
      d.mu.RLock()
      handlers := append([]DomainEventHandler(nil), d.handlers[event.EventName()]...)
      d.mu.RUnlock()
    
      // เรียก handler แต่ละตัว
      for _, handler := range handlers {
       err := func(h DomainEventHandler) error {
        // หาก handler ทำงานผิดพลาด จะคืน error พร้อมระบุ event ที่ผิด
        err := h.Handle(ctx, event)
        if err != nil {
         return fmt.Errorf("error handling event %s: %w", event.EventName(), err)
        }
        return nil
       }(handler)
    
       // หากมี error จาก handler ใด ๆ จะหยุดและ return เลย
       if err != nil {
        return err
       }
      }
     }
    
     // ถ้าไม่มี error เลย ส่ง nil กลับ
     return nil
    }
    ```

### Domain Event

- สร้าง domain event สำหรับเมื่อสร้างลูกค้าใหม่สำเร็จ

    > สร้างไฟล์ `customer/internal/domain/event/customer_created.go`
    >

    ```go
    package event
    
    import (
     "go-mma/shared/common/domain"
     "time"
    )
    
    // กำหนดชื่อ Event ที่ใช้ในระบบ (EventName)
    // เพื่อระบุชนิดของ Domain Event ว่าเป็น "CustomerCreated"
    const (
     CustomerCreatedDomainEventType domain.EventName = "CustomerCreated"
    )
    
    // CustomerCreatedDomainEvent คือ struct ที่เก็บข้อมูลของเหตุการณ์
    // ที่เกิดขึ้นเมื่อมีการสร้าง Customer ใหม่ในระบบ
    type CustomerCreatedDomainEvent struct {
     domain.BaseDomainEvent // ฝัง BaseDomainEvent ที่มีชื่อและเวลาเกิด event
     CustomerID int64       // ID ของ Customer ที่ถูกสร้าง
     Email      string      // Email ของ Customer ที่ถูกสร้าง
    }
    
    // NewCustomerCreatedDomainEvent สร้าง instance ใหม่ของ CustomerCreatedDomainEvent
    // โดยรับ customer ID และ email เป็น input และตั้งชื่อ event กับเวลาปัจจุบันอัตโนมัติ
    func NewCustomerCreatedDomainEvent(custID int64, email string) *CustomerCreatedDomainEvent {
     return &CustomerCreatedDomainEvent{
      BaseDomainEvent: domain.BaseDomainEvent{
       Name: CustomerCreatedDomainEventType, // กำหนดชื่อ event
       At:   time.Now(),                      // เวลาเกิด event ณ ปัจจุบัน
      },
      CustomerID: custID, // กำหนด customer ID
      Email:      email,  // กำหนด email
     }
    }
    ```

- ปรับโมเดล `Customer` ให้เป็น `Aggregate` เพื่อเพิ่มเหตุการณ์  “CustomerCreated” ณ ตอนสร้างโมเดล

    > แก้ไขไฟล์ `customer/internal/model/customer.go`
    >

    ```go
    package model
    
    import (
     "go-mma/modules/customer/internal/domain/event"
     "go-mma/shared/common/domain"
     "go-mma/shared/common/errs"
     "go-mma/shared/common/idgen"
     "time"
    )
    
    type Customer struct {
     ID        int64     `db:"id"` // tag db ใช้สำหรับ StructScan() ของ sqlx
     Email     string    `db:"email"`
     Credit    int       `db:"credit"`
     CreatedAt time.Time `db:"created_at"`
     UpdatedAt time.Time `db:"updated_at"`
     domain.Aggregate    // embed: เพื่อให้กลายเป็น Aggregate ของ Customer
    }
    
    func NewCustomer(email string, credit int) *Customer {
     customer := &Customer{
      ID:     idgen.GenerateTimeRandomID(),
      Email:  email,
      Credit: credit,
     }
    
     // เพิ่มเหตุการณ์ "CustomerCreated"
     customer.AddDomainEvent(event.NewCustomerCreatedDomainEvent(customer.ID, customer.Email))
    
     return customer
    }
    
    func (c *Customer) ReserveCredit(v int) error {
     // ...
    }
    
    func (c *Customer) ReleaseCredit(v int) {
     // ...
    }
    ```

### Domain Event Handler

สำหรับโค้ดที่รับ event “`CustomerCreated`” มาทำงานต่อ

> สร้างไฟล์ `customer/internal/domain/eventhandler/customer_created_handler.go`
>

```go
package eventhandler

import (
 "context"
 "go-mma/modules/customer/internal/domain/event"
 notiService "go-mma/modules/notification/service"
 "go-mma/shared/common/domain"
)

// customerCreatedDomainEventHandler คือ handler สำหรับจัดการ event ประเภท CustomerCreatedDomainEvent
type customerCreatedDomainEventHandler struct {
 notiSvc notiService.NotificationService // service สำหรับส่งการแจ้งเตือน (เช่น อีเมล)
}

// NewCustomerCreatedDomainEventHandler คือฟังก์ชันสร้าง instance ของ handler นี้
func NewCustomerCreatedDomainEventHandler(notiSvc notiService.NotificationService) domain.DomainEventHandler {
 return &customerCreatedDomainEventHandler{
  notiSvc: notiSvc,
 }
}

// Handle คือฟังก์ชันหลักที่ถูกเรียกเมื่อมี event ถูก dispatch มา
func (h *customerCreatedDomainEventHandler) Handle(ctx context.Context, evt domain.DomainEvent) error {
 // แปลง (type assert) event ที่รับมาเป็น pointer ของ CustomerCreatedDomainEvent
 e, ok := evt.(*event.CustomerCreatedDomainEvent)
 if !ok {
  // ถ้าไม่ใช่ event ประเภทนี้ ให้ส่ง error กลับไป
  return domain.ErrInvalidEvent
 }

 // เรียกใช้ service ส่งอีเมลต้อนรับลูกค้าใหม่
 if err := h.notiSvc.SendEmail(e.Email, "Welcome to our service!", map[string]any{
  "message": "Thank you for joining us! We are excited to have you as a member.",
 }); err != nil {
  // หากส่งอีเมลไม่สำเร็จ ส่ง error กลับไป
  return err
 }

 // ถ้าสำเร็จทั้งหมด ให้คืน nil (ไม่มี error)
 return nil
}
```

### Trigger Point

เป็นจุดที่ดึงเอา domain events ออกมาหลังจาก transaction logic ทั้งหมดทำงานเสร็จแล้ว แต่ยังไม่ได้ commit

> แก้ไขไฟล์ `customer/internal/feature/create/handler.go`
>

```go
package create

// ...

func (h *createCustomerCommandHandler) Handle(ctx context.Context, cmd *CreateCustomerCommand) (*CreateCustomerCommandResult, error) {
 // ...
 err := h.transactor.WithinTransaction(ctx, func(ctx context.Context, registerPostCommitHook func(transactor.PostCommitHook)) error {

  // ส่งไปที่ Repository Layer เพื่อบันทึกข้อมูลลงฐานข้อมูล
  if err := h.custRepo.Create(ctx, customer); err != nil {
   // error logging
   logger.Log.Error(err.Error())
   return err
  }
  
  // เพิ่มตรงนี้ หลังจากบันทึกสำเร็จแล้ว

  // ดึง domain events จาก customer model
  events := customer.PullDomainEvents()

  return nil
 })

 // ..
}
```

### Dispatch Events

เนื่องจาก domain event handler ที่สร้างมาจะเป็นการส่งอีเมล ซึ่งมี side-effect จึงเหมาะกับแบบ post-commit dispatch หรือ รอให้ DB commit ก่อนค่อย dispatch

- แก้ไขไฟล์ `customer/internal/feature/create/handler.go`

    ```go
    package create
    
    import (
     "context"
     "go-mma/modules/customer/domainerrors"
     "go-mma/modules/customer/internal/model"
     "go-mma/modules/customer/internal/repository"
     "go-mma/shared/common/domain" // เพิ่มตรงนี้
     "go-mma/shared/common/logger"
     "go-mma/shared/common/storage/sqldb/transactor"
    )
    
    type createCustomerCommandHandler struct {
     transactor transactor.Transactor
     custRepo   repository.CustomerRepository
     dispatcher domain.DomainEventDispatcher // เพิ่มตรงนี้ มีการใช้ dispatcher
    }
    
    func NewCreateCustomerCommandHandler(
     transactor transactor.Transactor,
     custRepo repository.CustomerRepository,
     dispatcher domain.DomainEventDispatcher, // เพิ่มตรงนี้
    ) *createCustomerCommandHandler {
     return &createCustomerCommandHandler{
      transactor: transactor,
      custRepo:   custRepo,
      dispatcher: dispatcher, // เพิ่มตรงนี้
     }
    }
    
    func (h *createCustomerCommandHandler) Handle(ctx context.Context, cmd *CreateCustomerCommand) (*CreateCustomerCommandResult, error) {
     // ...
     err := h.transactor.WithinTransaction(ctx, func(ctx context.Context, registerPostCommitHook func(transactor.PostCommitHook)) error {
      // ...
      // ดึง domain events จาก customer model
      events := customer.PullDomainEvents()
    
      // ให้ dispatch หลัง commit แล้ว
      registerPostCommitHook(func(ctx context.Context) error {
       return h.dispatcher.Dispatch(ctx, events)
      })
    
      return nil
     })
    
     // ..
    }
    ```

### Register domain event

เนื่องจาก domain events เป็นการทำงานเฉพาะในโมดูลนั้นๆ เท่านั้น ดังนั้น ให้สร้าง dispatcher แยกของแต่ละโมดูลได้เลย

- แก้ไขไฟล์ `customer/module.go`

    ```go
    package customer
    
    import (
     "go-mma/modules/customer/internal/domain/event"         // เพิ่มตรงนี้่
     "go-mma/modules/customer/internal/domain/eventhandler"  // เพิ่มตรงนี้่
     "go-mma/modules/customer/internal/feature/create"
     getbyid "go-mma/modules/customer/internal/feature/get-by-id"
     releasecredit "go-mma/modules/customer/internal/feature/release-credit"
     reservecredit "go-mma/modules/customer/internal/feature/reserve-credit"
     "go-mma/modules/customer/internal/repository"
     "go-mma/shared/common/domain"                           // เพิ่มตรงนี้่
     "go-mma/shared/common/mediator"
     "go-mma/shared/common/module"
     "go-mma/shared/common/registry"
    
     notiModule "go-mma/modules/notification"
     notiService "go-mma/modules/notification/service"
    
     "github.com/gofiber/fiber/v3"
    )
    
    // ...
    
    func (m *moduleImp) Init(reg registry.ServiceRegistry) error {
     // Resolve NotificationService from the registry
     // ...
    
     // เพิ่มตรงนี้่
     // สร้าง Domain Event Dispatcher สำหรับโมดูลนี้โดยเฉพาะ
     // เราจะไม่ใช้ dispatcher กลาง แต่จะสร้าง dispatcher แยกในแต่ละโมดูลแทน
     // เพื่อให้โมดูลนั้นๆ ควบคุมการลงทะเบียนและการจัดการ event handler ได้เองอย่างอิสระ
     dispatcher := domain.NewSimpleDomainEventDispatcher()
    
     // ลงทะเบียน handler สำหรับ event CustomerCreatedDomainEventType ใน dispatcher ของโมดูลนี้
     dispatcher.Register(event.CustomerCreatedDomainEventType, eventhandler.NewCustomerCreatedDomainEventHandler(notiSvc))
    
     // สร้าง repository ของโมดูลนี้
     repo := repository.NewCustomerRepository(m.mCtx.DBCtx)
    
     // ลงทะเบียน command handler และส่ง dispatcher เข้าไปใน handler ด้วย
     // เพื่อให้ handler สามารถ dispatch event ผ่าน dispatcher ของโมดูลนี้ได้
     mediator.Register(create.NewCreateCustomerCommandHandler(m.mCtx.Transactor, repo, dispatcher))
     
     // ...
    }
    ```

## Refactor เพิ่ม Integration Event

ในการทำ Integration Event ใน Event-Driven Architecture (EDA) มีหลาย รูปแบบ (patterns) ที่สามารถเลือกใช้ได้ ขึ้นอยู่กับความ ซับซ้อนของระบบ, ระดับการ decouple, และ ความน่าเชื่อถือที่ต้องการ โดยทั่วไปสามารถแบ่งได้เป็น 3 รูปแบบหลัก ๆ ดังนี้

1. **In-Memory Event Bus (Monolith)**

    **ลักษณะ**

    - Event ถูกส่งแบบ in-process (memory) ไปยัง handler ที่ลงทะเบียนไว้ใน runtime เดียวกัน
    - ใช้ในระบบ monolith หรือระบบที่แยกโมดูลแต่ยังรันใน process เดียว

    **ข้อดี**

    - ง่าย
    - เร็ว

    **ข้อเสีย**

    - ไม่ทนต่อ crash
    - ถ้า handler พังหรือ panic → ไม่มี retry
    - ไม่สามารถ scale ข้าม service/process ได้
2. **Outbox Pattern (Reliable Messaging in Monolith / Microservices)**

    **ลักษณะ**

    - เมื่อมี event เกิดขึ้น → บันทึกทั้ง business data + integration event ใน transaction เดียวกัน
    - Event ถูกเก็บใน outbox table
    - Worker (หรือ background process) คอยอ่านและส่งไปยัง message broker (Kafka, RabbitMQ)

    **ข้อดี**

    - ปลอดภัย (atomic): business data + event commit พร้อมกัน
    - ทนต่อ crash
    - Decouple services ได้ (publish ไป Kafka)

    **ข้อเสีย**

    - ต้องมี worker ดึงและส่ง
    - ซับซ้อนกว่า in-memory
3. **Change Data Capture (CDC)**

    **ลักษณะ**

    - ใช้ระบบอย่าง Debezium หรือ Kafka Connect ฟังการเปลี่ยนแปลงใน DB (ผ่าน WAL หรือ binlog)
    - เมื่อมี insert/update → สร้างเป็น event และส่งออกไป message broker

    **ข้อดี**

    - ไม่ต้องมี Worker (หรือ background process) คอยอ่านและส่งไปยัง message broker
    - มองเห็นทุกการเปลี่ยนแปลงของฐานข้อมูล

    **ข้อเสีย**

    - ต้องจัดการ schema evolution และ data format ให้ดี

### In-Memory Event Bus (Monolith)

ในบทความนี้จะเลือกใช้วิธีแบบ In-Memory Event Bus เพราะระบบเป็น Monolith และง่ายต่อการทำความเข้าใจ

การทำ Integration Event แบบ In-Memory Event Bus ภายใน Monolith คือการสื่อสารระหว่างโมดูล (bounded contexts) โดยไม่ใช้ messaging system ภายนอก เช่น Kafka หรือ RabbitMQ แต่ยังแยก "Integration Event" ออกจาก "Domain Event" เพื่อรักษา separation of concerns

มีองค์ประกอบหลัก ดังนี้

1. **Integration Event**
    - เป็น struct ที่ใช้สื่อสารข้ามโมดูล (context) ภายในระบบเดียวกัน
    - มี payload ที่ module ปลายทางต้องใช้ เช่น `CustomerCreatedIntegrationEvent`
2. **Integration Event Interface**
    - ใช้เป็น abstraction สำหรับ event ทั้งหมด เช่น: มี method `EventID()`หรือ `EventName()` หรือ `OccurredAt()`
3. **Event Bus (In-Memory Implementation)**
    - ตัวกลางในการ publish → ไปยัง handler ที่ลงทะเบียนไว้
    - เก็บ handler เป็น map จาก event name → handler list
4. **Register / Subscribe**
    - Module ที่สนใจ event ต้องลงทะเบียน handler ไว้กับ EventBus
5. **Publish**
    - เมื่อ module ต้นทางสร้าง event แล้วเรียก `eventBus.Publish(...)`
    - EventBus จะกระจาย event ไปยัง handler ที่ลงทะเบียนไว้
6. **Event Handlers**
    - แต่ละ handler มี logic ของตัวเอง เช่นส่งอีเมล

### สร้าง Integration Event Interface

- ใช้เป็น abstraction สำหรับ event ทั้งหมด เช่น: มี method `EventID()`หรือ `EventName()` หรือ `OccurredAt()`
- สร้างไฟล์ `common/eventbus/event.go`

    ```go
    package eventbus
    
    import (
     "time"
    )
    
    // EventName เป็นชนิดข้อมูลสำหรับชื่อ event
    type EventName string
    
    // Event คือ interface สำหรับ event ทั่วไปในระบบ
    // ต้องมี method สำหรับดึงข้อมูล ID, ชื่อ event, และเวลาที่เกิด event
    type Event interface {
     EventID() string       // คืนค่า ID ของ event (เช่น UUID หรือ ULID)
     EventName() EventName  // คืนค่าชื่อ event เช่น "CustomerCreated"
     OccurredAt() time.Time // เวลาที่ event นั้นเกิดขึ้น
    }
    
    // BaseEvent คือ struct พื้นฐานที่ใช้เก็บข้อมูล event ทั่วไป
    // สามารถนำไปฝังใน struct event ที่เฉพาะเจาะจงได้
    type BaseEvent struct {
     ID   string    // รหัส event แบบ unique (UUID/ULID)
     Name EventName // ชื่อของ event เช่น "CustomerCreated"
     At   time.Time // เวลาที่ event เกิดขึ้น
    }
    
    // EventID คืนค่า ID ของ event
    func (e BaseEvent) EventID() string {
     return e.ID
    }
    
    // EventName คืนค่าชื่อของ event
    func (e BaseEvent) EventName() EventName {
     return e.Name
    }
    
    // OccurredAt คืนค่าเวลาที่ event เกิดขึ้น
    func (e BaseEvent) OccurredAt() time.Time {
     return e.At
    }
    ```

### สร้าง EventBus (In-Memory Implementation)

สำหรับเป็นตัวกลางในการ publish → ไปยัง handler ที่ลงทะเบียนไว้

- สร้างไฟล์ `common/eventbus/eventbus.go`

    ```go
    package eventbus
    
    import (
     "context"
    )
    
    // IntegrationEventHandler คือ interface สำหรับ handler ที่จะรับผิดชอบการจัดการ Integration Event
    // จะต้องมี method Handle เพื่อรับ context และ event ที่ต้องการจัดการ
    type IntegrationEventHandler interface {
     Handle(ctx context.Context, event Event) error
    }
    
    // EventBus คือ interface สำหรับระบบ event bus ที่ใช้ในการ publish และ subscribe event ต่างๆ
    type EventBus interface {
     // Publish ใช้สำหรับส่ง event ออกไปยังระบบ event bus
     // โดยรับ context และ event ที่จะส่ง
     Publish(ctx context.Context, event Event) error
    
     // Subscribe ใช้สำหรับลงทะเบียน handler สำหรับ event ที่มีชื่อ eventName
     // เมื่อมี event ที่ตรงกับชื่อ eventName เข้ามา handler ที่ลงทะเบียนไว้จะถูกเรียกใช้
     Subscribe(eventName EventName, handler IntegrationEventHandler)
    }
    
    ```

- สร้างไฟล์ `common/eventbus/in_memory_eventbus.go`

    ```go
    package eventbus
    
    import (
     "context"
     "log"
     "sync"
    )
    
    // implementation ของ EventBus แบบง่าย ๆ ที่เก็บ subscriber ไว้ใน memory
    type inmemoryEventBus struct {
     subscribers map[EventName][]IntegrationEventHandler // เก็บ eventName กับ list ของ handler
     mu          sync.RWMutex                            // mutex สำหรับป้องกัน concurrent access
    }
    
    // สร้าง instance ใหม่ของ inmemoryEventBus พร้อม map subscribers ว่าง ๆ
    func NewInMemoryEventBus() EventBus {
     return &inmemoryEventBus{
      subscribers: make(map[EventName][]IntegrationEventHandler),
     }
    }
    
    // Subscribe ใช้ลงทะเบียน handler สำหรับ event ที่มีชื่อ eventName
    // โดยจะเพิ่ม handler เข้าไปใน map subscribers
    func (eb *inmemoryEventBus) Subscribe(eventName EventName, handler IntegrationEventHandler) {
     eb.mu.Lock()
     defer eb.mu.Unlock()
    
     // เพิ่ม handler เข้า slice ของ eventName นั้น ๆ
     eb.subscribers[eventName] = append(eb.subscribers[eventName], handler)
    }
    
    // Publish ส่ง event ไปยัง handler ทุกตัวที่ subscribe event ชื่อเดียวกัน
    func (eb *inmemoryEventBus) Publish(ctx context.Context, event Event) error {
     eb.mu.RLock()
     defer eb.mu.RUnlock()
    
     // หา handler ที่ลงทะเบียนกับ event นี้
     handlers, ok := eb.subscribers[event.EventName()]
     if !ok {
      // ไม่มี handler สำหรับ event นี้ ก็ return nil
      return nil
     }
    
     // สร้าง context ใหม่ที่อาจมีข้อมูลเพิ่มเติมสำหรับ event bus
     busCtx := context.WithValue(ctx, "name", "context in event bus")
    
     // เรียก handler ทุกตัวแบบ asynchronous (goroutine) เพื่อไม่บล็อกการทำงาน
     for _, handler := range handlers {
      go func(h IntegrationEventHandler) {
       // เรียก handle event และ log error ถ้ามี
       err := h.Handle(busCtx, event)
       if err != nil {
        log.Printf("error handling event %s: %v", event.EventName(), err)
       }
      }(handler)
     }
     return nil
    }
    ```

### สร้าง Integration Event

เนื่องจาก integration event จะต้องใช้ร่วมกันหลายโมดูล ให้สร้างโปรเจกต์ใหม่ชื่อ messaging ใน shared

- สร้างโปรเจกต์ `messaging`

    ```bash
    mkdir -p src/shared/messaging
    cd src/shared/messaging
    go mod init go-mma/shared/messaging
    ```

    เพิ่มลง workspace ด้วย

    ```go
    {
      "folders": [
        // ...,
        {
          "path": "src/shared/messaging"
        }
      ],
      "settings": {}
    }
    ```

    เพิ่ม module replace สำหรับโปรเจกต์ `common`

    ```go
    module go-mma/shared/messaging
    
    go 1.24.1
    
    replace go-mma/shared/common v0.0.0 => ../common
    
    require go-mma/shared/common v0.0.0
    ```

- เพิ่ม module replace ในทุกโมดูลโปรเจกต์ และ `app`

    ```go
    // app
    replace go-mma/shared/messaging v0.0.0 => ../shared/messaging
    
    // customer
    replace go-mma/shared/messaging v0.0.0 => ../../shared/messaging
    
    // order
    replace go-mma/shared/messaging v0.0.0 => ../../shared/messaging
    
    // notification
    replace go-mma/shared/messaging v0.0.0 => ../../shared/messaging
    ```

- สร้างไฟล์ `messaging/customer_created.go`

    ```go
    package messaging
    
    import (
     "go-mma/shared/common/eventbus"
     "go-mma/shared/common/idgen"
     "time"
    )
    
    const (
     CustomerCreatedIntegrationEventName eventbus.EventName = "CustomerCreated"
    )
    
    type CustomerCreatedIntegrationEvent struct {
     eventbus.BaseEvent
     CustomerID int64  `json:"customer_id"`
     Email      string `json:"email"`
    }
    
    func NewCustomerCreatedIntegrationEvent(customerID int64, email string) *CustomerCreatedIntegrationEvent {
     return &CustomerCreatedIntegrationEvent{
      BaseEvent: eventbus.BaseEvent{
       ID:   idgen.GenerateUUIDLikeID(),
       Name: CustomerCreatedIntegrationEventName,
       At:   time.Now(),
      },
      CustomerID: customerID,
      Email:      email,
     }
    }
    ```

- ติดตั้ง dependencies ด้วย `go mod tidy`

### สร้าง Integration Event Handler

สำหรับโค้ดที่รับ event “`CustomerCreated`” มาทำงานต่อ โดยใยที่นี่จะทำที่โมดูล notification เพื่อส่ง welcome email

- สร้างไฟล์ `notification/internal/integration/customer/welcome_email_handler.go` (สื่อว่า integration จากโมดูล customer)

    ```go
    package customer
    
    import (
     "context"
     "fmt"
     "go-mma/modules/notification/service"
     "go-mma/shared/common/eventbus"
     "go-mma/shared/messaging"
    )
    
    type welcomeEmailHandler struct {
     notiService service.NotificationService
    }
    
    func NewWelcomeEmailHandler(notiService service.NotificationService) *welcomeEmailHandler {
     return &welcomeEmailHandler{
      notiService: notiService,
     }
    }
    
    func (h *welcomeEmailHandler) Handle(ctx context.Context, evt eventbus.Event) error {
     e, ok := evt.(*messaging.CustomerCreatedIntegrationEvent) // ใช้ pointer
     if !ok {
      return fmt.Errorf("invalid event type")
     }
    
     return h.notiService.SendEmail(e.Email, "Welcome to our service!", map[string]any{
      "message": "Thank you for joining us! We are excited to have you as a member.",
     })
    }
    ```

- ติดตั้ง dependencies ด้วย `go mod tidy`

### สร้าง Integration Event Publisher

เดิมใน CustomerCreatedDomainEventHandler จะมีการเรียก notiService เพื่อส่งอีเมลโดยตรง เราจะเปลี่ยนส่งนี้ให้ส่งไปเป็น integration event แทน

- แก้ไขไฟล์  `customer/internal/domain/eventhandler/customer_created_handler.go`

    ```go
    package eventhandler
    
    import (
     "context"
     "go-mma/modules/customer/internal/domain/event"
     "go-mma/shared/common/domain"
     "go-mma/shared/common/eventbus"
     "go-mma/shared/messaging"
    )
    
    type customerCreatedDomainEventHandler struct {
     eventBus eventbus.EventBus // เปลี่ยนมาใช้ eventbus
    }
    
    // เปลี่ยนมาใช้ eventbus
    func NewCustomerCreatedDomainEventHandler(eventBus eventbus.EventBus) domain.DomainEventHandler {
     return &customerCreatedDomainEventHandler{
      eventBus: eventBus, // เปลี่ยนมาใช้ eventbus
     }
    }
    
    func (h *customerCreatedDomainEventHandler) Handle(ctx context.Context, evt domain.DomainEvent) error {
     e, ok := evt.(*event.CustomerCreatedDomainEvent) // ใช้ pointer
    
     if !ok {
      return domain.ErrInvalidEvent
     }
    
     // สร้าง IntegrationEvent จาก Domain Event
     integrationEvent := messaging.NewCustomerCreatedIntegrationEvent(
      e.CustomerID,
      e.Email,
     )
    
     return h.eventBus.Publish(ctx, integrationEvent)
    }
    ```

- ติดตั้ง dependencies ด้วย `go mod tidy`
- แก้ไขไฟล์  `common/module/module.go` เพื่อให้ `Init()` รองรับ event bus

    ```go
    type Module interface {
     APIVersion() string
     Init(reg registry.ServiceRegistry, eventBus eventbus.EventBus) error // รับ eventBus เพิ่ม
     RegisterRoutes(r fiber.Router)
    }
    ```

- แก้ไขไฟล์  `customer/module.go` เพื่อลบ notification service ออก

    ```go
    func (m *moduleImp) Init(reg registry.ServiceRegistry, eventBus eventbus.EventBus) error {
     // เอา notiSvc ออก
     
     // Register domain event handler
     dispatcher := domain.NewSimpleDomainEventDispatcher()
     dispatcher.Register(event.CustomerCreatedDomainEventType, eventhandler.NewCustomerCreatedDomainEventHandler(eventBus)) // ส่ง eventBus เข้าไปแทน
    
     repo := repository.NewCustomerRepository(m.mCtx.DBCtx)
    
     mediator.Register(create.NewCreateCustomerCommandHandler(m.mCtx.Transactor, repo, dispatcher))
     mediator.Register(getbyid.NewGetCustomerByIDQueryHandler(repo))
     mediator.Register(reservecredit.NewReserveCreditCommandHandler(m.mCtx.Transactor, repo))
     mediator.Register(releasecredit.NewReleaseCreditCommandHandler(m.mCtx.Transactor, repo))
    
     return nil
    }
    ```

- ต้องแก้ไข `Init()` ที่โมดูล `order` กับ `notification` ด้วย
- แก้ไฟล์ `app/application/application.go` เพื่อให้ส่ง event bus เข้าไปตอน init module

    ```go
    package application
    
    import (
      // ...
     "go-mma/shared/common/eventbus"
      // ...
    )
    
    type Application struct {
     config          config.Config
     httpServer      HTTPServer
     serviceRegistry registry.ServiceRegistry
     eventBus        eventbus.EventBus // เพ่ิม
    }
    
    func New(config config.Config) *Application {
     return &Application{
      config:          config,
      httpServer:      newHTTPServer(config),
      serviceRegistry: registry.NewServiceRegistry(),
      eventBus:        eventbus.NewInMemoryEventBus(), // เพ่ิม
     }
    }
    
    // ...
    
    func (app *Application) initModule(m module.Module) error {
     return m.Init(app.serviceRegistry, app.eventBus) // เพ่ิมส่ง eventBus
    }
    
    // ...
    ```

  - ติดตั้ง dependencies ด้วย `go mod tidy` ที่โปรเจกต์ `app`

### สร้าง Register / Subscribe

ให้โมดูล notification คอยรับ integration event

- แก้ไขไฟล์ `notification/module.go`

    ```go
    func (m *moduleImp) Init(reg registry.ServiceRegistry, eventBus eventbus.EventBus) error {
     m.notiSvc = service.NewNotificationService()
    
     // subscribe to integration events
     eventBus.Subscribe(messaging.CustomerCreatedIntegrationEventName, customer.NewWelcomeEmailHandler(m.notiSvc))
    
     return nil
    }
    ```

เพียงเท่านี้ก็สามารถใช้ integration event แบบ in-memory event bus ได้แล้ว แต่อย่าลืมว่าวิธีมีข้อเสียคือ ถ้า handler พังหรือ panic จะไม่มี retry อาจทำให้เกิด inconsistency เช่น ลูกค้าถูกสร้างแล้ว (`INSERT INTO customers`) แต่ไม่ส่งอีเมลต้อนรับ วิธีแก้ ได้แก่

- อาจเพิ่ม retry logic ตอนส่งอีเมล
- ใช้แนวทาง Hybrid Approach คือ ใช้ Domain Event → แปลง (map) เป็น Integration Event → เขียน Outbox table (ต้องทำใน transaction เดียวกับ business data)→ ใช้ CDC tools ส่ง event

---

## สรุป

**Modular Monolith** เป็นสถาปัตยกรรมที่ลงตัวอย่างยิ่งสำหรับหลายๆ โปรเจกต์ในปัจจุบัน โดยเฉพาะอย่างยิ่งเมื่อคุณต้องการความรวดเร็วในการพัฒนาในช่วงเริ่มต้น แต่ก็ยังคงความยืดหยุ่นในการปรับขนาดและปรับเปลี่ยนไปสู่ **Microservices** ในอนาคต

แนวทางนี้ช่วยให้คุณได้ประโยชน์จากการรวมโค้ดเบสไว้ในที่เดียว ทำให้การพัฒนา, ทดสอบ, และ Deployment ทำได้ง่ายขึ้น ลดความซับซ้อนในการจัดการโครงสร้างพื้นฐานที่ Microservices มักจะต้องเจอ ในขณะเดียวกันก็ยังคงรักษาหลักการของความเป็น **โมดูลาร์** ผ่านการกำหนดขอบเขตที่ชัดเจนและการบังคับใช้ **การแยกข้อมูล** ระหว่างโมดูลต่างๆ

ด้วยกฎเกณฑ์ที่ชัดเจนในการจัดการข้อมูล เช่น การที่แต่ละโมดูลเข้าถึงได้เฉพาะข้อมูลของตนเอง และไม่มีการแชร์ตารางโดยตรง Modular Monolith จึงส่งเสริมการออกแบบที่ **หลวมและยืดหยุ่น** ทำให้ง่ายต่อการบำรุงรักษาและเพิ่มฟีเจอร์ใหม่ๆ โดยไม่ก่อให้เกิดผลกระทบที่คาดไม่ถึง

ดังนั้น หากคุณกำลังมองหาสถาปัตยกรรมที่ช่วยให้คุณ **เริ่มต้นได้อย่างรวดเร็ว จัดการง่าย และพร้อมเติบโตในอนาคต** Modular Monolith คือทางเลือกที่คุณควรพิจารณาเป็นอันดับต้นๆ เลยครับ

---

## บทเสริม

## สร้าง API Document ด้วย Swagger

เมื่อพัฒนา API Service สิ่งที่ขาดไม่ได้คือ **API Documentation** ซึ่งช่วยให้เข้าใจว่าแต่ละ endpoint ทำงานอย่างไร รับ request แบบใด และตอบกลับ response ในรูปแบบไหน

หนึ่งในเครื่องมือยอดนิยมคือ **Swagger** และในภาษา Go มักใช้แพ็กเกจ [`swaggo/swag`](https://github.com/swaggo/swag) สำหรับ generate Swagger 2.0 documentation จาก comment ในโค้ด

### การใช้งานกับ Fiber

โดยทั่วไป หากใช้ **Fiber v2** จะสามารถใช้ middleware [`fiber-swagger`](https://github.com/swaggo/fiber-swagger) ได้ทันที ([อ่านเพิ่มเติมที่นี่](https://somprasongd.work/blog/go/golang-api-p10-doc))

แต่ในกรณีนี้เราใช้ **Fiber v3** จึงต้อง fork และปรับปรุงให้รองรับเอง โดยใช้เวอร์ชันนี้แทน

👉 [`github.com/somprasongd/fiber-swagger`](https://github.com/somprasongd/fiber-swagger)

### โครงสร้างแบบ Mono Repository

ในโปรเจกต์นี้ใช้โครงสร้างแบบ [Mono Repository](#รวมโค้ดทั้งหมดไว้ใน-mono-repository-อย่างเป็นระบบ) โดย

- โมดูลย่อยแต่ละตัว (เช่น `customer`, `order`) จะระบุ API ด้วย SWAG comment ไว้ภายใน
- ตัวโปรเจกต์หลัก (`app`) จะเป็นผู้เรียก `swag init` เพื่อรวมเอกสารทั้งหมด

### การเขียน SWAG Comment

ใส่ comment แบบ [Swagger annotations](https://github.com/swaggo/swag/tree/v2?tab=readme-ov-file#declarative-comments-format) ไว้เหนือฟังก์ชัน handler ดังนี้

- `/customers` – Create Customer

    > ใน `customer/internal/feature/create/endpoint.go`
    >

    ```go
    // CreateCustomer godoc
    // @Summary  Create Customer
    // @Description Create Customer
    // @Tags   Customer
    // @Produce  json
    // @Param   customer body CreateCustomerRequest true "Create Data"
    // @Failure  401
    // @Failure  500
    // @Success  201 {object} CreateCustomerResponse
    // @Router   /customers [post]
    func createCustomerHTTPHandler(c fiber.Ctx) error {
      // ...
    }
    ```

- `/orders` – Create Order

    > ใน `order/internal/feature/create/endpoint.go`
    >

    ```go
    // CreateOrder godoc
    // @Summary  Create Order
    // @Description Create Order
    // @Tags   Order
    // @Produce  json
    // @Param   order body CreateOrderRequest true "Create Data"
    // @Failure  401
    // @Failure  500
    // @Success  201 {object} CreateOrderResponse
    // @Router   /orders [post]
    func createOrderHTTPHandler(c fiber.Ctx) error {
      // ...
    }
    ```

- `/orders/{orderID}` – Cancel Order

    > ใน `order/internal/feature/cancel/endpoint.go`
    >

    ```go
    // CancelOrder godoc
    // @Summary  Cancel Order
    // @Description Cancel Order By Order ID
    // @Tags   Order
    // @Produce  json
    // @Param   orderID path string true "order id"
    // @Failure  401
    // @Failure  404
    // @Failure  500
    // @Success  204
    // @Router   /orders/{orderID} [delete]
    func cancelOrderHTTPHandler(c fiber.Ctx) error {
      // ...
    }
    ```

### ขั้นตอนการ Generate Swagger Documentation

- ติดตั้ง `swag` CLI

    ```bash
    go install github.com/swaggo/swag/v2/cmd/swag@latest
    ```

- เพิ่ม Makefile สำหรับ generate docs

    > แก้ไขไฟล์ `Makefile`
    >

    ```makefile
    .PHONY: doc
    doc:
     swag fmt -d src && \
     cd src/app && \
     swag init --parseDependency  --parseDependencyLevel 3 -g cmd/api/main.go -o docs
    ```

  - `swag fmt` : จัดรูปแบบ comment ให้อ่านง่าย
  - `--parseDependency  --parseDependencyLevel 3` : ให้ `swag` ตาม dependency ลงลึกถึงโมดูลย่อย
- รันคำสั่ง `make doc` จะได้ `docs` folder กับ  `docs/docs.go`
- สร้างไฟล์ middleware สำหรับ Swagger UI

    > ใน `application/middleware/swagger.go`
    >

    ```go
    package middleware
    
    import (
     "fmt"
     "go-mma/build"
     "go-mma/config"
     "go-mma/docs"
     "strings"
    
     "github.com/gofiber/fiber/v3"
     fiberSwagger "github.com/somprasongd/fiber-swagger"
    )
    
    func APIDoc(config config.Config) fiber.Handler {
     //Swagger Doc details
     host := removeProtocol(config.GatewayHost)
     basePath := config.GatewayBasePath
     schemas := []string{"http", "https"}
    
     if len(host) == 0 {
      host = fmt.Sprintf("localhost:%d", config.HTTPPort)
     }
    
     if len(basePath) == 0 {
      basePath = "/api/v1"
     }
    
     docs.SwaggerInfo.Title = "Go MMA Example API"
     docs.SwaggerInfo.Description = "This is a sample server GO MMA server."
     docs.SwaggerInfo.Version = build.Version
     docs.SwaggerInfo.Host = host
     docs.SwaggerInfo.BasePath = basePath
     docs.SwaggerInfo.Schemes = schemas
    
     return fiberSwagger.WrapHandler
    }
    
    func removeProtocol(url string) string {
     url = strings.TrimPrefix(url, "http://")
     url = strings.TrimPrefix(url, "https://")
     return url
    }
    ```

- อัปเดต `config/config.go` เพิ่ม `GatewayHost` และ `GatewayBasePath`

    ```go
    // ..
    
    type Config struct {
     HTTPPort        int
     GracefulTimeout time.Duration
     DSN             string
     GatewayHost     string // <-- เพิ่ม
     GatewayBasePath string // <-- เพิ่ม
    }
    
    func Load() (*Config, error) {
     config := &Config{
      HTTPPort:        env.GetIntDefault("HTTP_PORT", 8090),
      GracefulTimeout: env.GetDurationDefault("GRACEFUL_TIMEOUT", 5*time.Second),
      DSN:             env.Get("DB_DSN"),
      GatewayHost:     env.Get("GATEWAY_HOST"),  // <-- เพิ่ม
      GatewayBasePath: env.GetDefault("GATEWAY_BASEURL", "/api/v1"),  // <-- เพิ่ม
     }
     // ...
    }
    
    // ...
    ```

- เพิ่ม route สำหรับ Swagger UI

    > ใน `application/http.go`
    >

    ```go
    // ...
    
    func newHTTPServer(config config.Config) HTTPServer {
     return &httpServer{
      config: config,
      app:    newFiber(config), // เพิ่มส่ง config
     }
    }
    
    func newFiber(config config.Config) *fiber.App { // เพิ่มรับ config
     app := fiber.New(fiber.Config{
      AppName: fmt.Sprintf("Go MMA version %s", build.Version),
     })
    
     // global middleware
      // ...
    
     app.Get("/docs/*", middleware.APIDoc(config)) // <-- เพิ่ม
    
     app.Get("/", func(c fiber.Ctx) error {
      return c.JSON(map[string]string{"version": build.Version, "time": build.Time})
     })
    
     return app
    }
    
    // ...
    ```

- รันคำสั่ง `make run` แล้วเปิด [http://localhost:8090/docs/](http://localhost:8090/docs/index.html) เพื่อดู Swagger UI

> **หมายเหตุ:** ทุกครั้งที่มีการเพิ่มหรือแก้ไข SWAG comment ต้องรัน `make doc` ใหม่ เพื่อ regenerate เอกสาร
>
