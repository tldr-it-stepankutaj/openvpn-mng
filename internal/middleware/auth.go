package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/config"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/dto"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/models"
)

const (
	AuthUserKey = "auth_user"
)

// Claims represents JWT claims
type Claims struct {
	UserID   string      `json:"user_id"`
	Username string      `json:"username"`
	Role     models.Role `json:"role"`
	jwt.RegisteredClaims
}

// AuthMiddleware creates authentication middleware
func AuthMiddleware(cfg *config.AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// Try to get from cookie (for web frontend)
			authHeader, _ = c.Cookie("token")
			if authHeader == "" {
				c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
					Error:   "Unauthorized",
					Message: "Missing authentication token",
					Code:    http.StatusUnauthorized,
				})
				c.Abort()
				return
			}
		} else {
			// Remove "Bearer " prefix
			if strings.HasPrefix(authHeader, "Bearer ") {
				authHeader = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		// Parse token
		token, err := jwt.ParseWithClaims(authHeader, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Invalid or expired token",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*Claims)
		if !ok {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Invalid token claims",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// Set user in context
		c.Set(AuthUserKey, &dto.AuthUser{
			ID:       claims.UserID,
			Username: claims.Username,
			Role:     claims.Role,
		})

		c.Next()
	}
}

// RequireRole creates middleware that requires specific roles
func RequireRole(roles ...models.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		authUser, exists := c.Get(AuthUserKey)
		if !exists {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error:   "Unauthorized",
				Message: "User not authenticated",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		user := authUser.(*dto.AuthUser)
		for _, role := range roles {
			if user.Role == role {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "Forbidden",
			Message: "Insufficient permissions",
			Code:    http.StatusForbidden,
		})
		c.Abort()
	}
}

// RequireAdmin creates middleware that requires admin role
func RequireAdmin() gin.HandlerFunc {
	return RequireRole(models.RoleAdmin)
}

// RequireManagerOrAdmin creates middleware that requires manager or admin role
func RequireManagerOrAdmin() gin.HandlerFunc {
	return RequireRole(models.RoleManager, models.RoleAdmin)
}

// GetAuthUser returns the authenticated user from context
func GetAuthUser(c *gin.Context) *dto.AuthUser {
	authUser, exists := c.Get(AuthUserKey)
	if !exists {
		return nil
	}
	return authUser.(*dto.AuthUser)
}

// GetAuthUserID returns the authenticated user ID as UUID
func GetAuthUserID(c *gin.Context) uuid.UUID {
	authUser := GetAuthUser(c)
	if authUser == nil {
		return uuid.Nil
	}
	id, err := uuid.Parse(authUser.ID)
	if err != nil {
		return uuid.Nil
	}
	return id
}
