## ใช้งาน Database Transaction อย่างไรให้ถูกต้อง

ในตัวอย่างโค้ดล่าสุด การสร้างออเดอร์ใหม่มีการเรียกใช้งาน repository หลายครั้ง เช่น

1. หักเครดิตจากลูกค้า
2. บันทึกคำสั่งซื้อ (order)

หากคำสั่งแรกสำเร็จ แต่คำสั่งที่สองล้มเหลว ข้อมูลในระบบจะไม่สมบูรณ์และอาจก่อให้เกิดปัญหาตามมา เช่น เครดิตถูกหักไปแล้วแต่ไม่มีคำสั่งซื้อเกิดขึ้น

```go
+------------------+        +-------------------+        +------------------+
| OrderService     |        | CustomerRepo      |        | OrderRepo        |
|------------------|        |-------------------|        |------------------|
| CreateOrder()    |        | FindCustomerByID()|        | SaveOrder()      |
|                  |        | ReserveCredit()   |        |                  |
+--------+---------+        +--------+----------+        +--------+---------+
         |                           |                            |
         | Find Customer By ID       |                            |
         +-------------------------> |                            |
         |                           |                            |
         |     Reserve Credit        |                            |
         +-------------------------> |                            |
         |                           |                            |
         |   Save Order              |                            |
         +------------------------------------------------------->|
         |                           |                            |
         |   ❌ FAIL (DB Error)      |                            |
         |<-------------------------------------------------------+
         |                           |                            |
         |                           |                            |
         |                           |                            |
         |                           |      ⚠️ CREDIT ALREADY DEDUCTED
         |                           |      ❌ ORDER NOT CREATED
         |                           |
         |  ❌ Inconsistent state!
```

### Database Transaction คืออะไร?

**Database Transaction** คือกลุ่มของคำสั่ง (เช่น `INSERT`, `UPDATE`, `DELETE`) ที่ทำงานกับฐานข้อมูล ซึ่งจะถูกมองว่าเป็น **"หน่วยการทำงานเดียวที่แบ่งแยกไม่ได้"**

หัวใจสำคัญของมันคือคุณสมบัติ **ACID** ที่รับประกันความน่าเชื่อถือของการเปลี่ยนแปลงข้อมูล:

- **Atomicity (ทำงานพร้อมกันทั้งหมด หรือไม่เลย):** การดำเนินการทั้งหมดใน Transaction จะต้องสำเร็จทั้งหมด (เรียกว่า **Commit**) หรือถ้ามีข้อผิดพลาดแม้แต่อันเดียว ทุกอย่างที่ทำไปก่อนหน้าจะถูกยกเลิกทั้งหมด (เรียกว่า **Rollback**) กลับสู่สถานะเริ่มต้น
- **Consistency (ความสอดคล้อง):** Transaction จะต้องทำให้ข้อมูลเปลี่ยนจากสถานะที่ถูกต้องหนึ่ง ไปยังอีกสถานะที่ถูกต้องหนึ่งเสมอ จะไม่มีการทิ้งข้อมูลให้อยู่ในสถานะครึ่งๆ กลางๆ
- **Isolation (การแยกตัว):** Transaction ที่ทำงานพร้อมกันหลายๆ อัน จะต้องไม่กวนกัน ผลลัพธ์จะต้องเหมือนกับว่า Transaction เหล่านั้นทำงานเรียงกันทีละอัน
- **Durability (ความคงทน):** เมื่อ Transaction ถูก Commit แล้ว ข้อมูลนั้นจะถูกบันทึกอย่างถาวรและจะไม่สูญหายไป แม้ว่าจะเกิดไฟดับหรือระบบล่มก็ตาม

**สรุป:** Database Transaction เป็นกลไกใน **ระดับฐานข้อมูล** ที่รับประกันว่าการเขียน/ลบ/แก้ไขข้อมูลจะเสร็จสมบูรณ์หรือล้มเหลวไปพร้อมกันทั้งหมด

### Unit of Work คืออะไร?

**Unit of Work (UoW)** เป็น **Design Pattern** ที่ใช้ใน **Service Layer** ใช้จัดกลุ่มของหลายๆ operation (insert, update, delete) ให้อยู่ใน **transaction เดียวกัน** โดยมุ่งเน้นที่ **ความถูกต้องของข้อมูล (Consistency)** และ **การ rollback อัตโนมัติเมื่อเกิดข้อผิดพลาด**

**องค์ประกอบของ Unit of Work**

