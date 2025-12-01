package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/models"
)

// CreateVpnSessionRequest represents a request to create a new VPN session
type CreateVpnSessionRequest struct {
	UserID      uuid.UUID `json:"user_id" binding:"required"`
	VpnIP       string    `json:"vpn_ip" binding:"required,max=45"`
	ClientIP    string    `json:"client_ip,omitempty" binding:"max=45"`
	ConnectedAt time.Time `json:"connected_at" binding:"required"`
}

// UpdateVpnSessionRequest represents a request to update a VPN session (disconnect)
type UpdateVpnSessionRequest struct {
	DisconnectedAt   time.Time                `json:"disconnected_at" binding:"required"`
	BytesReceived    int64                    `json:"bytes_received"`
	BytesSent        int64                    `json:"bytes_sent"`
	DisconnectReason *models.DisconnectReason `json:"disconnect_reason,omitempty"`
}

// CreateVpnTrafficStatsRequest represents a request to create traffic stats
type CreateVpnTrafficStatsRequest struct {
	SessionID          uuid.UUID `json:"session_id" binding:"required"`
	Timestamp          time.Time `json:"timestamp" binding:"required"`
	BytesReceivedDelta int64     `json:"bytes_received_delta"`
	BytesSentDelta     int64     `json:"bytes_sent_delta"`
}

// VpnSessionResponse represents a VPN session in API responses
type VpnSessionResponse struct {
	ID               uuid.UUID                `json:"id"`
	UserID           uuid.UUID                `json:"user_id"`
	User             *UserResponse            `json:"user,omitempty"`
	VpnIP            string                   `json:"vpn_ip"`
	ClientIP         string                   `json:"client_ip,omitempty"`
	ConnectedAt      time.Time                `json:"connected_at"`
	DisconnectedAt   *time.Time               `json:"disconnected_at,omitempty"`
	BytesReceived    int64                    `json:"bytes_received"`
	BytesSent        int64                    `json:"bytes_sent"`
	TotalBytes       int64                    `json:"total_bytes"`
	DisconnectReason *models.DisconnectReason `json:"disconnect_reason,omitempty"`
	Duration         string                   `json:"duration"`
	IsActive         bool                     `json:"is_active"`
}

// VpnSessionListResponse represents a paginated list of VPN sessions
type VpnSessionListResponse struct {
	Sessions   []VpnSessionResponse `json:"sessions"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	TotalPages int                  `json:"total_pages"`
}

// VpnSessionFilter represents filters for VPN session queries
type VpnSessionFilter struct {
	UserID    *uuid.UUID `form:"user_id"`
	VpnIP     string     `form:"vpn_ip"`
	IsActive  *bool      `form:"is_active"`
	StartDate *time.Time `form:"start_date"`
	EndDate   *time.Time `form:"end_date"`
	Page      int        `form:"page,default=1"`
	PageSize  int        `form:"page_size,default=20"`
}

// VpnTrafficStatsResponse represents traffic stats in API responses
type VpnTrafficStatsResponse struct {
	ID                 uuid.UUID `json:"id"`
	SessionID          uuid.UUID `json:"session_id"`
	Timestamp          time.Time `json:"timestamp"`
	BytesReceivedDelta int64     `json:"bytes_received_delta"`
	BytesSentDelta     int64     `json:"bytes_sent_delta"`
	TotalBytesDelta    int64     `json:"total_bytes_delta"`
}

// VpnTrafficStatsListResponse represents a paginated list of traffic stats
type VpnTrafficStatsListResponse struct {
	Stats      []VpnTrafficStatsResponse `json:"stats"`
	Total      int64                     `json:"total"`
	Page       int                       `json:"page"`
	PageSize   int                       `json:"page_size"`
	TotalPages int                       `json:"total_pages"`
}

// VpnTrafficStatsFilter represents filters for traffic stats queries
type VpnTrafficStatsFilter struct {
	SessionID *uuid.UUID `form:"session_id"`
	StartDate *time.Time `form:"start_date"`
	EndDate   *time.Time `form:"end_date"`
	Page      int        `form:"page,default=1"`
	PageSize  int        `form:"page_size,default=20"`
}

// VpnUsageStats represents aggregated usage statistics
type VpnUsageStats struct {
	TotalSessions      int64 `json:"total_sessions"`
	ActiveSessions     int64 `json:"active_sessions"`
	TotalBytesReceived int64 `json:"total_bytes_received"`
	TotalBytesSent     int64 `json:"total_bytes_sent"`
	TotalBytes         int64 `json:"total_bytes"`
}

// UserVpnUsageResponse represents VPN usage for a specific user
type UserVpnUsageResponse struct {
	UserID             uuid.UUID  `json:"user_id"`
	Username           string     `json:"username"`
	TotalSessions      int64      `json:"total_sessions"`
	TotalBytesReceived int64      `json:"total_bytes_received"`
	TotalBytesSent     int64      `json:"total_bytes_sent"`
	TotalBytes         int64      `json:"total_bytes"`
	LastConnectedAt    *time.Time `json:"last_connected_at,omitempty"`
}

// ToVpnSessionResponse converts a VpnSession model to VpnSessionResponse DTO
func ToVpnSessionResponse(session *models.VpnSession) *VpnSessionResponse {
	if session == nil {
		return nil
	}

	response := &VpnSessionResponse{
		ID:               session.ID,
		UserID:           session.UserID,
		VpnIP:            session.VpnIP,
		ClientIP:         session.ClientIP,
		ConnectedAt:      session.ConnectedAt,
		DisconnectedAt:   session.DisconnectedAt,
		BytesReceived:    session.BytesReceived,
		BytesSent:        session.BytesSent,
		TotalBytes:       session.TotalBytes(),
		DisconnectReason: session.DisconnectReason,
		Duration:         session.Duration().String(),
		IsActive:         session.IsActive(),
	}

	if session.User != nil {
		response.User = ToUserResponse(session.User)
	}

	return response
}

// ToVpnSessionResponseList converts a slice of VpnSession models to VpnSessionResponse DTOs
func ToVpnSessionResponseList(sessions []models.VpnSession) []VpnSessionResponse {
	responses := make([]VpnSessionResponse, len(sessions))
	for i, session := range sessions {
		responses[i] = *ToVpnSessionResponse(&session)
	}
	return responses
}

// ToVpnTrafficStatsResponse converts a VpnTrafficStats model to VpnTrafficStatsResponse DTO
func ToVpnTrafficStatsResponse(stats *models.VpnTrafficStats) *VpnTrafficStatsResponse {
	if stats == nil {
		return nil
	}

	return &VpnTrafficStatsResponse{
		ID:                 stats.ID,
		SessionID:          stats.SessionID,
		Timestamp:          stats.Timestamp,
		BytesReceivedDelta: stats.BytesReceivedDelta,
		BytesSentDelta:     stats.BytesSentDelta,
		TotalBytesDelta:    stats.TotalBytesDelta(),
	}
}

// ToVpnTrafficStatsResponseList converts a slice of VpnTrafficStats models to VpnTrafficStatsResponse DTOs
func ToVpnTrafficStatsResponseList(stats []models.VpnTrafficStats) []VpnTrafficStatsResponse {
	responses := make([]VpnTrafficStatsResponse, len(stats))
	for i, s := range stats {
		responses[i] = *ToVpnTrafficStatsResponse(&s)
	}
	return responses
}
