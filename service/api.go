package service

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Set CORS permissions
func SetupCORS(r *gin.Engine) {
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		MaxAge: 12 * time.Hour,
	}))
}

// Setup playground routes
func SetupRoutes(r *gin.Engine) {
	r.POST("/", func(c *gin.Context) {
		// Request struct for body unmarshaling
		type request struct{ Src []string }
		// Unmarshal body
		var body request
		if err := c.Bind(&body); err != nil {
			c.Error(err)
			return
		}
	})
}