1. Start / Begin: เริ่มต้น transaction
2. Register Changes: เก็บรายการ operation ที่จะทำ (insert, update, delete)
3. Commit: ถ้าทุกอย่างผ่าน → commit DB
4. Rollback: ถ้า error เกิดขึ้น → ยกเลิกทั้งหมด (rollback)
5. Post-Commit Hook: รัน side effects (เช่น send email) **หลังจาก** commit สำเร็จ

### สร้าง Transactor(Unit of Work) สำหรับจัดการ Database Transaction

เพื่อให้การทำงานกับหลาย repository ในหนึ่ง business flow (เช่น การหักเครดิตลูกค้าและสร้างออเดอร์) มีความ atomic มากขึ้น เราสามารถใช้ **Transactor** เพื่อควบคุม transaction และทำ rollback อัตโนมัติเมื่อเกิด error

<aside>
💡

โค้ดส่วนนี้จะถูกดัดแปลงมาจาก <https://github.com/Thiht/transactor>

</aside>

- สร้าง Interface กลาง `DBTX` เพื่อให้ repository ทั้งหมดสามารถทำงานได้กับทั้ง `*sqlx.DB` และ `*sqlx.Tx`

    > สร้างไฟล์ `util/storage/sqldb/transactor/types.go`
    >

    ```go
    package transactor
    
    import (
     "context"
     "database/sql"
    
     "github.com/jmoiron/sqlx"
    )
    
    // DBTX คือ interface กลางระหว่าง *sqlx.DB และ *sqlx.Tx
    // รวม method ของทั้ง database/sql และ sqlx เพื่อให้สามารถใช้ interchangeably ได้
    type DBTX interface {
     // database/sql methods
     ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
     PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
     QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
     QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
    
     Exec(query string, args ...any) (sql.Result, error)
     Prepare(query string) (*sql.Stmt, error)
     Query(query string, args ...any) (*sql.Rows, error)
     QueryRow(query string, args ...any) *sql.Row
    
     // sqlx methods
     GetContext(ctx context.Context, dest any, query string, args ...any) error
     MustExecContext(ctx context.Context, query string, args ...any) sql.Result
     NamedExecContext(ctx context.Context, query string, arg any) (sql.Result, error)
     PrepareNamedContext(ctx context.Context, query string) (*sqlx.NamedStmt, error)
     PreparexContext(ctx context.Context, query string) (*sqlx.Stmt, error)
     QueryRowxContext(ctx context.Context, query string, args ...any) *sqlx.Row
     QueryxContext(ctx context.Context, query string, args ...any) (*sqlx.Rows, error)
     SelectContext(ctx context.Context, dest any, query string, args ...any) error
    
     Get(dest any, query string, args ...any) error
     MustExec(query string, args ...any) sql.Result
     NamedExec(query string, arg any) (sql.Result, error)
     NamedQuery(query string, arg any) (*sqlx.Rows, error)
     PrepareNamed(query string) (*sqlx.NamedStmt, error)
     Preparex(query string) (*sqlx.Stmt, error)
     QueryRowx(query string, args ...any) *sqlx.Row
     Queryx(query string, args ...any) (*sqlx.Rows, error)
     Select(dest any, query string, args ...any) error
    
     Rebind(query string) string
     BindNamed(query string, arg any) (string, []any, error)
     DriverName() string
    }
    
    // ขยาย interface DBTX โดยเพิ่ม method สำหรับเริ่ม transaction
    type sqlxDB interface {
     DBTX
     BeginTxx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error)
    }
    
    // interface สำหรับจัดการ transaction lifecycle
    type sqlxTx interface {
     Commit() error
     Rollback() error
    }
    
    // compile-time check เพื่อให้แน่ใจว่า type เหล่านี้ implement interface ที่กำหนด
    var (
     _ DBTX   = &sqlx.DB{}
     _ DBTX   = &sqlx.Tx{}
     _ sqlxDB = &sqlx.DB{}
     _ sqlxTx = &sqlx.Tx{}
    )
    
    type (
     // ใช้เป็น key สำหรับ context value ของ transactor
     transactorKey struct{}
    
     // DBTXContext ใช้ดึง DBTX ปัจจุบันจาก context (อาจเป็น transaction หรือ db ปกติ)
     DBTXContext func(context.Context) DBTX
    )
    
    // ฝัง sqlxDB (เช่น tx) ลงใน context เพื่อใช้ใน logic ต่อไป
    func txToContext(ctx context.Context, tx sqlxDB) context.Context {
     return context.WithValue(ctx, transactorKey{}, tx)
    }
    
    // ดึง sqlxDB ออกมาจาก context ถ้ามี
    func txFromContext(ctx context.Context) sqlxDB {
     if tx, ok := ctx.Value(transactorKey{}).(sqlxDB); ok {
      return tx
     }
     return nil
    }
    ```

    `DBTX` จะถูก inject ผ่าน context เพื่อให้ repository ไม่ต้องรู้ว่าใช้ DB หรือ Tx อยู่

