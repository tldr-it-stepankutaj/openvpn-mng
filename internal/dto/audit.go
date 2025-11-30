package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/models"
)

// AuditLogResponse represents an audit log entry in API responses
type AuditLogResponse struct {
	ID         uuid.UUID          `json:"id"`
	UserID     uuid.UUID          `json:"user_id"`
	User       *UserResponse      `json:"user,omitempty"`
	Action     models.AuditAction `json:"action"`
	EntityType string             `json:"entity_type"`
	EntityID   *uuid.UUID         `json:"entity_id,omitempty"`
	OldValues  string             `json:"old_values,omitempty"`
	NewValues  string             `json:"new_values,omitempty"`
	IPAddress  string             `json:"ip_address,omitempty"`
	UserAgent  string             `json:"user_agent,omitempty"`
	Details    string             `json:"details,omitempty"`
	CreatedAt  time.Time          `json:"created_at"`
}

// AuditLogListResponse represents a paginated list of audit logs
type AuditLogListResponse struct {
	Logs       []AuditLogResponse `json:"logs"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}

// AuditLogFilter represents filters for audit log queries
type AuditLogFilter struct {
	UserID     *uuid.UUID          `form:"user_id"`
	Action     *models.AuditAction `form:"action"`
	EntityType string              `form:"entity_type"`
	EntityID   *uuid.UUID          `form:"entity_id"`
	StartDate  *time.Time          `form:"start_date"`
	EndDate    *time.Time          `form:"end_date"`
	Page       int                 `form:"page,default=1"`
	PageSize   int                 `form:"page_size,default=20"`
}

// ToAuditLogResponse converts an AuditLog model to AuditLogResponse DTO
func ToAuditLogResponse(log *models.AuditLog) *AuditLogResponse {
	if log == nil {
		return nil
	}

	response := &AuditLogResponse{
		ID:         log.ID,
		UserID:     log.UserID,
		Action:     log.Action,
		EntityType: log.EntityType,
		EntityID:   log.EntityID,
		OldValues:  log.OldValues,
		NewValues:  log.NewValues,
		IPAddress:  log.IPAddress,
		UserAgent:  log.UserAgent,
		Details:    log.Details,
		CreatedAt:  log.CreatedAt,
	}

	if log.User != nil {
		response.User = ToUserResponse(log.User)
	}

	return response
}

// ToAuditLogResponseList converts a slice of AuditLog models to AuditLogResponse DTOs
func ToAuditLogResponseList(logs []models.AuditLog) []AuditLogResponse {
	responses := make([]AuditLogResponse, len(logs))
	for i, log := range logs {
		responses[i] = *ToAuditLogResponse(&log)
	}
	return responses
}
