## การทำ Input Sanitization เพื่อป้องกัน XSS

เมื่อรับข้อมูลจากผู้ใช้ผ่าน API ความเสี่ยงสำคัญอย่างหนึ่งคือ **Cross-Site Scripting (XSS)** ที่ผู้โจมตีอาจฝัง malicious script ไว้ใน input เพื่อให้ระบบ execute หรือแสดงผลโดยไม่ได้ตั้งใจ

การทำ **Input Sanitization** จึงเป็นขั้นตอนสำคัญในการป้องกัน XSS โดยการทำความสะอาด (clean) หรือ escape ข้อมูลที่อาจเป็นอันตรายก่อนนำไปใช้งานต่อ

### โครงสร้าง Sanitizer

ในโปรเจกต์นี้เราสร้าง `sanitizer` package ไว้ที่ `shared/common/sanitizer/` ซึ่งประกอบด้วย:

```
shared/common/sanitizer/
├── sanitizer.go       # ฟังก์ชันพื้นฐานสำหรับ sanitize string
└── struct.go          # ฟังก์ชันสำหรับ sanitize struct ด้วย reflection
```

### ฟังก์ชันพื้นฐาน

**`sanitizer.go`** ประกอบด้วยฟังก์ชันหลักสำหรับ sanitize string:

```go
// SanitizeInput ลบ HTML tags และ escape special characters
func SanitizeInput(input string) string {
    s := strings.TrimSpace(input)
    s = removeHTMLTags(s)      // ลบ <script>, <iframe>, ฯลฯ
    s = html.EscapeString(s)   // Escape & < > " '
    return s
}

// SanitizeEmail ทำความสะอาด email address
func SanitizeEmail(email string) string {
    s := removeHTMLTags(email)
    s = strings.TrimSpace(s)
    s = strings.ToLower(s)     // Normalize เป็น lowercase
    return s
}
```

### Sanitize Struct ด้วย struct tags

**`struct.go`** ใช้ reflection เพื่อ sanitize fields ตาม struct tags:

```go
// SanitizeStruct ใช้ sanitize tag เพื่อกำหนดวิธีการ sanitize แต่ละ field
func SanitizeStruct(s interface{})
```

รองรับ struct tags ดังนี้:

| Tag | การทำงาน |
|-----|----------|
| `sanitize:"email"` | ลบ HTML tags, trim, lowercase |
| `sanitize:"xss"` | ลบ HTML tags, escape special characters |
| `sanitize:"trim"` | ตัด whitespace หน้า-หลัง |
| `sanitize:"lower"` | แปลงเป็นตัวพิมพ์เล็ก |

### ตัวอย่างการใช้งาน

**1. กำหนด struct tags:**

```go
// shared/modules/customer/internal/feature/create/dto.go
type CreateCustomerRequest struct {
    Email  string `json:"email" sanitize:"email"`
    Name   string `json:"name" sanitize:"xss"`    // ← ป้องกัน XSS
    Credit int    `json:"credit"`
}
```

**2. เรียกใช้ใน Endpoint:**

```go
// shared/modules/customer/internal/feature/create/endpoint.go
func createCustomerHTTPHandler(c fiber.Ctx) error {
    var req CreateCustomerRequest
    if err := c.Bind().Body(&req); err != nil {
        return errs.InputValidationError(err.Error())
    }

    // Sanitize inputs ก่อน validation
    sanitizer.SanitizeStruct(&req)

    // จากนั้นค่อย validate
    if err := req.Validate(); err != nil {
        return errs.InputValidationError(err.Error())
    }
    
    // ... ดำเนินการต่อ
}
```

> **สำคัญ:** ควรเรียก `SanitizeStruct` **ก่อน** `Validate()` เพื่อให้ข้อมูลอยู่ในรูปแบบมาตรฐานก่อนตรวจสอบ

### รองรับ Nested Struct และ Complex Types

`SanitizeStruct` รองรับการทำงานแบบ recursive:

```go
type Address struct {
    Street  string `json:"street" sanitize:"xss"`
    City    string `json:"city" sanitize:"trim"`
    Country string `json:"country" sanitize:"lower"`
}

type PersonWithAddress struct {
    Name    string  `json:"name" sanitize:"xss"`
    Email   string  `json:"email" sanitize:"email"`
    Address Address `json:"address"`  // ← Nested struct จะถูก sanitize ด้วย
}
```

รองรับประเภทข้อมูลที่ซับซ้อน:

- **Nested struct** – sanitize fields ภายในโดยอัตโนมัติ
- **Pointer to struct** – dereference และ sanitize
- **Slice/Array** – sanitize ทุก element ตาม tag ที่กำหนด
- **Map** – sanitize string values (xss sanitize โดย default)
- **Pointer slice** – sanitize ทุก element แม้เป็น pointer

### ตัวอย่างการใช้งานกับ Slice และ Map

```go
type PersonWithSlice struct {
    Name     string   `json:"name" sanitize:"xss"`
    Tags     []string `json:"tags" sanitize:"trim"`    // ทุก tag จะถูก trim
    Aliases  []string `json:"aliases" sanitize:"lower"` // ทุก alias จะถูก lowercase
    Comments []string `json:"comments" sanitize:"xss"`  // ทุก comment จะถูก xss sanitize
}

type PersonWithMap struct {
    Name     string            `json:"name" sanitize:"xss"`
    Metadata map[string]string `json:"metadata"`  // values จะถูก xss sanitize
}
```

### Best Practices

1. **Sanitize ก่อน Validate** – ข้อมูลต้องอยู่ในรูปแบบมาตรฐานก่อนตรวจสอบความถูกต้อง

2. **กำหนด tag ให้ครอบคลุม** – ทุก string field ที่รับจากผู้ใช้ควรมี sanitize tag

3. **เลือก sanitize type ให้เหมาะสม**:
   - `email` – สำหรับ email addresses
   - `xss` – สำหรับข้อความทั่วไปที่อาจมี HTML
   - `trim` – สำหรับข้อความที่ไม่ต้องการ whitespace ส่วนเกิน
   - `lower` – สำหรับ case-insensitive values

4. **ไม่พึ่งพา sanitization เพียงอย่างเดียว** – ควรใช้ร่วมกับ validation และ output encoding เมื่อแสดงผล

---
