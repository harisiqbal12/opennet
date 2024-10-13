package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/harisiqbal12/opennet/internal/api/handlers"
)

func SetupRoutes(app fiber.Router) {
	app.Get("/healthz", handlers.Health)
	app.Get("/connect", handlers.Connect)
}
