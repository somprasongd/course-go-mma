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

---

