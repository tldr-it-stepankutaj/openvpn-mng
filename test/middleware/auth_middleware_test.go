package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/config"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/dto"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/middleware"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/models"
)

func createTestToken(userID uuid.UUID, username string, role models.Role, secret string, expiry time.Duration) string {
	claims := &middleware.Claims{
		UserID:   userID.String(),
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.AuthConfig{
		JWTSecret:   "test-secret-key-for-testing-minimum-32-chars",
		TokenExpiry: 24,
	}

	t.Run("valid bearer token", func(t *testing.T) {
		userID := uuid.New()
		token := createTestToken(userID, "testuser", models.RoleUser, cfg.JWTSecret, time.Hour)

		router := gin.New()
		router.Use(middleware.AuthMiddleware(cfg, nil))
		router.GET("/test", func(c *gin.Context) {
			authUser := middleware.GetAuthUser(c)
			c.JSON(http.StatusOK, gin.H{
				"user_id":  authUser.ID,
				"username": authUser.Username,
				"role":     authUser.Role,
			})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("valid cookie token", func(t *testing.T) {
		userID := uuid.New()
		token := createTestToken(userID, "testuser", models.RoleUser, cfg.JWTSecret, time.Hour)

		router := gin.New()
		router.Use(middleware.AuthMiddleware(cfg, nil))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		req.AddCookie(&http.Cookie{Name: "token", Value: token})
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("missing token", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware.AuthMiddleware(cfg, nil))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid token", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware.AuthMiddleware(cfg, nil))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("expired token", func(t *testing.T) {
		userID := uuid.New()
		token := createTestToken(userID, "testuser", models.RoleUser, cfg.JWTSecret, -time.Hour)

		router := gin.New()
		router.Use(middleware.AuthMiddleware(cfg, nil))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("wrong secret", func(t *testing.T) {
		userID := uuid.New()
		token := createTestToken(userID, "testuser", models.RoleUser, "wrong-secret", time.Hour)

		router := gin.New()
		router.Use(middleware.AuthMiddleware(cfg, nil))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestRequireRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	setupRouter := func(roles ...models.Role) *gin.Engine {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			// Simulate authenticated user
			roleHeader := c.GetHeader("X-Test-Role")
			role := models.RoleUser
			if roleHeader != "" {
				role = models.Role(roleHeader)
			}
			authUser := &dto.AuthUser{
				ID:       uuid.New().String(),
				Username: "testuser",
				Role:     role,
			}
			c.Set(middleware.AuthUserKey, authUser)
			c.Next()
		})
		router.Use(middleware.RequireRole(roles...))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
		return router
	}

	t.Run("allows matching role", func(t *testing.T) {
		router := setupRouter(models.RoleAdmin)

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Test-Role", string(models.RoleAdmin))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("allows any of multiple roles", func(t *testing.T) {
		router := setupRouter(models.RoleAdmin, models.RoleManager)

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Test-Role", string(models.RoleManager))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("denies non-matching role", func(t *testing.T) {
		router := setupRouter(models.RoleAdmin)

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Test-Role", string(models.RoleUser))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestRequireAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	setupRouter := func(role models.Role) *gin.Engine {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set(middleware.AuthUserKey, &dto.AuthUser{
				ID:       uuid.New().String(),
				Username: "testuser",
				Role:     role,
			})
			c.Next()
		})
		router.Use(middleware.RequireAdmin())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
		return router
	}

	t.Run("allows admin", func(t *testing.T) {
		router := setupRouter(models.RoleAdmin)

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("denies manager", func(t *testing.T) {
		router := setupRouter(models.RoleManager)

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("denies user", func(t *testing.T) {
		router := setupRouter(models.RoleUser)

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestRequireManagerOrAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	setupRouter := func(role models.Role) *gin.Engine {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set(middleware.AuthUserKey, &dto.AuthUser{
				ID:       uuid.New().String(),
				Username: "testuser",
				Role:     role,
			})
			c.Next()
		})
		router.Use(middleware.RequireManagerOrAdmin())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
		return router
	}

	t.Run("allows admin", func(t *testing.T) {
		router := setupRouter(models.RoleAdmin)

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("allows manager", func(t *testing.T) {
		router := setupRouter(models.RoleManager)

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("denies user", func(t *testing.T) {
		router := setupRouter(models.RoleUser)

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestGetAuthUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns auth user when set", func(t *testing.T) {
		router := gin.New()
		router.GET("/test", func(c *gin.Context) {
			expectedUser := &dto.AuthUser{
				ID:       uuid.New().String(),
				Username: "testuser",
				Role:     models.RoleUser,
			}
			c.Set(middleware.AuthUserKey, expectedUser)

			authUser := middleware.GetAuthUser(c)
			require.NotNil(t, authUser)
			assert.Equal(t, expectedUser.ID, authUser.ID)
			assert.Equal(t, expectedUser.Username, authUser.Username)
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("returns nil when not set", func(t *testing.T) {
		router := gin.New()
		router.GET("/test", func(c *gin.Context) {
			authUser := middleware.GetAuthUser(c)
			assert.Nil(t, authUser)
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestGetAuthUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns user ID when valid", func(t *testing.T) {
		expectedID := uuid.New()

		router := gin.New()
		router.GET("/test", func(c *gin.Context) {
			c.Set(middleware.AuthUserKey, &dto.AuthUser{
				ID:       expectedID.String(),
				Username: "testuser",
				Role:     models.RoleUser,
			})

			userID := middleware.GetAuthUserID(c)
			assert.Equal(t, expectedID, userID)
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("returns nil UUID when not authenticated", func(t *testing.T) {
		router := gin.New()
		router.GET("/test", func(c *gin.Context) {
			userID := middleware.GetAuthUserID(c)
			assert.Equal(t, uuid.Nil, userID)
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("returns nil UUID for invalid ID format", func(t *testing.T) {
		router := gin.New()
		router.GET("/test", func(c *gin.Context) {
			c.Set(middleware.AuthUserKey, &dto.AuthUser{
				ID:       "invalid-uuid",
				Username: "testuser",
				Role:     models.RoleUser,
			})

			userID := middleware.GetAuthUserID(c)
			assert.Equal(t, uuid.Nil, userID)
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
