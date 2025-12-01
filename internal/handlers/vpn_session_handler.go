package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/dto"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/services"
)

// VpnSessionHandler handles VPN session related requests
type VpnSessionHandler struct {
	sessionService *services.VpnSessionService
	statsService   *services.VpnTrafficStatsService
}

// NewVpnSessionHandler creates a new VPN session handler
func NewVpnSessionHandler() *VpnSessionHandler {
	return &VpnSessionHandler{
		sessionService: services.NewVpnSessionService(),
		statsService:   services.NewVpnTrafficStatsService(),
	}
}

// List godoc
// @Summary      List VPN sessions
// @Description  Get a paginated list of VPN sessions (ADMIN only)
// @Tags         vpn-sessions
// @Accept       json
// @Produce      json
// @Param        user_id    query    string  false  "Filter by user ID"
// @Param        vpn_ip     query    string  false  "Filter by VPN IP"
// @Param        is_active  query    bool    false  "Filter by active status"
// @Param        start_date query    string  false  "Filter by start date (RFC3339)"
// @Param        end_date   query    string  false  "Filter by end date (RFC3339)"
// @Param        page       query    int     false  "Page number"     default(1)
// @Param        page_size  query    int     false  "Page size"       default(20)
// @Success      200        {object} dto.VpnSessionListResponse
// @Failure      400        {object} map[string]string
// @Failure      401        {object} map[string]string
// @Failure      403        {object} map[string]string
// @Security     BearerAuth
// @Router       /api/v1/vpn/sessions [get]
func (h *VpnSessionHandler) List(c *gin.Context) {
	var filter dto.VpnSessionFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sessions, total, err := h.sessionService.List(&filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sessions"})
		return
	}

	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}

	c.JSON(http.StatusOK, dto.VpnSessionListResponse{
		Sessions:   dto.ToVpnSessionResponseList(sessions),
		Total:      total,
		Page:       filter.Page,
		PageSize:   pageSize,
		TotalPages: services.CalculateTotalPages(total, pageSize),
	})
}

// Get godoc
// @Summary      Get VPN session
// @Description  Get a VPN session by ID (ADMIN only)
// @Tags         vpn-sessions
// @Accept       json
// @Produce      json
// @Param        id   path     string  true  "Session ID"
// @Success      200  {object} dto.VpnSessionResponse
// @Failure      400  {object} map[string]string
// @Failure      401  {object} map[string]string
// @Failure      403  {object} map[string]string
// @Failure      404  {object} map[string]string
// @Security     BearerAuth
// @Router       /api/v1/vpn/sessions/{id} [get]
func (h *VpnSessionHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	session, err := h.sessionService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	c.JSON(http.StatusOK, dto.ToVpnSessionResponse(session))
}

// Create godoc
// @Summary      Create VPN session
// @Description  Create a new VPN session (called by VPN server)
// @Tags         vpn-sessions
// @Accept       json
// @Produce      json
// @Param        session  body     dto.CreateVpnSessionRequest  true  "Session data"
// @Success      201      {object} dto.VpnSessionResponse
// @Failure      400      {object} map[string]string
// @Failure      401      {object} map[string]string
// @Security     BearerAuth
// @Router       /api/v1/vpn/sessions [post]
func (h *VpnSessionHandler) Create(c *gin.Context) {
	var req dto.CreateVpnSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, err := h.sessionService.Create(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	c.JSON(http.StatusCreated, dto.ToVpnSessionResponse(session))
}

// Disconnect godoc
// @Summary      Disconnect VPN session
// @Description  Update a VPN session with disconnect info (called by VPN server)
// @Tags         vpn-sessions
// @Accept       json
// @Produce      json
// @Param        id       path     string                       true  "Session ID"
// @Param        session  body     dto.UpdateVpnSessionRequest  true  "Disconnect data"
// @Success      200      {object} dto.VpnSessionResponse
// @Failure      400      {object} map[string]string
// @Failure      401      {object} map[string]string
// @Failure      404      {object} map[string]string
// @Security     BearerAuth
// @Router       /api/v1/vpn/sessions/{id}/disconnect [put]
func (h *VpnSessionHandler) Disconnect(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	var req dto.UpdateVpnSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, err := h.sessionService.Disconnect(id, &req)
	if err != nil {
		if err == services.ErrSessionNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update session"})
		return
	}

	c.JSON(http.StatusOK, dto.ToVpnSessionResponse(session))
}

// GetActive godoc
// @Summary      Get active VPN sessions
// @Description  Get all currently active VPN sessions (ADMIN only)
// @Tags         vpn-sessions
// @Accept       json
// @Produce      json
// @Success      200  {object} []dto.VpnSessionResponse
// @Failure      401  {object} map[string]string
// @Failure      403  {object} map[string]string
// @Security     BearerAuth
// @Router       /api/v1/vpn/sessions/active [get]
func (h *VpnSessionHandler) GetActive(c *gin.Context) {
	sessions, err := h.sessionService.GetActiveSessions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch active sessions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"sessions": dto.ToVpnSessionResponseList(sessions)})
}

