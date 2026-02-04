package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/models"
)

// VpnClientConfigRequest represents a request to create or update VPN client config
type VpnClientConfigRequest struct {
	ServerAddress   string `json:"server_address" binding:"required,min=1,max=255"`
	ServerPort      int    `json:"server_port" binding:"required,min=1,max=65535"`
	Protocol        string `json:"protocol" binding:"required,oneof=udp tcp"`
	CACert          string `json:"ca_cert" binding:"required"`
	TLSKey          string `json:"tls_key,omitempty"`
	TLSKeyDirection int    `json:"tls_key_direction" binding:"omitempty,min=0,max=1"`
	Template        string `json:"template" binding:"required"`
	ConfigName      string `json:"config_name" binding:"required,min=1,max=100"`
}

// VpnClientConfigResponse represents VPN client config in API responses
type VpnClientConfigResponse struct {
	ID              uuid.UUID  `json:"id"`
	ServerAddress   string     `json:"server_address"`
	ServerPort      int        `json:"server_port"`
	Protocol        string     `json:"protocol"`
	CACert          string     `json:"ca_cert"`
	TLSKey          string     `json:"tls_key,omitempty"`
	TLSKeyDirection int        `json:"tls_key_direction"`
	Template        string     `json:"template"`
	ConfigName      string     `json:"config_name"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at,omitempty"`
	UpdatedBy       *uuid.UUID `json:"updated_by,omitempty"`
}

// VpnClientConfigPreviewResponse represents the preview of generated .ovpn content
type VpnClientConfigPreviewResponse struct {
	Content  string `json:"content"`
	Filename string `json:"filename"`
}

// DefaultTemplateResponse represents the default template response
type DefaultTemplateResponse struct {
	Template string `json:"template"`
}

// ToVpnClientConfigResponse converts a VpnClientConfig model to VpnClientConfigResponse DTO
func ToVpnClientConfigResponse(config *models.VpnClientConfig) *VpnClientConfigResponse {
	if config == nil {
		return nil
	}

	return &VpnClientConfigResponse{
		ID:              config.ID,
		ServerAddress:   config.ServerAddress,
		ServerPort:      config.ServerPort,
		Protocol:        config.Protocol,
		CACert:          config.CACert,
		TLSKey:          config.TLSKey,
		TLSKeyDirection: config.TLSKeyDirection,
		Template:        config.Template,
		ConfigName:      config.ConfigName,
		CreatedAt:       config.CreatedAt,
		UpdatedAt:       config.UpdatedAt,
		UpdatedBy:       config.UpdatedBy,
	}
}
