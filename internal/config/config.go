package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	API      APIConfig      `yaml:"api"`
	Auth     AuthConfig     `yaml:"auth"`
	Logging  LoggingConfig  `yaml:"logging"`
	VPN      VPNConfig      `yaml:"vpn"`
	Security SecurityConfig `yaml:"security"`
}

// SecurityConfig represents security-related configuration
type SecurityConfig struct {
	RateLimitEnabled   bool `yaml:"rate_limit_enabled"`   // default: true
	RateLimitRequests  int  `yaml:"rate_limit_requests"`  // max requests per window, default: 5
	RateLimitWindow    int  `yaml:"rate_limit_window"`    // window in seconds, default: 60
	RateLimitBurst     int  `yaml:"rate_limit_burst"`     // burst size, default: 10
	LockoutMaxAttempts int  `yaml:"lockout_max_attempts"` // default: 5
	LockoutDuration    int  `yaml:"lockout_duration"`     // minutes, default: 15
}

// VPNConfig represents VPN network configuration
type VPNConfig struct {
	Network  string `yaml:"network"`   // CIDR notation, e.g., "10.8.0.0/24"
	ServerIP string `yaml:"server_ip"` // Server IP (reserved), e.g., "10.8.0.1"
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Output string `yaml:"output"` // "stdout", "file", "both" (default: stdout)
	Path   string `yaml:"path"`   // Directory for log files (default: current directory)
	Format string `yaml:"format"` // "text" or "json" (default: text)
	Level  string `yaml:"level"`  // "debug", "info", "warn", "error" (default: info)
}

// ServerConfig represents server-specific configuration
type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Type     string `yaml:"type"` // "postgres" or "mysql"
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	SSLMode  string `yaml:"sslmode"` // for postgres
}

// APIConfig represents REST API configuration
type APIConfig struct {
	Enabled           bool     `yaml:"enabled"`
	SwaggerEnabled    bool     `yaml:"swagger_enabled"`
	SwaggerAllowedIPs []string `yaml:"swagger_allowed_ips"` // CIDR notation: "0.0.0.0/0" for all, "192.168.1.0/24" for subnet
	VpnToken          string   `yaml:"vpn_token"`           // Token for VPN server authentication (X-VPN-Token header)
}

// AuthConfig represents authentication configuration
type AuthConfig struct {
	JWTSecret     string `yaml:"jwt_secret"`
	TokenExpiry   int    `yaml:"token_expiry"`   // in hours
	SessionExpiry int    `yaml:"session_expiry"` // in hours
}

// Load loads configuration from a YAML file with environment variable overrides
func Load(path string) (*Config, error) {
	var config Config

	// Try to load from YAML file first
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			// If file doesn't exist and we have env vars, continue with defaults
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to read config file: %w", err)
			}
		} else {
			if err := yaml.Unmarshal(data, &config); err != nil {
				return nil, fmt.Errorf("failed to parse config file: %w", err)
			}
		}
	}

	// Override with environment variables
	loadEnvOverrides(&config)

	// Set defaults
	if config.Server.Host == "" {
		config.Server.Host = "0.0.0.0"
	}
	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}
	if config.Auth.TokenExpiry == 0 {
		config.Auth.TokenExpiry = 24
	}
	if config.Auth.SessionExpiry == 0 {
		config.Auth.SessionExpiry = 8
	}

	// Logging defaults
	if config.Logging.Output == "" {
		config.Logging.Output = "stdout"
	}
	if config.Logging.Format == "" {
		config.Logging.Format = "text"
	}
	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}

	// Security defaults
	if !config.Security.RateLimitEnabled && os.Getenv("SECURITY_RATE_LIMIT_ENABLED") == "" {
		config.Security.RateLimitEnabled = true
	}
	if config.Security.RateLimitRequests == 0 {
		config.Security.RateLimitRequests = 5
	}
	if config.Security.RateLimitWindow == 0 {
		config.Security.RateLimitWindow = 60
	}
	if config.Security.RateLimitBurst == 0 {
		config.Security.RateLimitBurst = 10
	}
	if config.Security.LockoutMaxAttempts == 0 {
		config.Security.LockoutMaxAttempts = 5
	}
	if config.Security.LockoutDuration == 0 {
		config.Security.LockoutDuration = 15
	}

	// Database defaults
	if config.Database.Type == "" {
		config.Database.Type = "postgres"
	}
	if config.Database.Port == 0 {
		if config.Database.Type == "mysql" {
			config.Database.Port = 3306
		} else {
			config.Database.Port = 5432
		}
	}

	return &config, nil
}

