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