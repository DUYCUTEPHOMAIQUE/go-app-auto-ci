package service

import (
	"testing"

	"go-app-auto-ci/internal/model"
)

func TestCreateUser(t *testing.T) {
	// Reset global state before each test
	users = []model.User{}
	nextUserID = 1

	tests := []struct {
		name    string
		req     model.CreateUserRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid user creation",
			req: model.CreateUserRequest{
				Username:  "testuser",
				Email:     "test@example.com",
				FirstName: "Test",
				LastName:  "User",
				Age:       25,
			},
			wantErr: false,
		},
		{
			name: "duplicate email",
			req: model.CreateUserRequest{
				Username:  "testuser2",
				Email:     "test@example.com", // Same email as first test
				FirstName: "Test2",
				LastName:  "User2",
				Age:       30,
			},
			wantErr: true,
			errMsg:  "email already exists",
		},
		{
			name: "duplicate username",
			req: model.CreateUserRequest{
				Username:  "testuser", // Same username as first test
				Email:     "test2@example.com",
				FirstName: "Test3",
				LastName:  "User3",
				Age:       35,
			},
			wantErr: true,
			errMsg:  "username already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := CreateUser(tt.req)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CreateUser() expected error but got none")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("CreateUser() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("CreateUser() error = %v, want nil", err)
					return
				}
				if user == nil {
					t.Errorf("CreateUser() returned nil user")
					return
				}
				if user.Username != tt.req.Username {
					t.Errorf("CreateUser() username = %v, want %v", user.Username, tt.req.Username)
				}
				if user.Email != tt.req.Email {
					t.Errorf("CreateUser() email = %v, want %v", user.Email, tt.req.Email)
				}
				if user.ID == 0 {
					t.Errorf("CreateUser() ID should not be 0")
				}
			}
		})
	}
}

func TestGetUserByID(t *testing.T) {
	// Reset global state
	users = []model.User{}
	nextUserID = 1

	// Create a test user
	req := model.CreateUserRequest{
		Username:  "testuser",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Age:       25,
	}
	createdUser, err := CreateUser(req)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	tests := []struct {
		name    string
		id      int
		wantErr bool
	}{
		{
			name:    "existing user",
			id:      createdUser.ID,
			wantErr: false,
		},
		{
			name:    "non-existing user",
			id:      999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := GetUserByID(tt.id)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetUserByID() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("GetUserByID() error = %v, want nil", err)
					return
				}
				if user.ID != tt.id {
					t.Errorf("GetUserByID() ID = %v, want %v", user.ID, tt.id)
				}
			}
		})
	}
}

func TestGetAllUsers(t *testing.T) {
	// Reset global state
	users = []model.User{}
	nextUserID = 1

	// Initially should be empty
	allUsers := GetAllUsers()
	if len(allUsers) != 0 {
		t.Errorf("GetAllUsers() length = %v, want 0", len(allUsers))
	}

	// Create some test users
	req1 := model.CreateUserRequest{
		Username:  "user1",
		Email:     "user1@example.com",
		FirstName: "User",
		LastName:  "One",
		Age:       25,
	}
	req2 := model.CreateUserRequest{
		Username:  "user2",
		Email:     "user2@example.com",
		FirstName: "User",
		LastName:  "Two",
		Age:       30,
	}

	_, err := CreateUser(req1)
	if err != nil {
		t.Fatalf("Failed to create test user 1: %v", err)
	}
	_, err = CreateUser(req2)
	if err != nil {
		t.Fatalf("Failed to create test user 2: %v", err)
	}

	// Now should have 2 users
	allUsers = GetAllUsers()
	if len(allUsers) != 2 {
		t.Errorf("GetAllUsers() length = %v, want 2", len(allUsers))
	}
}

func TestDeleteUser(t *testing.T) {
	// Reset global state
	users = []model.User{}
	nextUserID = 1

	// Create a test user
	req := model.CreateUserRequest{
		Username:  "testuser",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Age:       25,
	}
	createdUser, err := CreateUser(req)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Verify user exists
	allUsers := GetAllUsers()
	if len(allUsers) != 1 {
		t.Fatalf("Expected 1 user, got %v", len(allUsers))
	}

	// Delete the user
	err = DeleteUser(createdUser.ID)
	if err != nil {
		t.Errorf("DeleteUser() error = %v, want nil", err)
	}

	// Verify user is deleted
	allUsers = GetAllUsers()
	if len(allUsers) != 0 {
		t.Errorf("GetAllUsers() length = %v, want 0", len(allUsers))
	}

	// Try to delete non-existing user
	err = DeleteUser(999)
	if err == nil {
		t.Errorf("DeleteUser() expected error but got none")
	}
}
