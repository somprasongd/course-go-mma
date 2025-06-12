package customercontract

import (
	"context"
	"go-mma/shared/common/registry"
)

const (
	CreditManagerKey registry.ServiceKey = "customer:contract:credit"
)

type CustomerInfo struct {
	ID     int64  `json:"id"`
	Email  string `json:"email"`
	Credit int    `json:"credit"`
}

func NewCustomerInfo(id int64, email string, credit int) *CustomerInfo {
	return &CustomerInfo{ID: id, Email: email, Credit: credit}
}

type CustomerReader interface {
	GetCustomerByID(ctx context.Context, id int64) (*CustomerInfo, error)
}

type CreditManager interface {
	CustomerReader // embed เพื่อ reuse
	ReserveCredit(ctx context.Context, id int64, amount int) error
	ReleaseCredit(ctx context.Context, id int64, amount int) error
}
