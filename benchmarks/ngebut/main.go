package main

import (
	"log"

	"github.com/ryanbekhen/ngebut"
)

func main() {
	// Create a new ngebut app
	app := ngebut.New()

	// Simple route that returns a string
	app.GET("/", func(c *ngebut.Ctx) {
		c.String("Hello, World!")
	})

	// Route that returns JSON
	app.GET("/json", func(c *ngebut.Ctx) {
		c.JSON(map[string]interface{}{
			"message": "Hello, World!",
			"status": "success",
		})
	})

	// Route with path parameter
	app.GET("/users/:id", func(c *ngebut.Ctx) {
		id := c.Param("id")
		c.JSON(map[string]interface{}{
			"user_id": id,
			"message": "User details retrieved",
		})
	})

	// Start the server
	log.Println("Ngebut server starting on http://localhost:3000")
	if err := app.Listen(":3000"); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}