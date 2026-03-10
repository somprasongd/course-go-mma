## Refactor เพิ่ม Integration Event

ในการทำ Integration Event ใน Event-Driven Architecture (EDA) มีหลาย รูปแบบ (patterns) ที่สามารถเลือกใช้ได้ ขึ้นอยู่กับความ ซับซ้อนของระบบ, ระดับการ decouple, และ ความน่าเชื่อถือที่ต้องการ โดยทั่วไปสามารถแบ่งได้เป็น 3 รูปแบบหลัก ๆ ดังนี้

1. **In-Memory Event Bus (Monolith)**

    **ลักษณะ**

    - Event ถูกส่งแบบ in-process (memory) ไปยัง handler ที่ลงทะเบียนไว้ใน runtime เดียวกัน
    - ใช้ในระบบ monolith หรือระบบที่แยกโมดูลแต่ยังรันใน process เดียว

    **ข้อดี**

    - ง่าย
    - เร็ว

    **ข้อเสีย**

    - ไม่ทนต่อ crash
    - ถ้า handler พังหรือ panic → ไม่มี retry
    - ไม่สามารถ scale ข้าม service/process ได้
2. **Outbox Pattern (Reliable Messaging in Monolith / Microservices)**

    **ลักษณะ**

    - เมื่อมี event เกิดขึ้น → บันทึกทั้ง business data + integration event ใน transaction เดียวกัน
    - Event ถูกเก็บใน outbox table
    - Worker (หรือ background process) คอยอ่านและส่งไปยัง message broker (Kafka, RabbitMQ)

    **ข้อดี**

    - ปลอดภัย (atomic): business data + event commit พร้อมกัน
    - ทนต่อ crash
    - Decouple services ได้ (publish ไป Kafka)

    **ข้อเสีย**

    - ต้องมี worker ดึงและส่ง
    - ซับซ้อนกว่า in-memory
3. **Change Data Capture (CDC)**

    **ลักษณะ**

    - ใช้ระบบอย่าง Debezium หรือ Kafka Connect ฟังการเปลี่ยนแปลงใน DB (ผ่าน WAL หรือ binlog)
    - เมื่อมี insert/update → สร้างเป็น event และส่งออกไป message broker

    **ข้อดี**

    - ไม่ต้องมี Worker (หรือ background process) คอยอ่านและส่งไปยัง message broker
    - มองเห็นทุกการเปลี่ยนแปลงของฐานข้อมูล

    **ข้อเสีย**

    - ต้องจัดการ schema evolution และ data format ให้ดี

### In-Memory Event Bus (Monolith)

ในบทความนี้จะเลือกใช้วิธีแบบ In-Memory Event Bus เพราะระบบเป็น Monolith และง่ายต่อการทำความเข้าใจ

การทำ Integration Event แบบ In-Memory Event Bus ภายใน Monolith คือการสื่อสารระหว่างโมดูล (bounded contexts) โดยไม่ใช้ messaging system ภายนอก เช่น Kafka หรือ RabbitMQ แต่ยังแยก "Integration Event" ออกจาก "Domain Event" เพื่อรักษา separation of concerns

มีองค์ประกอบหลัก ดังนี้

1. **Integration Event**
    - เป็น struct ที่ใช้สื่อสารข้ามโมดูล (context) ภายในระบบเดียวกัน
    - มี payload ที่ module ปลายทางต้องใช้ เช่น `CustomerCreatedIntegrationEvent`
2. **Integration Event Interface**
    - ใช้เป็น abstraction สำหรับ event ทั้งหมด เช่น: มี method `EventID()`หรือ `EventName()` หรือ `OccurredAt()`
3. **Event Bus (In-Memory Implementation)**
    - ตัวกลางในการ publish → ไปยัง handler ที่ลงทะเบียนไว้
    - เก็บ handler เป็น map จาก event name → handler list
