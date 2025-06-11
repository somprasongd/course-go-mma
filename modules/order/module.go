package order

import (
	"go-mma/modules/order/handler"
	"go-mma/modules/order/repository"
	"go-mma/modules/order/service"
	"go-mma/util/module"

	custRepository "go-mma/modules/customer/repository"
	notiService "go-mma/modules/notification/service"

	"github.com/gofiber/fiber/v3"
)

func NewModule(mCtx *module.ModuleContext) module.Module {
	return &moduleImp{mCtx}
}

type moduleImp struct {
	mCtx *module.ModuleContext
}

func (m *moduleImp) APIVersion() string {
	return "v1"
}

func (m *moduleImp) RegisterRoutes(router fiber.Router) {
	// wiring dependencies
	repoCust := custRepository.NewCustomerRepository(m.mCtx.DBCtx)
	repoOrder := repository.NewOrderRepository(m.mCtx.DBCtx)
	svcNoti := notiService.NewNotificationService()
	svc := service.NewOrderService(m.mCtx.Transactor, repoCust, repoOrder, svcNoti)
	hdl := handler.NewOrderHandler(svc)

	orders := router.Group("/orders")
	orders.Post("", hdl.CreateOrder)
	orders.Delete("/:orderID", hdl.CancelOrder)
}
