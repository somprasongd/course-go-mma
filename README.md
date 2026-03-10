# แนวทางการสร้างโปรเจกต์แบบ Modular Monolith

> หากคุณเคยเริ่มพัฒนาแอปพลิเคชันด้วยโครงสร้างแบบ Monolith แล้วพบว่าระบบเริ่มยุ่งเหยิงเมื่อทีมเติบโตขึ้น หรือเมื่อฟีเจอร์ใหม่ๆ เริ่มทับซ้อนกัน — Modular Monolith อาจเป็นคำตอบที่เหมาะสม

บทความนี้จะพาคุณไล่เรียงตั้งแต่พื้นฐานของการพัฒนาเว็บแอปในรูปแบบ Monolith ไปสู่การออกแบบระบบแบบ **Modular Monolith** อย่างเป็นระบบ พร้อมแนวคิดและแนวปฏิบัติที่นำไปใช้ได้จริง โดยไม่ต้องกระโดดไปสู่ Microservices ตั้งแต่แรก

---

## สารบัญ

### บทนำและพื้นฐาน

| บท | หัวข้อ |
|:---:|--------|
| 1 | [โปรเจกต์นี้มีเป้าหมายอะไร](docs/01-goal.md) |
| 2 | [ออกแบบ Web Server ให้ดีตั้งแต่แรก](docs/02-web-server.md) |
| 3 | [การแยก logic ออกจาก routing ด้วย Handlers](docs/03-handlers.md) |
| 4 | [เชื่อมต่อฐานข้อมูลอย่างปลอดภัยและยืดหยุ่น](docs/04-database.md) |

### Layered Architecture

| บท | หัวข้อ |
|:---:|--------|
| 5 | [เริ่มต้นด้วย Layered Architecture ที่เข้าใจง่าย](docs/05-layered-architecture.md) |
| 6 | [ออกแบบระบบจัดการ Error ให้ตรวจสอบและแก้ไขง่าย](docs/06-error-handling.md) |
| 7 | [สร้างระบบส่งอีเมลแบบ Reusable ด้วย Notification Service](docs/07-notification-service.md) |
| 8 | [สร้างระบบจัดการออเดอร์ด้วย Layered Architecture](docs/08-order-management.md) |
| 9 | [ใช้งาน Database Transaction อย่างไรให้ถูกต้อง](docs/09-transaction.md) |
| 10 | [นำหลักการ Dependency Inversion มาใช้ในระบบจริง](docs/10-dependency-inversion.md) |

### Modular Architecture

| บท | หัวข้อ |
|:---:|--------|
| 11 | [แปลงโครงสร้างไปสู่ Modular Architecture อย่างเป็นขั้นตอน](docs/11-modular-architecture.md) |
| 12 | [แยกความรับผิดชอบด้วยการซ่อนรายละเอียดของ Subdomain](docs/12-subdomain.md) |
| 13 | [จัดการ Service ใน Monolith ด้วย Service Registry](docs/13-service-registry.md) |
| 14 | [ป้องกันการเข้าถึงข้ามโมดูลด้วยโฟลเดอร์ `internal`](docs/14-internal-folder.md) |
| 15 | [รวมโค้ดทั้งหมดไว้ใน Mono-Repository อย่างเป็นระบบ](docs/15-mono-repo.md) |
| 16 | [กำหนด Public API Contract ระหว่างโมดูล](docs/16-public-api.md) |
| 17 | [การแยกข้อมูลระหว่างโมดูล (Data Isolation)](docs/17-data-isolation.md) |
| 18 | [การจัดการโมดูลด้วย Feature-Based Structure และ CQRS](docs/18-feature-cqrs.md) |

### Event-Driven Architecture

| บท | หัวข้อ |
|:---:|--------|
| 19 | [เพิ่มความยืดหยุ่นด้วยแนวคิด Event-Driven Architecture](docs/19-event-driven.md) |
| 20 | [Refactor เพิ่ม Domain Event](docs/20-domain-event.md) |
| 21 | [Refactor เพิ่ม Integration Event](docs/21-integration-event.md) |

### สรุป

| บท | หัวข้อ |
|:---:|--------|
| 22 | [สรุป](docs/22-conclusion.md) |

---

## บทเสริม

| บท | หัวข้อ |
|:---:|--------|
| 23 | [สร้าง API Document ด้วย Swagger](docs/23-swagger.md) |
| 24 | [การล็อกข้อมูล (Database Lock)](docs/24-database-lock.md) |
| 25 | [Docker Compose สำหรับ Production](docs/25-docker-compose-prod.md) |
| 26 | [เชื่อมโยง Log ด้วย Request ID](docs/26-request-id.md) |
| 27 | [การทำ Observability](docs/27-observability.md) |
| 28 | [การทำ Input Sanitization เพื่อป้องกัน XSS](docs/28-input-sanitization.md) |