4. **Register / Subscribe**
    - Module ที่สนใจ event ต้องลงทะเบียน handler ไว้กับ EventBus
5. **Publish**
    - เมื่อ module ต้นทางสร้าง event แล้วเรียก `eventBus.Publish(...)`
    - EventBus จะกระจาย event ไปยัง handler ที่ลงทะเบียนไว้
6. **Event Handlers**
    - แต่ละ handler มี logic ของตัวเอง เช่นส่งอีเมล

### สร้าง Integration Event Interface

- ใช้เป็น abstraction สำหรับ event ทั้งหมด เช่น: มี method `EventID()`หรือ `EventName()` หรือ `OccurredAt()`
- สร้างไฟล์ `common/eventbus/event.go`

    ```go
    package eventbus
    
    import (
     "time"
    )
    
    // EventName เป็นชนิดข้อมูลสำหรับชื่อ event
    type EventName string
    
    // Event คือ interface สำหรับ event ทั่วไปในระบบ
    // ต้องมี method สำหรับดึงข้อมูล ID, ชื่อ event, และเวลาที่เกิด event
    type Event interface {
     EventID() string       // คืนค่า ID ของ event (เช่น UUID หรือ ULID)
     EventName() EventName  // คืนค่าชื่อ event เช่น "CustomerCreated"
     OccurredAt() time.Time // เวลาที่ event นั้นเกิดขึ้น
    }
    
    // BaseEvent คือ struct พื้นฐานที่ใช้เก็บข้อมูล event ทั่วไป
    // สามารถนำไปฝังใน struct event ที่เฉพาะเจาะจงได้
    type BaseEvent struct {
     ID   string    // รหัส event แบบ unique (UUID/ULID)
     Name EventName // ชื่อของ event เช่น "CustomerCreated"
     At   time.Time // เวลาที่ event เกิดขึ้น
    }
    
    // EventID คืนค่า ID ของ event
    func (e BaseEvent) EventID() string {
     return e.ID
    }
    
    // EventName คืนค่าชื่อของ event
    func (e BaseEvent) EventName() EventName {
     return e.Name
    }
    
    // OccurredAt คืนค่าเวลาที่ event เกิดขึ้น
    func (e BaseEvent) OccurredAt() time.Time {
     return e.At
    }
    ```

### สร้าง EventBus (In-Memory Implementation)

สำหรับเป็นตัวกลางในการ publish → ไปยัง handler ที่ลงทะเบียนไว้

- สร้างไฟล์ `common/eventbus/eventbus.go`

    ```go
    package eventbus
    
    import (
     "context"
    )
    
    // IntegrationEventHandler คือ interface สำหรับ handler ที่จะรับผิดชอบการจัดการ Integration Event
    // จะต้องมี method Handle เพื่อรับ context และ event ที่ต้องการจัดการ
    type IntegrationEventHandler interface {
     Handle(ctx context.Context, event Event) error
    }
    
    // EventBus คือ interface สำหรับระบบ event bus ที่ใช้ในการ publish และ subscribe event ต่างๆ
    type EventBus interface {
     // Publish ใช้สำหรับส่ง event ออกไปยังระบบ event bus
     // โดยรับ context และ event ที่จะส่ง
     Publish(ctx context.Context, event Event) error
    
     // Subscribe ใช้สำหรับลงทะเบียน handler สำหรับ event ที่มีชื่อ eventName
     // เมื่อมี event ที่ตรงกับชื่อ eventName เข้ามา handler ที่ลงทะเบียนไว้จะถูกเรียกใช้
     Subscribe(eventName EventName, handler IntegrationEventHandler)
    }
    
    ```

