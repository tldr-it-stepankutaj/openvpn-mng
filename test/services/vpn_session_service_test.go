package services_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/dto"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/models"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/services"
	"github.com/tldr-it-stepankutaj/openvpn-mng/test/testutil"
)

func TestVpnSessionService_Create(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewVpnSessionService()

	t.Run("successfully creates session", func(t *testing.T) {
		user := testutil.CreateTestRegularUser(t)
		connectedAt := time.Now()

		req := &dto.CreateVpnSessionRequest{
			UserID:      user.ID,
			VpnIP:       "10.8.0.100",
			ClientIP:    "203.0.113.50",
			ConnectedAt: connectedAt,
		}

		session, err := service.Create(req)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, session.ID)
		assert.Equal(t, user.ID, session.UserID)
		assert.Equal(t, "10.8.0.100", session.VpnIP)
		assert.Equal(t, "203.0.113.50", session.ClientIP)
		assert.Nil(t, session.DisconnectedAt)
	})
}

func TestVpnSessionService_GetByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewVpnSessionService()

	t.Run("successfully gets session", func(t *testing.T) {
		user := testutil.CreateTestRegularUser(t)
		testSession := testutil.CreateTestVpnSession(t, user.ID)

		session, err := service.GetByID(testSession.ID)
		require.NoError(t, err)
		assert.Equal(t, testSession.ID, session.ID)
	})

	t.Run("returns error for non-existent session", func(t *testing.T) {
		_, err := service.GetByID(uuid.New())
		assert.Error(t, err)
		assert.Equal(t, services.ErrSessionNotFound, err)
	})
}

func TestVpnSessionService_Disconnect(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewVpnSessionService()

	t.Run("successfully disconnects session", func(t *testing.T) {
		user := testutil.CreateTestRegularUser(t)
		testSession := testutil.CreateTestVpnSession(t, user.ID)
		disconnectedAt := time.Now()
		disconnectReason := models.DisconnectReasonUserRequest

		req := &dto.UpdateVpnSessionRequest{
			DisconnectedAt:   disconnectedAt,
			BytesReceived:    1024000,
			BytesSent:        512000,
			DisconnectReason: &disconnectReason,
		}

		session, err := service.Disconnect(testSession.ID, req)
		require.NoError(t, err)
		assert.NotNil(t, session.DisconnectedAt)
		assert.Equal(t, int64(1024000), session.BytesReceived)
		assert.Equal(t, int64(512000), session.BytesSent)
		assert.Equal(t, models.DisconnectReasonUserRequest, *session.DisconnectReason)
	})

	t.Run("returns error for non-existent session", func(t *testing.T) {
		req := &dto.UpdateVpnSessionRequest{
			DisconnectedAt: time.Now(),
		}

		_, err := service.Disconnect(uuid.New(), req)
		assert.Error(t, err)
	})
}

func TestVpnSessionService_List(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewVpnSessionService()

	// Create test sessions
	user1 := testutil.CreateTestRegularUser(t)
	user2 := testutil.CreateTestRegularUser(t)
	testutil.CreateTestVpnSession(t, user1.ID)
	testutil.CreateTestVpnSession(t, user1.ID)
	testutil.CreateTestVpnSession(t, user2.ID)

	t.Run("lists all sessions", func(t *testing.T) {
		filter := &dto.VpnSessionFilter{
			Page:     1,
			PageSize: 100,
		}

		sessions, total, err := service.List(filter)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(3))
		assert.GreaterOrEqual(t, len(sessions), 3)
	})

	t.Run("filters by user ID", func(t *testing.T) {
		filter := &dto.VpnSessionFilter{
			UserID:   &user1.ID,
			Page:     1,
			PageSize: 100,
		}

		sessions, total, err := service.List(filter)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Equal(t, 2, len(sessions))
		for _, session := range sessions {
			assert.Equal(t, user1.ID, session.UserID)
		}
	})

	t.Run("filters active sessions", func(t *testing.T) {
		isActive := true
		filter := &dto.VpnSessionFilter{
			IsActive: &isActive,
			Page:     1,
			PageSize: 100,
		}

		sessions, _, err := service.List(filter)
		require.NoError(t, err)
		for _, session := range sessions {
			assert.Nil(t, session.DisconnectedAt)
		}
	})

	t.Run("pagination works", func(t *testing.T) {
		filter := &dto.VpnSessionFilter{
			Page:     1,
			PageSize: 2,
		}

		sessions, _, err := service.List(filter)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(sessions), 2)
	})
}

func TestVpnSessionService_GetActiveSessions(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewVpnSessionService()

	t.Run("returns only active sessions", func(t *testing.T) {
		user := testutil.CreateTestRegularUser(t)

		// Create active session
		activeSession := testutil.CreateTestVpnSession(t, user.ID)

		// Create disconnected session
		disconnectedSession := testutil.CreateTestVpnSession(t, user.ID)
		disconnectedAt := time.Now()
		testutil.TestDB.Model(disconnectedSession).Update("disconnected_at", disconnectedAt)

		sessions, err := service.GetActiveSessions()
		require.NoError(t, err)

		// Check that only active session is returned
		var found bool
		for _, s := range sessions {
			if s.ID == activeSession.ID {
				found = true
			}
			assert.Nil(t, s.DisconnectedAt)
		}
		assert.True(t, found)
	})
}

