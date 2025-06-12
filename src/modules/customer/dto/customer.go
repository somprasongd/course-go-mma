package dto

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
		return errors.New("email is required")
	}
	if _, err := mail.ParseAddress(r.Email); err != nil {
		return errors.New("email is invalid")
	}
	if r.Credit <= 0 {
		errs = errors.Join(errs, errors.New("credit must be greater than 0"))
	}
	return errs
}

type CreateCustomerResponse struct {
	ID int64 `json:"id"`
}

func NewCreateCustomerResponse(id int64) *CreateCustomerResponse {
	return &CreateCustomerResponse{ID: id}
}

type CustomerInfo struct {
	ID     int64  `json:"id"`
	Email  string `json:"email"`
	Credit int    `json:"credit"`
}

func NewCustomerInfo(id int64, email string, credit int) *CustomerInfo {
	return &CustomerInfo{ID: id, Email: email, Credit: credit}
}
