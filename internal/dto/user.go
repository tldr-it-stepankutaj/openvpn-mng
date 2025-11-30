package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/models"
)

// CreateUserRequest represents a request to create a new user
type CreateUserRequest struct {
	Username   string      `json:"username" binding:"required,min=3,max=100"`
	Password   string      `json:"password" binding:"required,min=8"`
	ManagerID  *uuid.UUID  `json:"manager_id,omitempty"`
	FirstName  string      `json:"first_name" binding:"required,max=100"`
	MiddleName string      `json:"middle_name,omitempty" binding:"max=100"`
	LastName   string      `json:"last_name" binding:"required,max=100"`
	Email      string      `json:"email" binding:"required,email,max=255"`
	Telephone  string      `json:"telephone,omitempty" binding:"max=50"`
	Role       models.Role `json:"role" binding:"required,oneof=USER MANAGER ADMIN"`
}

// UpdateUserRequest represents a request to update a user
type UpdateUserRequest struct {
	ManagerID  *uuid.UUID  `json:"manager_id,omitempty"`
	FirstName  string      `json:"first_name,omitempty" binding:"max=100"`
	MiddleName string      `json:"middle_name,omitempty" binding:"max=100"`
	LastName   string      `json:"last_name,omitempty" binding:"max=100"`
	Email      string      `json:"email,omitempty" binding:"omitempty,email,max=255"`
	Telephone  string      `json:"telephone,omitempty" binding:"max=50"`
	Role       models.Role `json:"role,omitempty" binding:"omitempty,oneof=USER MANAGER ADMIN"`
}

// UpdatePasswordRequest represents a request to update user password
type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// UpdateProfileRequest represents a request to update own profile (for USER role)
type UpdateProfileRequest struct {
	FirstName  string `json:"first_name,omitempty" binding:"max=100"`
	MiddleName string `json:"middle_name,omitempty" binding:"max=100"`
	LastName   string `json:"last_name,omitempty" binding:"max=100"`
	Email      string `json:"email,omitempty" binding:"omitempty,email,max=255"`
	Telephone  string `json:"telephone,omitempty" binding:"max=50"`
}

// UserResponse represents a user in API responses
type UserResponse struct {
	ID         uuid.UUID     `json:"id"`
	Username   string        `json:"username"`
	ManagerID  *uuid.UUID    `json:"manager_id,omitempty"`
	Manager    *UserResponse `json:"manager,omitempty"`
	FirstName  string        `json:"first_name"`
	MiddleName string        `json:"middle_name,omitempty"`
	LastName   string        `json:"last_name"`
	Email      string        `json:"email"`
	Telephone  string        `json:"telephone,omitempty"`
	Role       models.Role   `json:"role"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  *time.Time    `json:"updated_at,omitempty"`
	CreatedBy  uuid.UUID     `json:"created_by"`
	UpdatedBy  *uuid.UUID    `json:"updated_by,omitempty"`
}

// UserListResponse represents a paginated list of users
type UserListResponse struct {
	Users      []UserResponse `json:"users"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

// ToUserResponse converts a User model to UserResponse DTO
func ToUserResponse(user *models.User) *UserResponse {
	if user == nil {
		return nil
	}

	response := &UserResponse{
		ID:         user.ID,
		Username:   user.Username,
		ManagerID:  user.ManagerID,
		FirstName:  user.FirstName,
		MiddleName: user.MiddleName,
		LastName:   user.LastName,
		Email:      user.Email,
		Telephone:  user.Telephone,
		Role:       user.Role,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
		CreatedBy:  user.CreatedBy,
		UpdatedBy:  user.UpdatedBy,
	}

	if user.Manager != nil {
		response.Manager = ToUserResponse(user.Manager)
	}

	return response
}

// ToUserResponseList converts a slice of User models to UserResponse DTOs
func ToUserResponseList(users []models.User) []UserResponse {
	responses := make([]UserResponse, len(users))
	for i, user := range users {
		responses[i] = *ToUserResponse(&user)
	}
	return responses
}
