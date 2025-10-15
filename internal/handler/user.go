package handler

import (
	"net/http"
	"strconv"

	"go-app-auto-ci/internal/model"
	"go-app-auto-ci/internal/service"

	"github.com/gin-gonic/gin"
)

// CreateUserHandler handles POST /users
func CreateUserHandler(c *gin.Context) {
	var req model.CreateUserRequest

	// Bind JSON request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	// Create user
	user, err := service.CreateUser(req)
	if err != nil {
		c.JSON(http.StatusConflict, model.ErrorResponse{
			Error:   "creation_error",
			Message: err.Error(),
		})
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, model.CreateUserResponse{
		Message: "User created successfully",
		User:    *user,
	})
}

// GetUserHandler handles GET /users/:id
func GetUserHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid user ID",
		})
		return
	}

	user, err := service.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{
			Error:   "user_not_found",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetAllUsersHandler handles GET /users
func GetAllUsersHandler(c *gin.Context) {
	users := service.GetAllUsers()
	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"count": len(users),
	})
}

// DeleteUserHandler handles DELETE /users/:id
func DeleteUserHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid user ID",
		})
		return
	}

	err = service.DeleteUser(id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{
			Error:   "user_not_found",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}
