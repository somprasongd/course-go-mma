package application

import (
	"context"
	"fmt"
	"go-mma/application/middleware"
	"go-mma/build"
	"go-mma/config"
	"go-mma/shared/common/logger"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

type HTTPServer interface {
	Start()
	Shutdown() error
	Group(prefix string) fiber.Router
}

type httpServer struct {
	config config.Config
	app    *fiber.App
}

func newHTTPServer(config config.Config) HTTPServer {
	return &httpServer{
		config: config,
		app:    newFiber(config),
	}
}

func newFiber(config config.Config) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: fmt.Sprintf("Go MMA version %s", build.Version),
	})

	// global middleware
	app.Use(middleware.Observability()) // จัดการ log + trace + metric
	app.Use(cors.New())                 // CORS ลำดับแรก เพื่อให้ OPTIONS request ผ่านได้เสมอ
	app.Use(recover.New())              // auto-recovers from panic (internal only)
	app.Use(middleware.ResponseError())

	app.Get("/docs/*", middleware.APIDoc(config))

	app.Get("/", func(c fiber.Ctx) error {
		return c.JSON(map[string]string{"version": build.Version, "time": build.Time})
	})

	return app
}

func (s *httpServer) Start() {
	go func() {
		logger.Log().Info(fmt.Sprintf("Starting server on port %d", s.config.HTTPPort))
		if err := s.app.Listen(fmt.Sprintf(":%d", s.config.HTTPPort)); err != nil && err != http.ErrServerClosed {
			logger.Log().Fatal(fmt.Sprintf("Error starting server: %v", err))
		}
	}()
}

func (s *httpServer) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.config.GracefulTimeout)
	defer cancel()
	return s.app.ShutdownWithContext(ctx)
}

// ใช้สำหรับสร้าง base url router เช่น /api/v1
func (s *httpServer) Group(prefix string) fiber.Router {
	return s.app.Group(prefix)
}
