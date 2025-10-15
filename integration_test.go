package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"go-app-auto-ci/internal/model"
	"go-app-auto-ci/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupIntegrationTest() *gin.Engine {
	gin.SetMode(gin.TestMode)

	// Reset service state
	service.ResetUsers()

	r := gin.Default()

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// User endpoints
	users := r.Group("/users")
	{
		users.POST("", func(c *gin.Context) {
			var req model.CreateUserRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, model.ErrorResponse{
					Error:   "validation_error",
					Message: err.Error(),
				})
				return
			}

			user, err := service.CreateUser(req)
			if err != nil {
				c.JSON(http.StatusConflict, model.ErrorResponse{
					Error:   "creation_error",
					Message: err.Error(),
				})
				return
			}

			c.JSON(http.StatusCreated, model.CreateUserResponse{
				Message: "User created successfully",
				User:    *user,
			})
		})

		users.GET("", func(c *gin.Context) {
			users := service.GetAllUsers()
			c.JSON(http.StatusOK, gin.H{
				"users": users,
				"count": len(users),
			})
		})

		users.GET("/:id", func(c *gin.Context) {
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
		})

		users.DELETE("/:id", func(c *gin.Context) {
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
		})
	}

	return r
}

func TestUserAPIIntegration(t *testing.T) {
	router := setupIntegrationTest()

	t.Run("Create and retrieve user", func(t *testing.T) {
		// Create user
		createReq := model.CreateUserRequest{
			Username:  "integrationuser",
			Email:     "integration@example.com",
			FirstName: "Integration",
			LastName:  "Test",
			Age:       28,
		}

		jsonPayload, _ := json.Marshal(createReq)
		req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonPayload))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var createResp model.CreateUserResponse
		err := json.Unmarshal(w.Body.Bytes(), &createResp)
		assert.NoError(t, err)
		assert.Equal(t, "User created successfully", createResp.Message)
		assert.Equal(t, createReq.Username, createResp.User.Username)
		assert.Equal(t, createReq.Email, createResp.User.Email)
		assert.NotZero(t, createResp.User.ID)

		// Retrieve user by ID
		req, _ = http.NewRequest("GET", "/users/1", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var user model.User
		err = json.Unmarshal(w.Body.Bytes(), &user)
		assert.NoError(t, err)
		assert.Equal(t, createReq.Username, user.Username)
		assert.Equal(t, createReq.Email, user.Email)
		assert.Equal(t, createReq.FirstName, user.FirstName)
		assert.Equal(t, createReq.LastName, user.LastName)
		assert.Equal(t, createReq.Age, user.Age)
	})

	t.Run("Get all users", func(t *testing.T) {
		// Create another user
		createReq := model.CreateUserRequest{
			Username:  "user2",
			Email:     "user2@example.com",
			FirstName: "User",
			LastName:  "Two",
			Age:       30,
		}

		jsonPayload, _ := json.Marshal(createReq)
		req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonPayload))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		// Get all users
		req, _ = http.NewRequest("GET", "/users", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(2), response["count"])

		users := response["users"].([]interface{})
		assert.Len(t, users, 2)
	})

	t.Run("Delete user", func(t *testing.T) {
		// Delete user with ID 1
		req, _ := http.NewRequest("DELETE", "/users/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "User deleted successfully", response["message"])

		// Verify user is deleted
		req, _ = http.NewRequest("GET", "/users/1", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		// Verify count is reduced
		req, _ = http.NewRequest("GET", "/users", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(1), response["count"])
	})

	t.Run("Health check", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "ok", response["status"])
	})
}

func TestUserAPIErrorCases(t *testing.T) {
	router := setupIntegrationTest()

	t.Run("Create user with invalid data", func(t *testing.T) {
		// Invalid email
		createReq := model.CreateUserRequest{
			Username:  "testuser",
			Email:     "invalid-email",
			FirstName: "Test",
			LastName:  "User",
			Age:       25,
		}

		jsonPayload, _ := json.Marshal(createReq)
		req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonPayload))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errorResp model.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "validation_error", errorResp.Error)
	})

	t.Run("Get non-existing user", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/users/999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var errorResp model.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "user_not_found", errorResp.Error)
	})

	t.Run("Delete non-existing user", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/users/999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var errorResp model.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.Equal(t, "user_not_found", errorResp.Error)
	})
}
