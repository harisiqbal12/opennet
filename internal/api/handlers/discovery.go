package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/harisiqbal12/opennet/pkg/node"
)

func Discovery(c *fiber.Ctx) error {

	network := c.Locals("network").(*node.Network)

	if network.Node != nil {
		return c.SendStatus(fiber.StatusOK)
	}

	return c.SendStatus(fiber.StatusOK)
}
