package module

import (
	"go-mma/util/storage/sqldb/transactor"

	"github.com/gofiber/fiber/v3"
)

type Module interface {
	APIVersion() string
	RegisterRoutes(r fiber.Router)
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