- สร้างไฟล์ `common/eventbus/in_memory_eventbus.go`

    ```go
    package eventbus
    
    import (
     "context"
     "log"
     "sync"
    )
    
    // implementation ของ EventBus แบบง่าย ๆ ที่เก็บ subscriber ไว้ใน memory
    type inmemoryEventBus struct {
     subscribers map[EventName][]IntegrationEventHandler // เก็บ eventName กับ list ของ handler
     mu          sync.RWMutex                            // mutex สำหรับป้องกัน concurrent access
    }
    
    // สร้าง instance ใหม่ของ inmemoryEventBus พร้อม map subscribers ว่าง ๆ
    func NewInMemoryEventBus() EventBus {
     return &inmemoryEventBus{
      subscribers: make(map[EventName][]IntegrationEventHandler),
     }
    }
    
    // Subscribe ใช้ลงทะเบียน handler สำหรับ event ที่มีชื่อ eventName
    // โดยจะเพิ่ม handler เข้าไปใน map subscribers
    func (eb *inmemoryEventBus) Subscribe(eventName EventName, handler IntegrationEventHandler) {
     eb.mu.Lock()
     defer eb.mu.Unlock()
    
     // เพิ่ม handler เข้า slice ของ eventName นั้น ๆ
     eb.subscribers[eventName] = append(eb.subscribers[eventName], handler)
    }
    
    // Publish ส่ง event ไปยัง handler ทุกตัวที่ subscribe event ชื่อเดียวกัน
    func (eb *inmemoryEventBus) Publish(ctx context.Context, event Event) error {
     eb.mu.RLock()
     defer eb.mu.RUnlock()
    
     // หา handler ที่ลงทะเบียนกับ event นี้
     handlers, ok := eb.subscribers[event.EventName()]
     if !ok {
      // ไม่มี handler สำหรับ event นี้ ก็ return nil
      return nil
     }
    
     // สร้าง context ใหม่ที่อาจมีข้อมูลเพิ่มเติมสำหรับ event bus
     busCtx := context.WithValue(ctx, "name", "context in event bus")
    
     // เรียก handler ทุกตัวแบบ asynchronous (goroutine) เพื่อไม่บล็อกการทำงาน
     for _, handler := range handlers {
      go func(h IntegrationEventHandler) {
       // เรียก handle event และ log error ถ้ามี
       err := h.Handle(busCtx, event)
       if err != nil {
        log.Printf("error handling event %s: %v", event.EventName(), err)
       }
      }(handler)
     }
     return nil
    }
    ```

### สร้าง Integration Event

เนื่องจาก integration event จะต้องใช้ร่วมกันหลายโมดูล ให้สร้างโปรเจกต์ใหม่ชื่อ messaging ใน shared

- สร้างโปรเจกต์ `messaging`

    ```bash
    mkdir -p src/shared/messaging
    cd src/shared/messaging
    go mod init go-mma/shared/messaging
    ```

    เพิ่มลง workspace ด้วย

    ```go
    {
      "folders": [
        // ...,
        {
          "path": "src/shared/messaging"
        }
      ],
      "settings": {}
    }
    ```

    เพิ่ม module replace สำหรับโปรเจกต์ `common`

    ```go
    module go-mma/shared/messaging
    
    go 1.24.1
    
    replace go-mma/shared/common v0.0.0 => ../common
    
    require go-mma/shared/common v0.0.0
    ```

- เพิ่ม module replace ในทุกโมดูลโปรเจกต์ และ `app`

    ```go
    // app
    replace go-mma/shared/messaging v0.0.0 => ../shared/messaging
    
    // customer
    replace go-mma/shared/messaging v0.0.0 => ../../shared/messaging
    
    // order
    replace go-mma/shared/messaging v0.0.0 => ../../shared/messaging
    
    // notification
    replace go-mma/shared/messaging v0.0.0 => ../../shared/messaging
    ```

