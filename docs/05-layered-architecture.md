## เริ่มต้นด้วย Layered Architecture ที่เข้าใจง่าย

ในการพัฒนา API หากเราเขียนโค้ดทั้งหมดไว้รวมกันในที่เดียว เช่น การเชื่อมต่อฐานข้อมูล การตรวจสอบข้อมูล และการจัดการ HTTP request ทั้งหมดไว้ใน handler เดียว จะทำให้ระบบ scale ได้ยาก ทดสอบลำบาก และเสี่ยงต่อ bug ที่ไม่คาดคิด

**ทางออกคือการใช้ Layered Architecture** ซึ่งช่วยแยกความรับผิดชอบของแต่ละส่วนออกจากกันอย่างชัดเจน ทำให้โค้ดมีโครงสร้างชัดเจน และดูแลต่อยอดในอนาคตได้ง่าย

เนื้อหาในส่วนนี้ประกอบด้วย

- Layered Architecture คืออะไร
- Repository Layer – ตัวกลางเชื่อมต่อกับฐานข้อมูล
- Service Layer – ประมวลผลลอจิกทางธุรกิจ
- Presentation Layer – จัดการ HTTP Request/Response
- ประกอบร่างระบบให้สมบูรณ์

### Layered Architecture คืออะไร

**Layered Architecture** คือรูปแบบการออกแบบซอฟต์แวร์ที่แยกโค้ดออกเป็น “ชั้น” หรือ “เลเยอร์” ตามหน้าที่ของมัน เช่น การเข้าถึงฐานข้อมูล, business logic และการจัดการ HTTP Request/Response โดยแต่ละเลเยอร์จะรับผิดชอบหน้าที่ของตัวเองอย่างชัดเจน

**โครงสร้าง Layered Architecture โดยทั่วไป**

```
Client/UI Layer        ← ผู้ใช้โต้ตอบกับระบบ
↓
Presentation Layer     ← Controller, API (จัดการคำขอจากผู้ใช้)
↓ ← DTO
Service Layer          ← Business Logic (กฎทางธุรกิจ)
↓ ← Model
Repository/Data Layer  ← จัดการกับฐานข้อมูล, external APIs
↓
Database/External APIs
```

**เมื่อนำมา implement ในโค้ดของเรา**

```
project/
│
├── handler/          ← Presentation Layer (HTTP handlers)
├── dto/              ← รับ/ส่งข้อมูลระหว่าง handler ↔ service
├── service/          ← Business Logic (core logic)
├── model/            ← ใช้จัดการข้อมูลภายในระบบ service ↔ repository
├── repository/       ← Data Access (DB queries)
└── main.go           ← Entry point (setup DI, server, etc)
```

### Repository Layer – ตัวกลางเชื่อมต่อกับฐานข้อมูล

**Repository Layer** มีหน้าที่หลักในการ ติดต่อกับฐานข้อมูลหรือแหล่งเก็บข้อมูล โดยรับคำสั่งจาก Service Layer แล้วทำหน้าที่ CRUD (Create, Read, Update, Delete) โดยไม่ให้เลเยอร์อื่นรู้ว่าข้อมูลมาจากที่ใด (Postgres, MySQL, Redis หรือแม้แต่ API)

