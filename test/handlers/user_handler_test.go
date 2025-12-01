package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/dto"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/handlers"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/middleware"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/models"
	"github.com/tldr-it-stepankutaj/openvpn-mng/test/testutil"
)

func setupUserRouter(authUser *dto.AuthUser) (*gin.Engine, *handlers.UserHandler) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := handlers.NewUserHandler()

	if authUser != nil {
		router.Use(func(c *gin.Context) {
			c.Set(middleware.AuthUserKey, authUser)
			c.Next()
		})
	}

	return router, handler
}

func TestUserHandler_Create(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	t.Run("admin creates user successfully", func(t *testing.T) {
		admin := testutil.CreateTestAdmin(t)
		router, handler := setupUserRouter(&dto.AuthUser{
			ID:       admin.ID.String(),
			Username: admin.Username,
			Role:     admin.Role,
		})
		router.POST("/api/v1/users", handler.Create)

		body := dto.CreateUserRequest{
			Username:  "newuser1",
			Password:  "password123",
			FirstName: "New",
			LastName:  "User",
			Email:     "newuser1@test.com",
			Role:      models.RoleUser,
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response dto.UserResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "newuser1", response.Username)
	})

	t.Run("manager creates user with self as manager", func(t *testing.T) {
		manager := testutil.CreateTestManager(t)
		router, handler := setupUserRouter(&dto.AuthUser{
			ID:       manager.ID.String(),
			Username: manager.Username,
			Role:     manager.Role,
		})
		router.POST("/api/v1/users", handler.Create)

		body := dto.CreateUserRequest{
			Username:  "manageduser1",
			Password:  "password123",
			FirstName: "Managed",
			LastName:  "User",
			Email:     "manageduser1@test.com",
			Role:      models.RoleUser,
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response dto.UserResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, manager.ID, *response.ManagerID)
	})

	t.Run("manager cannot create admin", func(t *testing.T) {
		manager := testutil.CreateTestManager(t)
		router, handler := setupUserRouter(&dto.AuthUser{
			ID:       manager.ID.String(),
			Username: manager.Username,
			Role:     manager.Role,
		})
		router.POST("/api/v1/users", handler.Create)

		body := dto.CreateUserRequest{
			Username:  "attemptadmin",
			Password:  "password123",
			FirstName: "Attempt",
			LastName:  "Admin",
			Email:     "attemptadmin@test.com",
			Role:      models.RoleAdmin,
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("fails with invalid body", func(t *testing.T) {
		admin := testutil.CreateTestAdmin(t)
		router, handler := setupUserRouter(&dto.AuthUser{
			ID:       admin.ID.String(),
			Username: admin.Username,
			Role:     admin.Role,
		})
		router.POST("/api/v1/users", handler.Create)

		req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("fails when user exists", func(t *testing.T) {
		admin := testutil.CreateTestAdmin(t)
		existingUser := testutil.CreateTestUserWithName(t, models.RoleUser, "existinguser")

		router, handler := setupUserRouter(&dto.AuthUser{
			ID:       admin.ID.String(),
			Username: admin.Username,
			Role:     admin.Role,
		})
		router.POST("/api/v1/users", handler.Create)

		body := dto.CreateUserRequest{
			Username:  existingUser.Username,
			Password:  "password123",
			FirstName: "Duplicate",
			LastName:  "User",
			Email:     "dup@test.com",
			Role:      models.RoleUser,
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})
}

func TestUserHandler_Get(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	t.Run("admin gets any user", func(t *testing.T) {
		admin := testutil.CreateTestAdmin(t)
		user := testutil.CreateTestRegularUser(t)

		router, handler := setupUserRouter(&dto.AuthUser{
			ID:       admin.ID.String(),
			Username: admin.Username,
			Role:     admin.Role,
		})
		router.GET("/api/v1/users/:id", handler.Get)

		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/users/%s", user.ID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.UserResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, user.ID, response.ID)
	})

	t.Run("user gets own profile", func(t *testing.T) {
		user := testutil.CreateTestRegularUser(t)

		router, handler := setupUserRouter(&dto.AuthUser{
			ID:       user.ID.String(),
			Username: user.Username,
			Role:     user.Role,
		})
		router.GET("/api/v1/users/:id", handler.Get)

		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/users/%s", user.ID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("user cannot get other user", func(t *testing.T) {
		user := testutil.CreateTestRegularUser(t)
		otherUser := testutil.CreateTestRegularUser(t)

		router, handler := setupUserRouter(&dto.AuthUser{
			ID:       user.ID.String(),
			Username: user.Username,
			Role:     user.Role,
		})
		router.GET("/api/v1/users/:id", handler.Get)

		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/users/%s", otherUser.ID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("manager gets subordinate", func(t *testing.T) {
		manager := testutil.CreateTestManager(t)
		subordinate := testutil.CreateManagedUser(t, manager.ID)

		router, handler := setupUserRouter(&dto.AuthUser{
			ID:       manager.ID.String(),
			Username: manager.Username,
			Role:     manager.Role,
		})
		router.GET("/api/v1/users/:id", handler.Get)

		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/users/%s", subordinate.ID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("manager cannot get non-subordinate", func(t *testing.T) {
		manager := testutil.CreateTestManager(t)
		otherUser := testutil.CreateTestRegularUser(t)

		router, handler := setupUserRouter(&dto.AuthUser{
			ID:       manager.ID.String(),
			Username: manager.Username,
			Role:     manager.Role,
		})
		router.GET("/api/v1/users/:id", handler.Get)

		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/users/%s", otherUser.ID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("returns 404 for non-existent user", func(t *testing.T) {
		admin := testutil.CreateTestAdmin(t)

		router, handler := setupUserRouter(&dto.AuthUser{
			ID:       admin.ID.String(),
			Username: admin.Username,
			Role:     admin.Role,
		})
		router.GET("/api/v1/users/:id", handler.Get)

		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/users/%s", uuid.New()), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("returns 400 for invalid UUID", func(t *testing.T) {
		admin := testutil.CreateTestAdmin(t)

		router, handler := setupUserRouter(&dto.AuthUser{
			ID:       admin.ID.String(),
			Username: admin.Username,
			Role:     admin.Role,
		})
		router.GET("/api/v1/users/:id", handler.Get)

		req, _ := http.NewRequest("GET", "/api/v1/users/invalid-uuid", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUserHandler_Update(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	t.Run("admin updates user", func(t *testing.T) {
		admin := testutil.CreateTestAdmin(t)
		user := testutil.CreateTestRegularUser(t)

		router, handler := setupUserRouter(&dto.AuthUser{
			ID:       admin.ID.String(),
			Username: admin.Username,
			Role:     admin.Role,
		})
		router.PUT("/api/v1/users/:id", handler.Update)

		body := dto.UpdateUserRequest{
			FirstName: "Updated",
			LastName:  "Name",
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/v1/users/%s", user.ID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.UserResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Updated", response.FirstName)
	})

	t.Run("user cannot update other user", func(t *testing.T) {
		user := testutil.CreateTestRegularUser(t)
		otherUser := testutil.CreateTestRegularUser(t)

		router, handler := setupUserRouter(&dto.AuthUser{
			ID:       user.ID.String(),
			Username: user.Username,
			Role:     user.Role,
		})
		router.PUT("/api/v1/users/:id", handler.Update)

		body := dto.UpdateUserRequest{
			FirstName: "Hacked",
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/v1/users/%s", otherUser.ID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("manager updates subordinate", func(t *testing.T) {
		manager := testutil.CreateTestManager(t)
		subordinate := testutil.CreateManagedUser(t, manager.ID)

		router, handler := setupUserRouter(&dto.AuthUser{
			ID:       manager.ID.String(),
			Username: manager.Username,
			Role:     manager.Role,
		})
		router.PUT("/api/v1/users/:id", handler.Update)

		body := dto.UpdateUserRequest{
			FirstName: "Updated",
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/v1/users/%s", subordinate.ID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("manager cannot assign admin role", func(t *testing.T) {
		manager := testutil.CreateTestManager(t)
		subordinate := testutil.CreateManagedUser(t, manager.ID)

		router, handler := setupUserRouter(&dto.AuthUser{
			ID:       manager.ID.String(),
			Username: manager.Username,
			Role:     manager.Role,
		})
		router.PUT("/api/v1/users/:id", handler.Update)

		body := dto.UpdateUserRequest{
			Role: models.RoleAdmin,
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/v1/users/%s", subordinate.ID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestUserHandler_Delete(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	t.Run("admin deletes user", func(t *testing.T) {
		admin := testutil.CreateTestAdmin(t)
		user := testutil.CreateTestRegularUser(t)

		router, handler := setupUserRouter(&dto.AuthUser{
			ID:       admin.ID.String(),
			Username: admin.Username,
			Role:     admin.Role,
		})
		router.DELETE("/api/v1/users/:id", handler.Delete)

		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/users/%s", user.ID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("returns 404 for non-existent user", func(t *testing.T) {
		admin := testutil.CreateTestAdmin(t)

		router, handler := setupUserRouter(&dto.AuthUser{
			ID:       admin.ID.String(),
			Username: admin.Username,
			Role:     admin.Role,
		})
		router.DELETE("/api/v1/users/:id", handler.Delete)

		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/users/%s", uuid.New()), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestUserHandler_List(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	t.Run("admin lists all users", func(t *testing.T) {
		admin := testutil.CreateTestAdmin(t)
		testutil.CreateTestRegularUser(t)
		testutil.CreateTestRegularUser(t)

		router, handler := setupUserRouter(&dto.AuthUser{
			ID:       admin.ID.String(),
			Username: admin.Username,
			Role:     admin.Role,
		})
		router.GET("/api/v1/users", handler.List)

		req, _ := http.NewRequest("GET", "/api/v1/users", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.UserListResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(response.Users), 2)
	})

	t.Run("pagination works", func(t *testing.T) {
		admin := testutil.CreateTestAdmin(t)

		router, handler := setupUserRouter(&dto.AuthUser{
			ID:       admin.ID.String(),
			Username: admin.Username,
			Role:     admin.Role,
		})
		router.GET("/api/v1/users", handler.List)

		req, _ := http.NewRequest("GET", "/api/v1/users?page=1&page_size=2", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.UserListResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(response.Users), 2)
		assert.Equal(t, 2, response.PageSize)
	})
}

func TestUserHandler_UpdateProfile(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	t.Run("user updates own profile", func(t *testing.T) {
		user := testutil.CreateTestRegularUser(t)

		router, handler := setupUserRouter(&dto.AuthUser{
			ID:       user.ID.String(),
			Username: user.Username,
			Role:     user.Role,
		})
		router.PUT("/api/v1/users/profile", handler.UpdateProfile)

		body := dto.UpdateProfileRequest{
			FirstName: "NewFirst",
			LastName:  "NewLast",
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("PUT", "/api/v1/users/profile", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.UserResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "NewFirst", response.FirstName)
		assert.Equal(t, "NewLast", response.LastName)
	})
}

func TestUserHandler_UpdatePassword(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	t.Run("user updates password", func(t *testing.T) {
		user := testutil.CreateTestUserWithName(t, models.RoleUser, "pwduser")

		router, handler := setupUserRouter(&dto.AuthUser{
			ID:       user.ID.String(),
			Username: user.Username,
			Role:     user.Role,
		})
		router.PUT("/api/v1/users/password", handler.UpdatePassword)

		body := dto.UpdatePasswordRequest{
			CurrentPassword: "testpassword123",
			NewPassword:     "newpassword456",
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("PUT", "/api/v1/users/password", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("fails with wrong current password", func(t *testing.T) {
		user := testutil.CreateTestUserWithName(t, models.RoleUser, "pwduser2")

		router, handler := setupUserRouter(&dto.AuthUser{
			ID:       user.ID.String(),
			Username: user.Username,
			Role:     user.Role,
		})
		router.PUT("/api/v1/users/password", handler.UpdatePassword)

		body := dto.UpdatePasswordRequest{
			CurrentPassword: "wrongpassword",
			NewPassword:     "newpassword456",
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("PUT", "/api/v1/users/password", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUserHandler_Groups(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	t.Run("get user groups", func(t *testing.T) {
		admin := testutil.CreateTestAdmin(t)
		user := testutil.CreateTestRegularUser(t)
		group := testutil.CreateTestGroup(t, admin.ID)

		// Add user to group using UserGroup join table
		userGroup := &models.UserGroup{
			UserID:    user.ID,
			GroupID:   group.ID,
			CreatedBy: admin.ID,
		}
		testutil.TestDB.Create(userGroup)

		router, handler := setupUserRouter(&dto.AuthUser{
			ID:       admin.ID.String(),
			Username: admin.Username,
			Role:     admin.Role,
		})
		router.GET("/api/v1/users/:id/groups", handler.GetGroups)

		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/users/%s/groups", user.ID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.UserGroupsResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, 1, len(response.Groups))
	})

	t.Run("add user to group", func(t *testing.T) {
		admin := testutil.CreateTestAdmin(t)
		user := testutil.CreateTestRegularUser(t)
		group := testutil.CreateTestGroup(t, admin.ID)

		router, handler := setupUserRouter(&dto.AuthUser{
			ID:       admin.ID.String(),
			Username: admin.Username,
			Role:     admin.Role,
		})
		router.POST("/api/v1/users/:id/groups", handler.AddGroup)

		body := dto.AddUserGroupRequest{
			GroupID: group.ID,
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/users/%s/groups", user.ID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("remove user from group", func(t *testing.T) {
		admin := testutil.CreateTestAdmin(t)
		user := testutil.CreateTestRegularUser(t)
		group := testutil.CreateTestGroup(t, admin.ID)

		// Add user to group first using UserGroup join table
		userGroup := &models.UserGroup{
			UserID:    user.ID,
			GroupID:   group.ID,
			CreatedBy: admin.ID,
		}
		testutil.TestDB.Create(userGroup)

		router, handler := setupUserRouter(&dto.AuthUser{
			ID:       admin.ID.String(),
			Username: admin.Username,
			Role:     admin.Role,
		})
		router.DELETE("/api/v1/users/:id/groups/:group_id", handler.RemoveGroup)

		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/users/%s/groups/%s", user.ID, group.ID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