- สร้างไฟล์ `messaging/customer_created.go`

    ```go
    package messaging
    
    import (
     "go-mma/shared/common/eventbus"
     "go-mma/shared/common/idgen"
     "time"
    )
    
    const (
     CustomerCreatedIntegrationEventName eventbus.EventName = "CustomerCreated"
    )
    
    type CustomerCreatedIntegrationEvent struct {
     eventbus.BaseEvent
     CustomerID int64  `json:"customer_id"`
     Email      string `json:"email"`
    }
    
    func NewCustomerCreatedIntegrationEvent(customerID int64, email string) *CustomerCreatedIntegrationEvent {
     return &CustomerCreatedIntegrationEvent{
      BaseEvent: eventbus.BaseEvent{
       ID:   idgen.GenerateUUIDLikeID(),
       Name: CustomerCreatedIntegrationEventName,
       At:   time.Now(),
      },
      CustomerID: customerID,
      Email:      email,
     }
    }
    ```

- ติดตั้ง dependencies ด้วย `go mod tidy`

### สร้าง Integration Event Handler

สำหรับโค้ดที่รับ event “`CustomerCreated`” มาทำงานต่อ โดยใยที่นี่จะทำที่โมดูล notification เพื่อส่ง welcome email

- สร้างไฟล์ `notification/internal/integration/customer/welcome_email_handler.go` (สื่อว่า integration จากโมดูล customer)

    ```go
    package customer
    
    import (
     "context"
     "fmt"
     "go-mma/modules/notification/service"
     "go-mma/shared/common/eventbus"
     "go-mma/shared/messaging"
    )
    
    type welcomeEmailHandler struct {
     notiService service.NotificationService
    }
    
    func NewWelcomeEmailHandler(notiService service.NotificationService) *welcomeEmailHandler {
     return &welcomeEmailHandler{
      notiService: notiService,
     }
    }
    
    func (h *welcomeEmailHandler) Handle(ctx context.Context, evt eventbus.Event) error {
     e, ok := evt.(*messaging.CustomerCreatedIntegrationEvent) // ใช้ pointer
     if !ok {
      return fmt.Errorf("invalid event type")
     }
    
     return h.notiService.SendEmail(e.Email, "Welcome to our service!", map[string]any{
      "message": "Thank you for joining us! We are excited to have you as a member.",
     })
    }
    ```

- ติดตั้ง dependencies ด้วย `go mod tidy`

### สร้าง Integration Event Publisher

เดิมใน CustomerCreatedDomainEventHandler จะมีการเรียก notiService เพื่อส่งอีเมลโดยตรง เราจะเปลี่ยนส่งนี้ให้ส่งไปเป็น integration event แทน

- แก้ไขไฟล์  `customer/internal/domain/eventhandler/customer_created_handler.go`

    ```go
    package eventhandler
    
    import (
     "context"
     "go-mma/modules/customer/internal/domain/event"
     "go-mma/shared/common/domain"
     "go-mma/shared/common/eventbus"
     "go-mma/shared/messaging"
    )
    
    type customerCreatedDomainEventHandler struct {
     eventBus eventbus.EventBus // เปลี่ยนมาใช้ eventbus
    }
    
    // เปลี่ยนมาใช้ eventbus
    func NewCustomerCreatedDomainEventHandler(eventBus eventbus.EventBus) domain.DomainEventHandler {
     return &customerCreatedDomainEventHandler{
      eventBus: eventBus, // เปลี่ยนมาใช้ eventbus
     }
    }
    
    func (h *customerCreatedDomainEventHandler) Handle(ctx context.Context, evt domain.DomainEvent) error {
     e, ok := evt.(*event.CustomerCreatedDomainEvent) // ใช้ pointer
    
     if !ok {
      return domain.ErrInvalidEvent
     }
    
     // สร้าง IntegrationEvent จาก Domain Event
     integrationEvent := messaging.NewCustomerCreatedIntegrationEvent(
      e.CustomerID,
      e.Email,
     )
    
     return h.eventBus.Publish(ctx, integrationEvent)
    }
    ```

