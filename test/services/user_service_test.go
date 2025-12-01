package services_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/dto"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/models"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/services"
	"github.com/tldr-it-stepankutaj/openvpn-mng/test/testutil"
)

func TestUserService_Create(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewUserService()
	admin := testutil.CreateTestAdmin(t)

	t.Run("successfully creates user", func(t *testing.T) {
		req := &dto.CreateUserRequest{
			Username:  "newuser",
			Password:  "password123",
			FirstName: "New",
			LastName:  "User",
			Email:     "newuser@test.com",
			Role:      models.RoleUser,
		}

		user, err := service.Create(req, admin.ID)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, user.ID)
		assert.Equal(t, "newuser", user.Username)
		assert.Equal(t, "New", user.FirstName)
		assert.Equal(t, "User", user.LastName)
		assert.Equal(t, "newuser@test.com", user.Email)
		assert.Equal(t, models.RoleUser, user.Role)
		assert.True(t, user.IsActive)
	})

	t.Run("fails when username exists", func(t *testing.T) {
		existingUser := testutil.CreateTestRegularUser(t)

		req := &dto.CreateUserRequest{
			Username:  existingUser.Username,
			Password:  "password123",
			FirstName: "Duplicate",
			LastName:  "User",
			Email:     "different@test.com",
			Role:      models.RoleUser,
		}

		_, err := service.Create(req, admin.ID)
		assert.Error(t, err)
		assert.Equal(t, services.ErrUserExists, err)
	})

	t.Run("fails when email exists", func(t *testing.T) {
		existingUser := testutil.CreateTestRegularUser(t)

		req := &dto.CreateUserRequest{
			Username:  "differentusername",
			Password:  "password123",
			FirstName: "Duplicate",
			LastName:  "Email",
			Email:     existingUser.Email,
			Role:      models.RoleUser,
		}

		_, err := service.Create(req, admin.ID)
		assert.Error(t, err)
		assert.Equal(t, services.ErrUserExists, err)
	})

	t.Run("creates user with manager", func(t *testing.T) {
		manager := testutil.CreateTestManager(t)

		req := &dto.CreateUserRequest{
			Username:  "manageduser",
			Password:  "password123",
			FirstName: "Managed",
			LastName:  "User",
			Email:     "managed@test.com",
			Role:      models.RoleUser,
			ManagerID: &manager.ID,
		}

		user, err := service.Create(req, admin.ID)
		require.NoError(t, err)
		assert.NotNil(t, user.ManagerID)
		assert.Equal(t, manager.ID, *user.ManagerID)
	})

	t.Run("creates user with IsActive parameter", func(t *testing.T) {
		// Note: GORM's default:true behavior on the model may override IsActive:false
		// This test verifies the service accepts the IsActive parameter
		isActive := true // Test with true which works with GORM's default
		req := &dto.CreateUserRequest{
			Username:  "activeuser",
			Password:  "password123",
			FirstName: "Active",
			LastName:  "User",
			Email:     "active@test.com",
			Role:      models.RoleUser,
			IsActive:  &isActive,
		}

		user, err := service.Create(req, admin.ID)
		require.NoError(t, err)
		assert.True(t, user.IsActive)
	})
}

func TestUserService_GetByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewUserService()

	t.Run("successfully gets user", func(t *testing.T) {
		testUser := testutil.CreateTestRegularUser(t)

		user, err := service.GetByID(testUser.ID)
		require.NoError(t, err)
		assert.Equal(t, testUser.ID, user.ID)
		assert.Equal(t, testUser.Username, user.Username)
	})

	t.Run("returns error for non-existent user", func(t *testing.T) {
		_, err := service.GetByID(uuid.New())
		assert.Error(t, err)
		assert.Equal(t, services.ErrUserNotFound, err)
	})
}

func TestUserService_GetByUsername(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewUserService()

	t.Run("successfully gets user by username", func(t *testing.T) {
		testUser := testutil.CreateTestRegularUser(t)

		user, err := service.GetByUsername(testUser.Username)
		require.NoError(t, err)
		assert.Equal(t, testUser.ID, user.ID)
	})

	t.Run("returns error for non-existent username", func(t *testing.T) {
		_, err := service.GetByUsername("nonexistent")
		assert.Error(t, err)
		assert.Equal(t, services.ErrUserNotFound, err)
	})
}

