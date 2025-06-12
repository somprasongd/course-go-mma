package model

import (
	"go-mma/shared/common/idgen"
	"time"
)

type Order struct {
	ID         int64      `db:"id"`
	CustomerID int64      `db:"customer_id"`
	OrderTotal int        `db:"order_total"`
	CreatedAt  time.Time  `db:"created_at"`
	CanceledAt *time.Time `db:"canceled_at"` // nullable
}

func NewOrder(customerID int64, orderTotal int) *Order {
	return &Order{
		ID:         idgen.GenerateTimeRandomID(),
		CustomerID: customerID,
		OrderTotal: orderTotal,
	}
}