- เพิ่มการรองรับ Nested Transactions
เราสามารถเลือกกลยุทธ์การจัดการ nested transactions ได้ 2 แบบ
  - **ไม่รองรับ** (เหมาะกับระบบที่ควบคุม transaction flow เอง)

        > สร้างไฟล์ `util/storage/sqldb/transactor/nested_transactions_none.go`
        >

        ```go
        package transactor
        
        import (
         "context"
         "database/sql"
         "errors"
        
         "github.com/jmoiron/sqlx"
        )
        
        // NestedTransactionsNone ป้องกันไม่ให้เกิด nested transaction
        // เหมาะสำหรับระบบที่ไม่รองรับหรือไม่ต้องการ nested tx หรือเพื่อป้องกัน logic ซ้อนผิดพลาด
        func NestedTransactionsNone(db sqlxDB, tx *sqlx.Tx) (sqlxDB, sqlxTx) {
         switch typedDB := db.(type) {
         case *sqlx.DB:
          // ถ้าเป็น root DB -> wrap tx ด้วย nestedTransactionNone เพื่อป้องกัน nested
          return &nestedTransactionNone{tx}, tx
        
         case *nestedTransactionNone:
          // ถ้า tx ถูก wrap แล้ว -> คืนอันเดิมไป (กัน nested)
          return typedDB, typedDB
        
         default:
          panic("unsupported type") // ไม่รองรับ type อื่น
         }
        }
        
        // nestedTransactionNone เป็น struct ที่ wrap *sqlx.Tx
        // และ override method ที่เกี่ยวกับการเริ่ม/commit/rollback transaction
        // ให้ return error เสมอ เพื่อ block การทำ nested transaction
        type nestedTransactionNone struct {
         *sqlx.Tx
        }
        
        // Override BeginTxx เพื่อ block nested tx
        func (t *nestedTransactionNone) BeginTxx(_ context.Context, _ *sql.TxOptions) (*sqlx.Tx, error) {
         return nil, errors.New("nested transactions are not supported")
        }
        
        // Override Commit เพื่อป้องกันการ commit nested tx
        func (t *nestedTransactionNone) Commit() error {
         return errors.New("nested transactions are not supported")
        }
        
        // Override Rollback เพื่อป้องกันการ rollback nested tx
        func (t *nestedTransactionNone) Rollback() error {
         return errors.New("nested transactions are not supported")
        }
        ```

  - **ใช้ Savepoints** (เหมาะกับระบบที่อาจซ้อน transaction ได้)

        > สร้างไฟล์ `util/storage/sqldb/transactor/nested_transactions_savepoint.go`
        >

        ```go
        package transactor
        
        import (
         "context"
         "database/sql"
         "fmt"
         "strconv"
         "sync/atomic"
        
         "github.com/jmoiron/sqlx"
        )
        
        // NestedTransactionsSavepoints ใช้ savepoint เพื่อจำลอง nested transaction
        // รองรับ DB ที่สนับสนุน savepoint เช่น PostgreSQL, MySQL, MariaDB, SQLite
        func NestedTransactionsSavepoints(db sqlxDB, tx *sqlx.Tx) (sqlxDB, sqlxTx) {
         switch typedDB := db.(type) {
         case *sqlx.DB:
          // เริ่ม nested transaction จาก root db
          return &nestedTransactionSavepoints{Tx: tx}, tx
        
         case *nestedTransactionSavepoints:
          // ซ้อน nested อีกชั้น (เพิ่ม depth)
          nestedTransaction := &nestedTransactionSavepoints{
           Tx:    tx,
           depth: typedDB.depth + 1,
          }
          return nestedTransaction, nestedTransaction
        
         default:
          panic("unsupported type") // ไม่รองรับ type อื่น
         }
        }
        
        // struct สำหรับจัดการ nested transaction ด้วย savepoint
        type nestedTransactionSavepoints struct {
         *sqlx.Tx
         depth int64       // ลำดับของ nested level (ใช้ในการตั้งชื่อ savepoint)
         done  atomic.Bool // เพื่อป้องกันไม่ให้ Commit() หรือ Rollback() ถูกเรียกซ้ำ
        }
        
        // BeginTxx สร้าง savepoint ใหม่ตามลำดับ depth
        func (t *nestedTransactionSavepoints) BeginTxx(ctx context.Context, _ *sql.TxOptions) (*sqlx.Tx, error) {
         if _, err := t.ExecContext(ctx, "SAVEPOINT sp_"+strconv.FormatInt(t.depth+1, 10)); err != nil {
          return nil, fmt.Errorf("failed to create savepoint: %w", err)
         }
         return t.Tx, nil
        }
        
        // Commit จะ release savepoint ที่เกี่ยวข้องกับ level นี้
        // ใช้ CompareAndSwap เพื่อกันการ commit ซ้ำ
        func (t *nestedTransactionSavepoints) Commit() error {
         if !t.done.CompareAndSwap(false, true) {
          return sql.ErrTxDone // ป้องกันการ commit ซ้ำ
         }
         if _, err := t.Exec("RELEASE SAVEPOINT sp_" + strconv.FormatInt(t.depth, 10)); err != nil {
          return fmt.Errorf("failed to release savepoint: %w", err)
         }
         return nil
        }
        
        // Rollback จะ rollback ไปยัง savepoint ของ level นี้
        // และ mark ว่า transaction นี้จบแล้ว
        func (t *nestedTransactionSavepoints) Rollback() error {
         if !t.done.CompareAndSwap(false, true) {
          return sql.ErrTxDone // ป้องกัน rollback ซ้ำ
         }
         if _, err := t.Exec("ROLLBACK TO SAVEPOINT sp_" + strconv.FormatInt(t.depth, 10)); err != nil {
          return fmt.Errorf("failed to rollback to savepoint: %w", err)
         }
         return nil
        }
        ```

