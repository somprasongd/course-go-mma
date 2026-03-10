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

