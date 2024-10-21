package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/harisiqbal12/opennet/pkg/node"
)

func Publish(c *fiber.Ctx) error {

	network := c.Locals("network").(*node.Network)

	var msg struct {
		Topic   string `json:"topic"`
		Content string `json:"content"`
	}

	if err := c.BodyParser(&msg); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := network.PublishMessage(msg.Topic, msg.Content); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": fmt.Sprintf("Message published"),
	})
}
