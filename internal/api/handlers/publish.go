package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func Publish(c *fiber.Ctx) error {
	var msg struct {
		Topic   string `json:"topic"`
		Content string `json:"content"`
	}

	if err := c.BodyParser(&msg); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": fmt.Sprintf("Message published"),
	})
}
