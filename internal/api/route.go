package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/harisiqbal12/opennet/internal/api/handlers"
)

func SetupRoutes(app fiber.Router) {
	app.Post("/publish", handlers.Publish)
	app.Post("/connect", handlers.Connect)
}
