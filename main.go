package main

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
)

var (
	Version = "local-dev"
	Time    = "n/a"
)

func main() {
	app := fiber.New(fiber.Config{
		AppName: fmt.Sprintf("Go MMA version %s", Version),
	})

	app.Get("/", func(c fiber.Ctx) error {
		// การตอบกลับด้วย JSON
		return c.JSON(map[string]string{"version": Version, "time": Time})
	})

	app.Listen(":8090")
}
