package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/dto"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/middleware"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/services"
)

// NetworkHandler handles network requests
type NetworkHandler struct {
	networkService *services.NetworkService
	auditLogger    *middleware.AuditLogger
}

// NewNetworkHandler creates a new network handler
func NewNetworkHandler() *NetworkHandler {
	return &NetworkHandler{
		networkService: services.NewNetworkService(),
		auditLogger:    middleware.NewAuditLogger(),
	}
}

// Create godoc
// @Summary Create network
// @Description Create a new network (Admin only)
// @Tags networks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateNetworkRequest true "Network data"
// @Success 201 {object} dto.NetworkResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /api/v1/networks [post]
func (h *NetworkHandler) Create(c *gin.Context) {
	var req dto.CreateNetworkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	createdBy := middleware.GetAuthUserID(c)
	network, err := h.networkService.Create(&req, createdBy)
	if err != nil {
		if err == services.ErrNetworkExists {
			c.JSON(http.StatusConflict, dto.ErrorResponse{
				Error:   "Conflict",
				Message: "Network with this name already exists",
				Code:    http.StatusConflict,
			})
			return
		}
		if err == services.ErrInvalidCIDR {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "Bad Request",
				Message: "Invalid CIDR format. Use format like '192.168.1.0/24' or '10.0.0.1'",
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

	h.auditLogger.LogCreate(c, "network", network.ID, network)

	c.JSON(http.StatusCreated, dto.ToNetworkResponse(network))
}

// Get godoc
// @Summary Get network
// @Description Get a network by ID (Admin only)
// @Tags networks
// @Produce json
// @Security BearerAuth
// @Param id path string true "Network ID"
// @Success 200 {object} dto.NetworkResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/networks/{id} [get]
func (h *NetworkHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid network ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	network, err := h.networkService.GetByID(id)
	if err != nil {
		if err == services.ErrNetworkNotFound {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "Not Found",
				Message: "Network not found",
				Code:    http.StatusNotFound,
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

	c.JSON(http.StatusOK, dto.ToNetworkResponse(network))
}

// Update godoc
// @Summary Update network
// @Description Update a network (Admin only)
// @Tags networks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Network ID"
// @Param request body dto.UpdateNetworkRequest true "Network data"
// @Success 200 {object} dto.NetworkResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/networks/{id} [put]
func (h *NetworkHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid network ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var req dto.UpdateNetworkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	oldNetwork, err := h.networkService.GetByID(id)
	if err != nil {
		if err == services.ErrNetworkNotFound {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "Not Found",
				Message: "Network not found",
				Code:    http.StatusNotFound,
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

	updatedBy := middleware.GetAuthUserID(c)
	network, err := h.networkService.Update(id, &req, updatedBy)
	if err != nil {
		if err == services.ErrInvalidCIDR {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "Bad Request",
				Message: "Invalid CIDR format. Use format like '192.168.1.0/24' or '10.0.0.1'",
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

	h.auditLogger.LogUpdate(c, "network", network.ID, oldNetwork, network)

	c.JSON(http.StatusOK, dto.ToNetworkResponse(network))
}

// Delete godoc
// @Summary Delete network
// @Description Delete a network (Admin only)
// @Tags networks
// @Produce json
// @Security BearerAuth
// @Param id path string true "Network ID"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/networks/{id} [delete]
func (h *NetworkHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid network ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	oldNetwork, err := h.networkService.GetByID(id)
	if err != nil {
		if err == services.ErrNetworkNotFound {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "Not Found",
				Message: "Network not found",
				Code:    http.StatusNotFound,
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

	if err := h.networkService.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	h.auditLogger.LogDelete(c, "network", id, oldNetwork)

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Network deleted successfully",
	})
}

// List godoc
// @Summary List networks
// @Description List networks with pagination (Admin only)
// @Tags networks
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.NetworkListResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /api/v1/networks [get]
func (h *NetworkHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	networks, total, err := h.networkService.List(page, pageSize)
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

	c.JSON(http.StatusOK, dto.NetworkListResponse{
		Networks:   dto.ToNetworkResponseList(networks),
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// AddGroup godoc
// @Summary Add group to network
// @Description Add a group to a network (Admin only)
// @Tags networks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Network ID"
// @Param request body dto.AddGroupToNetworkRequest true "Group data"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/networks/{id}/groups [post]
func (h *NetworkHandler) AddGroup(c *gin.Context) {
	idStr := c.Param("id")
	networkID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid network ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var req dto.AddGroupToNetworkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	createdBy := middleware.GetAuthUserID(c)
	if err := h.networkService.AddGroupToNetwork(networkID, req.GroupID, createdBy); err != nil {
		if err == services.ErrNetworkNotFound {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "Not Found",
				Message: "Network not found",
				Code:    http.StatusNotFound,
			})
			return
		}
		if err == services.ErrGroupNotFound {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "Not Found",
				Message: "Group not found",
				Code:    http.StatusNotFound,
			})
			return
		}
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	h.auditLogger.LogCreate(c, "network_group", networkID, map[string]interface{}{
		"network_id": networkID,
		"group_id":   req.GroupID,
	})

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Group added to network successfully",
	})
}

// RemoveGroup godoc
// @Summary Remove group from network
// @Description Remove a group from a network (Admin only)
// @Tags networks
// @Produce json
// @Security BearerAuth
// @Param id path string true "Network ID"
// @Param group_id path string true "Group ID"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /api/v1/networks/{id}/groups/{group_id} [delete]
func (h *NetworkHandler) RemoveGroup(c *gin.Context) {
	networkIDStr := c.Param("id")
	networkID, err := uuid.Parse(networkIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid network ID",
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

	if err := h.networkService.RemoveGroupFromNetwork(networkID, groupID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	h.auditLogger.LogDelete(c, "network_group", networkID, map[string]interface{}{
		"network_id": networkID,
		"group_id":   groupID,
	})

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: "Group removed from network successfully",
	})
}

// GetGroups godoc
// @Summary Get network groups
// @Description Get all groups assigned to a network (Admin only)
// @Tags networks
// @Produce json
// @Security BearerAuth
// @Param id path string true "Network ID"
// @Success 200 {array} dto.GroupResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/networks/{id}/groups [get]
func (h *NetworkHandler) GetGroups(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid network ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	groups, err := h.networkService.GetNetworkGroups(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, dto.ToGroupResponseList(groups))
}
