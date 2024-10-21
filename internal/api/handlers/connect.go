package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/harisiqbal12/opennet/pkg/node"
)

func Connect(c *fiber.Ctx) error {
	network := c.Locals("network").(*node.Network)

	var payload struct {
		Address string `json:"address"`
	}

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := network.PublishMessage("connection", payload.Address); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err,
		})
	}

	return c.Status(fiber.StatusOK).JSON(
		fiber.Map{
			"status":  "connecting",
			"address": payload.Address,
			"message": "Request send for conencting",
		},
	)

}
