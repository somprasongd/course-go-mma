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

