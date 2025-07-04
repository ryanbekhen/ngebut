package main

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

func main() {
	// Create a new echo instance
	e := echo.New()

	// Simple route that returns a string
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	// Route that returns JSON
	e.GET("/json", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Hello, World!",
			"status":  "success",
		})
	})

	// Route with path parameter
	e.GET("/users/:id", func(c echo.Context) error {
		id := c.Param("id")
		return c.JSON(http.StatusOK, map[string]interface{}{
			"user_id": id,
			"message": "User details retrieved",
		})
	})

	// Start the server
	log.Println("Echo server starting on http://localhost:3000")
	if err := e.Start(":3000"); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}