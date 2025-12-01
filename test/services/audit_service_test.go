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

func TestAuditService_List(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewAuditService()
	admin := testutil.CreateTestAdmin(t)

	// Create test audit logs
	testutil.CreateTestAuditLog(t, admin.ID)
	testutil.CreateTestAuditLog(t, admin.ID)
	testutil.CreateTestAuditLog(t, admin.ID)

	t.Run("lists all audit logs", func(t *testing.T) {
		filter := &dto.AuditLogFilter{
			Page:     1,
			PageSize: 100,
		}

		logs, total, err := service.List(filter)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(3))
		assert.GreaterOrEqual(t, len(logs), 3)
	})

	t.Run("filters by user ID", func(t *testing.T) {
		user := testutil.CreateTestRegularUser(t)
		testutil.CreateTestAuditLog(t, user.ID)

		filter := &dto.AuditLogFilter{
			UserID:   &user.ID,
			Page:     1,
			PageSize: 100,
		}

		logs, total, err := service.List(filter)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, 1, len(logs))
		assert.Equal(t, user.ID, logs[0].UserID)
	})

	t.Run("filters by action", func(t *testing.T) {
		// Create audit log with different action
		entityID := uuid.New()
		auditLog := &models.AuditLog{
			UserID:     admin.ID,
			Action:     models.AuditActionDelete,
			EntityType: "test",
			EntityID:   &entityID,
		}
		testutil.TestDB.Create(auditLog)

		action := models.AuditActionDelete
		filter := &dto.AuditLogFilter{
			Action:   &action,
			Page:     1,
			PageSize: 100,
		}

		logs, _, err := service.List(filter)
		require.NoError(t, err)
		for _, log := range logs {
			assert.Equal(t, models.AuditActionDelete, log.Action)
		}
	})

	t.Run("filters by entity type", func(t *testing.T) {
		filter := &dto.AuditLogFilter{
			EntityType: "user",
			Page:       1,
			PageSize:   100,
		}

		logs, _, err := service.List(filter)
		require.NoError(t, err)
		for _, log := range logs {
			assert.Equal(t, "user", log.EntityType)
		}
	})

	t.Run("pagination works", func(t *testing.T) {
		filter := &dto.AuditLogFilter{
			Page:     1,
			PageSize: 2,
		}

		logs, _, err := service.List(filter)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(logs), 2)
	})

	t.Run("handles default pagination values", func(t *testing.T) {
		filter := &dto.AuditLogFilter{
			Page:     0, // Should default to 1
			PageSize: 0, // Should default to 20
		}

		logs, _, err := service.List(filter)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(logs), 20)
	})

	t.Run("limits page size to 100", func(t *testing.T) {
		filter := &dto.AuditLogFilter{
			Page:     1,
			PageSize: 500, // Should be limited to 100
		}

		_, _, err := service.List(filter)
		require.NoError(t, err)
	})
}

func TestAuditService_GetByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewAuditService()
	admin := testutil.CreateTestAdmin(t)

	t.Run("successfully gets audit log", func(t *testing.T) {
		testLog := testutil.CreateTestAuditLog(t, admin.ID)

		log, err := service.GetByID(testLog.ID)
		require.NoError(t, err)
		assert.Equal(t, testLog.ID, log.ID)
	})

	t.Run("returns error for non-existent log", func(t *testing.T) {
		_, err := service.GetByID(uuid.New())
		assert.Error(t, err)
		assert.Equal(t, services.ErrAuditLogNotFound, err)
	})
}

