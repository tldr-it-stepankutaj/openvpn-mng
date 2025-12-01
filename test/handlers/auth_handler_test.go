package handlers_test

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
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/services"
	"github.com/tldr-it-stepankutaj/openvpn-mng/test/testutil"
)

func TestAuthHandler_Login(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	gin.SetMode(gin.TestMode)

	cfg := &config.AuthConfig{
		JWTSecret:     "test-secret-key-for-testing-minimum-32-chars",
		TokenExpiry:   24,
		SessionExpiry: 24,
	}
	handler := handlers.NewAuthHandler(cfg)

	t.Run("successful login", func(t *testing.T) {
		user := testutil.CreateTestUserWithName(t, models.RoleUser, "logintest")

		router := gin.New()
		router.POST("/api/v1/auth/login", handler.Login)

		body := dto.LoginRequest{
			Username: "logintest",
			Password: "testpassword123",
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.LoginResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.NotEmpty(t, response.Token)
		assert.Equal(t, user.Username, response.User.Username)
	})

	t.Run("invalid credentials", func(t *testing.T) {
		testutil.CreateTestUserWithName(t, models.RoleUser, "logintest2")

		router := gin.New()
		router.POST("/api/v1/auth/login", handler.Login)

		body := dto.LoginRequest{
			Username: "logintest2",
			Password: "wrongpassword",
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("non-existent user", func(t *testing.T) {
		router := gin.New()
		router.POST("/api/v1/auth/login", handler.Login)

		body := dto.LoginRequest{
			Username: "nonexistent",
			Password: "password123",
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid request body", func(t *testing.T) {
		router := gin.New()
		router.POST("/api/v1/auth/login", handler.Login)

		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("inactive user", func(t *testing.T) {
		// Create user first, then deactivate to bypass GORM default
		hashedPassword, _ := services.HashPassword("testpassword123")
		admin := testutil.CreateTestAdmin(t)
		inactiveUser := &models.User{
			Username:  "inactivelogin2",
			Password:  hashedPassword,
			FirstName: "Inactive",
			LastName:  "User",
			Email:     "inactivelogin2@test.com",
			Role:      models.RoleUser,
			IsActive:  true,
			CreatedBy: admin.ID,
		}
		testutil.TestDB.Create(inactiveUser)
		// Deactivate user after creation
		testutil.TestDB.Model(inactiveUser).Update("is_active", false)

		router := gin.New()
		router.POST("/api/v1/auth/login", handler.Login)

		body := dto.LoginRequest{
			Username: "inactivelogin2",
			Password: "testpassword123",
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response dto.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response.Message, "inactive")
	})
}

func TestAuthHandler_Logout(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	gin.SetMode(gin.TestMode)

	cfg := &config.AuthConfig{
		JWTSecret:     "test-secret-key-for-testing-minimum-32-chars",
		TokenExpiry:   24,
		SessionExpiry: 24,
	}
	handler := handlers.NewAuthHandler(cfg)

	t.Run("successful logout with auth", func(t *testing.T) {
		user := testutil.CreateTestUserWithName(t, models.RoleUser, "logouttest")

		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set(middleware.AuthUserKey, &dto.AuthUser{
				ID:       user.ID.String(),
				Username: user.Username,
				Role:     user.Role,
			})
			c.Next()
		})
		router.POST("/api/v1/auth/logout", handler.Logout)

		req, _ := http.NewRequest("POST", "/api/v1/auth/logout", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.SuccessResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response.Message, "logged out")
	})

	t.Run("logout without auth still succeeds", func(t *testing.T) {
		router := gin.New()
		router.POST("/api/v1/auth/logout", handler.Logout)

		req, _ := http.NewRequest("POST", "/api/v1/auth/logout", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestAuthHandler_Me(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	gin.SetMode(gin.TestMode)

	cfg := &config.AuthConfig{
		JWTSecret:     "test-secret-key-for-testing-minimum-32-chars",
		TokenExpiry:   24,
		SessionExpiry: 24,
	}
	handler := handlers.NewAuthHandler(cfg)

	t.Run("returns current user", func(t *testing.T) {
		user := testutil.CreateTestUserWithName(t, models.RoleUser, "metest")

		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set(middleware.AuthUserKey, &dto.AuthUser{
				ID:       user.ID.String(),
				Username: user.Username,
				Role:     user.Role,
			})
			c.Next()
		})
		router.GET("/api/v1/auth/me", handler.Me)

		req, _ := http.NewRequest("GET", "/api/v1/auth/me", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.UserResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, user.ID, response.ID)
		assert.Equal(t, user.Username, response.Username)
	})

	t.Run("returns unauthorized without auth", func(t *testing.T) {
		router := gin.New()
		router.GET("/api/v1/auth/me", handler.Me)

		req, _ := http.NewRequest("GET", "/api/v1/auth/me", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
