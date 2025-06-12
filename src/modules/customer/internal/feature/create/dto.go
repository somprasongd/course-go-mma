package create

import (
	"errors"
	"net/mail"
)

type CreateCustomerRequest struct {
	Email  string `json:"email"`
	Credit int    `json:"credit"`
}

func (r *CreateCustomerRequest) Validate() error {
	var errs error
	if r.Email == "" {
		errs = errors.Join(errs, errors.New("email is required"))
	}
	if _, err := mail.ParseAddress(r.Email); err != nil {
		errs = errors.Join(errs, errors.New("email is invalid"))
	}
	if r.Credit <= 0 {
		errs = errors.Join(errs, errors.New("credit must be greater than 0"))
	}
	return errs
}

type CreateCustomerResponse struct {
	ID int64 `json:"id"`
}
