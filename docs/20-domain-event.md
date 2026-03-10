## Refactor เพิ่ม Domain Event

การทำ Domain Event ให้สมบูรณ์ในระบบที่ใช้ DDD (Domain-Driven Design) และ Event-Driven Architecture มีองค์ประกอบหลัก ดังนี้

1. **Domain Event**
    - เป็น struct ที่บรรยายเหตุการณ์ที่ “เกิดขึ้นแล้ว” ใน domain
    - อยู่ใน layer `domain` หรือ `internal/domain/event`
2. **Aggregate/Entity ที่สร้าง Event**
    - Entity เช่น `Customer` ต้องมีช่องทางในการเก็บ domain events (เช่น slice `[]DomainEvent`)
    - เมื่อเกิดเหตุการณ์ ให้ `append()` ลงไป
3. **DomainEvent Interface**
    - ใช้เป็น abstraction สำหรับ event ทั้งหมด เช่น: มี method `EventName()` หรือ `OccurredAt()`
4. **Event Dispatcher**
    - ดึง events จาก aggregate แล้ว dispatch ไปยังผู้รับ (handler)
5. **Event Handler**
    - โค้ดที่รับ event และทำงานตอบสนอง
    - อยู่ใน layer `domain` หรือ `internal/domain/eventhandler`
6. **Trigger Point**
    - จุดที่ pull domain events เพื่อนำไปส่งผ่าน dispatcher (มักอยู่หลัง transaction สำเร็จ)
7. **Dispatch Events มี 2 แนวทางหลัก**
    - ภายใน transaction (immediate dispatch) เหมาะกับ use case ที่ event handler แค่ปรับ state ภายใน เช่น update model อื่น ซึ่งจะ coupling กับ transaction logic ถ้า event handler fail จะต้อง rollback transaction ด้วย
    - หลังจาก commit แล้ว คือ ดึง domain events → รอ DB commit → dispatch เช่น post-commit hook เหมาะกับ handler ที่มี side-effect เช่น ส่งอีเมล, call external service แต่ต้องมีการจัดการ error และ retry เอง แยกออกมาจาก transaction logic

### DomainEvent Interface

ใช้เป็น abstraction สำหรับ event ทั้งหมด เช่น: มี method `EventName()` หรือ `OccurredAt()`

- สร้างไฟล์ `common/domain/event.go`

    ```go
    package domain
    
    import "time"
    
    // EventName คือ alias ของ string เพื่อใช้แทนชื่อ event เช่น "CustomerCreated", "OrderPlaced" เป็นต้น
    type EventName string
    
    // DomainEvent เป็น interface สำหรับ event ที่เกิดขึ้นใน domain (Domain Event ตาม DDD)
    // ใช้เพื่อให้สามารถบันทึกหรือส่ง event ได้ โดยไม่ต้องรู้โครงสร้างภายใน
    type DomainEvent interface {
     EventName() EventName     // คืนชื่อ event
     OccurredAt() time.Time    // คืนเวลาที่ event เกิด
    }
    
    // BaseDomainEvent เป็น struct พื้นฐานที่ implement DomainEvent
    // ใช้ฝังใน struct อื่นๆ ที่เป็น event เพื่อ reuse method ได้
    type BaseDomainEvent struct {
     Name EventName   // ชื่อของ event เช่น "UserRegistered"
     At   time.Time   // เวลาที่ event นี้เกิดขึ้น
    }
    
    // EventName คืนชื่อของ event นี้
    func (e BaseDomainEvent) EventName() EventName {
     return e.Name
    }
    
    // OccurredAt คืนเวลาที่ event นี้เกิดขึ้น
    // ใช้แบบ value receiver ด้วยเหตุผลเดียวกันกับข้างต้น
    func (e BaseDomainEvent) OccurredAt() time.Time {
     return e.At
    }
    
    ```

    <aside>
    💡

    ใช้ **value receiver** เพราะ struct เล็ก ไม่มี mutation และปลอดภัยในเชิง concurrent

    </aside>

### Aggregate

`Aggregate` เป็น **ฐานแม่แบบ (base struct)** สำหรับ aggregate root ทุกตัว เช่น `Customer`, `Order` ฯลฯ จะทำหน้าที่ **บันทึกเหตุการณ์ที่เกิดขึ้น (Domain Events)** เพื่อให้ layer ภายนอก (เช่น Application หรือ Infrastructure) นำไปประมวลผลต่อ