- ติดตั้ง dependencies ด้วย `go mod tidy`
- แก้ไขไฟล์  `common/module/module.go` เพื่อให้ `Init()` รองรับ event bus

    ```go
    type Module interface {
     APIVersion() string
     Init(reg registry.ServiceRegistry, eventBus eventbus.EventBus) error // รับ eventBus เพิ่ม
     RegisterRoutes(r fiber.Router)
    }
    ```

- แก้ไขไฟล์  `customer/module.go` เพื่อลบ notification service ออก

    ```go
    func (m *moduleImp) Init(reg registry.ServiceRegistry, eventBus eventbus.EventBus) error {
     // เอา notiSvc ออก
     
     // Register domain event handler
     dispatcher := domain.NewSimpleDomainEventDispatcher()
     dispatcher.Register(event.CustomerCreatedDomainEventType, eventhandler.NewCustomerCreatedDomainEventHandler(eventBus)) // ส่ง eventBus เข้าไปแทน
    
     repo := repository.NewCustomerRepository(m.mCtx.DBCtx)
    
     mediator.Register(create.NewCreateCustomerCommandHandler(m.mCtx.Transactor, repo, dispatcher))
     mediator.Register(getbyid.NewGetCustomerByIDQueryHandler(repo))
     mediator.Register(reservecredit.NewReserveCreditCommandHandler(m.mCtx.Transactor, repo))
     mediator.Register(releasecredit.NewReleaseCreditCommandHandler(m.mCtx.Transactor, repo))
    
     return nil
    }
    ```

- ต้องแก้ไข `Init()` ที่โมดูล `order` กับ `notification` ด้วย
- แก้ไฟล์ `app/application/application.go` เพื่อให้ส่ง event bus เข้าไปตอน init module

    ```go
    package application
    
    import (
      // ...
     "go-mma/shared/common/eventbus"
      // ...
    )
    
    type Application struct {
     config          config.Config
     httpServer      HTTPServer
     serviceRegistry registry.ServiceRegistry
     eventBus        eventbus.EventBus // เพ่ิม
    }
    
    func New(config config.Config) *Application {
     return &Application{
      config:          config,
      httpServer:      newHTTPServer(config),
      serviceRegistry: registry.NewServiceRegistry(),
      eventBus:        eventbus.NewInMemoryEventBus(), // เพ่ิม
     }
    }
    
    // ...
    
    func (app *Application) initModule(m module.Module) error {
     return m.Init(app.serviceRegistry, app.eventBus) // เพ่ิมส่ง eventBus
    }
    
    // ...
    ```

  - ติดตั้ง dependencies ด้วย `go mod tidy` ที่โปรเจกต์ `app`

### สร้าง Register / Subscribe

ให้โมดูล notification คอยรับ integration event

- แก้ไขไฟล์ `notification/module.go`

    ```go
    func (m *moduleImp) Init(reg registry.ServiceRegistry, eventBus eventbus.EventBus) error {
     m.notiSvc = service.NewNotificationService()
    
     // subscribe to integration events
     eventBus.Subscribe(messaging.CustomerCreatedIntegrationEventName, customer.NewWelcomeEmailHandler(m.notiSvc))
    
     return nil
    }
    ```

เพียงเท่านี้ก็สามารถใช้ integration event แบบ in-memory event bus ได้แล้ว แต่อย่าลืมว่าวิธีมีข้อเสียคือ ถ้า handler พังหรือ panic จะไม่มี retry อาจทำให้เกิด inconsistency เช่น ลูกค้าถูกสร้างแล้ว (`INSERT INTO customers`) แต่ไม่ส่งอีเมลต้อนรับ วิธีแก้ ได้แก่

- อาจเพิ่ม retry logic ตอนส่งอีเมล
- ใช้แนวทาง Hybrid Approach คือ ใช้ Domain Event → แปลง (map) เป็น Integration Event → เขียน Outbox table (ต้องทำใน transaction เดียวกับ business data)→ ใช้ CDC tools ส่ง event

---

