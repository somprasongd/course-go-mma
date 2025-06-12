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