- สร้างไฟล์ `common/domain/aggregate.go`

    ```go
    package domain
    
    // Aggregate เป็น struct พื้นฐานสำหรับ aggregate root ทั้งหมดใน DDD
    // ใช้เพื่อเก็บรวบรวม domain events ที่เกิดขึ้นภายใน aggregate
    type Aggregate struct {
     domainEvents []DomainEvent // เก็บรายการของ event ที่เกิดขึ้นใน aggregate นี้
    }
    
    // AddDomainEvent ใช้สำหรับเพิ่ม domain event เข้าไปใน aggregate
    // ฟังก์ชันนี้จะถูกเรียกภายใน method อื่น ๆ ของ aggregate เมื่อต้องการประกาศว่า event บางอย่างได้เกิดขึ้นแล้ว
    func (a *Aggregate) AddDomainEvent(dv DomainEvent) {
     // สร้าง slice เปล่าหากยังไม่มี
     if a.domainEvents == nil {
      a.domainEvents = make([]DomainEvent, 0)
     }
    
     // เพิ่ม event ลงใน slice
     a.domainEvents = append(a.domainEvents, dv)
    }
    
    // PullDomainEvents จะดึง domain events ทั้งหมดออกจาก aggregate
    // พร้อมกับเคลียร์ events เหล่านั้นจาก memory (เพราะ events ถูกส่งออกไปแล้ว)
    // เหมาะสำหรับใช้ใน layer ที่ทำการ publish หรือ persist event
    func (a *Aggregate) PullDomainEvents() []DomainEvent {
     events := a.domainEvents    // ดึง event ทั้งหมดที่บันทึกไว้
     a.domainEvents = nil        // เคลียร์ event list เพื่อป้องกันการส่งซ้ำ
     return events
    }
    ```

### Event Dispatcher

สำหรับการ register handler และ dispatch ไปยังผู้รับ (handler)

- สร้างไฟล์ `common/domain/event_dispatcher.go`

    ```go
    package domain
    
    import (
     "context"
     "fmt"
     "sync"
    )
    
    // Error ที่ใช้ตรวจสอบความถูกต้องของ event
    var (
     ErrInvalidEvent = fmt.Errorf("invalid domain event")
    )
    
    // DomainEventHandler คือ interface ที่ทุก handler ของ event ต้อง implement
    // โดยจะมี method เดียวคือ Handle เพื่อรับ event และทำงานตาม logic ที่ต้องการ
    type DomainEventHandler interface {
     Handle(ctx context.Context, event DomainEvent) error
    }
    
    // DomainEventDispatcher คือ interface สำหรับระบบที่ทำหน้าที่กระจาย (dispatch) event
    // โดยสามารถ register handler สำหรับแต่ละ EventName และ dispatch หลาย event พร้อมกันได้
    type DomainEventDispatcher interface {
     Register(eventType EventName, handler DomainEventHandler)
     Dispatch(ctx context.Context, events []DomainEvent) error
    }
    
    // simpleDomainEventDispatcher เป็น implementation ง่าย ๆ ของ DomainEventDispatcher
    // ใช้ map เก็บ handler แยกตาม EventName
    type simpleDomainEventDispatcher struct {
     handlers map[EventName][]DomainEventHandler // แผนที่ของ EventName ไปยัง handler หลายตัว
     mu       sync.RWMutex                       // ใช้ mutex เพื่อป้องกัน concurrent read/write
    }
    
    // NewSimpleDomainEventDispatcher สร้าง instance ใหม่ของ dispatcher
    func NewSimpleDomainEventDispatcher() DomainEventDispatcher {
     return &simpleDomainEventDispatcher{
      handlers: make(map[EventName][]DomainEventHandler),
     }
    }
    
    // Register ใช้สำหรับลงทะเบียน handler กับ EventName
    // Handler จะถูกเรียกเมื่อมี event นั้น ๆ ถูก dispatch
    func (d *simpleDomainEventDispatcher) Register(eventType EventName, handler DomainEventHandler) {
     d.mu.Lock()
     defer d.mu.Unlock()
    
     // เพิ่ม handler ไปยัง slice ของ event นั้น ๆ
     d.handlers[eventType] = append(d.handlers[eventType], handler)
    }
    
    // Dispatch รับ slice ของ event แล้ว dispatch ไปยัง handler ที่ลงทะเบียนไว้
    // ถ้ามี handler มากกว่าหนึ่งตัวสำหรับ event เดียวกัน จะเรียกทุกตัว
    func (d *simpleDomainEventDispatcher) Dispatch(ctx context.Context, events []DomainEvent) error {
     for _, event := range events {
      // อ่าน handler ของ event นี้ (copy slice เพื่อป้องกัน concurrent modification)
      d.mu.RLock()
      handlers := append([]DomainEventHandler(nil), d.handlers[event.EventName()]...)
      d.mu.RUnlock()
    
      // เรียก handler แต่ละตัว
      for _, handler := range handlers {
       err := func(h DomainEventHandler) error {
        // หาก handler ทำงานผิดพลาด จะคืน error พร้อมระบุ event ที่ผิด
        err := h.Handle(ctx, event)
        if err != nil {
         return fmt.Errorf("error handling event %s: %w", event.EventName(), err)
        }
        return nil
       }(handler)
    
       // หากมี error จาก handler ใด ๆ จะหยุดและ return เลย
       if err != nil {
        return err
       }
      }
     }
    
     // ถ้าไม่มี error เลย ส่ง nil กลับ
     return nil
    }
    ```

