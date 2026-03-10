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

