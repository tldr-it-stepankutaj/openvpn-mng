package services_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/config"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/models"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/services"
	"github.com/tldr-it-stepankutaj/openvpn-mng/test/testutil"
)

func TestAuthService_Authenticate(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	cfg := &config.AuthConfig{
		JWTSecret:   "test-secret-key-for-testing",
		TokenExpiry: 24,
	}
	service := services.NewAuthService(cfg)

	t.Run("successfully authenticates valid user", func(t *testing.T) {
		user := testutil.CreateTestUserWithName(t, models.RoleUser, "authuser")

		token, authenticatedUser, err := service.Authenticate("authuser", "testpassword123")
		require.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.Equal(t, user.ID, authenticatedUser.ID)
		assert.Equal(t, user.Username, authenticatedUser.Username)
	})

	t.Run("fails with invalid username", func(t *testing.T) {
		_, _, err := service.Authenticate("nonexistent", "password")
		assert.Error(t, err)
		assert.Equal(t, services.ErrInvalidCredentials, err)
	})

	t.Run("fails with invalid password", func(t *testing.T) {
		testutil.CreateTestUserWithName(t, models.RoleUser, "authuser2")

		_, _, err := service.Authenticate("authuser2", "wrongpassword")
		assert.Error(t, err)
		assert.Equal(t, services.ErrInvalidCredentials, err)
	})

	t.Run("fails for inactive user", func(t *testing.T) {
		// Create user first, then deactivate
		hashedPassword, _ := services.HashPassword("testpassword123")
		inactiveUser := &models.User{
			Username:  "inactiveauth",
			Password:  hashedPassword,
			FirstName: "Inactive",
			LastName:  "User",
			Email:     "inactiveauth@test.com",
			Role:      models.RoleUser,
			IsActive:  true,
			CreatedBy: testutil.CreateTestAdmin(t).ID,
		}
		testutil.TestDB.Create(inactiveUser)
		// Deactivate user after creation to bypass GORM default
		testutil.TestDB.Model(inactiveUser).Update("is_active", false)

		_, _, err := service.Authenticate("inactiveauth", "testpassword123")
		assert.Error(t, err)
		assert.Equal(t, services.ErrUserInactive, err)
	})

	t.Run("fails for user not yet valid", func(t *testing.T) {
		// Create user with valid_from in the future
		hashedPassword, _ := services.HashPassword("testpassword123")
		futureDate := time.Now().AddDate(0, 0, 7) // 7 days in future
		notYetValidUser := &models.User{
			Username:  "notyetvalid",
			Password:  hashedPassword,
			FirstName: "NotYet",
			LastName:  "Valid",
			Email:     "notyetvalid@test.com",
			Role:      models.RoleUser,
			IsActive:  true,
			ValidFrom: &futureDate,
			CreatedBy: testutil.CreateTestAdmin(t).ID,
		}
		testutil.TestDB.Create(notYetValidUser)

		_, _, err := service.Authenticate("notyetvalid", "testpassword123")
		assert.Error(t, err)
		assert.Equal(t, services.ErrUserNotYetValid, err)
	})

	t.Run("fails for expired user", func(t *testing.T) {
		// Create user with valid_to in the past
		hashedPassword, _ := services.HashPassword("testpassword123")
		pastDate := time.Now().AddDate(0, 0, -7) // 7 days ago
		expiredUser := &models.User{
			Username:  "expireduser",
			Password:  hashedPassword,
			FirstName: "Expired",
			LastName:  "User",
			Email:     "expireduser@test.com",
			Role:      models.RoleUser,
			IsActive:  true,
			ValidTo:   &pastDate,
			CreatedBy: testutil.CreateTestAdmin(t).ID,
		}
		testutil.TestDB.Create(expiredUser)

		_, _, err := service.Authenticate("expireduser", "testpassword123")
		assert.Error(t, err)
		assert.Equal(t, services.ErrUserExpired, err)
	})

	t.Run("succeeds for user within validity period", func(t *testing.T) {
		hashedPassword, _ := services.HashPassword("testpassword123")
		pastDate := time.Now().AddDate(0, 0, -7)  // 7 days ago
		futureDate := time.Now().AddDate(0, 0, 7) // 7 days in future
		validUser := &models.User{
			Username:  "validperioduser",
			Password:  hashedPassword,
			FirstName: "Valid",
			LastName:  "Period",
			Email:     "validperiod@test.com",
			Role:      models.RoleUser,
			IsActive:  true,
			ValidFrom: &pastDate,
			ValidTo:   &futureDate,
			CreatedBy: testutil.CreateTestAdmin(t).ID,
		}
		testutil.TestDB.Create(validUser)

		token, user, err := service.Authenticate("validperioduser", "testpassword123")
		require.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.Equal(t, "validperioduser", user.Username)
	})
}

func TestHashPassword(t *testing.T) {
	t.Run("successfully hashes password", func(t *testing.T) {
		hash, err := services.HashPassword("mypassword")
		require.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.NotEqual(t, "mypassword", hash)
	})

	t.Run("different passwords produce different hashes", func(t *testing.T) {
		hash1, _ := services.HashPassword("password1")
		hash2, _ := services.HashPassword("password2")
		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("same password produces different hashes (bcrypt salt)", func(t *testing.T) {
		hash1, _ := services.HashPassword("samepassword")
		hash2, _ := services.HashPassword("samepassword")
		assert.NotEqual(t, hash1, hash2) // bcrypt uses random salt
	})
}

func TestVerifyPassword(t *testing.T) {
	t.Run("verifies correct password", func(t *testing.T) {
		hash, _ := services.HashPassword("correctpassword")
		assert.True(t, services.VerifyPassword("correctpassword", hash))
	})

	t.Run("rejects incorrect password", func(t *testing.T) {
		hash, _ := services.HashPassword("correctpassword")
		assert.False(t, services.VerifyPassword("wrongpassword", hash))
	})

	t.Run("rejects empty password", func(t *testing.T) {
		hash, _ := services.HashPassword("correctpassword")
		assert.False(t, services.VerifyPassword("", hash))
	})
}

func TestGetUserByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	t.Run("returns user by ID", func(t *testing.T) {
		testUser := testutil.CreateTestRegularUser(t)

		user, err := services.GetUserByID(testUser.ID)
		require.NoError(t, err)
		assert.Equal(t, testUser.ID, user.ID)
		assert.Equal(t, testUser.Username, user.Username)
	})

	t.Run("returns error for non-existent ID", func(t *testing.T) {
		_, err := services.GetUserByID(uuid.New())
		assert.Error(t, err)
	})
}
