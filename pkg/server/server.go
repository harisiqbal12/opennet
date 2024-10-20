package server

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	routes "github.com/harisiqbal12/opennet/internal/api"
	"github.com/harisiqbal12/opennet/pkg/node"
)

func StartServer(network *node.Network) {
	app := fiber.New()

	api := app.Group("/api/v1", func(c *fiber.Ctx) error {

		c.Locals("network", network)

		return c.Next()
	})

	routes.SetupRoutes(api)

	var PORT string = os.Getenv("PORT")

	if len(PORT) == 0 {
		PORT = "3000"
	}

	log.Fatal(app.Listen(":" + PORT))
}
