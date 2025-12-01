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

// GroupHandler handles group requests
type GroupHandler struct {
	groupService *services.GroupService
	auditLogger  *middleware.AuditLogger
}

// NewGroupHandler creates a new group handler
func NewGroupHandler() *GroupHandler {
	return &GroupHandler{
		groupService: services.NewGroupService(),
		auditLogger:  middleware.NewAuditLogger(),
	}
}

// Create godoc
// @Summary Create group
// @Description Create a new group (Admin only)
// @Tags groups
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateGroupRequest true "Group data"
// @Success 201 {object} dto.GroupResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /api/v1/groups [post]
func (h *GroupHandler) Create(c *gin.Context) {
	var req dto.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	createdBy := middleware.GetAuthUserID(c)
	group, err := h.groupService.Create(&req, createdBy)
	if err != nil {
		if err == services.ErrGroupExists {
			c.JSON(http.StatusConflict, dto.ErrorResponse{
				Error:   "Conflict",
				Message: "Group already exists",
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

	h.auditLogger.LogCreate(c, "group", group.ID, group)

	c.JSON(http.StatusCreated, dto.ToGroupResponse(group))
}

// Get godoc
// @Summary Get group
// @Description Get a group by ID
// @Tags groups
// @Produce json
// @Security BearerAuth
// @Param id path string true "Group ID"
// @Success 200 {object} dto.GroupResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/groups/{id} [get]
func (h *GroupHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid group ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	group, err := h.groupService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Not Found",
			Message: "Group not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, dto.ToGroupResponse(group))
}

// Update godoc
// @Summary Update group
// @Description Update a group (Admin only)
// @Tags groups
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Group ID"
// @Param request body dto.UpdateGroupRequest true "Group data"
// @Success 200 {object} dto.GroupResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/groups/{id} [put]
func (h *GroupHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid group ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var req dto.UpdateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	oldGroup, err := h.groupService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Not Found",
			Message: "Group not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	updatedBy := middleware.GetAuthUserID(c)
	group, err := h.groupService.Update(id, &req, updatedBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	h.auditLogger.LogUpdate(c, "group", group.ID, oldGroup, group)

	c.JSON(http.StatusOK, dto.ToGroupResponse(group))
}

// Delete godoc
// @Summary Delete group
// @Description Delete a group (Admin only)
// @Tags groups
// @Produce json
// @Security BearerAuth
// @Param id path string true "Group ID"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/groups/{id} [delete]
func (h *GroupHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid group ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	oldGroup, err := h.groupService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Not Found",
			Message: "Group not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	if err := h.groupService.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	h.auditLogger.LogDelete(c, "group", id, oldGroup)

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Group deleted successfully",
	})
}

// List godoc
// @Summary List groups
// @Description List groups with pagination
// @Tags groups
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.GroupListResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/v1/groups [get]
func (h *GroupHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	groups, total, err := h.groupService.List(page, pageSize)
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

	c.JSON(http.StatusOK, dto.GroupListResponse{
		Groups:     dto.ToGroupResponseList(groups),
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// AddUser godoc
// @Summary Add user to group
// @Description Add a user to a group. Admin can add anyone. Manager can only add their subordinates.
// @Tags groups
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Group ID"
// @Param request body dto.AddUserToGroupRequest true "User data"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/groups/{id}/users [post]
func (h *GroupHandler) AddUser(c *gin.Context) {
	idStr := c.Param("id")
	groupID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid group ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var req dto.AddUserToGroupRequest
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

	// MANAGER can only add their subordinates to groups
	if authUser.Role == models.RoleManager {
		userService := services.NewUserService()
		targetUser, err := userService.GetByID(req.UserID)
		if err != nil {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "Not Found",
				Message: "User not found",
				Code:    http.StatusNotFound,
			})
			return
		}
		if targetUser.ManagerID == nil || *targetUser.ManagerID != authUserID {
			c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Error:   "Forbidden",
				Message: "You can only add your subordinates to groups",
				Code:    http.StatusForbidden,
			})
			return
		}
	}

	if err := h.groupService.AddUserToGroup(groupID, req.UserID, authUserID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	h.auditLogger.LogCreate(c, "user_group", groupID, map[string]interface{}{
		"group_id": groupID,
		"user_id":  req.UserID,
	})

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "User added to group successfully",
	})
}

// RemoveUser godoc
// @Summary Remove user from group
// @Description Remove a user from a group. Admin can remove anyone. Manager can only remove their subordinates.
// @Tags groups
// @Produce json
// @Security BearerAuth
// @Param id path string true "Group ID"
// @Param user_id path string true "User ID"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /api/v1/groups/{id}/users/{user_id} [delete]
func (h *GroupHandler) RemoveUser(c *gin.Context) {
	groupIDStr := c.Param("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid group ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid user ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	authUser := middleware.GetAuthUser(c)
	authUserID := middleware.GetAuthUserID(c)

	// MANAGER can only remove their subordinates from groups
	if authUser.Role == models.RoleManager {
		userService := services.NewUserService()
		targetUser, err := userService.GetByID(userID)
		if err != nil {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "Not Found",
				Message: "User not found",
				Code:    http.StatusNotFound,
			})
			return
		}
		if targetUser.ManagerID == nil || *targetUser.ManagerID != authUserID {
			c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Error:   "Forbidden",
				Message: "You can only remove your subordinates from groups",
				Code:    http.StatusForbidden,
			})
			return
		}
	}

	if err := h.groupService.RemoveUserFromGroup(groupID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	h.auditLogger.LogDelete(c, "user_group", groupID, map[string]interface{}{
		"group_id": groupID,
		"user_id":  userID,
	})

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "User removed from group successfully",
	})
}

