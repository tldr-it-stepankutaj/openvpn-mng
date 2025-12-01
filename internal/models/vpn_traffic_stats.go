package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// VpnTrafficStats represents periodic traffic statistics for a VPN session
type VpnTrafficStats struct {
	ID                 uuid.UUID   `gorm:"type:uuid;primary_key" json:"id"`
	SessionID          uuid.UUID   `gorm:"type:uuid;not null;index" json:"session_id"`
	Session            *VpnSession `gorm:"foreignKey:SessionID" json:"session,omitempty"`
	Timestamp          time.Time   `gorm:"not null;index" json:"timestamp"`
	BytesReceivedDelta int64       `gorm:"default:0" json:"bytes_received_delta"`
	BytesSentDelta     int64       `gorm:"default:0" json:"bytes_sent_delta"`
}

// BeforeCreate hook to generate UUID before creating a new stats entry
func (s *VpnTrafficStats) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// TableName returns the table name for the VpnTrafficStats model
func (VpnTrafficStats) TableName() string {
	return "vpn_traffic_stats"
}

// TotalBytesDelta returns total bytes transferred in this interval
func (s *VpnTrafficStats) TotalBytesDelta() int64 {
	return s.BytesReceivedDelta + s.BytesSentDelta
}
