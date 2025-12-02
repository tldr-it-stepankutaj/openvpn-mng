// @title			OpenVPN Manager API
// @version		1.0
// @description	REST API for OpenVPN user, group, and network management
//
// @contact.name	API Support
// @contact.email	support@example.com
//
// @license.name	MIT
// @license.url	https://opensource.org/licenses/MIT
//
// @host		localhost:8080
// @BasePath	/
//
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description				Type "Bearer" followed by a space and JWT token. Example: "Bearer eyJhbGciOiJIUzI1NiIs..."
//
// @securityDefinitions.apikey	VpnToken
// @in							header
// @name						X-VPN-Token
// @description				VPN server authentication token configured in API settings

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/config"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/database"
	applogger "github.com/tldr-it-stepankutaj/openvpn-mng/internal/logger"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/models"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/routes"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/services"

	_ "github.com/tldr-it-stepankutaj/openvpn-mng/docs" // Swagger docs
)

var (
	configPath string
	version    = "1.0.1"
)

func init() {
	flag.StringVar(&configPath, "config", "config.yaml", "Path to configuration file")
}

func main() {
	flag.Parse()

	// Load configuration first (needed for logger setup)
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	if err := applogger.Initialize(&cfg.Logging); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer applogger.Close()

	// Print version
	applogger.Info("OpenVPN Manager starting", "version", version)
	applogger.Info("Configuration loaded", "path", configPath)
	applogger.Debug("Logging configuration", "output", cfg.Logging.Output, "format", cfg.Logging.Format, "level", cfg.Logging.Level)

	// Initialize database
	if err := database.Initialize(&cfg.Database, &cfg.Logging); err != nil {
		applogger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	applogger.Info("Connected to database", "type", cfg.Database.Type, "host", cfg.Database.Host, "port", cfg.Database.Port)

	// Run migrations
	if err := database.Migrate(); err != nil {
		applogger.Error("Failed to run database migrations", "error", err)
		os.Exit(1)
	}

	// Create a default admin user if not exists
	createDefaultAdmin()

	// Set Gin mode
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create a Gin router with our custom logger middleware
	r := gin.New()
	r.Use(applogger.GinLogger())
	r.Use(applogger.GinRecovery())

	// Setup routes
	routes.SetupRoutes(r, cfg)
	applogger.Info("Routes configured")

	// API status
	if cfg.API.Enabled {
		applogger.Info("REST API enabled", "path", "/api/v1")
		if cfg.API.VpnToken != "" {
			applogger.Info("VPN Auth API enabled", "path", "/api/v1/vpn-auth")
		}
		if cfg.API.SwaggerEnabled {
			applogger.Info("Swagger UI enabled", "path", "/swagger/index.html")
		}
	} else {
		applogger.Info("REST API disabled")
	}

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	applogger.Info("Starting server", "address", addr)

	if err := r.Run(addr); err != nil {
		applogger.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}

func createDefaultAdmin() {
	userService := services.NewUserService()

	// Check if admin exists
	_, err := userService.GetByUsername("admin")
	if err == nil {
		applogger.Debug("Default admin user already exists")
		return
	}

	// Create admin user
	hashedPassword, err := services.HashPassword("admin123")
	if err != nil {
		applogger.Error("Failed to hash password", "error", err)
		return
	}

	adminID := uuid.New()
	admin := &models.User{
		ID:        adminID,
		Username:  "admin",
		Password:  hashedPassword,
		FirstName: "System",
		LastName:  "Administrator",
		Email:     "admin@localhost",
		Role:      models.RoleAdmin,
		CreatedBy: adminID,
	}

	if err := database.GetDB().Create(admin).Error; err != nil {
		applogger.Error("Failed to create default admin user", "error", err)
		return
	}

	applogger.Info("Default admin user created", "username", "admin", "password", "admin123")
	applogger.Warn("Please change the default admin password immediately!")
}
