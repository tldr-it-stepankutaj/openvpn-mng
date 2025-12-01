package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/models"
)

// CreateNetworkRequest represents the request to create a network
type CreateNetworkRequest struct {
	Name        string `json:"name" binding:"required" example:"Office Network"`
	CIDR        string `json:"cidr" binding:"required" example:"192.168.1.0/24"`
	Description string `json:"description" example:"Main office network segment"`
}

// UpdateNetworkRequest represents the request to update a network
type UpdateNetworkRequest struct {
	Name        string `json:"name" example:"Office Network Updated"`
	CIDR        string `json:"cidr" example:"192.168.2.0/24"`
	Description string `json:"description" example:"Updated description"`
}

// NetworkResponse represents the network response
type NetworkResponse struct {
	ID          uuid.UUID       `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name        string          `json:"name" example:"Office Network"`
	CIDR        string          `json:"cidr" example:"192.168.1.0/24"`
	Description string          `json:"description" example:"Main office network segment"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	CreatedBy   uuid.UUID       `json:"created_by"`
	UpdatedBy   *uuid.UUID      `json:"updated_by,omitempty"`
	Groups      []GroupResponse `json:"groups,omitempty"`
}

// NetworkListResponse represents the paginated network list response
type NetworkListResponse struct {
	Networks   []NetworkResponse `json:"networks"`
	Total      int64             `json:"total" example:"100"`
	Page       int               `json:"page" example:"1"`
	PageSize   int               `json:"page_size" example:"20"`
	TotalPages int               `json:"total_pages" example:"5"`
}

// AddNetworkToGroupRequest represents the request to add a network to a group
type AddNetworkToGroupRequest struct {
	NetworkID uuid.UUID `json:"network_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// AddGroupToNetworkRequest represents the request to add a group to a network
type AddGroupToNetworkRequest struct {
	GroupID uuid.UUID `json:"group_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// ToNetworkResponse converts a Network model to NetworkResponse
func ToNetworkResponse(network *models.Network) NetworkResponse {
	response := NetworkResponse{
		ID:          network.ID,
		Name:        network.Name,
		CIDR:        network.CIDR,
		Description: network.Description,
		CreatedAt:   network.CreatedAt,
		UpdatedAt:   network.UpdatedAt,
		CreatedBy:   network.CreatedBy,
		UpdatedBy:   network.UpdatedBy,
	}

	if len(network.Groups) > 0 {
		response.Groups = make([]GroupResponse, len(network.Groups))
		for i, g := range network.Groups {
			response.Groups[i] = *ToGroupResponse(&g)
		}
	}

	return response
}

// ToNetworkResponseList converts a slice of Network models to NetworkResponse slice
func ToNetworkResponseList(networks []models.Network) []NetworkResponse {
	responses := make([]NetworkResponse, len(networks))
	for i, n := range networks {
		responses[i] = ToNetworkResponse(&n)
	}
	return responses
}
