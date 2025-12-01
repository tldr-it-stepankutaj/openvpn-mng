package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/config"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/dto"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/handlers"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/middleware"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/models"
	"github.com/tldr-it-stepankutaj/openvpn-mng/test/testutil"
)

func setupIntegrationRouter(cfg *config.AuthConfig) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery())

	authHandler := handlers.NewAuthHandler(cfg)
	userHandler := handlers.NewUserHandler()
	groupHandler := handlers.NewGroupHandler()
	networkHandler := handlers.NewNetworkHandler()

	// Public routes
	router.POST("/api/v1/auth/login", authHandler.Login)

	// Protected routes
	protected := router.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(cfg))
	{
		protected.POST("/auth/logout", authHandler.Logout)
		protected.GET("/auth/me", authHandler.Me)

		// User routes
		users := protected.Group("/users")
		{
			users.GET("", userHandler.List)
			users.POST("", middleware.RequireManagerOrAdmin(), userHandler.Create)
			users.GET("/:id", userHandler.Get)
			users.PUT("/:id", middleware.RequireManagerOrAdmin(), userHandler.Update)
			users.DELETE("/:id", middleware.RequireAdmin(), userHandler.Delete)
			users.PUT("/profile", userHandler.UpdateProfile)
			users.PUT("/password", userHandler.UpdatePassword)
			users.GET("/:id/groups", userHandler.GetGroups)
			users.POST("/:id/groups", middleware.RequireManagerOrAdmin(), userHandler.AddGroup)
			users.DELETE("/:id/groups/:group_id", middleware.RequireManagerOrAdmin(), userHandler.RemoveGroup)
		}

		// Group routes
		groups := protected.Group("/groups")
		groups.Use(middleware.RequireManagerOrAdmin())
		{
			groups.GET("", groupHandler.List)
			groups.POST("", groupHandler.Create)
			groups.GET("/:id", groupHandler.Get)
			groups.PUT("/:id", groupHandler.Update)
			groups.DELETE("/:id", middleware.RequireAdmin(), groupHandler.Delete)
			groups.GET("/:id/users", groupHandler.GetUsers)
			groups.GET("/:id/networks", groupHandler.GetNetworks)
			groups.POST("/:id/networks", groupHandler.AddNetwork)
			groups.DELETE("/:id/networks/:network_id", groupHandler.RemoveNetwork)
		}

		// Network routes
		networks := protected.Group("/networks")
		networks.Use(middleware.RequireManagerOrAdmin())
		{
			networks.GET("", networkHandler.List)
			networks.POST("", networkHandler.Create)
			networks.GET("/:id", networkHandler.Get)
			networks.PUT("/:id", networkHandler.Update)
			networks.DELETE("/:id", middleware.RequireAdmin(), networkHandler.Delete)
		}
	}

	return router
}

