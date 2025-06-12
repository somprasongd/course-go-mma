package module

import (
	"go-mma/shared/common/eventbus"
	"go-mma/shared/common/registry"
	"go-mma/shared/common/storage/sqldb/transactor"

	"github.com/gofiber/fiber/v3"
)

type Module interface {
	APIVersion() string
	Init(reg registry.ServiceRegistry, eventBus eventbus.EventBus) error // รับ eventBus เพิ่ม
	RegisterRoutes(r fiber.Router)
}

// แยกออกมาเพราะว่า บางโมดูลอาจไม่ต้อง export service
type ServiceProvider interface {
	Services() []registry.ProvidedService
}

type ModuleContext struct {
	Transactor transactor.Transactor
	DBCtx      transactor.DBTXContext
}

func NewModuleContext(transactor transactor.Transactor, dbCtx transactor.DBTXContext) *ModuleContext {
	return &ModuleContext{
		Transactor: transactor,
		DBCtx:      dbCtx,
	}
}
