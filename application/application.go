package application

import (
	"fmt"
	"go-mma/config"
	"go-mma/util/logger"
	"go-mma/util/storage/sqldb"
)

type Application struct {
	config     config.Config
	httpServer HTTPServer
	dbCtx      sqldb.DBContext
}

func New(config config.Config, dbCtx sqldb.DBContext) *Application {
	return &Application{
		config:     config,
		httpServer: newHTTPServer(config),
		dbCtx:      dbCtx,
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
