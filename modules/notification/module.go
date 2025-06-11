package notification

import (
	"go-mma/util/module"

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
	// โมดูลนี้ยังไม่มี routes
}
