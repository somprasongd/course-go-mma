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

