package main

import (
	"github.com/gin-gonic/gin"
	"github.com/ternary-club/backend/service"
)

// Initialization
func main() {
	// New Gin instance
	r := gin.Default()

	// Setup API
	service.SetupCORS(r)
	service.SetupRoutes(r)

	// Update binaries
	service.Update()

	// Start engine
	r.Run()
}
