package main

import (
	"go-app-auto-ci/internal/handler"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.GET("/health", handler.HealthHandler)
	r.Run(":8080")
}
