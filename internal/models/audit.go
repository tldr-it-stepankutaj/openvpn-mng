package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditAction represents the type of action performed
type AuditAction string

const (
	AuditActionCreate AuditAction = "CREATE"
	AuditActionRead   AuditAction = "READ"
	AuditActionUpdate AuditAction = "UPDATE"
	AuditActionDelete AuditAction = "DELETE"
	AuditActionLogin  AuditAction = "LOGIN"
	AuditActionLogout AuditAction = "LOGOUT"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	ID         uuid.UUID   `gorm:"type:uuid;primary_key" json:"id"`
	UserID     uuid.UUID   `gorm:"type:uuid;not null;index" json:"user_id"`
	User       *User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Action     AuditAction `gorm:"size:20;not null;index" json:"action"`
	EntityType string      `gorm:"size:50;not null;index" json:"entity_type"` // user, group, etc.
	EntityID   *uuid.UUID  `gorm:"type:uuid;index" json:"entity_id,omitempty"`
	OldValues  string      `gorm:"type:text" json:"old_values,omitempty"` // JSON string of old values
	NewValues  string      `gorm:"type:text" json:"new_values,omitempty"` // JSON string of new values
	IPAddress  string      `gorm:"size:45" json:"ip_address,omitempty"`
	UserAgent  string      `gorm:"size:500" json:"user_agent,omitempty"`
	Details    string      `gorm:"type:text" json:"details,omitempty"` // Additional details
	CreatedAt  time.Time   `gorm:"autoCreateTime;index" json:"created_at"`
}

// BeforeCreate hook to generate UUID before creating a new audit log
func (a *AuditLog) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// TableName returns the table name for the AuditLog model
func (AuditLog) TableName() string {
	return "audit_logs"
}
