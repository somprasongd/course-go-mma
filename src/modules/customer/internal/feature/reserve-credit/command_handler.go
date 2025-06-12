package reservecredit

import (
	"context"
	"go-mma/modules/customer/domainerrors"
	"go-mma/modules/customer/internal/repository"
	"go-mma/shared/common/errs"
	"go-mma/shared/common/logger"
	"go-mma/shared/common/mediator"
	"go-mma/shared/common/storage/sqldb/transactor"
	"go-mma/shared/contract/customercontract"
)

type reserveCreditCommandHandler struct {
	transactor transactor.Transactor
	custRepo   repository.CustomerRepository
}

func NewReserveCreditCommandHandler(
	transactor transactor.Transactor,
	repo repository.CustomerRepository) *reserveCreditCommandHandler {
	return &reserveCreditCommandHandler{
		transactor: transactor,
		custRepo:   repo,
	}
}

func (h *reserveCreditCommandHandler) Handle(ctx context.Context, cmd *customercontract.ReserveCreditCommand) (*mediator.NoResponse, error) {
	err := h.transactor.WithinTransaction(ctx, func(ctx context.Context, registerPostCommitHook func(transactor.PostCommitHook)) error {
		customer, err := h.custRepo.FindByID(ctx, cmd.CustomerID)
		if err != nil {
			logger.Log.Error(err.Error())
			return err
		}

		if customer == nil {
			return domainerrors.ErrCustomerNotFound
		}

		if err := customer.ReserveCredit(cmd.CreditAmount); err != nil {
			return err
		}

		if err := h.custRepo.UpdateCredit(ctx, customer); err != nil {
			logger.Log.Error(err.Error())
			return errs.DatabaseFailureError(err.Error())
		}

		return nil
	})

	return nil, err
}
