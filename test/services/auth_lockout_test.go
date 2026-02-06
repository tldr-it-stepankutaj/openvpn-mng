package services_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/apperror"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/config"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/models"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/services"
	"github.com/tldr-it-stepankutaj/openvpn-mng/test/testutil"
)

func TestAuthService_AccountLockout(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	authCfg := &config.AuthConfig{
		JWTSecret:   "test-secret-key-for-testing",
		TokenExpiry: 24,
	}
	secCfg := &config.SecurityConfig{
		LockoutMaxAttempts: 3,
		LockoutDuration:    15,
	}
	service := services.NewAuthServiceWithSecurity(authCfg, secCfg)

	t.Run("account locks after max failed attempts", func(t *testing.T) {
		testutil.CreateTestUserWithName(t, models.RoleUser, "locktest1")

		// Fail login 3 times
		for i := 0; i < 3; i++ {
			_, _, err := service.Authenticate("locktest1", "wrongpassword")
			assert.Error(t, err)
			assert.Equal(t, services.ErrInvalidCredentials, err)
		}

		// Next attempt should return locked error even with wrong password
		_, _, err := service.Authenticate("locktest1", "wrongpassword")
		require.Error(t, err)
		appErr, ok := err.(*apperror.AppError)
		require.True(t, ok)
		assert.Equal(t, 429, appErr.Code)
		assert.Contains(t, appErr.Message, "locked")
	})

	t.Run("locked account rejects correct password", func(t *testing.T) {
		testutil.CreateTestUserWithName(t, models.RoleUser, "locktest2")

		// Fail login 3 times
		for i := 0; i < 3; i++ {
			service.Authenticate("locktest2", "wrongpassword")
		}

		// Correct password should still be rejected
		_, _, err := service.Authenticate("locktest2", "testpassword123")
		require.Error(t, err)
		appErr, ok := err.(*apperror.AppError)
		require.True(t, ok)
		assert.Equal(t, 429, appErr.Code)
	})

	t.Run("successful login resets failed counter", func(t *testing.T) {
		testutil.CreateTestUserWithName(t, models.RoleUser, "locktest3")

		// Fail 2 times (below threshold)
		for i := 0; i < 2; i++ {
			service.Authenticate("locktest3", "wrongpassword")
		}

		// Succeed
		token, _, err := service.Authenticate("locktest3", "testpassword123")
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		// Fail 2 more times (counter should have been reset)
		for i := 0; i < 2; i++ {
			_, _, err := service.Authenticate("locktest3", "wrongpassword")
			assert.Error(t, err)
			assert.Equal(t, services.ErrInvalidCredentials, err)
		}

		// Should still be able to login (not locked, only 2 failures since reset)
		token, _, err = service.Authenticate("locktest3", "testpassword123")
		require.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("lockout expires after duration", func(t *testing.T) {
		testutil.CreateTestUserWithName(t, models.RoleUser, "locktest4")

		// Fail login 3 times to trigger lockout
		for i := 0; i < 3; i++ {
			service.Authenticate("locktest4", "wrongpassword")
		}

		// Manually set lockout to past time to simulate expiry
		pastTime := time.Now().Add(-1 * time.Minute)
		testutil.TestDB.Model(&models.User{}).Where("username = ?", "locktest4").
			Update("locked_until", pastTime)

		// Should be able to login now
		token, _, err := service.Authenticate("locktest4", "testpassword123")
		require.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("lockout does not apply without security config", func(t *testing.T) {
		serviceNoSec := services.NewAuthService(authCfg)
		testutil.CreateTestUserWithName(t, models.RoleUser, "locktest5")

		// Fail many times
		for i := 0; i < 10; i++ {
			serviceNoSec.Authenticate("locktest5", "wrongpassword")
		}

		// Should still be able to login (no lockout)
		token, _, err := serviceNoSec.Authenticate("locktest5", "testpassword123")
		require.NoError(t, err)
		assert.NotEmpty(t, token)
	})
}
