## เพิ่มความยืดหยุ่นด้วยแนวคิด Event-Driven Architecture

ในโค้ดปัจจุบัน เช่น ภายใน `CreateCustomerCommandHandler` ซึ่งมีหน้าที่สร้างลูกค้าใหม่ เรายังใช้รูปแบบการเขียนแบบ **Imperative Style** — คือการสั่งงานโดยตรง เช่น

```go
h.notiSvc.SendEmail(...)
```

```bash
+-------------------------------+
| CreateCustomerCommandHandler  |
+-------------------------------+
           |
           | creates Customer
           v
+-------------------------------+
|  CustomerRepository           |
+-------------------------------+
           |
           | persists to DB
           v
+-------------------------------+
|  NotificationService          |
+-------------------------------+
           |
           | sends welcome email
           v
       External System
```

แม้จะทำงานได้ดีในระบบขนาดเล็ก แต่แนวทางนี้มีข้อจำกัด

- ตัว Handler ผูกกับ service อื่นโดยตรง (tight coupling)
- หากต้องเพิ่ม action อื่น เช่น ส่ง SMS หรือบันทึก audit log → ต้องแก้โค้ดตรงนี้เพิ่ม
- ขัดกับหลัก Single Responsibility

### แนวทางใหม่: เปลี่ยนเป็น Event-Driven

แทนที่จะ “สั่งให้ทำงาน” โดยตรง → ให้ "ประกาศเหตุการณ์" แล้วให้ส่วนอื่นในระบบมารับฟังและจัดการ

ข้อดีของแนวทางนี้

- **ทำให้ระบบ หลวมตัว (loosely coupled):** Handler ไม่รู้ว่ามีใครฟัง event หรือใครจะทำอะไร
- **ขยายระบบง่าย:** อยากเพิ่ม logic ใหม่ แค่เพิ่ม event handler โดยไม่ต้องแก้ handler เดิม
- **พร้อมสำหรับ scaling:** สามารถแยกบาง handler ไปทำงาน async หรือใช้ message queue ได้ทันที

### Domain Events

- เป็น event ที่เกิดภายใน domain (bounded context)
- ใช้เพื่อแจ้งว่า *"สิ่งนี้เกิดขึ้นแล้ว"* (เช่น `CustomerCreated`)
- ถูก publish และ consume *ในกระบวนการเดียวกัน* (in-process)
- มักใช้กับ business logic ภายใน

### Integration Events

- ถูกใช้เพื่อสื่อสารข้าม bounded context / microservice
- ใช้ messaging system เช่น Kafka, RabbitMQ
- มักเกิดจาก domain event แล้วถูกแปลง (map) เป็น integration event
- ทำงานแบบ async

### โครงสร้างหลังเพิ่ม Domain Events กับ Integration Events

ตัวอย่าง การแยก logic การส่งอีเมลออกจาก Handler และรองรับการทำงานแบบ asynchronous ซึ่งจะมีขั้นตอนการทำงานแบบนี้

```bash
+-------------------------------+
| CreateCustomerCommandHandler  |
+-------------------------------+
           |
           | creates Customer
           v
+-------------------------+
|  Customer Entity        |
|  + AddDomainEvent()     |
+-------------------------+
           |
           | emits domain event
           v
+-------------------------------+
| DomainEventDispatcher         |
| (in-process, synchronous)     |
+-------------------------------+
           |
           | calls domain handler
           v
+------------------------------------------+
| CustomerCreatedDomainEventHandler        |
| - Converts to Integration Event          |
| - Calls EventBus.Publish()               |
+------------------------------------------+
           |
           | emits async message (Kafka/Outbox)
           v
+------------------------------+
|  NotificationService         |
| (another module/microservice)|
+------------------------------+
           |
           | sends welcome email
           v
       External System
```

1. `CreateCustomerHandler` → สร้าง Customer และเพิ่ม `CustomerCreated` domain event
2. `DomainEventDispatcher` → dispatch event นี้ให้ `CustomerCreatedDomainEventHandler`
3. Handler → สร้าง `CustomerCreatedIntegrationEvent` แล้วส่งผ่าน EventBus
4. ระบบภายนอก (เช่น Notification Module) consume แล้วจัดการเรื่อง Email

