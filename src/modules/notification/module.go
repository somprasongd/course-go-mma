package notification

import (
	"go-mma/modules/notification/internal/integration/customer"
	"go-mma/modules/notification/service"
	"go-mma/shared/common/eventbus"
	"go-mma/shared/common/module"
	"go-mma/shared/common/registry"
	"go-mma/shared/messaging"

	"github.com/gofiber/fiber/v3"
)

const (
	NotificationServiceKey registry.ServiceKey = "NotificationService"
)

func NewModule(mCtx *module.ModuleContext) module.Module {
	return &moduleImp{mCtx: mCtx}
}

type moduleImp struct {
	mCtx    *module.ModuleContext
	notiSvc service.NotificationService
}

func (m *moduleImp) APIVersion() string {
	return "v1"
}

func (m *moduleImp) Init(reg registry.ServiceRegistry, eventBus eventbus.EventBus) error {
	m.notiSvc = service.NewNotificationService()

	// subscribe to integration events
	eventBus.Subscribe(messaging.CustomerCreatedIntegrationEventName, customer.NewWelcomeEmailHandler(m.notiSvc))

	return nil
}

func (m *moduleImp) Services() []registry.ProvidedService {
	return []registry.ProvidedService{
		{Key: NotificationServiceKey, Value: m.notiSvc},
	}
}

func (m *moduleImp) RegisterRoutes(router fiber.Router) {
	// โมดูลนี้ยังไม่มี routes
}
