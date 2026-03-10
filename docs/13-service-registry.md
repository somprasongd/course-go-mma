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