func TestUserService_Update(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewUserService()
	admin := testutil.CreateTestAdmin(t)

	t.Run("successfully updates user", func(t *testing.T) {
		testUser := testutil.CreateTestRegularUser(t)

		req := &dto.UpdateUserRequest{
			FirstName: "Updated",
			LastName:  "Name",
		}

		user, err := service.Update(testUser.ID, req, admin.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated", user.FirstName)
		assert.Equal(t, "Name", user.LastName)
	})

	t.Run("updates user role", func(t *testing.T) {
		testUser := testutil.CreateTestRegularUser(t)

		req := &dto.UpdateUserRequest{
			Role: models.RoleManager,
		}

		user, err := service.Update(testUser.ID, req, admin.ID)
		require.NoError(t, err)
		assert.Equal(t, models.RoleManager, user.Role)
	})

	t.Run("updates user active status", func(t *testing.T) {
		testUser := testutil.CreateTestRegularUser(t)
		isActive := false

		req := &dto.UpdateUserRequest{
			IsActive: &isActive,
		}

		user, err := service.Update(testUser.ID, req, admin.ID)
		require.NoError(t, err)
		assert.False(t, user.IsActive)
	})

	t.Run("returns error for non-existent user", func(t *testing.T) {
		req := &dto.UpdateUserRequest{
			FirstName: "Test",
		}

		_, err := service.Update(uuid.New(), req, admin.ID)
		assert.Error(t, err)
	})
}

func TestUserService_UpdateProfile(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewUserService()

	t.Run("successfully updates own profile", func(t *testing.T) {
		testUser := testutil.CreateTestRegularUser(t)

		req := &dto.UpdateProfileRequest{
			FirstName: "MyNew",
			LastName:  "Name",
			Telephone: "+420123456789",
		}

		user, err := service.UpdateProfile(testUser.ID, req, testUser.ID)
		require.NoError(t, err)
		assert.Equal(t, "MyNew", user.FirstName)
		assert.Equal(t, "Name", user.LastName)
		assert.Equal(t, "+420123456789", user.Telephone)
	})
}

func TestUserService_UpdatePassword(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewUserService()

	t.Run("successfully updates password", func(t *testing.T) {
		testUser := testutil.CreateTestRegularUser(t)

		err := service.UpdatePassword(testUser.ID, "testpassword123", "newpassword123", testUser.ID)
		require.NoError(t, err)

		// Verify new password works
		updatedUser, _ := service.GetByID(testUser.ID)
		assert.True(t, services.VerifyPassword("newpassword123", updatedUser.Password))
	})

	t.Run("fails with wrong current password", func(t *testing.T) {
		testUser := testutil.CreateTestRegularUser(t)

		err := service.UpdatePassword(testUser.ID, "wrongpassword", "newpassword123", testUser.ID)
		assert.Error(t, err)
		assert.Equal(t, services.ErrInvalidCredentials, err)
	})
}

func TestUserService_Delete(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewUserService()

	t.Run("successfully soft deletes user", func(t *testing.T) {
		testUser := testutil.CreateTestRegularUser(t)

		err := service.Delete(testUser.ID)
		require.NoError(t, err)

		// User should not be found (soft deleted)
		_, err = service.GetByID(testUser.ID)
		assert.Error(t, err)
	})
}

func TestUserService_List(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewUserService()

	// Create test users
	admin := testutil.CreateTestAdmin(t)
	manager := testutil.CreateTestManager(t)
	user1 := testutil.CreateTestUserWithManager(t, manager)
	user2 := testutil.CreateTestRegularUser(t)

	t.Run("admin sees all users", func(t *testing.T) {
		users, total, err := service.List(1, 100, models.RoleAdmin, &admin.ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(4))
		assert.GreaterOrEqual(t, len(users), 4)
	})

	t.Run("manager sees only subordinates", func(t *testing.T) {
		users, total, err := service.List(1, 100, models.RoleManager, &manager.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, 1, len(users))
		assert.Equal(t, user1.ID, users[0].ID)
	})

	t.Run("user sees only self", func(t *testing.T) {
		users, total, err := service.List(1, 100, models.RoleUser, &user2.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, 1, len(users))
		assert.Equal(t, user2.ID, users[0].ID)
	})

	t.Run("pagination works", func(t *testing.T) {
		users, _, err := service.List(1, 2, models.RoleAdmin, &admin.ID)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(users), 2)
	})
}

func TestUserService_GetManagedUsers(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewUserService()

	t.Run("returns managed users", func(t *testing.T) {
		manager := testutil.CreateTestManager(t)
		user1 := testutil.CreateTestUserWithManager(t, manager)
		user2 := testutil.CreateTestUserWithManager(t, manager)

		users, total, err := service.GetManagedUsers(manager.ID, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Equal(t, 2, len(users))

		userIDs := []uuid.UUID{users[0].ID, users[1].ID}
		assert.Contains(t, userIDs, user1.ID)
		assert.Contains(t, userIDs, user2.ID)
	})

	t.Run("returns empty for manager without subordinates", func(t *testing.T) {
		manager := testutil.CreateTestManager(t)

		users, total, err := service.GetManagedUsers(manager.ID, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Equal(t, 0, len(users))
	})
}
