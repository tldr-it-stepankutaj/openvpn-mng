package dto

import (
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/models"
)

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token     string        `json:"token"`
	ExpiresIn int           `json:"expires_in"` // in seconds
	User      *UserResponse `json:"user"`
}

// AuthUser represents the authenticated user context
type AuthUser struct {
	ID       string      `json:"id"`
	Username string      `json:"username"`
	Role     models.Role `json:"role"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code,omitempty"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
