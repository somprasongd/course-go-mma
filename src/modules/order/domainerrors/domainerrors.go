package domainerrors

import "go-mma/shared/common/errs"

var (
	ErrNoOrderID = errs.ResourceNotFoundError("the order with given id was not found")
)
