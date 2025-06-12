package customer

import (
	"go-mma/modules/customer/handler"
	"go-mma/modules/customer/internal/repository"
	"go-mma/modules/customer/service"
	"go-mma/shared/common/module"
	"go-mma/shared/common/registry"
	"go-mma/shared/contract/customercontract"

	notiModule "go-mma/modules/notification"
	notiService "go-mma/modules/notification/service"

	"github.com/gofiber/fiber/v3"
)

func NewModule(mCtx *module.ModuleContext) module.Module {
	return &moduleImp{mCtx: mCtx}
}

type moduleImp struct {
	mCtx    *module.ModuleContext
	custSvc service.CustomerService
}

func (m *moduleImp) APIVersion() string {
	return "v1"
}

func (m *moduleImp) Init(reg registry.ServiceRegistry) error {
	// Resolve NotificationService from the registry
	notiSvc, err := registry.ResolveAs[notiService.NotificationService](reg, notiModule.NotificationServiceKey)
	if err != nil {
		return err
	}

	repo := repository.NewCustomerRepository(m.mCtx.DBCtx)
	m.custSvc = service.NewCustomerService(m.mCtx.Transactor, repo, notiSvc)

	return nil
}

func (m *moduleImp) Services() []registry.ProvidedService {
	return []registry.ProvidedService{
		{Key: customercontract.CreditManagerKey, Value: m.custSvc},
	}
}

func (m *moduleImp) RegisterRoutes(router fiber.Router) {
	// wiring dependencies
	hdl := handler.NewCustomerHandler(m.custSvc)

	customers := router.Group("/customers")
	customers.Post("", hdl.CreateCustomer)
}
