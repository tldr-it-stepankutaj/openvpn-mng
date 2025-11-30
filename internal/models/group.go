package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Group represents a group in the system (e.g., IT, HR, Finance)
type Group struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	Name        string         `gorm:"uniqueIndex;size:100;not null" json:"name"`
	Description string         `gorm:"size:500" json:"description,omitempty"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   *time.Time     `gorm:"autoUpdateTime" json:"updated_at,omitempty"`
	CreatedBy   uuid.UUID      `gorm:"type:uuid;not null" json:"created_by"`
	UpdatedBy   *uuid.UUID     `gorm:"type:uuid" json:"updated_by,omitempty"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate hook to generate UUID before creating a new group
func (g *Group) BeforeCreate(tx *gorm.DB) error {
	if g.ID == uuid.Nil {
		g.ID = uuid.New()
	}
	return nil
}

// TableName returns the table name for the Group model
func (Group) TableName() string {
	return "groups"
}

// UserGroup represents the many-to-many relationship between users and groups
type UserGroup struct {
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	GroupID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"group_id"`
	User      User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Group     Group     `gorm:"foreignKey:GroupID" json:"group,omitempty"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	CreatedBy uuid.UUID `gorm:"type:uuid;not null" json:"created_by"`
}

// TableName returns the table name for the UserGroup model
func (UserGroup) TableName() string {
	return "user_groups"
}