- สร้าง Transactor ซึ่งตัว `Transactor` จะทำหน้าที่เริ่ม transaction, inject context, และ commit/rollback โดยอัตโนมัติ

    > สร้างไฟล์ `util/storage/sqldb/transactor/transactor.go`
    >

    ```go
    // Ref: https://github.com/Thiht/transactor/blob/main/sqlx/transactor.go
    package transactor
    
    import (
     "context"
     "fmt"
     "go-mma/util/logger"
    
     "github.com/jmoiron/sqlx"
    )
    
    // PostCommitHook คือฟังก์ชันที่สามารถลงทะเบียนเพื่อให้ทำงานหลังจาก Commit เสร็จแล้ว
    type PostCommitHook func(ctx context.Context) error
    
    // Transactor คือ interface สำหรับการจัดการ Transaction
    type Transactor interface {
     // WithinTransaction ใช้สำหรับรันโค้ดใน Transaction
     // และสามารถลงทะเบียน PostCommitHook เพื่อรันหลังจาก Commit ได้
     WithinTransaction(
      ctx context.Context,
      txFunc func(ctxWithTx context.Context, registerPostCommitHook func(PostCommitHook)) error,
     ) error
    }
    
    // กำหนด alias สำหรับประเภทที่ใช้เพื่อ inject dependencies
    type (
     sqlxDBGetter               func(context.Context) sqlxDB // ดึง *sqlx.DB หรือ *sqlx.Tx จาก context
     nestedTransactionsStrategy func(sqlxDB, *sqlx.Tx) (sqlxDB, sqlxTx) // กลยุทธ์สำหรับจัดการ nested transaction
    )
    
    // sqlxTransactor คือ implementation ของ Transactor
    type sqlxTransactor struct {
     sqlxDBGetter               // วิธีดึง DB/Tx จาก context
     nestedTransactionsStrategy // กลยุทธ์ในการจัดการ nested transaction
    }
    
    // Option ใช้สำหรับกำหนดค่าเพิ่มเติมให้ sqlxTransactor เช่น กลยุทธ์ nested transaction
    type Option func(*sqlxTransactor)
    
    // New สร้าง instance ของ Transactor และ DBTXContext
    func New(db *sqlx.DB, opts ...Option) (Transactor, DBTXContext) {
     t := &sqlxTransactor{
      // default: ถ้า context มี tx ให้ใช้ tx, ไม่งั้นใช้ db
      sqlxDBGetter: func(ctx context.Context) sqlxDB {
       if tx := txFromContext(ctx); tx != nil {
        return tx
       }
       return db
      },
      nestedTransactionsStrategy: NestedTransactionsNone, // default strategy: ไม่รองรับ nested transaction
     }
    
     // apply ตัวเลือกเพิ่มเติม (ถ้ามี)
     for _, opt := range opts {
      opt(t)
     }
    
     // สร้างฟังก์ชันสำหรับดึง DBTX จาก context
     dbGetter := func(ctx context.Context) DBTX {
      if tx := txFromContext(ctx); tx != nil {
       return tx
      }
      return db
     }
    
     return t, dbGetter
    }
    
    // WithNestedTransactionStrategy ใช้เปลี่ยนกลยุทธ์ nested transaction (เช่น Savepoints)
    func WithNestedTransactionStrategy(strategy nestedTransactionsStrategy) Option {
     return func(t *sqlxTransactor) {
      t.nestedTransactionsStrategy = strategy
     }
    }
    
    // WithinTransaction ใช้สำหรับรันฟังก์ชันหนึ่งใน transaction context
    func (t *sqlxTransactor) WithinTransaction(
     ctx context.Context,
     txFunc func(ctxWithTx context.Context, registerPostCommitHook func(PostCommitHook)) error,
    ) error {
     // ดึง database หรือ transaction object จาก context
     currentDB := t.sqlxDBGetter(ctx)
    
     // เริ่มต้น transaction ใหม่
     tx, err := currentDB.BeginTxx(ctx, nil)
     if err != nil {
      return fmt.Errorf("failed to begin transaction: %w", err)
     }
    
     var hooks []PostCommitHook // เก็บ hook ที่จะเรียกหลังจาก commit
    
     // ฟังก์ชันสำหรับให้ผู้ใช้ลงทะเบียน post-commit hook
     registerPostCommitHook := func(hook PostCommitHook) {
      hooks = append(hooks, hook)
     }
    
     // สร้าง nested transaction context ใหม่ (ใช้กลยุทธ์ที่กำหนดไว้)
     newDB, currentTX := t.nestedTransactionsStrategy(currentDB, tx)
    
     // เผื่อกรณี panic หรือ error แล้วไม่ได้ commit — จะ rollback ให้อัตโนมัติ
     defer func() {
      _ = currentTX.Rollback()
     }()
    
     // inject transaction object เข้า context ใหม่
     ctxWithTx := txToContext(ctx, newDB)
    
     // เรียกใช้ฟังก์ชันที่รับ context + hook registration
     if err := txFunc(ctxWithTx, registerPostCommitHook); err != nil {
      return err // ถ้า error, transaction จะถูก rollback ใน defer
     }
    
     // พยายาม commit
     if err := currentTX.Commit(); err != nil {
      return fmt.Errorf("failed to commit transaction: %w", err)
     }
    
     // หลังจาก commit เสร็จ ให้ run post-commit hooks แบบ async
     go func() {
      for _, hook := range hooks {
       func(h PostCommitHook) {
        defer func() {
         // ดัก panic เพื่อไม่ให้ crash และ log ไว้
         if r := recover(); r != nil {
          logger.Log.Error(fmt.Sprintf("post-commit hook panic: %v", r))
         }
        }()
        // ถ้า error จาก hook ก็ log ไว้
        if err := h(ctx); err != nil {
         logger.Log.Error(fmt.Sprintf("post-commit hook error: %v", err))
        }
       }(hook)
      }
     }()
    
     return nil
    }
    
    // IsWithinTransaction ใช้ตรวจสอบว่า context นี้อยู่ใน transaction หรือไม่
    func IsWithinTransaction(ctx context.Context) bool {
     return ctx.Value(transactorKey{}) != nil
    }
    
    ```

    **สรุป**

  - ใช้ `Transactor` เพื่อควบคุมหลาย DB operation ให้เป็น atomic unit
  - Inject `DBTX` ผ่าน context ทำให้ repository ไม่ต้องรู้ว่าอยู่ใน transaction หรือไม่
  - รองรับ nested transactions ด้วย savepoint หากจำเป็น

