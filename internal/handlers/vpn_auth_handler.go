package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/dto"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/services"
)

// VpnAuthHandler handles VPN authentication requests from OpenVPN server
type VpnAuthHandler struct {
	userService    *services.UserService
	groupService   *services.GroupService
	networkService *services.NetworkService
	sessionService *services.VpnSessionService
}

// NewVpnAuthHandler creates a new VPN auth handler
func NewVpnAuthHandler() *VpnAuthHandler {
	return &VpnAuthHandler{
		userService:    services.NewUserService(),
		groupService:   services.NewGroupService(),
		networkService: services.NewNetworkService(),
		sessionService: services.NewVpnSessionService(),
	}
}

// VpnAuthRequest represents a VPN authentication request
type VpnAuthRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// VpnAuthResponse represents a VPN authentication response
type VpnAuthResponse struct {
	Success  bool       `json:"success"`
	UserID   *uuid.UUID `json:"user_id,omitempty"`
	Username string     `json:"username,omitempty"`
	VpnIP    string     `json:"vpn_ip,omitempty"`
	Message  string     `json:"message,omitempty"`
}

// VpnUserResponse represents a user response for VPN
type VpnUserResponse struct {
	ID        uuid.UUID  `json:"id"`
	Username  string     `json:"username"`
	FirstName string     `json:"first_name,omitempty"`
	LastName  string     `json:"last_name,omitempty"`
	Email     string     `json:"email,omitempty"`
	IsActive  bool       `json:"is_active"`
	VpnIP     string     `json:"vpn_ip,omitempty"`
	ValidFrom *time.Time `json:"valid_from,omitempty"`
	ValidTo   *time.Time `json:"valid_to,omitempty"`
}

