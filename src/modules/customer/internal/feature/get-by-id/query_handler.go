package getbyid

import (
	"context"
	"go-mma/modules/customer/domainerrors"
	"go-mma/modules/customer/internal/model"
	"go-mma/modules/customer/internal/repository"
	"go-mma/shared/contract/customercontract"
)

type getCustomerByIDQueryHandler struct {
	custRepo repository.CustomerRepository
}

func NewGetCustomerByIDQueryHandler(custRepo repository.CustomerRepository) *getCustomerByIDQueryHandler {
	return &getCustomerByIDQueryHandler{
		custRepo: custRepo,
	}
}

func (h *getCustomerByIDQueryHandler) Handle(ctx context.Context, query *customercontract.GetCustomerByIDQuery) (*customercontract.GetCustomerByIDQueryResult, error) {
	customer, err := h.custRepo.FindByID(ctx, query.ID)
	if err != nil {
		return nil, err
	}
	if customer == nil {
		return nil, domainerrors.ErrCustomerNotFound
	}
	return h.newGetCustomerByIDQueryResult(customer), nil
}

func (h *getCustomerByIDQueryHandler) newGetCustomerByIDQueryResult(customer *model.Customer) *customercontract.GetCustomerByIDQueryResult {
	return &customercontract.GetCustomerByIDQueryResult{
		ID:     customer.ID,
		Email:  customer.Email,
		Credit: customer.Credit,
	}
}
