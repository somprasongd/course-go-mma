package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

var (
	Version = "local-dev"
	Time    = "n/a"
)

func main() {
	app := fiber.New(fiber.Config{
		AppName: fmt.Sprintf("Go MMA version %s", Version),
	})

	// กำหนด global middleware
	app.Use(cors.New())      // CORS ลำดับแรก เพื่อให้ OPTIONS request ผ่านได้เสมอ
	app.Use(requestid.New()) // สร้าง request id ใน request header สำหรับการ debug
	app.Use(recover.New())   // auto-recovers from panic (internal only)
	app.Use(logger.New())    // logs HTTP request

	app.Get("/", func(c fiber.Ctx) error {
		return c.JSON(map[string]string{"version": Version, "time": Time})
	})

	// แยกการทำ routing ให้ชัดเจน
	v1 := app.Group("/api/v1")

	// สร้างกลุ่มของ customer
	customers := v1.Group("/customers")
	{
		customers.Post("", func(c fiber.Ctx) error {
			// เพิ่มหน่วงเวลา 3 วินาที สำหรับทดสอบ Graceful Shutdown
			time.Sleep(3 * time.Second)
			return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": 1})
		})
	}

	// สร้างกลุ่มของ order
	orders := v1.Group("/orders")
	{
		orders.Post("", func(c fiber.Ctx) error {
			return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": 1})
		})

		orders.Delete("/:orderID", func(c fiber.Ctx) error {
			// การตอบกลับแค่ status code เพียงอย่างเดียว
			return c.SendStatus(fiber.StatusNoContent)
		})
	}

	// ย้ายมา run server ใน goroutine
	go func() {
		if err := app.Listen(":8090"); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// รอสัญญาณการปิด
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down...")

	// หยุดรับ request ใหม่ แล้ว รอให้ request เดิมทำงานเสร็จ ภายใน timeout ที่กำหนด (เช่น 5 วินาที)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Fatalf("Error shutting down server: %v", err)
	}

	// Optionally: แล้วค่อยปิด resource อื่นๆ เช่น DB connection, cleanup, etc.

	log.Println("Shutdown complete.")
}
