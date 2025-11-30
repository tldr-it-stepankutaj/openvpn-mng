package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	API      APIConfig      `yaml:"api"`
	Auth     AuthConfig     `yaml:"auth"`
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
}

// AuthConfig represents authentication configuration
type AuthConfig struct {
	JWTSecret     string `yaml:"jwt_secret"`
	TokenExpiry   int    `yaml:"token_expiry"`   // in hours
	SessionExpiry int    `yaml:"session_expiry"` // in hours
}

// Load loads configuration from a YAML file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

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

	return &config, nil
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
