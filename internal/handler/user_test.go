package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-app-auto-ci/internal/model"
	"go-app-auto-ci/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	users := r.Group("/users")
	{
		users.POST("", CreateUserHandler)
		users.GET("", GetAllUsersHandler)
		users.GET("/:id", GetUserHandler)
		users.DELETE("/:id", DeleteUserHandler)
	}

	return r
}

func resetServiceState() {
	// Reset service state for clean tests
	service.ResetUsers()
}

func TestCreateUserHandler(t *testing.T) {
	resetServiceState()
	router := setupTestRouter()

	tests := []struct {
		name           string
		payload        interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "valid user creation",
			payload: model.CreateUserRequest{
				Username:  "testuser",
				Email:     "test@example.com",
				FirstName: "Test",
				LastName:  "User",
				Age:       25,
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid email format",
			payload: model.CreateUserRequest{
				Username:  "testuser2",
				Email:     "invalid-email",
				FirstName: "Test",
				LastName:  "User",
				Age:       25,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "validation_error",
		},
		{
			name: "missing required fields",
			payload: model.CreateUserRequest{
				Username: "testuser3",
				// Missing other required fields
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "validation_error",
		},
		{
			name: "duplicate email",
			payload: model.CreateUserRequest{
				Username:  "testuser4",
				Email:     "test@example.com", // Same as first test
				FirstName: "Test",
				LastName:  "User",
				Age:       25,
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "creation_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonPayload, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonPayload))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var errorResp model.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &errorResp)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, errorResp.Error)
			} else if tt.expectedStatus == http.StatusCreated {
				var resp model.CreateUserResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "User created successfully", resp.Message)
				assert.NotZero(t, resp.User.ID)
				assert.Equal(t, tt.payload.(model.CreateUserRequest).Username, resp.User.Username)
			}
		})
	}
}

func TestGetUserHandler(t *testing.T) {
	resetServiceState()
	router := setupTestRouter()

	// Create a test user first
	createReq := model.CreateUserRequest{
		Username:  "testuser",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Age:       25,
	}
	_, err := service.CreateUser(createReq)
	assert.NoError(t, err)

	tests := []struct {
		name           string
		userID         string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "existing user",
			userID:         "1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "non-existing user",
			userID:         "999",
			expectedStatus: http.StatusNotFound,
			expectedError:  "user_not_found",
		},
		{
			name:           "invalid user ID",
			userID:         "invalid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/users/"+tt.userID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var errorResp model.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &errorResp)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, errorResp.Error)
			} else if tt.expectedStatus == http.StatusOK {
				var user model.User
				err := json.Unmarshal(w.Body.Bytes(), &user)
				assert.NoError(t, err)
				assert.Equal(t, user.ID, user.ID)
			}
		})
	}
}

func TestGetAllUsersHandler(t *testing.T) {
	resetServiceState()
	router := setupTestRouter()

	// Initially should be empty
	req, _ := http.NewRequest("GET", "/users", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(0), response["count"])

	// Create some test users
	createReq1 := model.CreateUserRequest{
		Username:  "user1",
		Email:     "user1@example.com",
		FirstName: "User",
		LastName:  "One",
		Age:       25,
	}
	createReq2 := model.CreateUserRequest{
		Username:  "user2",
		Email:     "user2@example.com",
		FirstName: "User",
		LastName:  "Two",
		Age:       30,
	}

	_, err = service.CreateUser(createReq1)
	assert.NoError(t, err)
	_, err = service.CreateUser(createReq2)
	assert.NoError(t, err)

	// Now should have 2 users
	req, _ = http.NewRequest("GET", "/users", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(2), response["count"])
}

func TestDeleteUserHandler(t *testing.T) {
	resetServiceState()
	router := setupTestRouter()

	// Create a test user
	createReq := model.CreateUserRequest{
		Username:  "testuser",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Age:       25,
	}
	_, err := service.CreateUser(createReq)
	assert.NoError(t, err)

	tests := []struct {
		name           string
		userID         string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "existing user",
			userID:         "1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "non-existing user",
			userID:         "999",
			expectedStatus: http.StatusNotFound,
			expectedError:  "user_not_found",
		},
		{
			name:           "invalid user ID",
			userID:         "invalid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("DELETE", "/users/"+tt.userID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var errorResp model.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &errorResp)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, errorResp.Error)
			} else if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "User deleted successfully", response["message"])
			}
		})
	}
}
