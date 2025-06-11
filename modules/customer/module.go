package customer

import (
	"go-mma/modules/customer/handler"
	"go-mma/modules/customer/repository"
	"go-mma/modules/customer/service"
	"go-mma/util/module"

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
	repo := repository.NewCustomerRepository(m.mCtx.DBCtx)
	svcNoti := notiService.NewNotificationService()
	svc := service.NewCustomerService(m.mCtx.Transactor, repo, svcNoti)
	hdl := handler.NewCustomerHandler(svc)

	customers := router.Group("/customers")
	customers.Post("", hdl.CreateCustomer)
}
