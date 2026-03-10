## เชื่อมต่อฐานข้อมูลอย่างปลอดภัยและยืดหยุ่น

จากโค้ดก่อนหน้า เรายังไม่สามารถดำเนินการได้สมบูรณ์ เพราะจำเป็นต้องเชื่อมต่อและบันทึกข้อมูลลงฐานข้อมูลก่อน ในส่วนนี้จะประกอบด้วยขั้นตอนสำคัญดังนี้

- ติดตั้ง PostgreSQL ด้วย Docker
- ออกแบบ Schema ให้เหมาะสม
- จัดการ Schema ด้วย Migration Tool
- ตั้งค่า Database Connection อย่างปลอดภัย
- สร้างฟังก์ชันสำหรับ Gen ID
- เชื่อมต่อ Database ด้วย Dependency Injection
- เพิ่มความสามารถในการสร้างข้อมูลลูกค้า (Insert Customer)

### ติดตั้ง PostgreSQL ด้วย Docker

ในบทความนี้จะใช้ PostgreSQL และติดตั้งด้วย docker

- สร้างไฟล์ `docker-compose.yml` สำหรับสร้าง postgres service

    ```yaml
    services:
      db:
        image: postgres:17-alpine
        container_name: go-mma-db
        volumes:
          - pg_data:/var/lib/postgresql/data
    
    volumes:
      pg_data:
    ```

- สร้างไฟล์ `docker-compose.dev.yml` สำหรับกำหนด environtment ของ `dev`

    ```yaml
    services:
      db:
        environment:
          POSTGRES_DB: go-mma-db
          POSTGRES_USER: postgres
          POSTGRES_PASSWORD: postgres
        ports:
          - 5433:5432
    ```

    <aside>
    💡

    สามารถเพิ่มสำหรับ environtment ของ `test` และ `production` ได้ภายหลัง

    </aside>

- รัน PostgreSQL Server ด้วย `Makefile` โดยเพิ่มคำสั่ง ดังนี้

    ```bash
    # ...
    
    .PHONY: devup
    devup:
     docker compose -f docker-compose.yml -f docker-compose.dev.yml up -d
    
    .PHONY: devdown
    devdown:
     docker compose -f docker-compose.yml -f docker-compose.dev.yml down
    ```

- รันคำสั่ง `make devup`

### ออกแบบ Schema ให้เหมาะสม

```sql
+------------------+           +------------------+
|   customers      |           |     orders       |
+------------------+           +------------------+
| id (PK)          |<--------+ | id (PK)          |
| email (UNIQUE)   |         | | customer_id (FK) |
| credit           |         | | order_total      |
| created_at       |         | | created_at       |
| updated_at       |         | | canceled_at       |
+------------------+         | +------------------+
                             |
      [1] ---------------- [*]
     1 customer      →   many orders
```

### จัดการ Schema ด้วย Migration Tool

**Database Migration** คือกระบวนการจัดการและติดตามการเปลี่ยนแปลงโครงสร้างฐานข้อมูล (เช่น ตาราง, คอลัมน์, ดัชนี ฯลฯ) ด้วยสคริปต์ที่สามารถรันซ้ำได้อย่างปลอดภัยในทุกสภาพแวดล้อม (dev, staging, production) เพื่อให้ทีมพัฒนาและระบบสามารถทำงานร่วมกันได้อย่างราบรื่น

- เพิ่ม Environment Variable สำหรับการเชื่อมต่อ

    เพิ่มตัวแปร `DB_DSN` ในไฟล์ `.env` เพื่อเก็บ connection string ของ PostgreSQL

    ```
    DB_DSN=postgres://postgres:postgres@localhost:5433/go-mma-db?sslmode=disable
    ```

- เพิ่มคำสั่ง Migration ใน Makefile
แก้ไข `Makefile` เพื่อเพิ่มคำสั่งสำหรับสร้างและรัน migration ผ่าน Docker

    ```makefile
    # ...
    
    .PHONY: mgc
    # Example: make mgc filename=create_customer
    mgc:
     docker run --rm -v $(ROOT_DIR)migrations:/migrations migrate/migrate -verbose create -ext sql -dir /migrations $(filename)
    
    .PHONY: mgu
    mgu:
     docker run --rm --network host -v $(ROOT_DIR)migrations:/migrations migrate/migrate -verbose -path=/migrations/ -database "$(DB_DSN)" up
    
    .PHONY: mgd
    mgd:
     docker run --rm --network host -v $(ROOT_DIR)migrations:/migrations migrate/migrate -verbose -path=/migrations/ -database $(DB_DSN) down 1
    ```

    <aside>
    💡

    หากใช้ Docker Desktop ต้องเปิดใช้งาน host networking โดยไปที่ `Settings → Resources → Network` และเลือก "Enable host networking" จากนั้นกด Apply & Restart

    </aside>