// LoadFromEnv loads configuration purely from environment variables
func LoadFromEnv() (*Config, error) {
	return Load("")
}

// loadEnvOverrides overrides config values with environment variables
func loadEnvOverrides(config *Config) {
	// Server configuration
	if v := os.Getenv("SERVER_HOST"); v != "" {
		config.Server.Host = v
	}
	if v := os.Getenv("SERVER_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			config.Server.Port = port
		}
	}

	// Database configuration
	if v := os.Getenv("DB_TYPE"); v != "" {
		config.Database.Type = v
	}
	if v := os.Getenv("DB_HOST"); v != "" {
		config.Database.Host = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			config.Database.Port = port
		}
	}
	if v := os.Getenv("DB_USERNAME"); v != "" {
		config.Database.Username = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		config.Database.Password = v
	}
	if v := os.Getenv("DB_DATABASE"); v != "" {
		config.Database.Database = v
	}
	if v := os.Getenv("DB_SSLMODE"); v != "" {
		config.Database.SSLMode = v
	}

	// Auth configuration
	if v := os.Getenv("AUTH_JWT_SECRET"); v != "" {
		config.Auth.JWTSecret = v
	}
	if v := os.Getenv("AUTH_TOKEN_EXPIRY"); v != "" {
		if expiry, err := strconv.Atoi(v); err == nil {
			config.Auth.TokenExpiry = expiry
		}
	}
	if v := os.Getenv("AUTH_SESSION_EXPIRY"); v != "" {
		if expiry, err := strconv.Atoi(v); err == nil {
			config.Auth.SessionExpiry = expiry
		}
	}

	// API configuration
	if v := os.Getenv("API_ENABLED"); v != "" {
		config.API.Enabled = strings.ToLower(v) == "true" || v == "1"
	}
	if v := os.Getenv("API_SWAGGER_ENABLED"); v != "" {
		config.API.SwaggerEnabled = strings.ToLower(v) == "true" || v == "1"
	}
	if v := os.Getenv("API_SWAGGER_ALLOWED_IPS"); v != "" {
		config.API.SwaggerAllowedIPs = strings.Split(v, ",")
	}
	if v := os.Getenv("API_VPN_TOKEN"); v != "" {
		config.API.VpnToken = v
	}

	// Logging configuration
	if v := os.Getenv("LOG_OUTPUT"); v != "" {
		config.Logging.Output = v
	}
	if v := os.Getenv("LOG_PATH"); v != "" {
		config.Logging.Path = v
	}
	if v := os.Getenv("LOG_FORMAT"); v != "" {
		config.Logging.Format = v
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		config.Logging.Level = v
	}

	// Security configuration
	if v := os.Getenv("SECURITY_RATE_LIMIT_ENABLED"); v != "" {
		config.Security.RateLimitEnabled = strings.ToLower(v) == "true" || v == "1"
	}
	if v := os.Getenv("SECURITY_RATE_LIMIT_REQUESTS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			config.Security.RateLimitRequests = n
		}
	}
	if v := os.Getenv("SECURITY_RATE_LIMIT_WINDOW"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			config.Security.RateLimitWindow = n
		}
	}
	if v := os.Getenv("SECURITY_RATE_LIMIT_BURST"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			config.Security.RateLimitBurst = n
		}
	}
	if v := os.Getenv("SECURITY_LOCKOUT_MAX_ATTEMPTS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			config.Security.LockoutMaxAttempts = n
		}
	}
	if v := os.Getenv("SECURITY_LOCKOUT_DURATION"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			config.Security.LockoutDuration = n
		}
	}

	// VPN configuration
	if v := os.Getenv("VPN_NETWORK"); v != "" {
		config.VPN.Network = v
	}
	if v := os.Getenv("VPN_SERVER_IP"); v != "" {
		config.VPN.ServerIP = v
	}
}

// GetDSN returns the database connection string
func (c *DatabaseConfig) GetDSN() string {
	switch c.Type {
	case "postgres":
		sslMode := c.SSLMode
		if sslMode == "" {
			sslMode = "disable"
		}
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			c.Host, c.Port, c.Username, c.Password, c.Database, sslMode)
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			c.Username, c.Password, c.Host, c.Port, c.Database)
	default:
		return ""
	}
}
