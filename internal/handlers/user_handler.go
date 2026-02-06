package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/apperror"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/config"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/dto"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/middleware"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/models"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/services"
)

// UserHandler handles user requests
type UserHandler struct {
	userService  *services.UserService
	groupService *services.GroupService
	vpnIPService *services.VPNIPService
	auditLogger  *middleware.AuditLogger
}

// NewUserHandler creates a new user handler
func NewUserHandler(vpnCfg *config.VPNConfig) *UserHandler {
	return &UserHandler{
		userService:  services.NewUserService(),
		groupService: services.NewGroupService(),
		vpnIPService: services.NewVPNIPService(vpnCfg),
		auditLogger:  middleware.NewAuditLogger(),
	}
}

// Create godoc
// @Summary Create user
// @Description Create a new user (Admin or Manager). Managers can only create users assigned to themselves.
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

	authUser := middleware.GetAuthUser(c)
	authUserID := middleware.GetAuthUserID(c)

	// MANAGER restrictions
	if authUser.Role == models.RoleManager {
		// Manager must set themselves as the manager of new user
		req.ManagerID = &authUserID
		// Manager cannot create ADMIN users
		if req.Role == models.RoleAdmin {
			c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Error:   "Forbidden",
				Message: "Managers cannot create admin users",
				Code:    http.StatusForbidden,
			})
			return
		}
	}

	// Handle VPN IP: validate if provided, auto-assign if empty
	if req.VpnIP != "" {
		// Validate provided IP
		if err := h.vpnIPService.ValidateIP(req.VpnIP); err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "Bad Request",
				Message: "VPN IP error: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
	} else {
		// Auto-assign next available IP
		nextIP, err := h.vpnIPService.GetNextAvailableIP()
		if err != nil && err != services.ErrVPNNetworkNotConfigured {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error:   "Internal Server Error",
				Message: "Failed to allocate VPN IP: " + err.Error(),
				Code:    http.StatusInternalServerError,
			})
			return
		}
		if nextIP != "" {
			req.VpnIP = nextIP
		}
	}

	user, err := h.userService.Create(&req, authUserID)
	if err != nil {
		apperror.HandleError(c, err)
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
// @Description Update a user. Admin can update anyone. Manager can update their subordinates. User cannot use this endpoint.
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

	// USER role cannot update other users (they should use /profile endpoint for themselves)
	if authUser.Role == models.RoleUser {
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "Forbidden",
			Message: "Access denied. Use /users/profile to update your own profile.",
			Code:    http.StatusForbidden,
		})
		return
	}

	// MANAGER can only update their subordinates
	if authUser.Role == models.RoleManager {
		if oldUser.ManagerID == nil || *oldUser.ManagerID != authUserID {
			c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Error:   "Forbidden",
				Message: "Access denied. You can only update your subordinates.",
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
		// Managers cannot change the manager_id to someone else
		if req.ManagerID != nil && *req.ManagerID != authUserID {
			c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Error:   "Forbidden",
				Message: "Cannot reassign user to another manager",
				Code:    http.StatusForbidden,
			})
			return
		}
	}

	// Validate VPN IP if provided
	if req.VpnIP != nil && *req.VpnIP != "" {
		if err := h.vpnIPService.ValidateIP(*req.VpnIP, id.String()); err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "Bad Request",
				Message: "VPN IP error: " + err.Error(),
				Code:    http.StatusBadRequest,
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
// @Description List users with pagination. Admin sees all users. Manager sees only subordinates. User sees only themselves.
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

// GetGroups godoc
// @Summary Get user groups
// @Description Get all groups a user belongs to with their networks
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} dto.UserGroupsResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/users/{id}/groups [get]
func (h *UserHandler) GetGroups(c *gin.Context) {
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

	groups, err := h.groupService.GetUserGroupsWithNetworks(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Convert to DTO
	response := make([]dto.UserGroupWithNetworks, len(groups))
	for i, g := range groups {
		networks := make([]dto.NetworkBasicInfo, len(g.Networks))
		for j, n := range g.Networks {
			networks[j] = dto.NetworkBasicInfo{
				ID:          n.ID,
				Name:        n.Name,
				CIDR:        n.CIDR,
				Description: n.Description,
			}
		}
		response[i] = dto.UserGroupWithNetworks{
			ID:          g.ID,
			Name:        g.Name,
			Description: g.Description,
			Networks:    networks,
		}
	}

	c.JSON(http.StatusOK, dto.UserGroupsResponse{
		Groups: response,
	})
}

// AddGroup godoc
// @Summary Add user to group
// @Description Add a user to a group
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param request body dto.AddUserGroupRequest true "Group ID"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/users/{id}/groups [post]
func (h *UserHandler) AddGroup(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var req dto.AddUserGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	authUserID := middleware.GetAuthUserID(c)

	if err := h.groupService.AddUserToGroup(req.GroupID, userID, authUserID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	h.auditLogger.Log(c, models.AuditActionUpdate, "user_group", &userID, nil, nil, "Added user to group")

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "User added to group successfully",
	})
}

// RemoveGroup godoc
// @Summary Remove user from group
// @Description Remove a user from a group
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param group_id path string true "Group ID"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /api/v1/users/{id}/groups/{group_id} [delete]
func (h *UserHandler) RemoveGroup(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	groupIDStr := c.Param("group_id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid group ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	if err := h.groupService.RemoveUserFromGroup(groupID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	h.auditLogger.Log(c, models.AuditActionUpdate, "user_group", &userID, nil, nil, "Removed user from group")

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "User removed from group successfully",
	})
}
