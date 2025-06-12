package repository

import (
	"context"
	"database/sql"
	"fmt"
	"go-mma/modules/customer/internal/model"
	"go-mma/shared/common/errs"
	"go-mma/shared/common/storage/sqldb/transactor"
	"time"
)

// --> Step 1: สร้าง interface
type CustomerRepository interface {
	Create(ctx context.Context, customer *model.Customer) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	FindByID(ctx context.Context, id int64) (*model.Customer, error)
	UpdateCredit(ctx context.Context, customer *model.Customer) error
}

type customerRepository struct { // --> Step 2: เปลี่ยนชื่อ struct เป็นตัวพิมพ์เล็ก
	dbCtx transactor.DBTXContext
}

// --> Step 3: return เป็น interface
func NewCustomerRepository(dbCtx transactor.DBTXContext) CustomerRepository {
	return &customerRepository{ // --> Step 4: เปลี่ยนชื่อ struct เป็นตัวพิมพ์เล็ก
		dbCtx: dbCtx,
	}
}

func (r *customerRepository) Create(ctx context.Context, customer *model.Customer) error {
	query := `
	INSERT INTO public.customers (id, email, credit)
	VALUES ($1, $2, $3)
	RETURNING *
	`

	// กำหนด timeout ของ query
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err := r.dbCtx(ctx). // <-- จะเป็น *sqlx.DB หรือ *sqlx.Tx ก็ได้
				QueryRowxContext(ctx, query, customer.ID, customer.Email, customer.Credit).
				StructScan(customer) // นำค่า created_at, updated_at ใส่ใน struct customer
	if err != nil {
		return errs.HandleDBError(fmt.Errorf("an error occurred while inserting customer: %w", err))
	}
	return nil
}

func (r *customerRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT 1 FROM public.customers WHERE email = $1 LIMIT 1`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var exists int
	err := r.dbCtx(ctx).
		QueryRowxContext(ctx, query, email).
		Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows { // หาไม่เจอแสดงว่ายังไม่มี email ในระบบแล้ว
			return false, nil
		}
		return false, errs.HandleDBError(fmt.Errorf("an error occurred while checking email: %w", err))
	}
	return true, nil // ถ้าไม่ error แสดงว่ามี email ในระบบแล้ว
}

func (r *customerRepository) FindByID(ctx context.Context, id int64) (*model.Customer, error) {
	query := `
	SELECT *
	FROM public.customers
	WHERE id = $1
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

func (r *customerRepository) UpdateCredit(ctx context.Context, m *model.Customer) error {
	query := `
	UPDATE public.customers
	SET credit = $2
	WHERE id = $1
	RETURNING *
`
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	err := r.dbCtx(ctx).QueryRowxContext(ctx, query, m.ID, m.Credit).StructScan(m)
	if err != nil {
		return errs.HandleDBError(fmt.Errorf("an error occurred while updating customer credit: %w", err))
	}
	return nil
}
