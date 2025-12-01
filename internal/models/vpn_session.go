package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DisconnectReason represents the reason for VPN disconnection
type DisconnectReason string

const (
	DisconnectReasonUserRequest    DisconnectReason = "USER_REQUEST"
	DisconnectReasonTimeout        DisconnectReason = "TIMEOUT"
	DisconnectReasonServerShutdown DisconnectReason = "SERVER_SHUTDOWN"
	DisconnectReasonError          DisconnectReason = "ERROR"
	DisconnectReasonAdminAction    DisconnectReason = "ADMIN_ACTION"
)

// VpnSession represents a VPN connection session
type VpnSession struct {
	ID               uuid.UUID         `gorm:"type:uuid;primary_key" json:"id"`
	UserID           uuid.UUID         `gorm:"type:uuid;not null;index" json:"user_id"`
	User             *User             `gorm:"foreignKey:UserID" json:"user,omitempty"`
	VpnIP            string            `gorm:"size:45;not null" json:"vpn_ip"`
	ClientIP         string            `gorm:"size:45" json:"client_ip,omitempty"`
	ConnectedAt      time.Time         `gorm:"not null;index" json:"connected_at"`
	DisconnectedAt   *time.Time        `json:"disconnected_at,omitempty"`
	BytesReceived    int64             `gorm:"default:0" json:"bytes_received"`
	BytesSent        int64             `gorm:"default:0" json:"bytes_sent"`
	DisconnectReason *DisconnectReason `gorm:"size:20" json:"disconnect_reason,omitempty"`
}

// BeforeCreate hook to generate UUID before creating a new session
func (s *VpnSession) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// TableName returns the table name for the VpnSession model
func (VpnSession) TableName() string {
	return "vpn_sessions"
}

// IsActive returns true if the session is still active (not disconnected)
func (s *VpnSession) IsActive() bool {
	return s.DisconnectedAt == nil
}

// Duration returns the duration of the session
func (s *VpnSession) Duration() time.Duration {
	if s.DisconnectedAt != nil {
		return s.DisconnectedAt.Sub(s.ConnectedAt)
	}
	return time.Since(s.ConnectedAt)
}

// TotalBytes returns total bytes transferred (received + sent)
func (s *VpnSession) TotalBytes() int64 {
	return s.BytesReceived + s.BytesSent
}
