package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Network represents a network or IP range that can be assigned to groups
type Network struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	Name        string         `gorm:"size:100;not null;uniqueIndex" json:"name"`
	CIDR        string         `gorm:"size:50;not null" json:"cidr"` // e.g., "192.168.1.0/24" or "10.0.0.1/32"
	Description string         `gorm:"size:500" json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedBy   uuid.UUID      `gorm:"type:uuid;not null" json:"created_by"`
	UpdatedBy   *uuid.UUID     `gorm:"type:uuid" json:"updated_by"`
	Creator     *User          `gorm:"foreignKey:CreatedBy;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"creator,omitempty"`
	Updater     *User          `gorm:"foreignKey:UpdatedBy;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"updater,omitempty"`
	Groups      []Group        `gorm:"many2many:network_groups;" json:"groups,omitempty"`
}

// BeforeCreate hook to set UUID
func (n *Network) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return nil
}

// NetworkGroup represents the many-to-many relationship between networks and groups
type NetworkGroup struct {
	NetworkID uuid.UUID `gorm:"type:uuid;primaryKey" json:"network_id"`
	GroupID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"group_id"`
	Network   *Network  `gorm:"foreignKey:NetworkID" json:"network,omitempty"`
	Group     *Group    `gorm:"foreignKey:GroupID" json:"group,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy uuid.UUID `gorm:"type:uuid;not null" json:"created_by"`
	Creator   *User     `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}