// GetUsers godoc
// @Summary Get group users
// @Description Get users in a group. Admin sees all. Manager sees only subordinates in the group. User sees only themselves if member.
// @Tags groups
// @Produce json
// @Security BearerAuth
// @Param id path string true "Group ID"
// @Success 200 {array} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/groups/{id}/users [get]
func (h *GroupHandler) GetUsers(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid group ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	users, err := h.groupService.GetGroupUsers(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	authUser := middleware.GetAuthUser(c)
	authUserID := middleware.GetAuthUserID(c)

	// Filter users based on role
	var filteredUsers []models.User
	switch authUser.Role {
	case models.RoleAdmin:
		// Admin sees all users in the group
		filteredUsers = users
	case models.RoleManager:
		// Manager sees only their subordinates in the group
		for _, u := range users {
			if u.ManagerID != nil && *u.ManagerID == authUserID {
				filteredUsers = append(filteredUsers, u)
			}
		}
	case models.RoleUser:
		// User sees only themselves if they are a member
		for _, u := range users {
			if u.ID == authUserID {
				filteredUsers = append(filteredUsers, u)
				break
			}
		}
	}

	c.JSON(http.StatusOK, dto.ToUserResponseList(filteredUsers))
}

// GetNetworks godoc
// @Summary Get group networks
// @Description Get networks assigned to a group
// @Tags groups
// @Produce json
// @Security BearerAuth
// @Param id path string true "Group ID"
// @Success 200 {array} dto.NetworkResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/groups/{id}/networks [get]
func (h *GroupHandler) GetNetworks(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid group ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	networks, err := h.groupService.GetGroupNetworks(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, dto.ToNetworkResponseList(networks))
}

// AddNetwork godoc
// @Summary Add network to group
// @Description Add a network to a group (Admin only)
// @Tags groups
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Group ID"
// @Param request body dto.AddNetworkToGroupRequest true "Network data"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/groups/{id}/networks [post]
func (h *GroupHandler) AddNetwork(c *gin.Context) {
	idStr := c.Param("id")
	groupID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid group ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var req dto.AddNetworkToGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	authUserID := middleware.GetAuthUserID(c)

	if err := h.groupService.AddNetworkToGroup(groupID, req.NetworkID, authUserID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	h.auditLogger.LogCreate(c, "network_group", groupID, map[string]interface{}{
		"group_id":   groupID,
		"network_id": req.NetworkID,
	})

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Network added to group successfully",
	})
}

// RemoveNetwork godoc
// @Summary Remove network from group
// @Description Remove a network from a group (Admin only)
// @Tags groups
// @Produce json
// @Security BearerAuth
// @Param id path string true "Group ID"
// @Param network_id path string true "Network ID"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /api/v1/groups/{id}/networks/{network_id} [delete]
func (h *GroupHandler) RemoveNetwork(c *gin.Context) {
	groupIDStr := c.Param("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid group ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	networkIDStr := c.Param("network_id")
	networkID, err := uuid.Parse(networkIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid network ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	if err := h.groupService.RemoveNetworkFromGroup(groupID, networkID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	h.auditLogger.LogDelete(c, "network_group", groupID, map[string]interface{}{
		"group_id":   groupID,
		"network_id": networkID,
	})

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Network removed from group successfully",
	})
}
