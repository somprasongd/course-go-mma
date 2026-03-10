## ป้องกันการเข้าถึงข้ามโมดูลด้วยโฟลเดอร์ `internal`

หลังจากที่เรา แยกขอบเขตของแต่ละ sub-domain (Encapsulation) แล้ว ปัญหาที่ยังเหลือคือ โค้ดในโมดูล order ยังสามารถ `import` `model` หรือ `repository` ของโมดูล customer ได้โดยตรง นั่นทำให้ละเมิดขอบเขต (boundary) ของโดเมนและสร้างความพึ่งพา (coupling) ที่ไม่พึงประสงค์

ในภาษา Go สามารถย้ายไฟล์ที่ “ห้ามภายนอกใช้” เข้าไปไว้ภายใต้โฟลเดอร์ **`internal`** ได้ ตัวคอมไพเลอร์จะบังคับไม่ให้ path นอกโฟลเดอร์แม่ (root) ของ `internal` ทำ `import` ได้เลย

```go
customer/
├── internal/
│   ├── model/        // โครงสร้างข้อมูลเฉพาะ customer
│   └── repository/   // DB logic ของ customer
└── service/          // business logic (export)
```

ถ้าโมดูลอื่น เช่น `order` พยายาม `import "go-mma/modules/customer/internal/repository"` จะขึ้นข้อความ error แบบนี้

```go
could not import go-mma/modules/customer/internal/repository (invalid use of internal package "go-mma/modules/customer/internal/repository")
```

