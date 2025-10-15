package service

import (
	"errors"
	"time"

	"go-app-auto-ci/internal/model"
)

// In-memory storage for demo purposes
var users []model.User
var nextUserID = 1

// CreateUser creates a new user
func CreateUser(req model.CreateUserRequest) (*model.User, error) {
	// Validate email uniqueness
	for _, user := range users {
		if user.Email == req.Email {
			return nil, errors.New("email already exists")
		}
		if user.Username == req.Username {
			return nil, errors.New("username already exists")
		}
	}

	// Create new user
	now := time.Now()
	user := model.User{
		ID:        nextUserID,
		Username:  req.Username,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Age:       req.Age,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Add to storage
	users = append(users, user)
	nextUserID++

	return &user, nil
}

// GetUserByID retrieves a user by ID
func GetUserByID(id int) (*model.User, error) {
	for _, user := range users {
		if user.ID == id {
			return &user, nil
		}
	}
	return nil, errors.New("user not found")
}

// GetAllUsers retrieves all users
func GetAllUsers() []model.User {
	return users
}

// DeleteUser deletes a user by ID
func DeleteUser(id int) error {
	for i, user := range users {
		if user.ID == id {
			users = append(users[:i], users[i+1:]...)
			return nil
		}
	}
	return errors.New("user not found")
}

// ResetUsers resets the users slice and nextUserID for testing purposes
func ResetUsers() {
	users = []model.User{}
	nextUserID = 1
}
