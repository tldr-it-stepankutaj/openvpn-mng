package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WellKnownVpnClientConfigID is the fixed UUID for the single-row VPN client config
// This ensures only one configuration exists in the database
var WellKnownVpnClientConfigID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

// DefaultVpnClientTemplate is the default OpenVPN client configuration template
const DefaultVpnClientTemplate = `client
dev tun
proto {{PROTOCOL}}
remote {{SERVER_ADDRESS}} {{SERVER_PORT}}
resolv-retry infinite
nobind
persist-key
persist-tun
remote-cert-tls server
auth-user-pass
verb 3

<ca>
{{CA_CERT}}
</ca>
{{#TLS_KEY}}
<tls-auth>
{{TLS_KEY}}
</tls-auth>
key-direction {{TLS_KEY_DIRECTION}}
{{/TLS_KEY}}`

// VpnClientConfig represents the global OpenVPN client configuration
type VpnClientConfig struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	ServerAddress   string         `gorm:"size:255;not null" json:"server_address"`
	ServerPort      int            `gorm:"not null;default:1194" json:"server_port"`
	Protocol        string         `gorm:"size:10;not null;default:'udp'" json:"protocol"`
	CACert          string         `gorm:"type:text;not null" json:"ca_cert"`
	TLSKey          string         `gorm:"type:text" json:"tls_key,omitempty"`
	TLSKeyDirection int            `gorm:"default:1" json:"tls_key_direction"`
	Template        string         `gorm:"type:text;not null" json:"template"`
	ConfigName      string         `gorm:"size:100;not null;default:'client'" json:"config_name"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       *time.Time     `gorm:"autoUpdateTime" json:"updated_at,omitempty"`
	UpdatedBy       *uuid.UUID     `gorm:"type:uuid" json:"updated_by,omitempty"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate hook to set the well-known UUID
func (v *VpnClientConfig) BeforeCreate(tx *gorm.DB) error {
	v.ID = WellKnownVpnClientConfigID
	return nil
}

// TableName returns the table name for the VpnClientConfig model
func (VpnClientConfig) TableName() string {
	return "vpn_client_configs"
}

// HasTLSKey returns true if TLS key is configured
func (v *VpnClientConfig) HasTLSKey() bool {
	return v.TLSKey != ""
}

// GetFilename returns the filename for the .ovpn file download
func (v *VpnClientConfig) GetFilename() string {
	if v.ConfigName != "" {
		return v.ConfigName + ".ovpn"
	}
	return "client.ovpn"
}