### ปรับปรุง Repository Layer

ปรับปรุงโค้ดใน layer ของ repository เพื่อรองรับการใช้งาน transaction ร่วมกันข้ามหลาย repository โดยเปลี่ยนจากการ inject ตัว database connection (`*sql.DB` หรือ `sqldb.DBContext`) มาเป็นการใช้ interface ใหม่ที่ชื่อว่า `transactor.DBContext`

**สิ่งที่เปลี่ยนแปลง**

- เดิมใช้ `dbCtx sqldb.DBContext`
- เปลี่ยนเป็น `dbCtx transactor.DBContext`
- เวลาจะเรียกใช้ `db.QueryContext(...)`, `db.ExecContext(...)` หรืออื่น ๆ ให้ดึง `sql.DB` หรือ `sql.Tx` จาก `context.Context` โดยใช้ `r.dbCtx(ctx)` แทน

**ปรับปรุงโค้ด**

- แก้ไขไฟล์ `repository/customer.go`

    ```go
    package repository
    
    import (
     "context"
     "database/sql"
     "fmt"
     "go-mma/model"
     "go-mma/util/errs"
     "go-mma/util/storage/sqldb/transactor"
     "time"
    )
    
    type CustomerRepository struct {
     dbCtx transactor.DBTXContext // <-- ตรงนี้
    }
    
    func NewCustomerRepository(dbCtx transactor.DBTXContext) // <-- ตรงนี้
    *CustomerRepository {
     // ...
    }
    
    func (r *CustomerRepository) Create(ctx context.Context, customer *model.Customer) error {
     // ...
    
     err := r.dbCtx(ctx). // <-- ตรงนี้ จะเป็น *sqlx.DB หรือ *sqlx.Tx ก็ได้
     
     // ...
    }
    
    func (r *CustomerRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
     // ...
     
     err := r.dbCtx(ctx). // <-- ตรงนี้ จะเป็น *sqlx.DB หรือ *sqlx.Tx ก็ได้
     
     // ...
    }
    
    func (r *CustomerRepository) FindByID(ctx context.Context, id int64) (*model.Customer, error) {
     // ...
     
     err := r.dbCtx(ctx). // <-- ตรงนี้ จะเป็น *sqlx.DB หรือ *sqlx.Tx ก็ได้
     
     // ...
    }
    
    func (r *CustomerRepository) UpdateCredit(ctx context.Context, m *model.Customer) error {
     // ...
    
     err := r.dbCtx(ctx). // <-- ตรงนี้ จะเป็น *sqlx.DB หรือ *sqlx.Tx ก็ได้
     
     // ...
    }
    
    ```

