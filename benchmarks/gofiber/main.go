package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// Create a new fiber app
	app := fiber.New()

	// Simple route that returns a string
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	// Route that returns JSON
	app.Get("/json", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Hello, World!",
			"status":  "success",
		})
	})

	// Route with path parameter
	app.Get("/users/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		return c.JSON(fiber.Map{
			"user_id": id,
			"message": "User details retrieved",
		})
	})

	// Start the server
	log.Println("Fiber server starting on http://localhost:3000")
	if err := app.Listen(":3000"); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
