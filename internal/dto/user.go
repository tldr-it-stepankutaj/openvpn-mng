package dto

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/models"
)

// DateOnly is a custom type that can parse both date-only (2006-01-02) and full datetime formats
type DateOnly struct {
	time.Time
}

// UnmarshalJSON implements json.Unmarshaler for DateOnly
func (d *DateOnly) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), "\"")
	if s == "" || s == "null" {
		return nil
	}

	// Try date-only format first
	t, err := time.Parse("2006-01-02", s)
	if err == nil {
		d.Time = t
		return nil
	}

	// Try full datetime format
	t, err = time.Parse(time.RFC3339, s)
	if err == nil {
		d.Time = t
		return nil
	}

	// Try datetime without timezone
	t, err = time.Parse("2006-01-02T15:04:05", s)
	if err == nil {
		d.Time = t
		return nil
	}

	return err
}

// ToTimePtr converts DateOnly to *time.Time
func (d *DateOnly) ToTimePtr() *time.Time {
	if d == nil || d.Time.IsZero() {
		return nil
	}
	return &d.Time
}

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
	IsActive   *bool       `json:"is_active,omitempty"`
	ValidFrom  *DateOnly   `json:"valid_from,omitempty"`
	ValidTo    *DateOnly   `json:"valid_to,omitempty"`
	VpnIP      string      `json:"vpn_ip,omitempty" binding:"max=45"`
}

// UpdateUserRequest represents a request to update a user
type UpdateUserRequest struct {
	Username   string      `json:"username,omitempty" binding:"omitempty,min=3,max=100"`
	Password   string      `json:"password,omitempty" binding:"omitempty,min=8"`
	ManagerID  *uuid.UUID  `json:"manager_id,omitempty"`
	FirstName  string      `json:"first_name,omitempty" binding:"max=100"`
	MiddleName string      `json:"middle_name,omitempty" binding:"max=100"`
	LastName   string      `json:"last_name,omitempty" binding:"max=100"`
	Email      string      `json:"email,omitempty" binding:"omitempty,email,max=255"`
	Telephone  string      `json:"telephone,omitempty" binding:"max=50"`
	Role       models.Role `json:"role,omitempty" binding:"omitempty,oneof=USER MANAGER ADMIN"`
	IsActive   *bool       `json:"is_active,omitempty"`
	ValidFrom  *DateOnly   `json:"valid_from,omitempty"`
	ValidTo    *DateOnly   `json:"valid_to,omitempty"`
	VpnIP      *string     `json:"vpn_ip,omitempty" binding:"omitempty,max=45"`
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
	IsActive   bool          `json:"is_active"`
	ValidFrom  *time.Time    `json:"valid_from,omitempty"`
	ValidTo    *time.Time    `json:"valid_to,omitempty"`
	VpnIP      string        `json:"vpn_ip,omitempty"`
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
		IsActive:   user.IsActive,
		ValidFrom:  user.ValidFrom,
		ValidTo:    user.ValidTo,
		VpnIP:      user.VpnIP,
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

// AddUserGroupRequest represents a request to add a user to a group
type AddUserGroupRequest struct {
	GroupID uuid.UUID `json:"group_id" binding:"required"`
}

// UserGroupsResponse represents a user's groups with their networks
type UserGroupsResponse struct {
	Groups []UserGroupWithNetworks `json:"groups"`
}

// UserGroupWithNetworks represents a group with its networks for user view
type UserGroupWithNetworks struct {
	ID          uuid.UUID          `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Networks    []NetworkBasicInfo `json:"networks"`
}

// NetworkBasicInfo represents basic network info for user view
type NetworkBasicInfo struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	CIDR        string    `json:"cidr"`
	Description string    `json:"description,omitempty"`
}