- แก้ไขไฟล์ `repository/order.go`

    ```go
    package repository
    
    import (
     "context"
     "database/sql"
     "fmt"
     "go-mma/model"
     "go-mma/util/errs"
     "go-mma/util/storage/sqldb/transactor"
     "time"
    )
    
    type OrderRepository struct {
     dbCtx transactor.DBTXContext  // <-- ตรงนี้
    }
    
    func NewOrderRepository(dbCtx transactor.DBTXContext) // <-- ตรงนี้
    *OrderRepository {
     // ...
    }
    
    func (r *OrderRepository) Create(ctx context.Context, m *model.Order) error {
     // ...
    
     err := r.dbCtx(ctx). // <-- ตรงนี้ จะเป็น *sqlx.DB หรือ *sqlx.Tx ก็ได้
     
     // ...
    }
    
    func (r *OrderRepository) FindByID(ctx context.Context, id int64) (*model.Order, error) {
     // ...
     
     err := r.dbCtx(ctx). // <-- ตรงนี้ จะเป็น *sqlx.DB หรือ *sqlx.Tx ก็ได้
     
     // ...
    }
    
    func (r *OrderRepository) Cancel(ctx context.Context, id int64) error {
     // ...
     
     _, err := r.dbCtx(ctx). // <-- ตรงนี้ จะเป็น *sqlx.DB หรือ *sqlx.Tx ก็ได้
     
     // ...
    }
    ```

**สรุปแนวทาง**

- ใช้ `transactor.DBTXContext` แทนการเข้าถึง `sql.DB` โดยตรง
- `dbCtx(ctx)` จะ return `sql.Tx` ถ้ามี transaction, หรือ `sql.DB` ถ้าไม่มี
- ทำให้ repository ทุกตัวสามารถทำงานร่วมกันใน transaction เดียวได้ โดยไม่ต้องรู้ว่าใช้ `sql.Tx` หรือ `sql.DB`

### ปรับปรุง Service Layer

ในชั้น `service` เราจะรับ `transactor.Transactor` มาตั้งแต่ตอนสร้าง Service และจะใช้ `WithinTransaction()` เพื่อให้ทุกคำสั่งฐานข้อมูลอยู่ภายใต้ transaction เดียวกัน ซึ่งช่วยให้ rollback ได้หากเกิดข้อผิดพลาดในกระบวนการใดๆ

**ปรับปรุงโค้ด**

