package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// Set Gin to release mode
	gin.SetMode(gin.ReleaseMode)

	// Create a new gin router
	r := gin.New()

	// Simple route that returns a string
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello, World!")
	})

	// Route that returns JSON
	r.GET("/json", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello, World!",
			"status":  "success",
		})
	})

	// Route with path parameter
	r.GET("/users/:id", func(c *gin.Context) {
		id := c.Param("id")
		c.JSON(http.StatusOK, gin.H{
			"user_id": id,
			"message": "User details retrieved",
		})
	})

	// Start the server
	log.Println("Gin server starting on http://localhost:3000")
	if err := r.Run(":3000"); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}