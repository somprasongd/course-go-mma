package model

import (
	"go-mma/util/errs"
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

func (c *Customer) ReserveCredit(v int) error {
	newCredit := c.Credit - v
	if newCredit < 0 { // เมื่อตัดยอดติดลบแสดงว่า credit ไม่พอ
		return errs.BusinessRuleError("insufficient credit limit")
	}
	c.Credit = newCredit
	return nil
}

func (c *Customer) ReleaseCredit(v int) {
	if c.Credit <= 0 { // reset ยอดก่อนถ้าติดลบ
		c.Credit = 0
	}
	c.Credit = c.Credit + v
}
