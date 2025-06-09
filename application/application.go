package application

import (
	"fmt"
	"go-mma/config"
	"go-mma/util/logger"
)

type Application struct {
	config     config.Config
	httpServer HTTPServer
}

func New(config config.Config) *Application {
	return &Application{
		config:     config,
		httpServer: newHTTPServer(config),
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

func (app *Application) RegisterRoutes() {
	app.httpServer.RegisterRoutes()
}
