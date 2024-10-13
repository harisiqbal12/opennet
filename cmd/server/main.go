package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	routes "github.com/harisiqbal12/opennet/internal/api"
	"github.com/harisiqbal12/opennet/pkg/node"
)

func main() {
	fmt.Println("Starting Opennet!!")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// starting point
	app := fiber.New()

	networkChan := make(chan *node.Network)

	opts := node.NetworkOpts{
		Context:   ctx,
		Networkch: networkChan,
	}

	go node.CreateNetwork(opts)

	network := <-networkChan

	var api fiber.Router = app.Group("/api/v1", func(c *fiber.Ctx) error {

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
