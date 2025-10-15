package main

import (
	"go-app-auto-ci/internal/handler"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Health check endpoint
	r.GET("/health", handler.HealthHandler)

	// User endpoints
	users := r.Group("/users")
	{
		users.POST("", handler.CreateUserHandler)
		users.GET("", handler.GetAllUsersHandler)
		users.GET("/:id", handler.GetUserHandler)
		users.DELETE("/:id", handler.DeleteUserHandler)
	}

	r.Run(":8080")
}
