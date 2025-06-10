package application

import (
	"context"
	"fmt"
	"go-mma/application/middleware"
	"go-mma/build"
	"go-mma/config"
	"go-mma/handler"
	"go-mma/repository"
	"go-mma/service"
	"go-mma/util/logger"
	"go-mma/util/storage/sqldb"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

type HTTPServer interface {
	Start()
	Shutdown() error
	RegisterRoutes(dbCtx sqldb.DBContext)
}

type httpServer struct {
	config config.Config
	app    *fiber.App
}

func newHTTPServer(config config.Config) HTTPServer {
	return &httpServer{
		config: config,
		app:    newFiber(),
	}
}

func newFiber() *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: fmt.Sprintf("Go MMA version %s", build.Version),
	})

	// global middleware
	app.Use(cors.New())                 // CORS ลำดับแรก เพื่อให้ OPTIONS request ผ่านได้เสมอ
	app.Use(requestid.New())            // สร้าง request id ใน request header สำหรับการ debug
	app.Use(recover.New())              // auto-recovers from panic (internal only)
	app.Use(middleware.RequestLogger()) // logs HTTP request
	app.Use(middleware.ResponseError())

	app.Get("/", func(c fiber.Ctx) error {
		return c.JSON(map[string]string{"version": build.Version, "time": build.Time})
	})

	return app
}

func (s *httpServer) Start() {
	go func() {
		logger.Log.Info(fmt.Sprintf("Starting server on port %d", s.config.HTTPPort))
		if err := s.app.Listen(fmt.Sprintf(":%d", s.config.HTTPPort)); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal(fmt.Sprintf("Error starting server: %v", err))
		}
	}()
}

func (s *httpServer) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.config.GracefulTimeout)
	defer cancel()
	return s.app.ShutdownWithContext(ctx)
}

func (s *httpServer) RegisterRoutes(dbCtx sqldb.DBContext) {
	v1 := s.app.Group("/api/v1")

	customers := v1.Group("/customers")
	{
		// กำหนด dependency ระหว่างเลเยอร์
		repo := repository.NewCustomerRepository(dbCtx)
		svcNoti := service.NewNotificationService() // <-- เพิ่มตรงนี้
		// ส่ง instance ของ repository, notiSvc เข้า service
		svc := service.NewCustomerService(repo, svcNoti) // <-- เพิ่มตรงนี้
		// ส่ง service เข้า handler
		hdlr := handler.NewCustomerHandler(svc)
		// Register routes เข้ากับ HTTP server
		customers.Post("", hdlr.CreateCustomer)
	}

	orders := v1.Group("/orders")
	{
		hdl := handler.NewOrderHandler()
		orders.Post("", hdl.CreateOrder)
		orders.Delete("/:orderID", hdl.CancelOrder)
	}
}
