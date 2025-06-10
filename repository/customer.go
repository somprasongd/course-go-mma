package repository

import (
	"context"
	"database/sql"
	"fmt"
	"go-mma/model"
	"go-mma/util/errs"
	"go-mma/util/storage/sqldb"
	"time"
)

type CustomerRepository struct {
	dbCtx sqldb.DBContext // ใช้งาน database ผ่าน DBContext interface
}

func NewCustomerRepository(dbCtx sqldb.DBContext) *CustomerRepository {
	return &CustomerRepository{
		dbCtx: dbCtx, // inject DBContext instance into CustomerRepository
	}
}

func (r *CustomerRepository) Create(ctx context.Context, customer *model.Customer) error {
	query := `
	INSERT INTO public.customers (id, email, credit)
	VALUES ($1, $2, $3)
	RETURNING *
	`

	// กำหนด timeout ของ query
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err := r.dbCtx.DB().
		QueryRowxContext(ctx, query, customer.ID, customer.Email, customer.Credit).
		StructScan(customer) // นำค่า created_at, updated_at ใส่ใน struct customer
	if err != nil {
		return errs.HandleDBError(fmt.Errorf("an error occurred while inserting customer: %w", err))
	}
	return nil
}

func (r *CustomerRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT 1 FROM public.customers WHERE email = $1 LIMIT 1`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var exists int
	err := r.dbCtx.DB().
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
