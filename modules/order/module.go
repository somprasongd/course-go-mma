package order

import (
	"go-mma/modules/order/handler"
	"go-mma/modules/order/repository"
	"go-mma/modules/order/service"
	"go-mma/util/module"
	"go-mma/util/registry"

	custModule "go-mma/modules/customer"
	custService "go-mma/modules/customer/service"
	notiModule "go-mma/modules/notification"
	notiService "go-mma/modules/notification/service"

	"github.com/gofiber/fiber/v3"
)

func NewModule(mCtx *module.ModuleContext) module.Module {
	return &moduleImp{mCtx: mCtx}
}

type moduleImp struct {
	mCtx     *module.ModuleContext
	orderSvc service.OrderService
}

func (m *moduleImp) APIVersion() string {
	return "v1"
}

func (m *moduleImp) Init(reg registry.ServiceRegistry) error {
	// Resolve CustomerService from the registry
	custSvc, err := registry.ResolveAs[custService.CustomerService](reg, custModule.CustomerServiceKey)
	if err != nil {
		return err
	}

	// Resolve NotificationService from the registry
	notiSvc, err := registry.ResolveAs[notiService.NotificationService](reg, notiModule.NotificationServiceKey)
	if err != nil {
		return err
	}

	repo := repository.NewOrderRepository(m.mCtx.DBCtx)
	m.orderSvc = service.NewOrderService(m.mCtx.Transactor, custSvc, repo, notiSvc)

	return nil
}

func (m *moduleImp) RegisterRoutes(router fiber.Router) {
	// wiring dependencies
	hdl := handler.NewOrderHandler(m.orderSvc)

	orders := router.Group("/orders")
	orders.Post("", hdl.CreateOrder)
	orders.Delete("/:orderID", hdl.CancelOrder)
}
