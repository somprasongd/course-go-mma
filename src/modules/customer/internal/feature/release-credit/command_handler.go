package releasecredit

import (
	"context"
	"go-mma/modules/customer/domainerrors"
	"go-mma/modules/customer/internal/repository"
	"go-mma/shared/common/logger"
	"go-mma/shared/common/mediator"
	"go-mma/shared/common/storage/sqldb/transactor"
	"go-mma/shared/contract/customercontract"
)

type releaseCreditCommandHandler struct {
	transactor transactor.Transactor
	custRepo   repository.CustomerRepository
}

func NewReleaseCreditCommandHandler(
	transactor transactor.Transactor,
	repo repository.CustomerRepository) *releaseCreditCommandHandler {
	return &releaseCreditCommandHandler{
		transactor: transactor,
		custRepo:   repo,
	}
}

func (h *releaseCreditCommandHandler) Handle(ctx context.Context, cmd *customercontract.ReleaseCreditCommand) (*mediator.NoResponse, error) {
	err := h.transactor.WithinTransaction(ctx, func(ctx context.Context, registerPostCommitHook func(transactor.PostCommitHook)) error {
		customer, err := h.custRepo.FindByID(ctx, cmd.CustomerID)
		if err != nil {
			logger.Log.Error(err.Error())
			return err
		}

		if customer == nil {
			return domainerrors.ErrCustomerNotFound
		}

		customer.ReleaseCredit(cmd.CreditAmount)

		if err := h.custRepo.UpdateCredit(ctx, customer); err != nil {
			logger.Log.Error(err.Error())
			return err
		}

		return nil
	})

	return nil, err
}
