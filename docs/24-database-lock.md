## การล็อกข้อมูล (Database Lock)

ใน PostgreSQL (และ RDBMS ส่วนใหญ่) เมื่อมีคำสั่ง `UPDATE` เกิดขึ้นกับแถวใดแถวหนึ่ง **ระบบจะทำการล็อกระดับแถว (row-level lock)** อัตโนมัติด้วยเหตุผลเพื่อป้องกันการเขียนซ้อน และกันการเกิด Deadlock

แต่ใน application logic ของเรามีการ `SELECT` ออกมาก่อน แล้วทำการแก้ไขค่า ก่อนสั่ง `UPDATE` กลับไปในฐานข้อมูล เช่น ในขั้นตอนการตัดยอดเครดิต กับคืนยอดเครดิต

```go
func (h *reserveCreditCommandHandler) Handle(ctx context.Context, cmd *customercontract.ReserveCreditCommand) (*mediator.NoResponse, error) {
 err := h.transactor.WithinTransaction(ctx, func(ctx context.Context, registerPostCommitHook func(transactor.PostCommitHook)) error {
   // ดึงข้อมูล customer จากฐานข้อมูล
  customer, err := h.custRepo.FindByID(ctx, cmd.CustomerID)
  if err != nil {
   logger.Log.Error(err.Error())
   return err
  }

  if customer == nil {
   return domainerrors.ErrCustomerNotFound
  }

    // แก้ไขค่า credit
  if err := customer.ReserveCredit(cmd.CreditAmount); err != nil {
   return err
  }

    // update ค่า credit ที่แก้ไขแล้ว
  if err := h.custRepo.UpdateCredit(ctx, customer); err != nil {
   logger.Log.Error(err.Error())
   return errs.DatabaseFailureError(err.Error())
  }

  return nil
 })

 return nil, err
}
```

ซึ่งหากระบบมี Concurrent Transactions สูง อาจทำให้ในขั้นตอนการ `SELECT` (ระบบไม่ได้ล็อกแถว) อาจได้ค่าเดียวกัน เมื่อถึงขั้นตอนการ `UPDATE` ทำให้ credit ไม่ถูกต้องได้

### วิธีแก้: ใช้ `SELECT ... FOR UPDATE`

ใน PostgreSQL มี syntax พิเศษ

```sql
SELECT * FROM FROM customer.customers WHERE id = $1 FOR UPDATE
```

- เมื่อรัน `SELECT ... FOR UPDATE` จะล็อกแถวทันที (row-level lock)
- Lock นี้กัน Transaction อื่นไม่ให้ `UPDATE` หรือ `DELETE` แถวนี้ได้ จนกว่าจะ `COMMIT` หรือ `ROLLBACK`
- ทำให้แน่ใจว่าค่า `credit` ที่อ่านมา จะเป็นฐานเดียวของการคำนวณต่อไป

PostgreSQL มี Lock Mode หลายแบบ

- `FOR UPDATE` = ล็อกเต็มที่ เหมือนบอกว่า “ฉันจะอาจแก้ค่าใดๆ ในแถวนี้” ใช้เมื่อ **จะอัปเดตค่าในแถว (รวมถึงคีย์)**
- `FOR NO KEY UPDATE` = เบากว่า `FOR UPDATE` นิดหน่อย ใช้กรณีจะอัปเดตค่าธรรมดา แต่ไม่อัปเดต PRIMARY KEY หรือ UNIQUE KEY ใช้เมื่อ **จะอัปเดตค่าในแถว แต่ไม่แตะคีย์ (เช่น credit, status)**

ดังนั้น เราจะใช้ `FOR NO KEY UPDATE` โดยเพิ่มฟังก์ชัน `FindByIDForUpdate` ใน `customer/repository/customer.go`

```go
type CustomerRepository interface {
 Create(ctx context.Context, customer *model.Customer) error
 ExistsByEmail(ctx context.Context, email string) (bool, error)
 FindByID(ctx context.Context, id int64) (*model.Customer, error)
 // <-- เพิ่มตรงนี้
 FindByIDForUpdate(ctx context.Context, id int64) (*model.Customer, error)
 UpdateCredit(ctx context.Context, customer *model.Customer) error
}

// <-- เพิ่มตรงนี้
func (r *customerRepository) FindByIDForUpdate(ctx context.Context, id int64) (*model.Customer, error) {
 query := `
 SELECT *
 FROM customer.customers
 WHERE id = $1
 FOR NO KEY UPDATE
`
 ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
 defer cancel()

 var customer model.Customer
 err := r.dbCtx(ctx).QueryRowxContext(ctx, query, id).StructScan(&customer)
 if err != nil {
  if err == sql.ErrNoRows {
   return nil, nil
  }
  return nil, errs.HandleDBError(fmt.Errorf("an error occurred while finding a customer by id: %w", err))
 }

 return &customer, nil
}
```

แล้วเปลี่ยนให้ command handler มาเรียกใช้

```go
// customer/internal/feature/reserve-credit/command_handler.go
func (h *reserveCreditCommandHandler) Handle(ctx context.Context, cmd *customercontract.ReserveCreditCommand) (*mediator.NoResponse, error) {
 err := h.transactor.WithinTransaction(ctx, func(ctx context.Context, registerPostCommitHook func(transactor.PostCommitHook)) error {
  customer, err := h.custRepo.FindByIDForUpdate(ctx, cmd.CustomerID)
  
  // ...
 })

 return nil, err
}

// customer/internal/feature/release-credit/command_handler.go
func (h *releaseCreditCommandHandler) Handle(ctx context.Context, cmd *customercontract.ReleaseCreditCommand) (*mediator.NoResponse, error) {
 err := h.transactor.WithinTransaction(ctx, func(ctx context.Context, registerPostCommitHook func(transactor.PostCommitHook)) error {
  customer, err := h.custRepo.FindByIDForUpdate(ctx, cmd.CustomerID)
  
  // ...
  })

 return nil, err
}
```

### สรุปการล็อกข้อมูล (Database Lock)

- **แค่ `UPDATE` ธรรมดา:** PostgreSQL จะล็อกตอนเขียน แต่ถ้า `SELECT` มาก่อน ค่าอาจซ้ำได้
- **`SELECT FOR UPDATE`:** ล็อกแถวตั้งแต่เริ่มอ่าน ป้องกันการอ่านค่าซ้ำในกรณี read-modify-write
- เลือก `FOR UPDATE` หรือ `FOR NO KEY UPDATE` ตามว่าอัปเดต Primary Key หรือไม่

---