func TestAuditService_GetByEntityID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewAuditService()
	admin := testutil.CreateTestAdmin(t)

	t.Run("returns logs for entity", func(t *testing.T) {
		entityID := uuid.New()

		// Create multiple logs for same entity
		for i := 0; i < 3; i++ {
			auditLog := &models.AuditLog{
				UserID:     admin.ID,
				Action:     models.AuditActionUpdate,
				EntityType: "user",
				EntityID:   &entityID,
			}
			testutil.TestDB.Create(auditLog)
		}

		logs, total, err := service.GetByEntityID("user", entityID, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Equal(t, 3, len(logs))

		for _, log := range logs {
			assert.Equal(t, entityID, *log.EntityID)
			assert.Equal(t, "user", log.EntityType)
		}
	})

	t.Run("returns empty for entity without logs", func(t *testing.T) {
		logs, total, err := service.GetByEntityID("user", uuid.New(), 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Equal(t, 0, len(logs))
	})
}

func TestAuditService_GetByUserID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewAuditService()

	t.Run("returns logs for user", func(t *testing.T) {
		user := testutil.CreateTestRegularUser(t)

		// Create multiple logs for same user
		for i := 0; i < 3; i++ {
			testutil.CreateTestAuditLog(t, user.ID)
		}

		logs, total, err := service.GetByUserID(user.ID, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Equal(t, 3, len(logs))

		for _, log := range logs {
			assert.Equal(t, user.ID, log.UserID)
		}
	})

	t.Run("returns empty for user without logs", func(t *testing.T) {
		user := testutil.CreateTestRegularUser(t)

		logs, total, err := service.GetByUserID(user.ID, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Equal(t, 0, len(logs))
	})
}

func TestAuditService_GetActions(t *testing.T) {
	service := services.NewAuditService()

	t.Run("returns all audit actions", func(t *testing.T) {
		actions := service.GetActions()

		expectedActions := []models.AuditAction{
			models.AuditActionCreate,
			models.AuditActionRead,
			models.AuditActionUpdate,
			models.AuditActionDelete,
			models.AuditActionLogin,
			models.AuditActionLogout,
		}

		assert.Equal(t, len(expectedActions), len(actions))
		for _, expected := range expectedActions {
			assert.Contains(t, actions, expected)
		}
	})
}

func TestAuditService_GetEntityTypes(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewAuditService()
	admin := testutil.CreateTestAdmin(t)

	t.Run("returns unique entity types", func(t *testing.T) {
		// Create logs with different entity types
		entityTypes := []string{"user", "group", "network"}
		for _, entityType := range entityTypes {
			entityID := uuid.New()
			auditLog := &models.AuditLog{
				UserID:     admin.ID,
				Action:     models.AuditActionCreate,
				EntityType: entityType,
				EntityID:   &entityID,
			}
			testutil.TestDB.Create(auditLog)
		}

		types, err := service.GetEntityTypes()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(types), 3)
	})
}

func TestAuditService_GetStatsByAction(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewAuditService()
	admin := testutil.CreateTestAdmin(t)

	t.Run("returns stats grouped by action", func(t *testing.T) {
		// Create logs with different actions
		entityID := uuid.New()
		actions := []models.AuditAction{
			models.AuditActionCreate,
			models.AuditActionCreate,
			models.AuditActionUpdate,
		}
		for _, action := range actions {
			auditLog := &models.AuditLog{
				UserID:     admin.ID,
				Action:     action,
				EntityType: "test",
				EntityID:   &entityID,
			}
			testutil.TestDB.Create(auditLog)
		}

		stats, err := service.GetStatsByAction()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, stats[models.AuditActionCreate], int64(2))
		assert.GreaterOrEqual(t, stats[models.AuditActionUpdate], int64(1))
	})
}

func TestAuditService_GetStatsByEntityType(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewAuditService()
	admin := testutil.CreateTestAdmin(t)

	t.Run("returns stats grouped by entity type", func(t *testing.T) {
		// Create logs with different entity types
		types := []string{"user", "user", "group"}
		for _, entityType := range types {
			entityID := uuid.New()
			auditLog := &models.AuditLog{
				UserID:     admin.ID,
				Action:     models.AuditActionCreate,
				EntityType: entityType,
				EntityID:   &entityID,
			}
			testutil.TestDB.Create(auditLog)
		}

		stats, err := service.GetStatsByEntityType()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, stats["user"], int64(2))
		assert.GreaterOrEqual(t, stats["group"], int64(1))
	})
}
