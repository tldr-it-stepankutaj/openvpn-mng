package dto_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/dto"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/models"
)

func TestDateOnly_UnmarshalJSON(t *testing.T) {
	t.Run("parses date-only format", func(t *testing.T) {
		jsonData := `{"valid_from": "2024-01-15"}`

		var req struct {
			ValidFrom *dto.DateOnly `json:"valid_from"`
		}

		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)
		require.NotNil(t, req.ValidFrom)
		assert.Equal(t, 2024, req.ValidFrom.Year())
		assert.Equal(t, time.January, req.ValidFrom.Month())
		assert.Equal(t, 15, req.ValidFrom.Day())
	})

	t.Run("parses RFC3339 format", func(t *testing.T) {
		jsonData := `{"valid_from": "2024-06-20T10:30:00Z"}`

		var req struct {
			ValidFrom *dto.DateOnly `json:"valid_from"`
		}

		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)
		require.NotNil(t, req.ValidFrom)
		assert.Equal(t, 2024, req.ValidFrom.Year())
		assert.Equal(t, time.June, req.ValidFrom.Month())
		assert.Equal(t, 20, req.ValidFrom.Day())
	})

	t.Run("parses datetime without timezone", func(t *testing.T) {
		jsonData := `{"valid_from": "2024-12-25T08:00:00"}`

		var req struct {
			ValidFrom *dto.DateOnly `json:"valid_from"`
		}

		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)
		require.NotNil(t, req.ValidFrom)
		assert.Equal(t, 2024, req.ValidFrom.Year())
		assert.Equal(t, time.December, req.ValidFrom.Month())
		assert.Equal(t, 25, req.ValidFrom.Day())
	})

	t.Run("handles null value", func(t *testing.T) {
		jsonData := `{"valid_from": null}`

		var req struct {
			ValidFrom *dto.DateOnly `json:"valid_from"`
		}

		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)
	})

	t.Run("handles empty string", func(t *testing.T) {
		jsonData := `{"valid_from": ""}`

		var req struct {
			ValidFrom *dto.DateOnly `json:"valid_from"`
		}

		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)
	})

	t.Run("returns error for invalid format", func(t *testing.T) {
		jsonData := `{"valid_from": "not-a-date"}`

		var req struct {
			ValidFrom *dto.DateOnly `json:"valid_from"`
		}

		err := json.Unmarshal([]byte(jsonData), &req)
		assert.Error(t, err)
	})
}

func TestDateOnly_ToTimePtr(t *testing.T) {
	t.Run("returns time pointer for valid date", func(t *testing.T) {
		d := &dto.DateOnly{Time: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)}

		ptr := d.ToTimePtr()
		require.NotNil(t, ptr)
		assert.Equal(t, 2024, ptr.Year())
		assert.Equal(t, time.January, ptr.Month())
		assert.Equal(t, 15, ptr.Day())
	})

	t.Run("returns nil for zero time", func(t *testing.T) {
		d := &dto.DateOnly{Time: time.Time{}}

		ptr := d.ToTimePtr()
		assert.Nil(t, ptr)
	})

	t.Run("returns nil for nil DateOnly", func(t *testing.T) {
		var d *dto.DateOnly

		ptr := d.ToTimePtr()
		assert.Nil(t, ptr)
	})
}

func TestToUserResponse(t *testing.T) {
	t.Run("converts user model to response", func(t *testing.T) {
		userID := uuid.New()
		managerID := uuid.New()
		now := time.Now()
		validFrom := time.Now().AddDate(0, 0, -7)
		validTo := time.Now().AddDate(0, 0, 7)

		user := &models.User{
			ID:         userID,
			Username:   "testuser",
			ManagerID:  &managerID,
			FirstName:  "Test",
			MiddleName: "Middle",
			LastName:   "User",
			Email:      "test@example.com",
			Telephone:  "123456789",
			Role:       models.RoleUser,
			IsActive:   true,
			ValidFrom:  &validFrom,
			ValidTo:    &validTo,
			VpnIP:      "10.8.0.50",
			CreatedAt:  now,
			CreatedBy:  managerID,
		}

		response := dto.ToUserResponse(user)

		require.NotNil(t, response)
		assert.Equal(t, userID, response.ID)
		assert.Equal(t, "testuser", response.Username)
		assert.Equal(t, &managerID, response.ManagerID)
		assert.Equal(t, "Test", response.FirstName)
		assert.Equal(t, "Middle", response.MiddleName)
		assert.Equal(t, "User", response.LastName)
		assert.Equal(t, "test@example.com", response.Email)
		assert.Equal(t, "123456789", response.Telephone)
		assert.Equal(t, models.RoleUser, response.Role)
		assert.True(t, response.IsActive)
		assert.NotNil(t, response.ValidFrom)
		assert.NotNil(t, response.ValidTo)
		assert.Equal(t, "10.8.0.50", response.VpnIP)
		assert.Equal(t, managerID, response.CreatedBy)
	})

	t.Run("returns nil for nil user", func(t *testing.T) {
		response := dto.ToUserResponse(nil)
		assert.Nil(t, response)
	})

	t.Run("includes manager when present", func(t *testing.T) {
		managerID := uuid.New()
		userID := uuid.New()

		manager := &models.User{
			ID:        managerID,
			Username:  "manager",
			FirstName: "Manager",
			LastName:  "User",
			Email:     "manager@example.com",
			Role:      models.RoleManager,
		}

		user := &models.User{
			ID:        userID,
			Username:  "testuser",
			ManagerID: &managerID,
			Manager:   manager,
			FirstName: "Test",
			LastName:  "User",
			Email:     "test@example.com",
			Role:      models.RoleUser,
			CreatedBy: managerID,
		}

		response := dto.ToUserResponse(user)

		require.NotNil(t, response)
		require.NotNil(t, response.Manager)
		assert.Equal(t, managerID, response.Manager.ID)
		assert.Equal(t, "manager", response.Manager.Username)
	})
}

