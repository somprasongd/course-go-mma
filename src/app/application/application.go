package application

import (
	"fmt"
	"go-mma/config"
	"go-mma/shared/common/eventbus"
	"go-mma/shared/common/logger"
	"go-mma/shared/common/module"
	"go-mma/shared/common/registry"
)

type Application struct {
	config          config.Config
	httpServer      HTTPServer
	serviceRegistry registry.ServiceRegistry
	eventBus        eventbus.EventBus
}

func New(config config.Config) *Application {
	return &Application{
		config:          config,
		httpServer:      newHTTPServer(config),
		serviceRegistry: registry.NewServiceRegistry(),
		eventBus:        eventbus.NewInMemoryEventBus(),
	}
}

func (app *Application) Run() error {
	app.httpServer.Start()

	return nil
}

func (app *Application) Shutdown() error {
	// Gracefully close fiber server
	logger.Log.Info("Shutting down server")
	if err := app.httpServer.Shutdown(); err != nil {
		logger.Log.Fatal(fmt.Sprintf("Error shutting down server: %v", err))
	}
	logger.Log.Info("Server stopped")

	return nil
}

func (app *Application) RegisterModules(modules ...module.Module) error {
	for _, m := range modules {
		// Initialize each module
		if err := app.initModule(m); err != nil {
			return fmt.Errorf("failed to init module [%T]: %w", m, err)
		}

		// ถ้าโมดูลเป็น ServiceProvider ให้เอา service มาลง registry
		if sp, ok := m.(module.ServiceProvider); ok {
			for _, p := range sp.Services() {
				app.serviceRegistry.Register(p.Key, p.Value)
			}
		}

		app.registerModuleRoutes(m)
	}

	return nil
}

// แยกเป็นฟังก์ชันตาม single-responsibility principle (SRP)
func (app *Application) initModule(m module.Module) error {
	return m.Init(app.serviceRegistry, app.eventBus)
}

func (app *Application) registerModuleRoutes(m module.Module) {
	prefix := app.buildGroupPrefix(m)
	group := app.httpServer.Group(prefix)
	m.RegisterRoutes(group)
}

func (app *Application) buildGroupPrefix(m module.Module) string {
	apiBase := "/api"
	version := m.APIVersion()
	if version != "" {
		return fmt.Sprintf("%s/%s", apiBase, version)
	}
	return apiBase
}