- แก้ไขไฟล์ `service/customer.go` ย้ายส่วนการบันทึกลูกค้าใหม่ กับส่งอีเมล มาไว้ใน `WithinTransaction`

    ```go
    package service
    
    // ...
    
    type CustomerService struct {
     transactor transactor.Transactor // <-- ตรงนี้
     custRepo   *repository.CustomerRepository
     notiSvc    *NotificationService
    }
    
    func NewCustomerService(
     transactor transactor.Transactor, // <-- ตรงนี้
     custRepo *repository.CustomerRepository,
     notiSvc *NotificationService,
    ) *CustomerService {
     return &CustomerService{
      transactor: transactor, // <-- ตรงนี้
      custRepo:   custRepo,
      notiSvc:    notiSvc,
     }
    }
    
    func (s *CustomerService) CreateCustomer(ctx context.Context, req *dto.CreateCustomerRequest) (*dto.CreateCustomerResponse, error) {
     // ตรวจสอบความถูกต้องตาม "กฎทางธุรกิจ" (Business Logic/Semantic Validation)
     // Business Logic: email ต้องไม่ซ้ำในฐานข้อมูล
     
     // แปลง DTO → Model
     customer := model.NewCustomer(req.Email, req.Credit)
     
      // <-- แก้ตรงนี้
     // ย้ายส่วนที่ติดต่อฐานข้อมูล กับส่งอีเมลมาทำงานใน WithinTransaction
     err = s.transactor.WithinTransaction(ctx, func(ctx context.Context, registerPostCommitHook func(transactor.PostCommitHook)) error {
         
      // Unit of Work: register change
      if err := s.custRepo.Create(ctx, customer); err != nil {
       // error logging
       logger.Log.Error(err.Error())
       return err // Unit of Work: rollback จะทำงาน
      }
    
      // ส่งอีเมลต้อนรับ
      if err := s.notiSvc.SendEmail(customer.Email, "Welcome to our service!", map[string]any{
       "message": "Thank you for joining us! We are excited to have you as a member.",
      }); err != nil {
       // error logging
       logger.Log.Error(err.Error())
       return err // Unit of Work: rollback จะทำงาน
      }
    
        // Unit of Work: commit จะเกิดขึ้นอัตโนมัติหลัง func สำเร็จ
      return nil
     })
    
     if err != nil {
      return nil, err
     }
    
     // สร้าง DTO Response
     // ...
    }
    ```

- แก้ไขไฟล์ `service/order.go` ย้ายส่วนที่ติดต่อฐานข้อมูล กับส่งอีเมล มาไว้ใน `WithinTransaction`

    ```go
    package service
    
    // ...
    
    type OrderService struct {
     transactor transactor.Transactor // <-- ตรงนี้
     custRepo   *repository.CustomerRepository
     orderRepo  *repository.OrderRepository
     notiSvc    *NotificationService
    }
    
    func NewOrderService(
     transactor transactor.Transactor, // <-- ตรงนี้
     custRepo *repository.CustomerRepository,
     orderRepo *repository.OrderRepository,
     notiSvc *NotificationService) *OrderService {
     return &OrderService{
      transactor: transactor, // <-- ตรงนี้
      custRepo:   custRepo,
      orderRepo:  orderRepo,
      notiSvc:    notiSvc,
     }
    }
    
    func (s *OrderService) CreateOrder(ctx context.Context, req *dto.CreateOrderRequest) (*dto.CreateOrderResponse, error) {
     // Business Logic: ตรวจสอบ customer id ในฐานข้อมูล
     // ...
    
        // <-- แก้ตรงนี้
     // ย้ายส่วนที่ติดต่อฐานข้อมูล กับส่งอีเมลมาทำงานใน WithinTransaction
     var order *model.Order
     err = s.transactor.WithinTransaction(ctx, func(ctx context.Context, registerPostCommitHook func(transactor.PostCommitHook)) error {
       
       // Business Logic Rule: ตัดยอด credit ถ้าไม่พอให้ error
      if err := customer.ReserveCredit(req.OrderTotal); err != nil {
       return err
      }
     
        // Unit of Work: register change
      // ตัดยอด credit ในตาราง customer
      if err := s.custRepo.UpdateCredit(ctx, customer); err != nil {
       logger.Log.Error(err.Error())
       return err // Unit of Work: rollback จะทำงาน
      }
    
      // สร้าง order ใหม่ DTO -> Model
      order = model.NewOrder(req.CustomerID, req.OrderTotal)
      // Unit of Work: register change
      // บันทึกลงฐานข้อมูล
      err = s.orderRepo.Create(ctx, order)
      if err != nil {
       logger.Log.Error(err.Error())
       return err // Unit of Work: rollback จะทำงาน
      }
    
      // ส่งอีเมลยืนยัน
      err = s.notiSvc.SendEmail(customer.Email, "Order Created", map[string]any{
       "order_id": order.ID,
       "total":    order.OrderTotal,
      })
      if err != nil {
       logger.Log.Error(err.Error())
       return err // Unit of Work: rollback จะทำงาน
      }
        
        // Unit of Work: commit จะเกิดขึ้นอัตโนมัติหลัง func สำเร็จ
      return nil
     })
    
     // จัดการ error จากใน transactor
     if err != nil {
      return nil, err
     }
    
     // สร้าง DTO Response
     // ...
    }
    ```