### Domain Event

- สร้าง domain event สำหรับเมื่อสร้างลูกค้าใหม่สำเร็จ

    > สร้างไฟล์ `customer/internal/domain/event/customer_created.go`
    >

    ```go
    package event
    
    import (
     "go-mma/shared/common/domain"
     "time"
    )
    
    // กำหนดชื่อ Event ที่ใช้ในระบบ (EventName)
    // เพื่อระบุชนิดของ Domain Event ว่าเป็น "CustomerCreated"
    const (
     CustomerCreatedDomainEventType domain.EventName = "CustomerCreated"
    )
    
    // CustomerCreatedDomainEvent คือ struct ที่เก็บข้อมูลของเหตุการณ์
    // ที่เกิดขึ้นเมื่อมีการสร้าง Customer ใหม่ในระบบ
    type CustomerCreatedDomainEvent struct {
     domain.BaseDomainEvent // ฝัง BaseDomainEvent ที่มีชื่อและเวลาเกิด event
     CustomerID int64       // ID ของ Customer ที่ถูกสร้าง
     Email      string      // Email ของ Customer ที่ถูกสร้าง
    }
    
    // NewCustomerCreatedDomainEvent สร้าง instance ใหม่ของ CustomerCreatedDomainEvent
    // โดยรับ customer ID และ email เป็น input และตั้งชื่อ event กับเวลาปัจจุบันอัตโนมัติ
    func NewCustomerCreatedDomainEvent(custID int64, email string) *CustomerCreatedDomainEvent {
     return &CustomerCreatedDomainEvent{
      BaseDomainEvent: domain.BaseDomainEvent{
       Name: CustomerCreatedDomainEventType, // กำหนดชื่อ event
       At:   time.Now(),                      // เวลาเกิด event ณ ปัจจุบัน
      },
      CustomerID: custID, // กำหนด customer ID
      Email:      email,  // กำหนด email
     }
    }
    ```

- ปรับโมเดล `Customer` ให้เป็น `Aggregate` เพื่อเพิ่มเหตุการณ์  “CustomerCreated” ณ ตอนสร้างโมเดล

    > แก้ไขไฟล์ `customer/internal/model/customer.go`
    >

    ```go
    package model
    
    import (
     "go-mma/modules/customer/internal/domain/event"
     "go-mma/shared/common/domain"
     "go-mma/shared/common/errs"
     "go-mma/shared/common/idgen"
     "time"
    )
    
    type Customer struct {
     ID        int64     `db:"id"` // tag db ใช้สำหรับ StructScan() ของ sqlx
     Email     string    `db:"email"`
     Credit    int       `db:"credit"`
     CreatedAt time.Time `db:"created_at"`
     UpdatedAt time.Time `db:"updated_at"`
     domain.Aggregate    // embed: เพื่อให้กลายเป็น Aggregate ของ Customer
    }
    
    func NewCustomer(email string, credit int) *Customer {
     customer := &Customer{
      ID:     idgen.GenerateTimeRandomID(),
      Email:  email,
      Credit: credit,
     }
    
     // เพิ่มเหตุการณ์ "CustomerCreated"
     customer.AddDomainEvent(event.NewCustomerCreatedDomainEvent(customer.ID, customer.Email))
    
     return customer
    }
    
    func (c *Customer) ReserveCredit(v int) error {
     // ...
    }
    
    func (c *Customer) ReleaseCredit(v int) {
     // ...
    }
    ```

