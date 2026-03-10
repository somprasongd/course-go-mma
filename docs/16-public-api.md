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