func TestIntegration_AuthFlow(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	cfg := &config.AuthConfig{
		JWTSecret:     "test-secret-key-for-testing-minimum-32-chars",
		TokenExpiry:   24,
		SessionExpiry: 24,
	}

	router := setupIntegrationRouter(cfg)

	t.Run("full authentication flow", func(t *testing.T) {
		// Create a test user
		user := testutil.CreateTestUserWithName(t, models.RoleUser, "integrationuser")

		// Step 1: Login
		loginBody := dto.LoginRequest{
			Username: "integrationuser",
			Password: "testpassword123",
		}
		jsonBody, _ := json.Marshal(loginBody)

		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var loginResponse dto.LoginResponse
		err := json.Unmarshal(w.Body.Bytes(), &loginResponse)
		require.NoError(t, err)
		assert.NotEmpty(t, loginResponse.Token)
		assert.Equal(t, user.Username, loginResponse.User.Username)

		token := loginResponse.Token

		// Step 2: Get current user
		req, _ = http.NewRequest("GET", "/api/v1/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var meResponse dto.UserResponse
		err = json.Unmarshal(w.Body.Bytes(), &meResponse)
		require.NoError(t, err)
		assert.Equal(t, user.ID, meResponse.ID)

		// Step 3: Logout
		req, _ = http.NewRequest("POST", "/api/v1/auth/logout", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
	})
}

func TestIntegration_AdminUserManagement(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	cfg := &config.AuthConfig{
		JWTSecret:     "test-secret-key-for-testing-minimum-32-chars",
		TokenExpiry:   24,
		SessionExpiry: 24,
	}

	router := setupIntegrationRouter(cfg)

	t.Run("admin creates and manages user", func(t *testing.T) {
		// Create admin user
		admin := testutil.CreateTestUserWithName(t, models.RoleAdmin, "adminuser")

		// Login as admin
		loginBody := dto.LoginRequest{
			Username: "adminuser",
			Password: "testpassword123",
		}
		jsonBody, _ := json.Marshal(loginBody)

		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var loginResponse dto.LoginResponse
		json.Unmarshal(w.Body.Bytes(), &loginResponse)
		token := loginResponse.Token

		// Step 1: Create user
		createBody := dto.CreateUserRequest{
			Username:  "newuser",
			Password:  "newpassword123",
			FirstName: "New",
			LastName:  "User",
			Email:     "newuser@test.com",
			Role:      models.RoleUser,
		}
		jsonBody, _ = json.Marshal(createBody)

		req, _ = http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusCreated, w.Code)

		var createdUser dto.UserResponse
		json.Unmarshal(w.Body.Bytes(), &createdUser)
		assert.Equal(t, "newuser", createdUser.Username)

		// Step 2: Update user
		updateBody := dto.UpdateUserRequest{
			FirstName: "Updated",
		}
		jsonBody, _ = json.Marshal(updateBody)

		req, _ = http.NewRequest("PUT", "/api/v1/users/"+createdUser.ID.String(), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var updatedUser dto.UserResponse
		json.Unmarshal(w.Body.Bytes(), &updatedUser)
		assert.Equal(t, "Updated", updatedUser.FirstName)

		// Step 3: List users
		req, _ = http.NewRequest("GET", "/api/v1/users", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var listResponse dto.UserListResponse
		json.Unmarshal(w.Body.Bytes(), &listResponse)
		assert.GreaterOrEqual(t, len(listResponse.Users), 2) // admin + created user

		// Step 4: Delete user
		req, _ = http.NewRequest("DELETE", "/api/v1/users/"+createdUser.ID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		// Verify user is deleted
		req, _ = http.NewRequest("GET", "/api/v1/users/"+createdUser.ID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusNotFound, w.Code)

		_ = admin // silence unused warning
	})
}

func TestIntegration_GroupNetworkManagement(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	cfg := &config.AuthConfig{
		JWTSecret:     "test-secret-key-for-testing-minimum-32-chars",
		TokenExpiry:   24,
		SessionExpiry: 24,
	}

	router := setupIntegrationRouter(cfg)

	t.Run("admin manages groups and networks", func(t *testing.T) {
		// Create and login as admin
		testutil.CreateTestUserWithName(t, models.RoleAdmin, "groupadmin")

		loginBody := dto.LoginRequest{
			Username: "groupadmin",
			Password: "testpassword123",
		}
		jsonBody, _ := json.Marshal(loginBody)

		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var loginResponse dto.LoginResponse
		json.Unmarshal(w.Body.Bytes(), &loginResponse)
		token := loginResponse.Token

		// Step 1: Create network
		createNetworkBody := dto.CreateNetworkRequest{
			Name:        "Test Network",
			CIDR:        "10.0.0.0/24",
			Description: "Test network for integration",
		}
		jsonBody, _ = json.Marshal(createNetworkBody)

		req, _ = http.NewRequest("POST", "/api/v1/networks", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusCreated, w.Code)

		var network dto.NetworkResponse
		json.Unmarshal(w.Body.Bytes(), &network)
		assert.Equal(t, "Test Network", network.Name)

		// Step 2: Create group
		createGroupBody := dto.CreateGroupRequest{
			Name:        "Test Group",
			Description: "Test group for integration",
		}
		jsonBody, _ = json.Marshal(createGroupBody)

		req, _ = http.NewRequest("POST", "/api/v1/groups", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusCreated, w.Code)

		var group dto.GroupResponse
		json.Unmarshal(w.Body.Bytes(), &group)
		assert.Equal(t, "Test Group", group.Name)

		// Step 3: Add network to group
		addNetworkBody := dto.AddNetworkToGroupRequest{
			NetworkID: network.ID,
		}
		jsonBody, _ = json.Marshal(addNetworkBody)

		req, _ = http.NewRequest("POST", "/api/v1/groups/"+group.ID.String()+"/networks", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		// Step 4: Get group networks
		req, _ = http.NewRequest("GET", "/api/v1/groups/"+group.ID.String()+"/networks", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var networks []dto.NetworkResponse
		json.Unmarshal(w.Body.Bytes(), &networks)
		require.Len(t, networks, 1)
		assert.Equal(t, network.ID, networks[0].ID)

		// Step 5: Remove network from group
		req, _ = http.NewRequest("DELETE", "/api/v1/groups/"+group.ID.String()+"/networks/"+network.ID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
	})
}

func TestIntegration_RoleBasedAccess(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	cfg := &config.AuthConfig{
		JWTSecret:     "test-secret-key-for-testing-minimum-32-chars",
		TokenExpiry:   24,
		SessionExpiry: 24,
	}

	router := setupIntegrationRouter(cfg)

	t.Run("user cannot create users", func(t *testing.T) {
		// Create and login as regular user
		testutil.CreateTestUserWithName(t, models.RoleUser, "regularuser")

		loginBody := dto.LoginRequest{
			Username: "regularuser",
			Password: "testpassword123",
		}
		jsonBody, _ := json.Marshal(loginBody)

		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var loginResponse dto.LoginResponse
		json.Unmarshal(w.Body.Bytes(), &loginResponse)
		token := loginResponse.Token

		// Try to create user
		createBody := dto.CreateUserRequest{
			Username:  "attemptuser",
			Password:  "password123",
			FirstName: "Attempt",
			LastName:  "User",
			Email:     "attempt@test.com",
			Role:      models.RoleUser,
		}
		jsonBody, _ = json.Marshal(createBody)

		req, _ = http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("user cannot access groups", func(t *testing.T) {
		// Create and login as regular user
		testutil.CreateTestUserWithName(t, models.RoleUser, "regularuser2")

		loginBody := dto.LoginRequest{
			Username: "regularuser2",
			Password: "testpassword123",
		}
		jsonBody, _ := json.Marshal(loginBody)

		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var loginResponse dto.LoginResponse
		json.Unmarshal(w.Body.Bytes(), &loginResponse)
		token := loginResponse.Token

		// Try to list groups
		req, _ = http.NewRequest("GET", "/api/v1/groups", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("manager cannot delete users", func(t *testing.T) {
		// Create manager and regular user
		testutil.CreateTestUserWithName(t, models.RoleManager, "manageruser")
		userToDelete := testutil.CreateTestRegularUser(t)

		loginBody := dto.LoginRequest{
			Username: "manageruser",
			Password: "testpassword123",
		}
		jsonBody, _ := json.Marshal(loginBody)

		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var loginResponse dto.LoginResponse
		json.Unmarshal(w.Body.Bytes(), &loginResponse)
		token := loginResponse.Token

		// Try to delete user
		req, _ = http.NewRequest("DELETE", "/api/v1/users/"+userToDelete.ID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestIntegration_UserProfileUpdate(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	cfg := &config.AuthConfig{
		JWTSecret:     "test-secret-key-for-testing-minimum-32-chars",
		TokenExpiry:   24,
		SessionExpiry: 24,
	}

	router := setupIntegrationRouter(cfg)

	t.Run("user updates own profile", func(t *testing.T) {
		// Create and login as regular user
		testutil.CreateTestUserWithName(t, models.RoleUser, "profileuser")

		loginBody := dto.LoginRequest{
			Username: "profileuser",
			Password: "testpassword123",
		}
		jsonBody, _ := json.Marshal(loginBody)

		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var loginResponse dto.LoginResponse
		json.Unmarshal(w.Body.Bytes(), &loginResponse)
		token := loginResponse.Token

		// Update profile
		updateBody := dto.UpdateProfileRequest{
			FirstName: "UpdatedFirst",
			LastName:  "UpdatedLast",
			Email:     "updated@test.com",
		}
		jsonBody, _ = json.Marshal(updateBody)

		req, _ = http.NewRequest("PUT", "/api/v1/users/profile", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var updatedUser dto.UserResponse
		json.Unmarshal(w.Body.Bytes(), &updatedUser)
		assert.Equal(t, "UpdatedFirst", updatedUser.FirstName)
		assert.Equal(t, "UpdatedLast", updatedUser.LastName)
		assert.Equal(t, "updated@test.com", updatedUser.Email)
	})

	t.Run("user changes password", func(t *testing.T) {
		// Create and login as regular user
		testutil.CreateTestUserWithName(t, models.RoleUser, "pwdchangeuser")

		loginBody := dto.LoginRequest{
			Username: "pwdchangeuser",
			Password: "testpassword123",
		}
		jsonBody, _ := json.Marshal(loginBody)

		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var loginResponse dto.LoginResponse
		json.Unmarshal(w.Body.Bytes(), &loginResponse)
		token := loginResponse.Token

		// Change password
		passwordBody := dto.UpdatePasswordRequest{
			CurrentPassword: "testpassword123",
			NewPassword:     "newpassword456",
		}
		jsonBody, _ = json.Marshal(passwordBody)

		req, _ = http.NewRequest("PUT", "/api/v1/users/password", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w = httptest.NewRecorder()

		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		// Verify can login with new password
		loginBody = dto.LoginRequest{
			Username: "pwdchangeuser",
			Password: "newpassword456",
		}
		jsonBody, _ = json.Marshal(loginBody)

		req, _ = http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
