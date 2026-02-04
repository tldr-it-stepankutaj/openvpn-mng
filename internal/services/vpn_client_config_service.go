package services

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/database"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/dto"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/models"
	"gorm.io/gorm"
)

var (
	ErrVpnClientConfigNotFound = errors.New("vpn client configuration not found")
	ErrInvalidCACert           = errors.New("invalid CA certificate: must be a valid PEM certificate")
	ErrInvalidTLSKey           = errors.New("invalid TLS key: must be a valid PEM key")
)

// VpnClientConfigService provides VPN client configuration management services
type VpnClientConfigService struct{}

// NewVpnClientConfigService creates a new VPN client config service
func NewVpnClientConfigService() *VpnClientConfigService {
	return &VpnClientConfigService{}
}

// Get retrieves the current VPN client configuration
func (s *VpnClientConfigService) Get() (*models.VpnClientConfig, error) {
	var config models.VpnClientConfig
	if err := database.GetDB().First(&config, "id = ?", models.WellKnownVpnClientConfigID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrVpnClientConfigNotFound
		}
		return nil, err
	}
	return &config, nil
}

// CreateOrUpdate creates or updates the VPN client configuration
func (s *VpnClientConfigService) CreateOrUpdate(req *dto.VpnClientConfigRequest, updatedBy uuid.UUID) (*models.VpnClientConfig, error) {
	// Validate CA certificate
	if !isValidPEMCertificate(req.CACert) {
		return nil, ErrInvalidCACert
	}

	// Validate TLS key if provided
	if req.TLSKey != "" && !isValidPEMKey(req.TLSKey) {
		return nil, ErrInvalidTLSKey
	}

	// Check if config exists
	existingConfig, err := s.Get()
	if err != nil && !errors.Is(err, ErrVpnClientConfigNotFound) {
		return nil, err
	}

	if existingConfig != nil {
		// Update existing config
		updates := map[string]any{
			"server_address":    req.ServerAddress,
			"server_port":       req.ServerPort,
			"protocol":          req.Protocol,
			"ca_cert":           req.CACert,
			"tls_key":           req.TLSKey,
			"tls_key_direction": req.TLSKeyDirection,
			"template":          req.Template,
			"config_name":       req.ConfigName,
			"updated_by":        updatedBy,
		}

		if err := database.GetDB().Model(existingConfig).Updates(updates).Error; err != nil {
			return nil, err
		}

		return s.Get()
	}

	// Create new config
	config := &models.VpnClientConfig{
		ServerAddress:   req.ServerAddress,
		ServerPort:      req.ServerPort,
		Protocol:        req.Protocol,
		CACert:          req.CACert,
		TLSKey:          req.TLSKey,
		TLSKeyDirection: req.TLSKeyDirection,
		Template:        req.Template,
		ConfigName:      req.ConfigName,
		UpdatedBy:       &updatedBy,
	}

	if err := database.GetDB().Create(config).Error; err != nil {
		return nil, err
	}

	return config, nil
}

// GenerateOvpnConfig generates the .ovpn configuration content
func (s *VpnClientConfigService) GenerateOvpnConfig() (string, string, error) {
	config, err := s.Get()
	if err != nil {
		return "", "", err
	}

	content := s.processTemplate(config)
	filename := config.GetFilename()

	return content, filename, nil
}

// processTemplate processes the template with the configuration values
func (s *VpnClientConfigService) processTemplate(config *models.VpnClientConfig) string {
	template := config.Template

	// Replace simple placeholders
	replacements := map[string]string{
		"{{SERVER_ADDRESS}}":    config.ServerAddress,
		"{{SERVER_PORT}}":       intToString(config.ServerPort),
		"{{PROTOCOL}}":          config.Protocol,
		"{{CA_CERT}}":           strings.TrimSpace(config.CACert),
		"{{TLS_KEY_DIRECTION}}": intToString(config.TLSKeyDirection),
	}

	for placeholder, value := range replacements {
		template = strings.ReplaceAll(template, placeholder, value)
	}

	// Handle conditional TLS_KEY section
	if config.HasTLSKey() {
		// Replace TLS_KEY placeholder
		template = strings.ReplaceAll(template, "{{TLS_KEY}}", strings.TrimSpace(config.TLSKey))
		// Remove conditional markers but keep content
		template = strings.ReplaceAll(template, "{{#TLS_KEY}}", "")
		template = strings.ReplaceAll(template, "{{/TLS_KEY}}", "")
	} else {
		// Remove entire conditional section including content
		re := regexp.MustCompile(`(?s)\{\{#TLS_KEY\}\}.*?\{\{/TLS_KEY\}\}`)
		template = re.ReplaceAllString(template, "")
	}

	// Clean up any trailing whitespace on lines and normalize line endings
	lines := strings.Split(template, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	template = strings.Join(lines, "\n")

	// Remove multiple consecutive blank lines
	re := regexp.MustCompile(`\n{3,}`)
	template = re.ReplaceAllString(template, "\n\n")

	return strings.TrimSpace(template) + "\n"
}

// GetDefaultTemplate returns the default OpenVPN client template
func (s *VpnClientConfigService) GetDefaultTemplate() string {
	return models.DefaultVpnClientTemplate
}

// isValidPEMCertificate checks if the string is a valid PEM certificate
func isValidPEMCertificate(cert string) bool {
	cert = strings.TrimSpace(cert)
	return strings.Contains(cert, "-----BEGIN CERTIFICATE-----") &&
		strings.Contains(cert, "-----END CERTIFICATE-----")
}

// isValidPEMKey checks if the string is a valid PEM key
func isValidPEMKey(key string) bool {
	key = strings.TrimSpace(key)
	// Check for various key formats
	hasBegin := strings.Contains(key, "-----BEGIN")
	hasEnd := strings.Contains(key, "-----END")
	return hasBegin && hasEnd
}

// intToString converts an int to string
func intToString(n int) string {
	return strconv.Itoa(n)
}
