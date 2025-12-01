package testutil

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/database"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/models"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestDB holds the test database connection
var TestDB *gorm.DB

// SetupTestDB initializes an in-memory SQLite database for testing
func SetupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Auto migrate all models
	err = db.AutoMigrate(
		&models.User{},
		&models.Group{},
		&models.Network{},
		&models.UserGroup{},
		&models.NetworkGroup{},
		&models.VpnSession{},
		&models.VpnTrafficStats{},
		&models.AuditLog{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	// Set the global database instance
	database.SetDB(db)
	TestDB = db

	return db
}

// CleanupTestDB cleans up the test database
func CleanupTestDB(t *testing.T, db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		t.Errorf("Failed to get underlying DB: %v", err)
		return
	}
	sqlDB.Close()
}

// CreateTestUser creates a test user with the given role
func CreateTestUser(t *testing.T, role models.Role) *models.User {
	return CreateTestUserWithName(t, role, fmt.Sprintf("testuser_%s", uuid.New().String()[:8]))
}

// CreateTestUserWithName creates a test user with a specific username
func CreateTestUserWithName(t *testing.T, role models.Role, username string) *models.User {
	hashedPassword, err := services.HashPassword("testpassword123")
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	user := &models.User{
		ID:        uuid.New(),
		Username:  username,
		Password:  hashedPassword,
		FirstName: "Test",
		LastName:  "User",
		Email:     fmt.Sprintf("%s@test.com", username),
		Role:      role,
		IsActive:  true,
		CreatedBy: uuid.New(),
	}

	if err := TestDB.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return user
}

// CreateTestAdmin creates a test admin user
func CreateTestAdmin(t *testing.T) *models.User {
	return CreateTestUser(t, models.RoleAdmin)
}

// CreateTestManager creates a test manager user
func CreateTestManager(t *testing.T) *models.User {
	return CreateTestUser(t, models.RoleManager)
}

// CreateTestRegularUser creates a test regular user
func CreateTestRegularUser(t *testing.T) *models.User {
	return CreateTestUser(t, models.RoleUser)
}

// CreateManagedUser creates a test user managed by the specified manager
func CreateManagedUser(t *testing.T, managerID uuid.UUID) *models.User {
	hashedPassword, err := services.HashPassword("testpassword123")
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	user := &models.User{
		ID:        uuid.New(),
		Username:  fmt.Sprintf("managed_user_%s", uuid.New().String()[:8]),
		Password:  hashedPassword,
		FirstName: "Managed",
		LastName:  "User",
		Email:     fmt.Sprintf("managed_%s@test.com", uuid.New().String()[:8]),
		Role:      models.RoleUser,
		IsActive:  true,
		ManagerID: &managerID,
		CreatedBy: managerID,
	}

	if err := TestDB.Create(user).Error; err != nil {
		t.Fatalf("Failed to create managed user: %v", err)
	}

	return user
}

// CreateTestUserWithManager creates a test user with a manager
func CreateTestUserWithManager(t *testing.T, manager *models.User) *models.User {
	hashedPassword, err := services.HashPassword("testpassword123")
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	user := &models.User{
		ID:        uuid.New(),
		Username:  fmt.Sprintf("managed_user_%s", uuid.New().String()[:8]),
		Password:  hashedPassword,
		FirstName: "Managed",
		LastName:  "User",
		Email:     fmt.Sprintf("managed_%s@test.com", uuid.New().String()[:8]),
		Role:      models.RoleUser,
		IsActive:  true,
		ManagerID: &manager.ID,
		CreatedBy: manager.ID,
	}

	if err := TestDB.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user with manager: %v", err)
	}

	return user
}

// CreateTestGroup creates a test group
func CreateTestGroup(t *testing.T, createdBy uuid.UUID) *models.Group {
	group := &models.Group{
		ID:          uuid.New(),
		Name:        fmt.Sprintf("Test Group %s", uuid.New().String()[:8]),
		Description: "Test group description",
		CreatedBy:   createdBy,
	}

	if err := TestDB.Create(group).Error; err != nil {
		t.Fatalf("Failed to create test group: %v", err)
	}

	return group
}

// CreateTestNetwork creates a test network
func CreateTestNetwork(t *testing.T, createdBy uuid.UUID) *models.Network {
	network := &models.Network{
		ID:          uuid.New(),
		Name:        fmt.Sprintf("Test Network %s", uuid.New().String()[:8]),
		CIDR:        fmt.Sprintf("192.168.%d.0/24", time.Now().UnixNano()%255),
		Description: "Test network description",
		CreatedBy:   createdBy,
	}

	if err := TestDB.Create(network).Error; err != nil {
		t.Fatalf("Failed to create test network: %v", err)
	}

	return network
}

// CreateTestVpnSession creates a test VPN session
func CreateTestVpnSession(t *testing.T, userID uuid.UUID) *models.VpnSession {
	session := &models.VpnSession{
		ID:          uuid.New(),
		UserID:      userID,
		VpnIP:       "10.8.0.100",
		ClientIP:    "203.0.113.50",
		ConnectedAt: time.Now(),
	}

	if err := TestDB.Create(session).Error; err != nil {
		t.Fatalf("Failed to create test VPN session: %v", err)
	}

	return session
}

// CreateTestAuditLog creates a test audit log
func CreateTestAuditLog(t *testing.T, userID uuid.UUID) *models.AuditLog {
	entityID := uuid.New()
	audit := &models.AuditLog{
		ID:         uuid.New(),
		UserID:     userID,
		Action:     models.AuditActionCreate,
		EntityType: "user",
		EntityID:   &entityID,
		Details:    "Test audit log",
		IPAddress:  "127.0.0.1",
		UserAgent:  "Test Agent",
	}

	if err := TestDB.Create(audit).Error; err != nil {
		t.Fatalf("Failed to create test audit log: %v", err)
	}

	return audit
}

// TimePtr returns a pointer to a time.Time
func TimePtr(t time.Time) *time.Time {
	return &t
}

// BoolPtr returns a pointer to a bool
func BoolPtr(b bool) *bool {
	return &b
}

// StringPtr returns a pointer to a string
func StringPtr(s string) *string {
	return &s
}

// UUIDPtr returns a pointer to a UUID
func UUIDPtr(u uuid.UUID) *uuid.UUID {
	return &u
}

// SkipIfShort skips the test if running in short mode
func SkipIfShort(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
}

// SetEnv sets an environment variable for the duration of a test
func SetEnv(t *testing.T, key, value string) func() {
	oldValue, hadOldValue := os.LookupEnv(key)
	os.Setenv(key, value)

	return func() {
		if hadOldValue {
			os.Setenv(key, oldValue)
		} else {
			os.Unsetenv(key)
		}
	}
}
