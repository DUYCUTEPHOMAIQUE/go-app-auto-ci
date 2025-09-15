package handler

import (
	"net/http"

	"go-app-auto-ci/internal/model"
	"go-app-auto-ci/internal/service"

	"github.com/gin-gonic/gin"
)

func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, model.HealthResponse{
		Status: service.GetHealthStatus(),
	})
}
