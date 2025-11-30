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
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/models"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/routes"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/services"

	_ "github.com/tldr-it-stepankutaj/openvpn-mng/docs" // Swagger docs
)

var (
	configPath string
	version    = "1.0.0"
)

func init() {
	flag.StringVar(&configPath, "config", "config.yaml", "Path to configuration file")
}

func main() {
	flag.Parse()

	// Print version
	log.Printf("OpenVPN Manager v%s starting...", version)

	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Configuration loaded from %s", configPath)

	// Initialize database
	if err := database.Initialize(&cfg.Database); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Printf("Connected to %s database at %s:%d", cfg.Database.Type, cfg.Database.Host, cfg.Database.Port)

	// Run migrations
	if err := database.Migrate(); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}
	log.Println("Database migrations completed")

	// Create a default admin user if not exists
	createDefaultAdmin()

	// Set Gin mode
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create a Gin router
	r := gin.Default()

	// Setup routes
	routes.SetupRoutes(r, cfg)
	log.Println("Routes configured")

	// API status
	if cfg.API.Enabled {
		log.Println("REST API enabled at /api/v1")
		if cfg.API.SwaggerEnabled {
			log.Println("Swagger UI enabled at /swagger/index.html")
		}
	} else {
		log.Println("REST API disabled")
	}

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting server on %s", addr)

	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func createDefaultAdmin() {
	userService := services.NewUserService()

	// Check if admin exists
	_, err := userService.GetByUsername("admin")
	if err == nil {
		log.Println("Default admin user already exists")
		return
	}

	// Create admin user
	hashedPassword, err := services.HashPassword("admin123")
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
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
		log.Printf("Failed to create default admin user: %v", err)
		return
	}

	log.Println("Default admin user created (username: admin, password: admin123)")
	log.Println("WARNING: Please change the default admin password immediately!")
}
