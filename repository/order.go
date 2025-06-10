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

// --> Step 1: สร้าง interface
type OrderRepository interface {
	Create(ctx context.Context, order *model.Order) error
	FindByID(ctx context.Context, id int64) (*model.Order, error)
	Cancel(ctx context.Context, id int64) error
}

type orderRepository struct { // --> Step 2: เปลี่ยนชื่อ struct เป็นตัวพิมพ์เล็ก
	dbCtx transactor.DBTXContext
}

// --> Step 3: return เป็น interface
func NewOrderRepository(dbCtx transactor.DBTXContext) OrderRepository {
	return &orderRepository{ // --> Step 4: เปลี่ยนชื่อ struct เป็นตัวพิมพ์เล็ก
		dbCtx: dbCtx,
	}
}

func (r *orderRepository) Create(ctx context.Context, m *model.Order) error {
	query := `
	INSERT INTO public.orders (
			id, customer_id, order_total
	)
	VALUES ($1, $2, $3)
	RETURNING *
	`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := r.dbCtx(ctx).QueryRowxContext(ctx, query, m.ID, m.CustomerID, m.OrderTotal).StructScan(m)
	if err != nil {
		return errs.HandleDBError(fmt.Errorf("an error occurred while inserting an order: %w", err))
	}
	return nil
}

func (r *orderRepository) FindByID(ctx context.Context, id int64) (*model.Order, error) {
	query := `
	SELECT *
	FROM public.orders
	WHERE id = $1
	AND canceled_at IS NULL -- รายออเดอร์ต้องยังไม่ถูกยกเลิก
`
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var order model.Order
	err := r.dbCtx(ctx).QueryRowxContext(ctx, query, id).StructScan(&order)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errs.HandleDBError(fmt.Errorf("an error occurred while finding a order by id: %w", err))
	}
	return &order, nil
}

func (r *orderRepository) Cancel(ctx context.Context, id int64) error {
	query := `
	UPDATE public.orders
	SET canceled_at = current_timestamp -- soft delete record
	WHERE id = $1
`
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, err := r.dbCtx(ctx).ExecContext(ctx, query, id)
	if err != nil {
		return errs.HandleDBError(fmt.Errorf("failed to cancel order: %w", err))
	}
	return nil
}