- สร้าง Migration สำหรับ Customer
ใช้คำสั่งด้านล่างเพื่อสร้าง migration สำหรับตาราง `customers`

    ```bash
    make mgc filename=create_customer
    ```

    ระบบจะสร้างไฟล์ 2 ไฟล์

    ```bash
    ./migrations/20250529103238_create_customer.up.sql
    ./migrations/20250529103238_create_customer.down.sql
    ```

    แก้ไขไฟล์ `create_customer.up.sql`

    ```sql
    CREATE TABLE public.customers (
     id BIGINT NOT NULL,
     email text NOT NULL,
     credit int4 NOT NULL,
     created_at timestamp DEFAULT CURRENT_TIMESTAMP NULL,
     updated_at timestamp DEFAULT CURRENT_TIMESTAMP NULL,
     CONSTRAINT customers_pkey PRIMARY KEY (id),
     CONSTRAINT customers_unique UNIQUE (email)
    );
    ```

    และไฟล์ `create_customer.down.sql`

    ```sql
    drop table public.customers;
    ```

- สร้าง Migration สำหรับ Order
ใช้คำสั่ง

    ```bash
    make mgc filename=create_order
    ```

    จะได้ไฟล์

    ```bash
    ./migrations/20250529103715_create_order.up.sql
    ./migrations/20250529103715_create_order.down.sql
    ```

    แก้ไขไฟล์ `create_order.up.sql`

    ```sql
    CREATE TABLE public.orders (
     id BIGINT NOT NULL,
     customer_id BIGINT NOT NULL,
     order_total int4 NOT NULL,
     created_at timestamp DEFAULT CURRENT_TIMESTAMP NULL,
     canceled_at timestamp NULL,
     CONSTRAINT orders_pkey PRIMARY KEY (id),
     CONSTRAINT fk_customer FOREIGN KEY (customer_id) REFERENCES public.customers(id)
    );
    ```

    และ `create_order.down.sql`

    ```sql
    drop table public.orders;
    ```

- รัน Migration เพื่อสร้างตาราง
ใช้คำสั่ง

    ```bash
    make mgu
    ```

    ระบบจะแสดงผลลัพธ์การรัน migration เช่น

    ```bash
    2025/05/29 10:39:20 Start buffering 20250529103238/u create_customer
    2025/05/29 10:39:20 Start buffering 20250529103715/u create_order
    2025/05/29 10:39:20 Read and execute 20250529103238/u create_customer
    2025/05/29 10:39:20 Finished 20250529103238/u create_customer (read 906.667µs, ran 2.125583ms)
    2025/05/29 10:39:20 Read and execute 20250529103715/u create_order
    2025/05/29 10:39:20 Finished 20250529103715/u create_order (read 3.458625ms, ran 1.860583ms)
    2025/05/29 10:39:20 Finished after 7.190625ms
    2025/05/29 10:39:20 Closing source and database
    ```

### ตั้งค่า Database Connection อย่างปลอดภัย

ก่อนจะเริ่มพัฒนา API จำเป็นต้องตั้งค่าการเชื่อมต่อกับฐานข้อมูลให้เรียบร้อย เพื่อให้ระบบสามารถใช้งานฐานข้อมูล PostgreSQL ได้อย่างมีประสิทธิภาพและปลอดภัย

- เพิ่มคอนฟิก `DB_DSN` ในไฟล์ `config/config.go` เพื่อระบุข้อมูลการเชื่อมต่อฐานข้อมูล

    ```go
    package config
    
    import (
     "errors"
     "go-mma/util/env"
     "time"
    )
    
    var (
     ErrInvalidHTTPPort = errors.New("HTTP_PORT must be a positive integer")
     ErrGracefulTimeout = errors.New("GRACEFUL_TIMEOUT must be a positive duration")
     ErrDSN             = errors.New("DB_DSN must be set") // เพิ่ม
    )
    
    type Config struct {
     HTTPPort        int
     GracefulTimeout time.Duration
     DSN             string  // เพิ่ม
    }
    
    func Load() (*Config, error) {
     config := &Config{
      HTTPPort:        env.GetIntDefault("HTTP_PORT", 8090),
      GracefulTimeout: env.GetDurationDefault("GRACEFUL_TIMEOUT", 5*time.Second),
      DSN:             env.Get("DB_DSN"),  // เพิ่ม
     }
     err := config.Validate()
     if err != nil {
      return nil, err
     }
     return config, err
    }
    
    func (c *Config) Validate() error {
     if c.HTTPPort <= 0 {
      return ErrInvalidHTTPPort
     }
     if c.GracefulTimeout <= 0 {
      return ErrGracefulTimeout
     }
     // เพิ่ม
     if len(c.DSN) == 0 {
      return ErrDSN
     }
    
     return nil
    }
    ```