### Post-Commit Hook

เราควรแยกโค้ดส่วนที่ทำให้เกิด side effects เช่น ส่งอีเมล, call external service มาทำงาน**หลังจาก** commit สำเร็จ เพื่อทำให้มั่นใจว่า data persist แล้ว จะไม่ถูก rollback เพราะ external failure

**ก่อนแก้ไข**

```go
// service/customer.go
// func (s *customerService) CreateCustomer(...)
err = s.transactor.WithinTransaction(ctx, func(ctx context.Context, registerPostCommitHook func(PostCommitHook)) error {
     
  // Unit of Work: register change
  if err := s.custRepo.Create(ctx, customer); err != nil {
   // error logging
   logger.Log.Error(err.Error())
   return err // Unit of Work: rollback จะทำงาน
  }

  // ส่งอีเมลต้อนรับ <-- side-effect
  if err := s.notiSvc.SendEmail(customer.Email, "Welcome to our service!", map[string]any{
   "message": "Thank you for joining us! We are excited to have you as a member.",
  }); err != nil {
   // error logging
   logger.Log.Error(err.Error())
   return err // Unit of Work: rollback จะทำงาน
  }

    // Unit of Work: commit จะเกิดขึ้นอัตโนมัติหลัง func สำเร็จ
  return nil
 })
```

**หลังแก้ไข**

```go
// service/customer.go
// func (s *customerService) CreateCustomer(...)
err = s.transactor.WithinTransaction(ctx, func(ctx context.Context, registerPostCommitHook func(PostCommitHook)) error {
     
  // Unit of Work: register change
  if err := s.custRepo.Create(ctx, customer); err != nil {
   // error logging
   logger.Log.Error(err.Error())
   return err // Unit of Work: rollback จะทำงาน
  }

  // เพิ่มส่งอีเมลต้อนรับ เข้าไปใน hook แทนการเรียกใช้งานทันที
  registerPostCommitHook(func(ctx context.Context) error {
   return s.notiSvc.SendEmail(customer.Email, "Welcome to our service!", map[string]any{
    "message": "Thank you for joining us! We are excited to have you as a member."})
  })

    // Unit of Work: commit จะเกิดขึ้นอัตโนมัติหลัง func สำเร็จ
  return nil
 })
```

แก้ไขที่ `func (s *orderService) CreateOrder(…)` ด้วย

### ปรับปรุง Dependency Injection

เพิ่มการ inject `transactor.Transactor` เข้าไปใน Service Layer

- แก้ไขไฟล์ `application/http.go`

    ```go
    func (s *httpServer) RegisterRoutes(dbCtx sqldb.DBContext) {
     v1 := s.app.Group("/api/v1")
    
      // สร้าง transactor กับ dbCtx จาก transactor
     transactor, dbtxCtx := transactor.New(dbCtx.DB()) // <-- ตรงนี้
     customers := v1.Group("/customers")
     {
      repo := repository.NewCustomerRepository(dbtxCtx) // <-- ตรงนี้
      svcNoti := service.NewNotificationService()
      svc := service.NewCustomerService(transactor, repo, svcNoti) // <-- ตรงนี้
      hdl := handler.NewCustomerHandler(svc)
      customers.Post("", hdl.CreateCustomer)
     }
    
     orders := v1.Group("/orders")
     {
      repoCust := repository.NewCustomerRepository(dbtxCtx) // <-- ตรงนี้
      repoOrder := repository.NewOrderRepository(dbtxCtx) // <-- ตรงนี้
      svcNoti := service.NewNotificationService()
      svcOrder := service.NewOrderService(transactor, repoCust, repoOrder, svcNoti) // <-- ตรงนี้
      hdl := handler.NewOrderHandler(svcOrder)
      orders.Post("", hdlr.CreateOrder)
      orders.Delete("/:orderID", hdl.CancelOrder)
     }
    }
    ```