- เริ่มจากการสร้าง Model โดยให้ไฟล์ชื่อ `model/customer.go` เพื่อใช้แทนโครงสร้างข้อมูลที่ตรงกับฐานข้อมูล

    ```go
    package model
    
    import (
     "go-mma/util/idgen"
     "time"
    )
    
    type Customer struct {
     ID          int64     `db:"id"` // tag db ใช้สำหรับ StructScan() ของ sqlx
     Email       string    `db:"email"`
     Credit      int       `db:"credit"`
     CreatedAt   time.Time `db:"created_at"`
     UpdatedAt   time.Time `db:"updated_at"`
    }
    
    func NewCustomer(email string, credit int) *Customer {
     return &Customer{
      ID:     idgen.GenerateTimeRandomID(),
      Email:  email,
      Credit: credit,
     }
    }
    ```

    <aside>
    💡

    สร้าง Model แบบ [Rich Model](https://somprasongd.work/blog/architecture/anemic-vs-rich-model-ddd)

    </aside>

- สร้าง Repository สำหรับ บันทึกลูกค้าใหม่ลงฐานข้อมูล ไว้ที่ไฟล์ `repository/customer.go`

    ```go
    package repository
    
    import (
     "context"
     "database/sql"
     "fmt"
     "go-mma/model"
     "go-mma/util/storage/sqldb"
     "time"
    )
    
    type CustomerRepository struct {
     dbCtx sqldb.DBContext // ใช้งาน database ผ่าน DBContext interface
    }
    
    func NewCustomerRepository(dbCtx sqldb.DBContext) *CustomerRepository {
     return &CustomerRepository{
      dbCtx: dbCtx, // inject DBContext instance into CustomerRepository
     }
    }
    
    func (r *CustomerRepository) Create(ctx context.Context, customer *model.Customer) error {
     query := `
     INSERT INTO public.customers (id, email, credit)
     VALUES ($1, $2, $3)
     RETURNING *
     `
    
     // กำหนด timeout ของ query
     ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
     defer cancel()
    
     err := r.dbCtx.DB().
      QueryRowxContext(ctx, query, customer.ID, customer.Email, customer.Credit).
      StructScan(customer) // นำค่า created_at, updated_at ใส่ใน struct customer
     if err != nil {
      return fmt.Errorf("an error occurred while inserting customer: %w", err)
     }
     return nil
    }
    
    func (r *CustomerRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
     query := `SELECT 1 FROM public.customers WHERE email = $1 LIMIT 1`
    
     ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
     defer cancel()
    
     var exists int
     err := r.dbCtx.DB().
      QueryRowxContext(ctx, query, email).
      Scan(&exists)
     if err != nil {
      if err == sql.ErrNoRows { // หาไม่เจอแสดงว่ายังไม่มี email ในระบบแล้ว
       return false, nil
      }
      return false, fmt.Errorf("an error occurred while checking email: %w", err)
     }
     return true, nil // ถ้าไม่ error แสดงว่ามี email ในระบบแล้ว
    }
    ```

    <aside>
    💡

    การใช้ `context.WithTimeout` เป็นแนวปฏิบัติมาตรฐานสำหรับระบบงานที่เกี่ยวข้องกับฐานข้อมูลหรือ external service

    </aside>

### Service Layer – ประมวลผลลอจิกทางธุรกิจ

**Service Layer** คือเลเยอร์ที่อยู่ตรงกลางระหว่าง Controller (หรือ Handler) กับ Repository

หน้าที่หลักของ Service Layer คือ รวมและควบคุม Business Logic ของแอปพลิเคชันไว้ในที่เดียว ดังนี้

- **รับ DTO**: รับ DTO จาก Handler เข้ามาเพื่อประมวลผล
- **ตรวจสอบ**: ตรวจสอบความถูกต้องตาม **"กฎทางธุรกิจ"** (Business Logic/Semantic Validation) ซึ่งมักจะต้องมีการประมวลผลหรือตรวจสอบกับข้อมูลส่วนอื่นๆ ในระบบ เช่น การตรวจสอบข้อมูลซ้ำในฐานข้อมูล
- **แปลงข้อมูล**: แปลง DTO → Model
- **เรียก Repository**: เพื่อทำ CRUD (Create, Read, Update, Delete) ตามเงื่อนไข
- **ส่งผลลัพธ์**: รับผลลัพธ์จาก Repository แล้วแปลงกลับเป็น DTO Response
- **จัดการ error**: แสดง error log แล้วส่งกลับไปให้ Controller (หรือ Handler) จัดการต่อ

**ขั้นตอนการสร้าง Service Layer**

- สร้าง DTO (Data Transfer Object) ไว้เป็นตัวกลางสำหรับรับ–ส่งข้อมูล ระหว่างชั้น Handler ↔ Service

    สร้างไฟล์ `dto/customer_request.go`

    ```go
    package dto
    
    type CreateCustomerRequest struct {
     Email  string `json:"email"`
     Credit int    `json:"credit"`
    }
    ```

    สร้างไฟล์ `dto/customer_response.go`

    ```go
    package dto
    
    type CreateCustomerResponse struct {
     ID int64 `json:"id"`
    }
    
    func NewCreateCustomerResponse(id int64) *CreateCustomerResponse {
     return &CreateCustomerResponse{ID: id}
    }
    ```

- สร้าง Service สำหรับควบคุม Business Logic ทั้งหมดในการสร้างลูกค้าใหม่

    สร้างไฟล์ `service/customer.go`

    ```go
    package service
    
    import (
     "context"
     "errors"
     "go-mma/dto"
     "go-mma/model"
     "go-mma/repository"
     "go-mma/util/logger"
    )
    
    var (
     ErrEmailExists = errors.New("email already exists")
    )
    
    type CustomerService struct {
     custRepo *repository.CustomerRepository
    }
    
    func NewCustomerService(custRepo *repository.CustomerRepository) *CustomerService {
     return &CustomerService{
      custRepo: custRepo,
     }
    }
    
    func (s *CustomerService) CreateCustomer(ctx context.Context, req *dto.CreateCustomerRequest) (*dto.CreateCustomerResponse, error) {
     // ตรวจสอบความถูกต้องตาม "กฎทางธุรกิจ" (Business Logic/Semantic Validation)
     // Rule: email ต้องไม่ซ้ำในฐานข้อมูล
     exists, err := s.custRepo.ExistsByEmail(ctx, req.Email)
     if err != nil {
      // error logging
      logger.Log.Error(err.Error())
      return nil, err
     }
    
     if exists {
      return nil, ErrEmailExists
     }
    
     // แปลง DTO → Model
     customer := model.NewCustomer(req.Email, req.Credit)
    
     // ส่งไปที่ Repository Layer เพื่อบันทึกข้อมูลลงฐานข้อมูล
     if err := s.custRepo.Create(ctx, customer); err != nil {
      // error logging
      logger.Log.Error(err.Error())
      return nil, err
     }
    
     // สร้าง DTO Response
     resp := dto.NewCreateCustomerResponse(customer.ID)
    
     return resp, nil
    }
    ```

### Presentation Layer – จัดการ HTTP Request/Response

**Presentation Layer (HTTP Handlers)** คือชั้นที่อยู่บนสุดของระบบในสถาปัตยกรรมแบบ Layered Architecture โดยทำหน้าที่เป็น “จุดเชื่อมต่อระหว่างผู้ใช้ (Client) กับระบบ” ผ่านโปรโตคอล เช่น HTTP หรือ WebSocket

หน้าที่หลักของ Presentation Layer (หรือ HTTP Handler)

- **รับคำขอ:** รับ HTTP Request จาก Client
- **แปลงข้อมูล:** แปลง JSON → DTO (ใช้ `BodyParser`, `Bind`, หรือ Unmarshal)
- **ตรวจสอบ:** ตรวจสอบความถูกต้องของ **"รูปแบบ"** และ **"โครงสร้าง"** ของข้อมูลที่ส่งเข้ามา (Input/Syntax Validation) เช่น ค่าว่าง, รูปแบบอีเมล, หรือประเภทข้อมูลที่ถูกต้อง
- **เรียก Service:** ส่ง DTO เข้า Service Layer เพื่อประมวลผล
- **ส่งผลลัพธ์:** รับผลลัพธ์จาก Service แล้วแปลงกลับเป็น JSON Response ส่งกลับไปให้ client
- **จัดการ error**: แปลง error จาก Service ให้เป็น HTTP response code เช่น `400`, `500`

**ขั้นตอนการสร้าง Presentation Layer (HTTP Handlers)**

- แก้ไขไฟล์ `dto/customer_request.go` เพื่อเพิ่ม validation เช่น credit ต้อง ≥ 0 ก่อนส่งให้ Service

    ```go
    package dto
    
    import (
     "errors"
     "net/mail"
    )
    
    // struct
    
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
    ```

- แก้ไขไฟล์ `handler/customer.go` เพื่อให้ทำงานตามหน้าที่ของ Presentation Layer

    ```go
    package handler
    
    import (
     "fmt"
     "go-mma/dto"
     "go-mma/service"
     "go-mma/util/logger"
    
     "github.com/gofiber/fiber/v3"
    )
    
    type CustomerHandler struct {
     custService *service.CustomerService
    }
    
    func NewCustomerHandler(custService *service.CustomerService) *CustomerHandler {
     return &CustomerHandler{
      custService: custService,
     }
    }
    
    func (h *CustomerHandler) CreateCustomer(c fiber.Ctx) error {
     // แปลง request body -> dto
     var req dto.CreateCustomerRequest
     if err := c.Bind().Body(&req); err != nil {
      // แปลงไม่ได้ให้ส่ง error 400
      return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
     }
    
     logger.Log.Info(fmt.Sprintf("Received customer: %v", req))
    
     // ตรวจสอบ input fields (e.g., value, format, etc.)
     if err := req.Validate(); err != nil {
      return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
     }
    
     // ส่งไปที่ Service Layer
     resp, err := h.custService.CreateCustomer(c.Context(), &req)
    
     // จัดการ error จาก Service Layer หากเกิดขึ้น
     if err != nil {
      return c.Status(500).JSON(fiber.Map{"error": err.Error()})
     }
    
     // ตอบกลับด้วย status code 201 (created) และข้อมูลแบบ JSON
     return c.Status(fiber.StatusCreated).JSON(resp)
    }
    ```

### ประกอบร่างระบบให้สมบูรณ์

เมื่อเราแยกเลเยอร์ต่าง ๆ ออกมาแล้ว ขั้นตอนสุดท้ายคือการ “ประกอบ” หรือ “wire” ส่วนต่าง ๆ เข้าด้วยกันในไฟล์ `application/http.go` โดยจะใช้การทำ Dependency Injection แต่ละเลเยอร์เข้าไป

```go
func (s *httpServer) RegisterRoutes(db sqldb.DBContext) {
 v1 := s.app.Group("/api/v1")

 customers := v1.Group("/customers")
 {
   // กำหนด dependency ระหว่างเลเยอร์
  repo := repository.NewCustomerRepository(dbCtx)
  // ส่ง instance ของ repository เข้า service
  svc := service.NewCustomerService(repo)
  // ส่ง service เข้า handler
  hdlr := handler.NewCustomerHandler(svc)
  // Register routes เข้ากับ HTTP server
  customers.Post("", hdlr.CreateCustomer)
 }

 // orders
}
```

