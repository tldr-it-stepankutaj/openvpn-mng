package database

import (
	"fmt"

	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/config"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Initialize initializes the database connection
func Initialize(cfg *config.DatabaseConfig) error {
	var dialector gorm.Dialector

	switch cfg.Type {
	case "postgres":
		dialector = postgres.Open(cfg.GetDSN())
	case "mysql":
		dialector = mysql.Open(cfg.GetDSN())
	default:
		return fmt.Errorf("unsupported database type: %s", cfg.Type)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	DB = db
	return nil
}

// Migrate runs database migrations
func Migrate() error {
	return DB.AutoMigrate(
		&models.User{},
		&models.Group{},
		&models.UserGroup{},
		&models.AuditLog{},
	)
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}

// Close closes the database connection
func Close() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
