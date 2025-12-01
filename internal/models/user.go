package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	Username     string         `gorm:"uniqueIndex;size:100;not null" json:"username"`
	Password     string         `gorm:"size:255;not null" json:"-"`
	ManagerID    *uuid.UUID     `gorm:"type:uuid;index" json:"manager_id,omitempty"`
	Manager      *User          `gorm:"foreignKey:ManagerID" json:"manager,omitempty"`
	FirstName    string         `gorm:"size:100;not null" json:"first_name"`
	MiddleName   string         `gorm:"size:100" json:"middle_name,omitempty"`
	LastName     string         `gorm:"size:100;not null" json:"last_name"`
	Email        string         `gorm:"uniqueIndex;size:255;not null" json:"email"`
	Telephone    string         `gorm:"size:50" json:"telephone,omitempty"`
	Role         Role           `gorm:"size:20;not null;default:'USER'" json:"role"`
	IsActive     bool           `gorm:"not null;default:true" json:"is_active"`
	ValidFrom    *time.Time     `gorm:"type:date" json:"valid_from,omitempty"`
	ValidTo      *time.Time     `gorm:"type:date" json:"valid_to,omitempty"`
	VpnIP        string         `gorm:"size:45" json:"vpn_ip,omitempty"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    *time.Time     `gorm:"autoUpdateTime" json:"updated_at,omitempty"`
	CreatedBy    uuid.UUID      `gorm:"type:uuid;not null" json:"created_by"`
	UpdatedBy    *uuid.UUID     `gorm:"type:uuid" json:"updated_by,omitempty"`
	ManagedUsers []User         `gorm:"foreignKey:ManagerID" json:"managed_users,omitempty"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate hook to generate UUID before creating a new user
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// TableName returns the table name for the User model
func (User) TableName() string {
	return "users"
}

// GetFullName returns the full name of the user
func (u *User) GetFullName() string {
	if u.MiddleName != "" {
		return u.FirstName + " " + u.MiddleName + " " + u.LastName
	}
	return u.FirstName + " " + u.LastName
}

// IsAdmin checks if the user is an admin
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// IsManager checks if the user is a manager
func (u *User) IsManager() bool {
	return u.Role == RoleManager
}

// CanManage checks if this user can manage another user
func (u *User) CanManage(targetUser *User) bool {
	// Admin can manage everyone
	if u.IsAdmin() {
		return true
	}

	// Manager can manage users assigned to them
	if u.IsManager() && targetUser.ManagerID != nil && *targetUser.ManagerID == u.ID {
		return true
	}

	// Users can manage themselves
	if u.ID == targetUser.ID {
		return true
	}

	return false
}

// IsValidForLogin checks if user can login based on is_active, valid_from, valid_to
func (u *User) IsValidForLogin() bool {
	if !u.IsActive {
		return false
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Check valid_from
	if u.ValidFrom != nil {
		validFromDate := time.Date(u.ValidFrom.Year(), u.ValidFrom.Month(), u.ValidFrom.Day(), 0, 0, 0, 0, u.ValidFrom.Location())
		if today.Before(validFromDate) {
			return false
		}
	}

	// Check valid_to
	if u.ValidTo != nil {
		validToDate := time.Date(u.ValidTo.Year(), u.ValidTo.Month(), u.ValidTo.Day(), 23, 59, 59, 0, u.ValidTo.Location())
		if today.After(validToDate) {
			return false
		}
	}

	return true
}
