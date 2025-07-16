package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go-mma/application"
	"go-mma/config"
	"go-mma/modules/customer"
	"go-mma/modules/notification"
	"go-mma/modules/order"
	"go-mma/shared/common/logger"
	"go-mma/shared/common/module"
	"go-mma/shared/common/observability"
	"go-mma/shared/common/storage/sqldb"
	"go-mma/shared/common/storage/sqldb/transactor"
)

// //  @securityDefinitions.apikey	ApiKeyAuth
// //  @in							header
// //  @name						Authorization
// //  @description				Enter the token with the `Bearer: ` prefix, e.g. "Bearer YOUR_JWT_TOKEN"
func main() {
	config, err := config.Load()
	if err != nil {
		panic(err.Error())
	}

	closeLog, err := logger.Init(config.AppName)
	if err != nil {
		panic(err.Error())
	}
	defer closeLog()

	shutdown, err := observability.InitOtlp(context.Background(), config.OtelExporterEnpoint, config.AppName)
	if err != nil {
		panic(err.Error())
	}
	defer shutdown(context.Background())

	dbCtx, closeDB, err := sqldb.NewDBContext(config.DSN)
	if err != nil {
		panic(err.Error())
	}
	defer func() { // ใช่ท่า IIFE เพราะต้องการแสดง error ถ้าปิดไม่ได้
		if err := closeDB(); err != nil {
			logger.Log().Error(fmt.Sprintf("Error closing database: %v", err))
		}
	}()

	app := application.New(*config)

	transactor, dbtxCtx := transactor.New(dbCtx.DB(),
		// เพิ่มใช้งาน nested transaction strategy ที่ใช้ Savepoints
		transactor.WithNestedTransactionStrategy(transactor.NestedTransactionsSavepoints))
	mCtx := module.NewModuleContext(transactor, dbtxCtx)
	app.RegisterModules(
		notification.NewModule(mCtx),
		customer.NewModule(mCtx),
		order.NewModule(mCtx),
	)

	app.Run()

	// รอสัญญาณการปิด
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Log().Info("Shutting down...")

	app.Shutdown()

	// Optionally: แล้วค่อยปิด resource อื่นๆ เช่น DB connection, cleanup, etc.

	logger.Log().Info("Shutdown complete.")
}
