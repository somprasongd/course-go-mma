## การแยกข้อมูลระหว่างโมดูล (Data Isolation)

หลังจากที่เราการกำหนดขอบเขตโมดูลและรูปแบบการสื่อสารที่ชัดเจนแล้ว สิ่งหนึ่งที่มองข้ามไม่ได้คือ **การแยกข้อมูลระหว่างโมดูล ([Data Isolation](https://somprasongd.work/blog/architecture/modular-monolith-data-isolation))** ซึ่งช่วยให้โมดูลมีความเป็นอิสระและมีการผูกแน่นที่หลวม (loosely coupled) สถาปัตยกรรมแบบ Modular Monolith มีกฎที่เข้มงวดสำหรับการรักษาความสมบูรณ์ของข้อมูล:

- แต่ละโมดูลสามารถเข้าถึงได้เฉพาะตารางของตนเองเท่านั้น
- ไม่มีการแชร์ตารางหรืออ็อบเจกต์ระหว่างโมดูล
- การ Join สามารถทำได้เฉพาะตารางภายในโมดูลเดียวกันเท่านั้น

ประโยชน์ของการออกแบบนี้คือการส่งเสริมความเป็นโมดูลาร์และการผูกแน่นที่หลวม ทำให้ง่ายต่อการเปลี่ยนแปลงระบบและลดผลข้างเคียงที่ไม่พึงประสงค์

### ระดับการแยกข้อมูล (Data Isolation Levels)

นี่คือสี่แนวทางในการแยกข้อมูลสำหรับ Modular Monoliths

- **Level 1 - Separate Table (ไม่แนะนำ)**

    เป็นวิธีการที่ง่ายที่สุดคือไม่มีการแยกข้อมูลในระดับฐานข้อมูล ทุกตารางของทุกโมดูลจะอยู่ในฐานข้อมูลเดียวกัน ทำให้ยากที่จะระบุว่าตารางใดเป็นของโมดูลใด

    ```bash
    ┌─────────────────────────────────┐
    │         Single Database         │
    │ ┌─────────────────────────────┐ │
    │ │   Module A - Schema public  │ │
    │ │  Table A1  Table A2  ...    │ │
    │ └─────────────────────────────┘ │
    │ ┌─────────────────────────────┐ │
    │ │   Module B - Schema public  │ │
    │ │  Table B1  Table B2  ...    │ │
    │ └─────────────────────────────┘ │
    │ ┌─────────────────────────────┐ │
    │ │   Module B - Schema public  │ │
    │ │  Table C1  Table C2  ...    │ │
    │ └─────────────────────────────┘ │
    └─────────────────────────────────┘
    ```

- **Level 2 - Separate Schema (Logical Isolation)**

    วิธีนี้เป็นการจัดกลุ่มตารางที่เกี่ยวข้องเข้าด้วยกันในฐานข้อมูลโดยใช้ **Database Schemas** แต่ละโมดูลมี Schema เฉพาะของตนเองที่มีตารางของโมดูลนั้นๆ ทำให้ง่ายต่อการจำแนกว่าตารางใดเป็นของโมดูลใด

    ```bash
    ┌─────────────────────────────────┐
    │         Single Database         │
    │ ┌─────────────────────────────┐ │
    │ │     Module A - Schema A     │ │
    │ │  Table A1  Table A2  ...    │ │
    │ └─────────────────────────────┘ │
    │ ┌─────────────────────────────┐ │
    │ │     Module B - Schema B     │ │
    │ │  Table B1  Table B2  ...    │ │
    │ └─────────────────────────────┘ │
    │ ┌─────────────────────────────┐ │
    │ │     Module C - Schema C     │ │
    │ │  Table C1  Table C2  ...    │ │
    │ └─────────────────────────────┘ │
    └─────────────────────────────────┘
    ```

- **Level 3 - Separate Database (Physical Isolation)**

    ระดับถัดไปคือการย้ายข้อมูลของแต่ละโมดูลไปยัง **ฐานข้อมูลที่แยกจากกัน** วิธีนี้มีการจำกัดมากกว่าการแยกข้อมูลด้วย Schema และเป็นทางเลือกที่ดีหากคุณต้องการกฎการแยกข้อมูลที่เข้มงวดระหว่างโมดูล

    ```bash
    ┌─────────────────┐   ┌─────────────────┐   ┌─────────────────┐
    │  Database A     │   │  Database B     │   │  Database C     │
    │ ┌─────────────┐ │   │ ┌─────────────┐ │   │ ┌─────────────┐ │
    │ │ Module A    │ │   │ │ Module B    │ │   │ │ Module C    │ │
    │ │ Table A1    │ │   │ │ Table B1    │ │   │ │ Table C1    │ │
    │ │ Table A2    │ │   │ │ Table B2    │ │   │ │ Table C2    │ │
    │ │     ...     │ │   │ │     ...     │ │   │ │     ...     │ │
    │ └─────────────┘ │   │ └─────────────┘ │   │ └─────────────┘ │
    └─────────────────┘   └─────────────────┘   └─────────────────┘
    ```

- **Level 4 - Different Persistence (Polyglot Persistence)**

    วิธีนี้ไปไกลกว่านั้นโดยการใช้ **ฐานข้อมูลประเภทที่แตกต่างกัน** สำหรับแต่ละโมดูล เพื่อแก้ปัญหาเฉพาะทางของแต่ละเรื่อง

    ```bash
    ┌─────────────────┐   ┌─────────────────┐   ┌─────────────────┐
    │  Database A     │   │  Database B     │   │  Database C     │
    │ (Relational DB) │   │ (Document DB)   │   │ (Graph DB)      │
    │ ┌─────────────┐ │   │ ┌─────────────┐ │   │ ┌─────────────┐ │
    │ │ Module A    │ │   │ │ Module B    │ │   │ │ Module C    │ │
    │ │ Table A1    │ │   │ │ Doc B1      │ │   │ │ Node C1     │ │
    │ │ Table A2    │ │   │ │ Doc B2      │ │   │ │ Edge C1     │ │
    │ │     ...     │ │   │ │     ...     │ │   │ │     ...     │ │
    │ └─────────────┘ │   │ └─────────────┘ │   │ └─────────────┘ │
    └─────────────────┘   └─────────────────┘   └─────────────────┘
    ```

### **Level 2 - Separate Schema (Logical Isolation)**

โดยทั่วไปแล้ว การเริ่มต้นด้วยการแยกข้อมูลแบบ Logical Isolation โดยใช้ Schemas เป็นวิธีที่ง่ายต่อการนำไปใช้และช่วยให้เข้าใจขอบเขตของระบบได้ดีขึ้น และคุณสามารถพิจารณาใช้ฐานข้อมูลแยกกัน (Separate Database) ในภายหลังได้ตามความต้องการของระบบที่เปลี่ยนแปลงไป

**ขั้นตอนการแยก schema**

- สร้างไฟล์ migration สำหรับการแยก schema

    ```bash
    make mgc filename=separate_schema
    ```

- เพิ่มคำสั่งสำหรับแยก schema

    > แก้ไขไฟล์ `xxx_separate_schema.up.sql`
    >

    ```sql
    -- สร้าง schema ใหม่สำหรับโมดูลลูกค้า
    CREATE SCHEMA customer;
    
    -- ย้ายตาราง 'customers' จาก schema 'public' ไปยัง schema 'customer'
    ALTER TABLE public.customers SET SCHEMA customer;
    
    -- สร้าง schema ใหม่สำหรับโมดูลคำสั่งซื้อ
    CREATE SCHEMA sales;
    
    -- ลบ Foreign Key Constraint เดิมก่อนย้ายตาราง
    ALTER TABLE public.orders DROP CONSTRAINT IF EXISTS fk_customer;
    
    -- ย้ายตาราง 'orders' จาก schema 'public' ไปยัง schema 'order'
    ALTER TABLE public.orders SET SCHEMA sales;
    
    -- หมายเหตุ: ตามหลัก Modular Monolith จะไม่มีการสร้าง Foreign Key ข้ามโมดูล
    -- การตรวจสอบความถูกต้องของ customer_id จะจัดการที่ระดับแอปพลิเคชัน
    ```

- เพิ่มคำสั่งสำหรับการย้อนกลับ

    > แก้ไขไฟล์ `xxx_separate_schema.down.sql`
    >

    ```sql
    -- ย้ายตาราง 'customers' กลับจาก schema 'customer' ไปยัง 'public'
    ALTER TABLE customer.customers SET SCHEMA public;
    
    -- ลบ schema 'customer' (จะสำเร็จเมื่อ schema ว่างเปล่า)
    DROP SCHEMA customer;
    
    -- ลบ Foreign Key Constraint บน 'sales.orders' ถ้ามีอยู่ (ป้องกันข้อผิดพลาด)
    ALTER TABLE sales.orders DROP CONSTRAINT IF EXISTS fk_customer;
    
    -- ย้ายตาราง 'orders' กลับจาก schema 'sales' ไปยัง 'public'
    ALTER TABLE sales.orders SET SCHEMA public;
    
    -- ลบ schema 'sales' (จะสำเร็จเมื่อ schema ว่างเปล่า)
    DROP SCHEMA sales;
    
    -- เพิ่ม Foreign Key Constraint กลับคืนที่ 'public.sales' โดยอ้างอิง 'public.customers'
    ALTER TABLE public.orders
    ADD CONSTRAINT fk_customer FOREIGN KEY (customer_id) REFERENCES public.customers(id);
    ```

- รันคำสั่ง `make mgu` เพิ่มสั่งสร้าง schema และย้ายตาราง
- ปรับปรุง `CustomerRepository` ให้ใช้ schema ที่ถูกต้อง `public` → `customer`

    > แก้ไขไฟล์ `customer/internal/repository/customer.go`
    >

    ```go
    package repository
    
    // ...
    
    func (r *customerRepository) Create(ctx context.Context, customer *model.Customer) error {
     query := `
     INSERT INTO customer.customers (id, email, credit)
     VALUES ($1, $2, $3)
     RETURNING *
     `
    
     // ...
     return nil
    }
    
    func (r *customerRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
     query := `SELECT 1 FROM customer.customers WHERE email = $1 LIMIT 1`
    
     // ...
     return true, nil
    }
    
    func (r *customerRepository) FindByID(ctx context.Context, id int64) (*model.Customer, error) {
     query := `
     SELECT *
     FROM customer.customers
     WHERE id = $1
    `
     // ...
     return &customer, nil
    }
    
    func (r *customerRepository) UpdateCredit(ctx context.Context, m *model.Customer) error {
     query := `
     UPDATE customer.customers
     SET credit = $2
     WHERE id = $1
     RETURNING *
    `
      // ...
     return nil
    }
    
    ```

- ปรับปรุง `OrderRepository` ให้ใช้ schema ที่ถูกต้อง `public` → `sales`

    > แก้ไขไฟล์ `order/internal/repository/order.go`
    >

    ```go
    package repository
    
    // ...
    
    func (r *orderRepository) FindByID(ctx context.Context, id int64) (*model.Order, error) {
     query := `
     SELECT *
     FROM sales.orders
     WHERE id = $1
     AND canceled_at IS NULL -- รายออเดอร์ต้องยังไม่ถูกยกเลิก
    `
     // ...
     return &order, nil
    }
    
    func (r *orderRepository) Cancel(ctx context.Context, id int64) error {
     query := `
     UPDATE sales.orders
     SET canceled_at = current_timestamp -- soft delete record
     WHERE id = $1
    `
     // ...
     return nil
    }
    
    ```

