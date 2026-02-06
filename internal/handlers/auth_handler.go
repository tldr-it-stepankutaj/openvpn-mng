package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/apperror"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/config"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/dto"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/middleware"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/services"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	authService *services.AuthService
	auditLogger *middleware.AuditLogger
	config      *config.AuthConfig
	blacklist   *middleware.TokenBlacklist
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(cfg *config.AuthConfig, blacklist *middleware.TokenBlacklist, securityCfg ...*config.SecurityConfig) *AuthHandler {
	var authService *services.AuthService
	if len(securityCfg) > 0 && securityCfg[0] != nil {
		authService = services.NewAuthServiceWithSecurity(cfg, securityCfg[0])
	} else {
		authService = services.NewAuthService(cfg)
	}
	return &AuthHandler{
		authService: authService,
		auditLogger: middleware.NewAuditLogger(),
		config:      cfg,
		blacklist:   blacklist,
	}
}

// Login godoc
// @Summary Login user
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.LoginResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	token, user, err := h.authService.Authenticate(req.Username, req.Password)
	if err != nil {
		var appErr *apperror.AppError
		if errors.As(err, &appErr) && appErr.Code == http.StatusTooManyRequests {
			c.Header("Retry-After", strconv.Itoa(h.config.SessionExpiry*60))
		}
		apperror.HandleError(c, err)
		return
	}

	// Log login
	err = h.auditLogger.LogLogin(c, user.ID, "Successful login")
	if err != nil {
		return
	}

	// Set cookie for web frontend with SameSite=Lax for CSRF protection
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("token", token, h.config.SessionExpiry*3600, "/", "", false, true)

	c.JSON(http.StatusOK, dto.LoginResponse{
		Token:     token,
		ExpiresIn: h.config.TokenExpiry * 3600,
		User:      dto.ToUserResponse(user),
	})
}

// Logout godoc
// @Summary Logout user
// @Description Logout user and invalidate session
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.SuccessResponse
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	authUser := middleware.GetAuthUser(c)
	if authUser != nil {
		userID := middleware.GetAuthUserID(c)
		err := h.auditLogger.LogLogout(c, userID)
		if err != nil {
			return
		}
	}

	// Blacklist the token so it cannot be reused
	if h.blacklist != nil {
		rawToken := ""
		if authHeader := c.GetHeader("Authorization"); authHeader != "" {
			rawToken = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			rawToken, _ = c.Cookie("token")
		}
		if rawToken != "" {
			// Parse expiry from token to know when to remove from blacklist
			parser := jwt.NewParser()
			claims := &middleware.Claims{}
			_, _, _ = parser.ParseUnverified(rawToken, claims)
			if claims.ExpiresAt != nil {
				h.blacklist.Add(rawToken, claims.ExpiresAt.Time)
			}
		}
	}

	// Clear cookie
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("token", "", -1, "/", "", false, true)

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Successfully logged out",
	})
}

// Me godoc
// @Summary Get current user
// @Description Get the currently authenticated user
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.UserResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/v1/auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	authUser := middleware.GetAuthUser(c)
	if authUser == nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "Unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	user, err := services.GetUserByID(middleware.GetAuthUserID(c))
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Not Found",
			Message: "User not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, dto.ToUserResponse(user))
}
