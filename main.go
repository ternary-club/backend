package main

import "github.com/gin-gonic/gin"

// Initialization
func main() {
	// New Gin instance
	r := gin.Default()

	// Setup API
	SetupCORS(r)
	SetupRoutes(r)

	// Start engine
	r.Run()
}
