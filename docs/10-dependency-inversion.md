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

