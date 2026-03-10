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