- สร้างไฟล์ `util/storage/sqldb/sqldb.go` สำหรับจัดการการเชื่อมต่อกับฐานข้อมูล PostgreSQL โดยใช้ `sqlx`

    ```go
    package sqldb
    
    import (
     "github.com/jmoiron/sqlx"
     _ "github.com/lib/pq"
    )
    
    type closeDB func() error
    
    type DBContext interface {
     DB() *sqlx.DB
    }
    
    type dbContext struct {
     db *sqlx.DB
    }
    
    var _ DBContext = (*dbContext)(nil)
    
    func NewDBContext(dsn string) (DBContext, closeDB, error) {
     // this Pings the database trying to connect
     db, err := sqlx.Connect("postgres", dsn)
     if err != nil {
      return nil, nil, err
     }
     return &dbContext{db: db},
      func() error {
       return db.Close()
      },
      nil
    }
    
    func (c *dbContext) DB() *sqlx.DB {
     return c.db
    }
    ```

- รันคำสั่ง `go mod tidy` เพื่อดึง dependencies ที่จำเป็น
- ปรับปรุงไฟล์ `application/application.go` เพื่อเก็บอินสแตนซ์ของ database connection

    ```go
    type Application struct {
     config     config.Config
     httpServer HTTPServer
     dbCtx      sqldb.DBContext
    }
    
    func New(config config.Config, dbCtx sqldb.DBContext) *Application {
     return &Application{
      config:     config,
      httpServer: newHTTPServer(config),
      dbCtx:         dbCtx,
     }
    }
    ```

- เชื่อมต่อฐานข้อมูลใน `cmd/api/main.go` และส่งผ่านเข้าไปใน Application

    ```go
    func main() {
     // config
     
     dbCtx, closeDB, err := sqldb.NewDBContext(config.DSN)
     if err != nil {
      panic(err.Error())
     }
     defer func() { // ใช่ท่า IIFE เพราะต้องการแสดง error ถ้าปิดไม่ได้
      if err := closeDB(); err != nil {
       logger.Log.Error(fmt.Sprintf("Error closing database: %v", err))
      }
     }()
    
     app := application.New(*config, dbCtx)
     // ...
    }
    ```

### สร้างฟังก์ชันสำหรับ Gen ID

เพื่อให้สามารถสร้าง ID สำหรับลูกค้า (หรือ entity อื่น ๆ) ได้อย่างยืดหยุ่น จึงต้องแยกฟังก์ชันสำหรับการ generate ID ออกเป็น utility

- สร้างไฟล์ `util/idgen/idgen.go` เพื่อรวมฟังก์ชันสร้าง ID หลายรูปแบบ เช่น
  - `GenerateTimeRandomID`: สร้างเลขสุ่มแบบ `int64` โดยอิงจาก timestamp (ใช้ตัวนี้)
  - `GenerateTimeID`: สร้างเลข `int` โดยใช้เวลาปัจจุบัน
  - `GenerateTimeRandomIDBase36`: แปลง ID เป็น string แบบ base36
  - `GenerateUUIDLikeID`: สร้าง string ที่มีรูปแบบคล้าย UUID

    ```go
    package idgen
    
    import (
     "fmt"
     "math/rand"
     "strconv"
     "strings"
     "time"
    )
    
    // GenerateTimeRandomID สร้าง ID แบบ int64
    func GenerateTimeRandomID() int64 {
     timestamp := time.Now().UnixNano() >> 32
     randomPart := rand.Int63() & 0xFFFFFFFF
     return (timestamp << 32) | randomPart
    }
    
    // GenerateTimeID สร้าง ID แบบ int (ใช้ timestamp เป็นหลัก)
    func GenerateTimeID() int {
     // ใช้ timestamp Unix วินาที (int64) แปลงเป็น int (int32/64 ขึ้นกับระบบ)
     return int(time.Now().Unix())
    }
    
    // GenerateTimeRandomIDBase36 คืนค่า ID เป็น base36 string
    func GenerateTimeRandomIDBase36() string {
     id := GenerateTimeRandomID()
     return strconv.FormatInt(id, 36) // แปลงเลขฐาน 10 -> 36
    }
    
    // GenerateUUIDLikeID คืนค่าเป็น string แบบ UUID-like (แต่ไม่ใช่ UUID จริง)
    func GenerateUUIDLikeID() string {
     id := GenerateTimeRandomID()
    
     // แปลง int64 เป็น hex string ยาว 16 ตัว (64 bit)
     hex := fmt.Sprintf("%016x", uint64(id))
    
     // สร้าง UUID-like string รูปแบบ 8-4-4-4-12
     // แต่มีแค่ 16 hex chars แบ่งคร่าวๆ: 8-4-4 (เหลือไม่พอจริงๆ)
     // ดังนั้นเราจะเติม random เพิ่มเพื่อครบ 32 hex (128 bit) เหมือน UUID
    
     randPart := fmt.Sprintf("%016x", rand.Uint64())
    
     uuidLike := strings.Join([]string{
      hex[0:8],
      hex[8:12],
      hex[12:16],
      randPart[0:4],
      randPart[4:16],
     }, "-")
    
     return uuidLike
    }
    
    // ก่อน Go 1.20 ต้องเรียก เพื่อให้ได้เลขสุ่มไม่ซ้ำ
    // func init() {
    //  rand.Seed(time.Now().UnixNano())
    // }
    ```

