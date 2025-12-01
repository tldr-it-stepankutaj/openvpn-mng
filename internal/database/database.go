package database

import (
	"fmt"

	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/config"
	applogger "github.com/tldr-it-stepankutaj/openvpn-mng/internal/logger"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// Initialize initializes the database connection
func Initialize(dbCfg *config.DatabaseConfig, logCfg *config.LoggingConfig) error {
	var dialector gorm.Dialector

	switch dbCfg.Type {
	case "postgres":
		dialector = postgres.Open(dbCfg.GetDSN())
	case "mysql":
		dialector = mysql.Open(dbCfg.GetDSN())
	default:
		return fmt.Errorf("unsupported database type: %s", dbCfg.Type)
	}

	// Use custom GORM logger that integrates with our logging system
	gormLogger := applogger.NewGormLogger(logCfg)

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger:                                   gormLogger,
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	DB = db
	return nil
}

// Migrate runs database migrations
func Migrate() error {
	applogger.Info("Database migrations started")

	// Define tables to migrate
	tables := []struct {
		name  string
		model interface{}
	}{
		{"users", &models.User{}},
		{"groups", &models.Group{}},
		{"user_groups", &models.UserGroup{}},
		{"networks", &models.Network{}},
		{"network_groups", &models.NetworkGroup{}},
		{"audit_logs", &models.AuditLog{}},
		{"vpn_sessions", &models.VpnSession{}},
		{"vpn_traffic_stats", &models.VpnTrafficStats{}},
	}

	for _, t := range tables {
		applogger.Debug("Migrating table", "table", t.name)
		if err := DB.AutoMigrate(t.model); err != nil {
			return fmt.Errorf("failed to migrate %s: %w", t.name, err)
		}
	}

	applogger.Info("Database migrations completed")
	return nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}

// SetDB sets the database instance (used for testing)
func SetDB(db *gorm.DB) {
	DB = db
}

// Close closes the database connection
func Close() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
