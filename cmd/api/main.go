package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"go-mma/application"
	"go-mma/config"
	"go-mma/util/logger"
)

var (
	Version = "local-dev"
	Time    = "n/a"
)

func main() {
	closeLog, err := logger.Init()
	if err != nil {
		panic(err.Error())
	}
	defer closeLog()

	config, err := config.Load()
	if err != nil {
		log.Panic(err)
	}

	app := application.New(*config)
	app.RegisterRoutes()
	app.Run()

	// รอสัญญาณการปิด
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Log.Info("Shutting down...")

	app.Shutdown()

	// Optionally: แล้วค่อยปิด resource อื่นๆ เช่น DB connection, cleanup, etc.

	logger.Log.Info("Shutdown complete.")
}
