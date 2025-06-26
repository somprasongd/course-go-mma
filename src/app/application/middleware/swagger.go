package middleware

import (
	"fmt"
	"go-mma/build"
	"go-mma/config"
	"go-mma/docs"
	"strings"

	"github.com/gofiber/fiber/v3"
	fiberSwagger "github.com/somprasongd/fiber-swagger"
)

func APIDoc(config config.Config) fiber.Handler {
	//Swagger Doc details
	host := removeProtocol(config.GatewayHost)
	basePath := config.GatewayBasePath
	schemas := []string{"http", "https"}

	if len(host) == 0 {
		host = fmt.Sprintf("localhost:%d", config.HTTPPort)
	}

	if len(basePath) == 0 {
		basePath = "/api/v1"
	}

	docs.SwaggerInfo.Title = "Go MMA Example API"
	docs.SwaggerInfo.Description = "This is a sample server GO MMA server."
	docs.SwaggerInfo.Version = build.Version
	docs.SwaggerInfo.Host = host
	docs.SwaggerInfo.BasePath = basePath
	docs.SwaggerInfo.Schemes = schemas

	return fiberSwagger.WrapHandler
}

func removeProtocol(url string) string {
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")
	return url
}
