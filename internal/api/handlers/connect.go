package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/harisiqbal12/opennet/pkg/node"
)

func Connect(c *fiber.Ctx) error {
	network := c.Locals("network").(*node.Network)

	network.Discover()

	if network.Node != nil {
		peerCount := len(network.Node.Network().Peers())
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message":   fmt.Sprintf("Connected to %d peers", peerCount),
			"nodeId":    network.Node.ID(),
			"peerCount": peerCount,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": fmt.Sprintf("error %s", "An error occurred"),
	})
}