### Domain Event Handler

สำหรับโค้ดที่รับ event “`CustomerCreated`” มาทำงานต่อ

> สร้างไฟล์ `customer/internal/domain/eventhandler/customer_created_handler.go`
>

```go
package eventhandler

import (
 "context"
 "go-mma/modules/customer/internal/domain/event"
 notiService "go-mma/modules/notification/service"
 "go-mma/shared/common/domain"
)

// customerCreatedDomainEventHandler คือ handler สำหรับจัดการ event ประเภท CustomerCreatedDomainEvent
type customerCreatedDomainEventHandler struct {
 notiSvc notiService.NotificationService // service สำหรับส่งการแจ้งเตือน (เช่น อีเมล)
}

// NewCustomerCreatedDomainEventHandler คือฟังก์ชันสร้าง instance ของ handler นี้
func NewCustomerCreatedDomainEventHandler(notiSvc notiService.NotificationService) domain.DomainEventHandler {
 return &customerCreatedDomainEventHandler{
  notiSvc: notiSvc,
 }
}

// Handle คือฟังก์ชันหลักที่ถูกเรียกเมื่อมี event ถูก dispatch มา
func (h *customerCreatedDomainEventHandler) Handle(ctx context.Context, evt domain.DomainEvent) error {
 // แปลง (type assert) event ที่รับมาเป็น pointer ของ CustomerCreatedDomainEvent
 e, ok := evt.(*event.CustomerCreatedDomainEvent)
 if !ok {
  // ถ้าไม่ใช่ event ประเภทนี้ ให้ส่ง error กลับไป
  return domain.ErrInvalidEvent
 }

 // เรียกใช้ service ส่งอีเมลต้อนรับลูกค้าใหม่
 if err := h.notiSvc.SendEmail(e.Email, "Welcome to our service!", map[string]any{
  "message": "Thank you for joining us! We are excited to have you as a member.",
 }); err != nil {
  // หากส่งอีเมลไม่สำเร็จ ส่ง error กลับไป
  return err
 }

 // ถ้าสำเร็จทั้งหมด ให้คืน nil (ไม่มี error)
 return nil
}
```

### Trigger Point

เป็นจุดที่ดึงเอา domain events ออกมาหลังจาก transaction logic ทั้งหมดทำงานเสร็จแล้ว แต่ยังไม่ได้ commit

> แก้ไขไฟล์ `customer/internal/feature/create/handler.go`
>

```go
package create

// ...

func (h *createCustomerCommandHandler) Handle(ctx context.Context, cmd *CreateCustomerCommand) (*CreateCustomerCommandResult, error) {
 // ...
 err := h.transactor.WithinTransaction(ctx, func(ctx context.Context, registerPostCommitHook func(transactor.PostCommitHook)) error {

  // ส่งไปที่ Repository Layer เพื่อบันทึกข้อมูลลงฐานข้อมูล
  if err := h.custRepo.Create(ctx, customer); err != nil {
   // error logging
   logger.Log.Error(err.Error())
   return err
  }
  
  // เพิ่มตรงนี้ หลังจากบันทึกสำเร็จแล้ว

  // ดึง domain events จาก customer model
  events := customer.PullDomainEvents()

  return nil
 })

 // ..
}
```

### Dispatch Events

เนื่องจาก domain event handler ที่สร้างมาจะเป็นการส่งอีเมล ซึ่งมี side-effect จึงเหมาะกับแบบ post-commit dispatch หรือ รอให้ DB commit ก่อนค่อย dispatch