// VpnRouteResponse represents a network route for VPN
type VpnRouteResponse struct {
	CIDR        string `json:"cidr"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	GroupName   string `json:"group_name"`
}

// VpnUserRoutesResponse represents user routes response
type VpnUserRoutesResponse struct {
	UserID   uuid.UUID          `json:"user_id"`
	Username string             `json:"username"`
	VpnIP    string             `json:"vpn_ip,omitempty"`
	Routes   []VpnRouteResponse `json:"routes"`
}

// Authenticate godoc
// @Summary      Authenticate VPN user
// @Description  Authenticate a user for VPN connection (called by OpenVPN auth-user-pass-verify script)
// @Tags         vpn-auth
// @Accept       json
// @Produce      json
// @Param        credentials  body      VpnAuthRequest  true  "User credentials"
// @Success      200          {object}  VpnAuthResponse
// @Failure      400          {object}  dto.ErrorResponse
// @Failure      401          {object}  VpnAuthResponse
// @Security     VpnToken
// @Router       /api/v1/vpn-auth/authenticate [post]
func (h *VpnAuthHandler) Authenticate(c *gin.Context) {
	var req VpnAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Authenticate user
	user, err := services.AuthenticateUser(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, VpnAuthResponse{
			Success: false,
			Message: "Invalid credentials",
		})
		return
	}

	// Check if user is active
	if !user.IsActive {
		c.JSON(http.StatusUnauthorized, VpnAuthResponse{
			Success: false,
			Message: "User account is disabled",
		})
		return
	}

	// Check validity period
	now := time.Now()
	if user.ValidFrom != nil && now.Before(*user.ValidFrom) {
		c.JSON(http.StatusUnauthorized, VpnAuthResponse{
			Success: false,
			Message: "User account is not yet valid",
		})
		return
	}
	if user.ValidTo != nil && now.After(*user.ValidTo) {
		c.JSON(http.StatusUnauthorized, VpnAuthResponse{
			Success: false,
			Message: "User account has expired",
		})
		return
	}

	c.JSON(http.StatusOK, VpnAuthResponse{
		Success:  true,
		UserID:   &user.ID,
		Username: user.Username,
		VpnIP:    user.VpnIP,
	})
}

// GetUserByUsername godoc
// @Summary      Get user by username
// @Description  Get user details by username for VPN (called by OpenVPN scripts)
// @Tags         vpn-auth
// @Accept       json
// @Produce      json
// @Param        username  path      string  true  "Username"
// @Success      200       {object}  VpnUserResponse
// @Failure      404       {object}  dto.ErrorResponse
// @Security     VpnToken
// @Router       /api/v1/vpn-auth/users/by-username/{username} [get]
func (h *VpnAuthHandler) GetUserByUsername(c *gin.Context) {
	username := c.Param("username")

	user, err := h.userService.GetByUsername(username)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Not Found",
			Message: "User not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, VpnUserResponse{
		ID:        user.ID,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		IsActive:  user.IsActive,
		VpnIP:     user.VpnIP,
		ValidFrom: user.ValidFrom,
		ValidTo:   user.ValidTo,
	})
}

// GetUserByID godoc
// @Summary      Get user by ID
// @Description  Get user details by ID for VPN (called by OpenVPN scripts)
// @Tags         vpn-auth
// @Accept       json
// @Produce      json
// @Param        id  path      string  true  "User ID"
// @Success      200 {object}  VpnUserResponse
// @Failure      400 {object}  dto.ErrorResponse
// @Failure      404 {object}  dto.ErrorResponse
// @Security     VpnToken
// @Router       /api/v1/vpn-auth/users/{id} [get]
func (h *VpnAuthHandler) GetUserByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid user ID format",
			Code:    http.StatusBadRequest,
		})
		return
	}

	user, err := h.userService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Not Found",
			Message: "User not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, VpnUserResponse{
		ID:        user.ID,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		IsActive:  user.IsActive,
		VpnIP:     user.VpnIP,
		ValidFrom: user.ValidFrom,
		ValidTo:   user.ValidTo,
	})
}

// GetUserRoutes godoc
// @Summary      Get user routes
// @Description  Get network routes for a user based on their group memberships (called by OpenVPN client-connect script)
// @Tags         vpn-auth
// @Accept       json
// @Produce      json
// @Param        id  path      string  true  "User ID"
// @Success      200 {object}  VpnUserRoutesResponse
// @Failure      400 {object}  dto.ErrorResponse
// @Failure      404 {object}  dto.ErrorResponse
// @Security     VpnToken
// @Router       /api/v1/vpn-auth/users/{id}/routes [get]
func (h *VpnAuthHandler) GetUserRoutes(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid user ID format",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Get user
	user, err := h.userService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "Not Found",
			Message: "User not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	// Get user groups with networks
	groupsWithNetworks, err := h.groupService.GetUserGroupsWithNetworks(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to get user routes",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Collect unique routes
	routeMap := make(map[string]VpnRouteResponse)
	for _, group := range groupsWithNetworks {
		for _, network := range group.Networks {
			// Use CIDR as key to avoid duplicates
			if _, exists := routeMap[network.CIDR]; !exists {
				routeMap[network.CIDR] = VpnRouteResponse{
					CIDR:        network.CIDR,
					Name:        network.Name,
					Description: network.Description,
					GroupName:   group.Name,
				}
			}
		}
	}

	// Convert map to slice
	routes := make([]VpnRouteResponse, 0, len(routeMap))
	for _, route := range routeMap {
		routes = append(routes, route)
	}

	c.JSON(http.StatusOK, VpnUserRoutesResponse{
		UserID:   user.ID,
		Username: user.Username,
		VpnIP:    user.VpnIP,
		Routes:   routes,
	})
}

// CreateSession godoc
// @Summary      Create VPN session
// @Description  Create a new VPN session (called by OpenVPN client-connect script)
// @Tags         vpn-auth
// @Accept       json
// @Produce      json
// @Param        session  body      dto.CreateVpnSessionRequest  true  "Session data"
// @Success      201      {object}  dto.VpnSessionResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Security     VpnToken
// @Router       /api/v1/vpn-auth/sessions [post]
func (h *VpnAuthHandler) CreateSession(c *gin.Context) {
	var req dto.CreateVpnSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	session, err := h.sessionService.Create(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to create session",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusCreated, dto.ToVpnSessionResponse(session))
}

// DisconnectSession godoc
// @Summary      Disconnect VPN session
// @Description  Update a VPN session with disconnect info (called by OpenVPN client-disconnect script)
// @Tags         vpn-auth
// @Accept       json
// @Produce      json
// @Param        id       path      string                       true  "Session ID"
// @Param        session  body      dto.UpdateVpnSessionRequest  true  "Disconnect data"
// @Success      200      {object}  dto.VpnSessionResponse
// @Failure      400      {object}  dto.ErrorResponse
// @Failure      404      {object}  dto.ErrorResponse
// @Security     VpnToken
// @Router       /api/v1/vpn-auth/sessions/{id}/disconnect [put]
func (h *VpnAuthHandler) DisconnectSession(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid session ID format",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var req dto.UpdateVpnSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	session, err := h.sessionService.Disconnect(id, &req)
	if err != nil {
		if err == services.ErrSessionNotFound {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "Not Found",
				Message: "Session not found",
				Code:    http.StatusNotFound,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to update session",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, dto.ToVpnSessionResponse(session))
}

// ListAllUsers godoc
// @Summary      List all active VPN users
// @Description  Get a list of all active users for VPN (used for firewall rules generation)
// @Tags         vpn-auth
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string][]VpnUserResponse
// @Security     VpnToken
// @Router       /api/v1/vpn-auth/users [get]
func (h *VpnAuthHandler) ListAllUsers(c *gin.Context) {
	// Get all active users (high limit)
	users, _, err := h.userService.List(1, 10000, "admin", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to fetch users",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Filter only active users with VPN IP
	var vpnUsers []VpnUserResponse
	for _, user := range users {
		if user.IsActive && user.VpnIP != "" {
			vpnUsers = append(vpnUsers, VpnUserResponse{
				ID:        user.ID,
				Username:  user.Username,
				FirstName: user.FirstName,
				LastName:  user.LastName,
				Email:     user.Email,
				IsActive:  user.IsActive,
				VpnIP:     user.VpnIP,
				ValidFrom: user.ValidFrom,
				ValidTo:   user.ValidTo,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"users": vpnUsers})
}
