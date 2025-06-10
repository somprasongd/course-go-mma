package model

import (
	"go-mma/util/idgen"
	"time"
)

type Customer struct {
	ID        int64     `db:"id"` // tag db ใช้สำหรับ StructScan() ของ sqlx
	Email     string    `db:"email"`
	Credit    int       `db:"credit"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func NewCustomer(email string, credit int) *Customer {
	return &Customer{
		ID:     idgen.GenerateTimeRandomID(),
		Email:  email,
		Credit: credit,
	}
}