func TestVpnSessionService_GetActiveSessionByUserID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewVpnSessionService()

	t.Run("returns active session for user", func(t *testing.T) {
		user := testutil.CreateTestRegularUser(t)
		activeSession := testutil.CreateTestVpnSession(t, user.ID)

		session, err := service.GetActiveSessionByUserID(user.ID)
		require.NoError(t, err)
		assert.Equal(t, activeSession.ID, session.ID)
	})

	t.Run("returns error when no active session", func(t *testing.T) {
		user := testutil.CreateTestRegularUser(t)

		// Create disconnected session
		disconnectedSession := testutil.CreateTestVpnSession(t, user.ID)
		disconnectedAt := time.Now()
		testutil.TestDB.Model(disconnectedSession).Update("disconnected_at", disconnectedAt)

		_, err := service.GetActiveSessionByUserID(user.ID)
		assert.Error(t, err)
		assert.Equal(t, services.ErrSessionNotFound, err)
	})
}

func TestVpnSessionService_GetUsageStats(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewVpnSessionService()

	t.Run("returns usage stats", func(t *testing.T) {
		user := testutil.CreateTestRegularUser(t)

		// Create sessions with traffic data
		session1 := testutil.CreateTestVpnSession(t, user.ID)
		testutil.TestDB.Model(session1).Updates(map[string]interface{}{
			"bytes_received": 1024,
			"bytes_sent":     512,
		})

		session2 := testutil.CreateTestVpnSession(t, user.ID)
		testutil.TestDB.Model(session2).Updates(map[string]interface{}{
			"bytes_received": 2048,
			"bytes_sent":     1024,
		})

		stats, err := service.GetUsageStats()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, stats.TotalSessions, int64(2))
		assert.GreaterOrEqual(t, stats.TotalBytesReceived, int64(3072))
		assert.GreaterOrEqual(t, stats.TotalBytesSent, int64(1536))
	})
}

func TestVpnTrafficStatsService_Create(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewVpnTrafficStatsService()

	t.Run("successfully creates traffic stats", func(t *testing.T) {
		user := testutil.CreateTestRegularUser(t)
		session := testutil.CreateTestVpnSession(t, user.ID)

		req := &dto.CreateVpnTrafficStatsRequest{
			SessionID:          session.ID,
			Timestamp:          time.Now(),
			BytesReceivedDelta: 1024,
			BytesSentDelta:     512,
		}

		stats, err := service.Create(req)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, stats.ID)
		assert.Equal(t, session.ID, stats.SessionID)
		assert.Equal(t, int64(1024), stats.BytesReceivedDelta)
		assert.Equal(t, int64(512), stats.BytesSentDelta)
	})
}

func TestVpnTrafficStatsService_List(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewVpnTrafficStatsService()

	// Create test data
	user := testutil.CreateTestRegularUser(t)
	session := testutil.CreateTestVpnSession(t, user.ID)

	for i := 0; i < 3; i++ {
		stats := &models.VpnTrafficStats{
			SessionID:          session.ID,
			Timestamp:          time.Now(),
			BytesReceivedDelta: int64(i * 1024),
			BytesSentDelta:     int64(i * 512),
		}
		testutil.TestDB.Create(stats)
	}

	t.Run("lists all stats", func(t *testing.T) {
		filter := &dto.VpnTrafficStatsFilter{
			Page:     1,
			PageSize: 100,
		}

		stats, total, err := service.List(filter)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(3))
		assert.GreaterOrEqual(t, len(stats), 3)
	})

	t.Run("filters by session ID", func(t *testing.T) {
		filter := &dto.VpnTrafficStatsFilter{
			SessionID: &session.ID,
			Page:      1,
			PageSize:  100,
		}

		stats, _, err := service.List(filter)
		require.NoError(t, err)
		for _, s := range stats {
			assert.Equal(t, session.ID, s.SessionID)
		}
	})
}

func TestVpnTrafficStatsService_GetBySessionID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	service := services.NewVpnTrafficStatsService()

	t.Run("returns stats for session", func(t *testing.T) {
		user := testutil.CreateTestRegularUser(t)
		session := testutil.CreateTestVpnSession(t, user.ID)

		// Create stats for this session
		for i := 0; i < 3; i++ {
			stats := &models.VpnTrafficStats{
				SessionID:          session.ID,
				Timestamp:          time.Now(),
				BytesReceivedDelta: int64(i * 1024),
				BytesSentDelta:     int64(i * 512),
			}
			testutil.TestDB.Create(stats)
		}

		stats, total, err := service.GetBySessionID(session.ID, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Equal(t, 3, len(stats))
	})

	t.Run("returns empty for session without stats", func(t *testing.T) {
		user := testutil.CreateTestRegularUser(t)
		session := testutil.CreateTestVpnSession(t, user.ID)

		stats, total, err := service.GetBySessionID(session.ID, 1, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Equal(t, 0, len(stats))
	})
}

func TestCalculateTotalPages(t *testing.T) {
	testCases := []struct {
		name     string
		total    int64
		pageSize int
		expected int
	}{
		{"exact division", 100, 10, 10},
		{"with remainder", 105, 10, 11},
		{"less than page size", 5, 10, 1},
		{"zero items", 0, 10, 0},
		{"default page size for zero", 100, 0, 5},      // defaults to 20
		{"default page size for negative", 100, -1, 5}, // defaults to 20
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := services.CalculateTotalPages(tc.total, tc.pageSize)
			assert.Equal(t, tc.expected, result)
		})
	}
}
