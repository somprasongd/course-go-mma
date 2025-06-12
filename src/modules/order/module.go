package order

import (
	"go-mma/modules/order/internal/feature/cancel"
	"go-mma/modules/order/internal/feature/create"
	"go-mma/modules/order/internal/repository"
	"go-mma/shared/common/eventbus"
	"go-mma/shared/common/mediator"
	"go-mma/shared/common/module"
	"go-mma/shared/common/registry"

	notiModule "go-mma/modules/notification"
	notiService "go-mma/modules/notification/service"

	"github.com/gofiber/fiber/v3"
)

func NewModule(mCtx *module.ModuleContext) module.Module {
	return &moduleImp{mCtx: mCtx}
}

type moduleImp struct {
	mCtx *module.ModuleContext
}

func (m *moduleImp) APIVersion() string {
	return "v1"
}

func (m *moduleImp) Init(reg registry.ServiceRegistry, eventBus eventbus.EventBus) error {

	// Resolve NotificationService from the registry
	notiSvc, err := registry.ResolveAs[notiService.NotificationService](reg, notiModule.NotificationServiceKey)
	if err != nil {
		return err
	}

	repo := repository.NewOrderRepository(m.mCtx.DBCtx)

	mediator.Register(create.NewCreateOrderCommandHandler(m.mCtx.Transactor, repo, notiSvc))
	mediator.Register(cancel.NewCancelOrderCommandHandler(m.mCtx.Transactor, repo))

	return nil
}

func (m *moduleImp) RegisterRoutes(router fiber.Router) {
	orders := router.Group("/orders")
	create.NewEndpoint(orders, "")
	cancel.NewEndpoint(orders, "/:orderID")
}