### เชื่อมต่อ Database ด้วย Dependency Injection

เพื่อให้ระบบมีความยืดหยุ่นและทดสอบง่ายขึ้น เราใช้หลักการ Dependency Injection (DI) ในการส่ง database connection เข้าไปยัง handler

- แก้ไขไฟล์ `handlers/customer.go` โดยปรับโครงสร้าง `CustomerHandler` เพื่อรับ `DBContext` ผ่าน constructor

    ```go
    type CustomerHandler struct {
     dbCtx sqldb.DBContext
    }
    
    func NewCustomerHandler(db sqldb.DBContext) *CustomerHandler {
     return &CustomerHandler{dbCtx: db}
    }
    ```

- แก้ไข `application/http.go` เพื่อส่ง `DBContext` ไปยัง handler

    ```go
    type HTTPServer interface {
     Start()
     Shutdown() error
     RegisterRoutes(dbCtx sqldb.DBContext)
    }
    
    // ...
    
    func (s *httpServer) RegisterRoutes(dbCtx sqldb.DBContext) {
     v1 := s.app.Group("/api/v1")
    
     customers := v1.Group("/customers")
     {
      hdlr := handlers.NewCustomerHandler(dbCtx)
      customers.Post("", hdlr.CreateCustomer)
     }
    
     // orders
    }
    ```

- เพิ่มเมธอด `RegisterRoutes()` ใน `application/application.go` เพื่อจัดการเส้นทาง HTTP โดยส่ง db context เข้าไป

    ```go
    func (app *Application) RegisterRoutes() {
     app.httpServer.RegisterRoutes(app.db)
    }
    ```

### เพิ่มความสามารถในการสร้างข้อมูลลูกค้า (Insert Customer)

หลังจากเชื่อมต่อฐานข้อมูลเรียบร้อยแล้ว ขั้นตอนต่อไปคือการเพิ่มความสามารถในการสร้างข้อมูลลูกค้าผ่าน HTTP POST

- แก้ไขไฟล์ `handlers/customer.go` เพื่อบันทึกข้อมูลลงตาราง customers

    ```go
    func (h *CustomerHandler) CreateCustomer(c fiber.Ctx) error {
     // กำหนด payload structure (DTO: Request)
     // แปลง request body -> dto
     // ตรวจสอบ input fields (e.g., value, format, etc.)
     // ตรวจสอบความถูกต้องตาม "กฎทางธุรกิจ" (Business Logic/Semantic Validation)
     // Rule: email ต้องไม่ซ้ำในฐานข้อมูล
     query := "SELECT 1 FROM public.customers where email = $1 LIMIT 1"
     ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
     defer cancel()
    
     var exists int
     if err := h.dbCtx.DB().QueryRowxContext(ctx, query, req.Email).Scan(&exists); err != nil {
      if err != sql.ErrNoRows {
       logger.Log.Error(fmt.Sprintf("error checking email: %v", err))
       return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "an error occurred while checking email"})
      }
     }
     if exists == 1 {
      return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "email already exists"})
     }
    
     // บันทึกลงฐานข้อมูล
     var id int64
     query = "INSERT INTO customers (id, email, credit) VALUES ($1, $2, $3) RETURNING id"
     ctxIns, cancelIns := context.WithTimeout(c.Context(), 10*time.Second)
     defer cancelIns()
     if err := h.dbCtx.DB().QueryRowxContext(ctxIns, query, idgen.GenerateTimeRandomID(), req.Email, req.Credit).Scan(&id); err != nil {
      logger.Log.Error(fmt.Sprintf("error insert customer: %v", err))
      return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "an error occurred while inserting customer"})
     }
    
     // กำหนดโครงสร้างข้อมูลที่จะส่งกลับไป (DTO: Response)
     // ตอบกลับด้วย status code 201 (created) และข้อมูลแบบ JSON
    }
    
    ```