func TestToUserResponseList(t *testing.T) {
	t.Run("converts list of users", func(t *testing.T) {
		users := []models.User{
			{
				ID:        uuid.New(),
				Username:  "user1",
				FirstName: "First",
				LastName:  "User",
				Email:     "user1@example.com",
				Role:      models.RoleUser,
				CreatedBy: uuid.New(),
			},
			{
				ID:        uuid.New(),
				Username:  "user2",
				FirstName: "Second",
				LastName:  "User",
				Email:     "user2@example.com",
				Role:      models.RoleManager,
				CreatedBy: uuid.New(),
			},
		}

		responses := dto.ToUserResponseList(users)

		require.Len(t, responses, 2)
		assert.Equal(t, "user1", responses[0].Username)
		assert.Equal(t, "user2", responses[1].Username)
	})

	t.Run("returns empty list for empty input", func(t *testing.T) {
		responses := dto.ToUserResponseList([]models.User{})
		require.Empty(t, responses)
	})
}

func TestCreateUserRequest(t *testing.T) {
	t.Run("parses valid request", func(t *testing.T) {
		jsonData := `{
			"username": "newuser",
			"password": "password123",
			"first_name": "New",
			"last_name": "User",
			"email": "new@example.com",
			"role": "USER",
			"valid_from": "2024-01-01",
			"valid_to": "2024-12-31"
		}`

		var req dto.CreateUserRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.Equal(t, "newuser", req.Username)
		assert.Equal(t, "password123", req.Password)
		assert.Equal(t, "New", req.FirstName)
		assert.Equal(t, "User", req.LastName)
		assert.Equal(t, "new@example.com", req.Email)
		assert.Equal(t, models.RoleUser, req.Role)
		require.NotNil(t, req.ValidFrom)
		require.NotNil(t, req.ValidTo)
	})

	t.Run("parses request with optional fields", func(t *testing.T) {
		jsonData := `{
			"username": "newuser",
			"password": "password123",
			"first_name": "New",
			"middle_name": "Middle",
			"last_name": "User",
			"email": "new@example.com",
			"telephone": "123456789",
			"role": "USER",
			"vpn_ip": "10.8.0.100"
		}`

		var req dto.CreateUserRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.Equal(t, "Middle", req.MiddleName)
		assert.Equal(t, "123456789", req.Telephone)
		assert.Equal(t, "10.8.0.100", req.VpnIP)
	})
}

func TestUpdateUserRequest(t *testing.T) {
	t.Run("parses partial update", func(t *testing.T) {
		jsonData := `{
			"first_name": "Updated"
		}`

		var req dto.UpdateUserRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.Equal(t, "Updated", req.FirstName)
		assert.Empty(t, req.LastName)
		assert.Empty(t, req.Username)
	})

	t.Run("parses boolean fields", func(t *testing.T) {
		jsonData := `{
			"is_active": false
		}`

		var req dto.UpdateUserRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		require.NotNil(t, req.IsActive)
		assert.False(t, *req.IsActive)
	})
}

func TestAddUserGroupRequest(t *testing.T) {
	t.Run("parses group ID", func(t *testing.T) {
		groupID := uuid.New()
		jsonData := `{"group_id": "` + groupID.String() + `"}`

		var req dto.AddUserGroupRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.Equal(t, groupID, req.GroupID)
	})
}
