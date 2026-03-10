## โปรเจกต์นี้มีเป้าหมายอะไร

เป็นโปรเจกต์ Monolith App แบบง่ายๆ ที่มีฟีเจอร์ ดังนี้

1. สร้างลูกค้าใหม่ พร้อมส่งอีเมลต้อนรับ
2. สั่งออเดอร์ พร้อมส่งอีเมลยืนยันคำสั่งออเดอร์
3. ยกเลิกออเดอร์

### ภาพรวมโปรเจกต์

```markdown
+------------+        +----------------------+        +-----------+
|   Client   | <----> |    Monolith App      | <----> | Database  |
+------------+        |----------------------|        +-----------+
                      |  Modules:            |
                      |    - customer        |
                      |    - order           |
                      |    - email           |
                      +----------------------+

1. สร้างลูกค้าใหม่ (POST /customers)
---------------------------------------
Client ----> Monolith: POST /customers {email, credit}
Monolith.customer --> Database: ตรวจสอบ email ซ้ำ?
  └─ ซ้ำ --> Monolith.customer --> Client: 409 Conflict (email already exists)
  └─ ไม่ซ้ำ:
      Monolith.customer --> Database: INSERT INTO customers
      Monolith.email --> ส่งอีเมลต้อนรับ
      Monolith.customer --> Client: 201 Created

2. สั่งออเดอร์ (POST /orders)
-------------------------------
Client ----> Monolith: POST /orders {customer_id, order_total}
Monolith.order --> Database: ตรวจสอบ customer_id
  └─ ไม่พบ --> Monolith.order --> Client: 404 Not Found (customer not found)
  └─ พบ:
      Monolith.order --> Database: ตรวจสอบ credit เพียงพอ?
          └─ ไม่พอ --> Monolith.order --> Client: 422 Unprocessable Entity (insufficient credit)
          └─ พอ:
              Monolith.order --> Database: INSERT INTO orders, UPDATE credit (หักยอด)
              Monolith.email --> ส่งอีเมลยืนยันออเดอร์
              Monolith.order --> Client: 201 Created

3. ยกเลิกออเดอร์ (DELETE /orders/:orderID)
---------------------------------------------
Client ----> Monolith: DELETE /orders/:orderID
Monolith.order --> Database: ตรวจสอบ orderID
  └─ ไม่พบ --> Monolith.order --> Client: 404 Not Found (order not found)
  └─ พบ:
      Monolith.order --> Database: DELETE order, UPDATE credit (คืนยอด)
      Monolith.order --> Client: 204 No Content
```

### API endpoint

ระบบนี้มี 3 API endpoint

- `POST /customers` – สร้างลูกค้าใหม่

    | JSON Field | Type | Required | Description |
    | --- | --- | --- | --- |
    | `email` | string | ✅ | อีเมลลูกค้า |
    | `credit` | integer | ✅ | เครดิตเริ่มต้น |

    **Response**

    | Status Code | Description |
    | --- | --- |
    | `201` | สร้างสำเร็จ |
    | `400` | payload ไม่ครบ หรือข้อมูลไม่ถูกต้อง เช่น ไม่ส่ง `email` หรือ `email` ผิดรูปแบบ |
    | `422` | ไม่ผ่าน business rule เช่น `credit` ≤ 0 |
    | `409` | อีเมลนี้มีอยู่แล้วในระบบ (Conflict) |

- `POST /orders` – สร้างออเดอร์

    | JSON Field | Type | Required | Description |
    | --- | --- | --- | --- |
    | `customer_id` | integer | ✅ | ID ลูกค้า |
    | `order_total` | integer | ✅ | ยอดรวมออเดอร์ |

    **Response**

    | Status Code | Description |
    | --- | --- |
    | `201` | สร้างออเดอร์เรียบร้อย |
    | `400` | ไม่ส่ง `customer_id` หรือ `order_total` ≤ 0 |
    | `404` | ไม่พบลูกค้า (`customer_id` ไม่ถูกต้อง) |
    | `422` | เครดิตไม่เพียงพอในการสั่งออเดอร์ |

- `DELETE /orders/:orderID` – ยกเลิกออเดอร์

    | Path Param | Type | Required | Description |
    | --- | --- | --- | --- |
    | `orderID` | integer | ✅ | ID ออเดอร์ |

    **Response**

    | Status Code | Description |
    | --- | --- |
    | `204` | ลบออเดอร์สำเร็จ (No Content) |
    | `404` | ไม่พบออเดอร์นี้ในระบบ |

---

