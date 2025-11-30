package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/dto"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/middleware"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/models"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/services"
)

// UserHandler handles user requests
type UserHandler struct {
	userService *services.UserService
	auditLogger *middleware.AuditLogger
}

// NewUserHandler creates a new user handler
func NewUserHandler() *UserHandler {
	return &UserHandler{
		userService: services.NewUserService(),
		auditLogger: middleware.NewAuditLogger(),
	}
}

// Create godoc
// @Summary Create user
// @Description Create a new user (Admin only)
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateUserRequest true "User data"
// @Success 201 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /api/v1/users [post]
func (h *UserHandler) Create(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	createdBy := middleware.GetAuthUserID(c)
	user, err := h.userService.Create(&req, createdBy)
	if err != nil {
		if err == services.ErrUserExists {
			c.JSON(http.StatusConflict, dto.ErrorResponse{
				Error:   "Conflict",
				Message: "User already exists",
				Code:    http.StatusConflict,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	h.auditLogger.LogCreate(c, "user", user.ID, user)

	c.JSON(http.StatusCreated, dto.ToUserResponse(user))
}

// Get godoc
// @Summary Get user
// @Description Get a user by ID
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/users/{id} [get]
func (h *UserHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Check access
	authUser := middleware.GetAuthUser(c)
	authUserID := middleware.GetAuthUserID(c)

	user, err := h.userService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Not Found",
			Message: "User not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	// Check permissions
	if authUser.Role == models.RoleUser && authUserID != user.ID {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "Forbidden",
			Message: "Access denied",
			Code:    http.StatusForbidden,
		})
		return
	}

	if authUser.Role == models.RoleManager {
		if authUserID != user.ID && (user.ManagerID == nil || *user.ManagerID != authUserID) {
			c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Error:   "Forbidden",
				Message: "Access denied",
				Code:    http.StatusForbidden,
			})
			return
		}
	}

	c.JSON(http.StatusOK, dto.ToUserResponse(user))
}

// Update godoc
// @Summary Update user
// @Description Update a user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param request body dto.UpdateUserRequest true "User data"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/users/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Get old values for audit
	oldUser, err := h.userService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Not Found",
			Message: "User not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	// Check permissions
	authUser := middleware.GetAuthUser(c)
	authUserID := middleware.GetAuthUserID(c)

	if authUser.Role == models.RoleManager {
		if oldUser.ManagerID == nil || *oldUser.ManagerID != authUserID {
			c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Error:   "Forbidden",
				Message: "Access denied",
				Code:    http.StatusForbidden,
			})
			return
		}
		// Managers cannot change roles to admin
		if req.Role == models.RoleAdmin {
			c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Error:   "Forbidden",
				Message: "Cannot assign admin role",
				Code:    http.StatusForbidden,
			})
			return
		}
	}

	updatedBy := authUserID
	user, err := h.userService.Update(id, &req, updatedBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	h.auditLogger.LogUpdate(c, "user", user.ID, oldUser, user)

	c.JSON(http.StatusOK, dto.ToUserResponse(user))
}

// UpdateProfile godoc
// @Summary Update own profile
// @Description Update own profile (limited fields)
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.UpdateProfileRequest true "Profile data"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/v1/users/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	authUserID := middleware.GetAuthUserID(c)

	oldUser, _ := h.userService.GetByID(authUserID)

	user, err := h.userService.UpdateProfile(authUserID, &req, authUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	h.auditLogger.LogUpdate(c, "user", user.ID, oldUser, user)

	c.JSON(http.StatusOK, dto.ToUserResponse(user))
}

// UpdatePassword godoc
// @Summary Update password
// @Description Update own password
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.UpdatePasswordRequest true "Password data"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/v1/users/password [put]
func (h *UserHandler) UpdatePassword(c *gin.Context) {
	var req dto.UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	authUserID := middleware.GetAuthUserID(c)

	err := h.userService.UpdatePassword(authUserID, req.CurrentPassword, req.NewPassword, authUserID)
	if err != nil {
		if err == services.ErrInvalidCredentials {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "Bad Request",
				Message: "Invalid current password",
				Code:    http.StatusBadRequest,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	h.auditLogger.Log(c, models.AuditActionUpdate, "user", &authUserID, nil, nil, "Password changed")

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Password updated successfully",
	})
}

// Delete godoc
// @Summary Delete user
// @Description Delete a user (Admin only)
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	oldUser, err := h.userService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Not Found",
			Message: "User not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	if err := h.userService.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	h.auditLogger.LogDelete(c, "user", id, oldUser)

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "User deleted successfully",
	})
}

// List godoc
// @Summary List users
// @Description List users with pagination
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.UserListResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /api/v1/users [get]
func (h *UserHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	authUser := middleware.GetAuthUser(c)
	authUserID := middleware.GetAuthUserID(c)

	users, total, err := h.userService.List(page, pageSize, authUser.Role, &authUserID)
	if err != nil {
		if err == services.ErrAccessDenied {
			c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Error:   "Forbidden",
				Message: "Access denied",
				Code:    http.StatusForbidden,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, dto.UserListResponse{
		Users:      dto.ToUserResponseList(users),
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}
