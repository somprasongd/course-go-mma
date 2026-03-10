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

