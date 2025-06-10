package dto

import "fmt"

type CreateOrderRequest struct {
	CustomerID int64 `json:"customer_id"`
	OrderTotal int   `json:"order_total"`
}

func (r *CreateOrderRequest) Validate() error {
	if r.CustomerID <= 0 {
		return fmt.Errorf("customer_id is required")
	}
	return nil
}

type CreateOrderResponse struct {
	ID int64 `json:"id"`
}

func NewCreateOrderResponse(id int64) *CreateOrderResponse {
	return &CreateOrderResponse{ID: id}
}