- แก้ไขไฟล์ `customer/internal/feature/create/handler.go`

    ```go
    package create
    
    import (
     "context"
     "go-mma/modules/customer/domainerrors"
     "go-mma/modules/customer/internal/model"
     "go-mma/modules/customer/internal/repository"
     "go-mma/shared/common/domain" // เพิ่มตรงนี้
     "go-mma/shared/common/logger"
     "go-mma/shared/common/storage/sqldb/transactor"
    )
    
    type createCustomerCommandHandler struct {
     transactor transactor.Transactor
     custRepo   repository.CustomerRepository
     dispatcher domain.DomainEventDispatcher // เพิ่มตรงนี้ มีการใช้ dispatcher
    }
    
    func NewCreateCustomerCommandHandler(
     transactor transactor.Transactor,
     custRepo repository.CustomerRepository,
     dispatcher domain.DomainEventDispatcher, // เพิ่มตรงนี้
    ) *createCustomerCommandHandler {
     return &createCustomerCommandHandler{
      transactor: transactor,
      custRepo:   custRepo,
      dispatcher: dispatcher, // เพิ่มตรงนี้
     }
    }
    
    func (h *createCustomerCommandHandler) Handle(ctx context.Context, cmd *CreateCustomerCommand) (*CreateCustomerCommandResult, error) {
     // ...
     err := h.transactor.WithinTransaction(ctx, func(ctx context.Context, registerPostCommitHook func(transactor.PostCommitHook)) error {
      // ...
      // ดึง domain events จาก customer model
      events := customer.PullDomainEvents()
    
      // ให้ dispatch หลัง commit แล้ว
      registerPostCommitHook(func(ctx context.Context) error {
       return h.dispatcher.Dispatch(ctx, events)
      })
    
      return nil
     })
    
     // ..
    }
    ```

### Register domain event

เนื่องจาก domain events เป็นการทำงานเฉพาะในโมดูลนั้นๆ เท่านั้น ดังนั้น ให้สร้าง dispatcher แยกของแต่ละโมดูลได้เลย

- แก้ไขไฟล์ `customer/module.go`

    ```go
    package customer
    
    import (
     "go-mma/modules/customer/internal/domain/event"         // เพิ่มตรงนี้่
     "go-mma/modules/customer/internal/domain/eventhandler"  // เพิ่มตรงนี้่
     "go-mma/modules/customer/internal/feature/create"
     getbyid "go-mma/modules/customer/internal/feature/get-by-id"
     releasecredit "go-mma/modules/customer/internal/feature/release-credit"
     reservecredit "go-mma/modules/customer/internal/feature/reserve-credit"
     "go-mma/modules/customer/internal/repository"
     "go-mma/shared/common/domain"                           // เพิ่มตรงนี้่
     "go-mma/shared/common/mediator"
     "go-mma/shared/common/module"
     "go-mma/shared/common/registry"
    
     notiModule "go-mma/modules/notification"
     notiService "go-mma/modules/notification/service"
    
     "github.com/gofiber/fiber/v3"
    )
    
    // ...
    
    func (m *moduleImp) Init(reg registry.ServiceRegistry) error {
     // Resolve NotificationService from the registry
     // ...
    
     // เพิ่มตรงนี้่
     // สร้าง Domain Event Dispatcher สำหรับโมดูลนี้โดยเฉพาะ
     // เราจะไม่ใช้ dispatcher กลาง แต่จะสร้าง dispatcher แยกในแต่ละโมดูลแทน
     // เพื่อให้โมดูลนั้นๆ ควบคุมการลงทะเบียนและการจัดการ event handler ได้เองอย่างอิสระ
     dispatcher := domain.NewSimpleDomainEventDispatcher()
    
     // ลงทะเบียน handler สำหรับ event CustomerCreatedDomainEventType ใน dispatcher ของโมดูลนี้
     dispatcher.Register(event.CustomerCreatedDomainEventType, eventhandler.NewCustomerCreatedDomainEventHandler(notiSvc))
    
     // สร้าง repository ของโมดูลนี้
     repo := repository.NewCustomerRepository(m.mCtx.DBCtx)
    
     // ลงทะเบียน command handler และส่ง dispatcher เข้าไปใน handler ด้วย
     // เพื่อให้ handler สามารถ dispatch event ผ่าน dispatcher ของโมดูลนี้ได้
     mediator.Register(create.NewCreateCustomerCommandHandler(m.mCtx.Transactor, repo, dispatcher))
     
     // ...
    }
    ```