// GetStats godoc
// @Summary      Get VPN usage statistics
// @Description  Get aggregated VPN usage statistics (ADMIN only)
// @Tags         vpn-sessions
// @Accept       json
// @Produce      json
// @Success      200  {object} dto.VpnUsageStats
// @Failure      401  {object} map[string]string
// @Failure      403  {object} map[string]string
// @Security     BearerAuth
// @Router       /api/v1/vpn/stats [get]
func (h *VpnSessionHandler) GetStats(c *gin.Context) {
	stats, err := h.sessionService.GetUsageStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetUserStats godoc
// @Summary      Get VPN usage statistics per user
// @Description  Get VPN usage statistics aggregated per user (ADMIN only)
// @Tags         vpn-sessions
// @Accept       json
// @Produce      json
// @Param        page       query    int     false  "Page number"     default(1)
// @Param        page_size  query    int     false  "Page size"       default(20)
// @Success      200        {object} map[string]interface{}
// @Failure      401        {object} map[string]string
// @Failure      403        {object} map[string]string
// @Security     BearerAuth
// @Router       /api/v1/vpn/stats/users [get]
func (h *VpnSessionHandler) GetUserStats(c *gin.Context) {
	page := 1
	pageSize := 20
	if p := c.Query("page"); p != "" {
		if _, err := c.GetQuery("page"); err {
			page = 1
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if _, err := c.GetQuery("page_size"); err {
			pageSize = 20
		}
	}

	stats, total, err := h.sessionService.GetUserUsageStats(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users":       stats,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": services.CalculateTotalPages(total, pageSize),
	})
}

// CreateTrafficStats godoc
// @Summary      Create traffic stats
// @Description  Create a new traffic stats entry (called by VPN server)
// @Tags         vpn-sessions
// @Accept       json
// @Produce      json
// @Param        stats  body     dto.CreateVpnTrafficStatsRequest  true  "Stats data"
// @Success      201    {object} dto.VpnTrafficStatsResponse
// @Failure      400    {object} map[string]string
// @Failure      401    {object} map[string]string
// @Security     BearerAuth
// @Router       /api/v1/vpn/traffic-stats [post]
func (h *VpnSessionHandler) CreateTrafficStats(c *gin.Context) {
	var req dto.CreateVpnTrafficStatsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stats, err := h.statsService.Create(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create traffic stats"})
		return
	}

	c.JSON(http.StatusCreated, dto.ToVpnTrafficStatsResponse(stats))
}

// ListTrafficStats godoc
// @Summary      List traffic stats
// @Description  Get a paginated list of traffic stats (ADMIN only)
// @Tags         vpn-sessions
// @Accept       json
// @Produce      json
// @Param        session_id  query    string  false  "Filter by session ID"
// @Param        start_date  query    string  false  "Filter by start date (RFC3339)"
// @Param        end_date    query    string  false  "Filter by end date (RFC3339)"
// @Param        page        query    int     false  "Page number"     default(1)
// @Param        page_size   query    int     false  "Page size"       default(20)
// @Success      200         {object} dto.VpnTrafficStatsListResponse
// @Failure      400         {object} map[string]string
// @Failure      401         {object} map[string]string
// @Failure      403         {object} map[string]string
// @Security     BearerAuth
// @Router       /api/v1/vpn/traffic-stats [get]
func (h *VpnSessionHandler) ListTrafficStats(c *gin.Context) {
	var filter dto.VpnTrafficStatsFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stats, total, err := h.statsService.List(&filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch traffic stats"})
		return
	}

	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}

	c.JSON(http.StatusOK, dto.VpnTrafficStatsListResponse{
		Stats:      dto.ToVpnTrafficStatsResponseList(stats),
		Total:      total,
		Page:       filter.Page,
		PageSize:   pageSize,
		TotalPages: services.CalculateTotalPages(total, pageSize),
	})
}
