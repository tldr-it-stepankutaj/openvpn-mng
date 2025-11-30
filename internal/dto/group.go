package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/models"
)

// CreateGroupRequest represents a request to create a new group
type CreateGroupRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Description string `json:"description,omitempty" binding:"max=500"`
}

// UpdateGroupRequest represents a request to update a group
type UpdateGroupRequest struct {
	Name        string `json:"name,omitempty" binding:"omitempty,min=2,max=100"`
	Description string `json:"description,omitempty" binding:"max=500"`
}

// GroupResponse represents a group in API responses
type GroupResponse struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
	CreatedBy   uuid.UUID  `json:"created_by"`
	UpdatedBy   *uuid.UUID `json:"updated_by,omitempty"`
}

// GroupListResponse represents a paginated list of groups
type GroupListResponse struct {
	Groups     []GroupResponse `json:"groups"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}

// AddUserToGroupRequest represents a request to add a user to a group
type AddUserToGroupRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
}

// UserGroupResponse represents a user-group relationship in API responses
type UserGroupResponse struct {
	UserID    uuid.UUID `json:"user_id"`
	GroupID   uuid.UUID `json:"group_id"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy uuid.UUID `json:"created_by"`
}

// ToGroupResponse converts a Group model to GroupResponse DTO
func ToGroupResponse(group *models.Group) *GroupResponse {
	if group == nil {
		return nil
	}

	return &GroupResponse{
		ID:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		CreatedAt:   group.CreatedAt,
		UpdatedAt:   group.UpdatedAt,
		CreatedBy:   group.CreatedBy,
		UpdatedBy:   group.UpdatedBy,
	}
}

// ToGroupResponseList converts a slice of Group models to GroupResponse DTOs
func ToGroupResponseList(groups []models.Group) []GroupResponse {
	responses := make([]GroupResponse, len(groups))
	for i, group := range groups {
		responses[i] = *ToGroupResponse(&group)
	}
	return responses
}
