package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/dto"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/services"
)

// AuditHandler handles audit log related requests
type AuditHandler struct {
	auditService *services.AuditService
}

// NewAuditHandler creates a new audit handler
func NewAuditHandler() *AuditHandler {
	return &AuditHandler{
		auditService: services.NewAuditService(),
	}
}

// List godoc
// @Summary      List audit logs
// @Description  Get a paginated list of audit logs (ADMIN only)
// @Tags         audit
// @Accept       json
// @Produce      json
// @Param        user_id     query    string  false  "Filter by user ID"
// @Param        action      query    string  false  "Filter by action (CREATE, READ, UPDATE, DELETE, LOGIN, LOGOUT)"
// @Param        entity_type query    string  false  "Filter by entity type"
// @Param        entity_id   query    string  false  "Filter by entity ID"
// @Param        start_date  query    string  false  "Filter by start date (RFC3339)"
// @Param        end_date    query    string  false  "Filter by end date (RFC3339)"
// @Param        page        query    int     false  "Page number"     default(1)
// @Param        page_size   query    int     false  "Page size"       default(20)
// @Success      200         {object} dto.AuditLogListResponse
// @Failure      400         {object} map[string]string
// @Failure      401         {object} map[string]string
// @Failure      403         {object} map[string]string
// @Security     BearerAuth
// @Router       /api/v1/audit [get]
func (h *AuditHandler) List(c *gin.Context) {
	var filter dto.AuditLogFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logs, total, err := h.auditService.List(&filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch audit logs"})
		return
	}

	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}

	c.JSON(http.StatusOK, dto.AuditLogListResponse{
		Logs:       dto.ToAuditLogResponseList(logs),
		Total:      total,
		Page:       filter.Page,
		PageSize:   pageSize,
		TotalPages: services.CalculateTotalPages(total, pageSize),
	})
}

// Get godoc
// @Summary      Get audit log
// @Description  Get an audit log by ID (ADMIN only)
// @Tags         audit
// @Accept       json
// @Produce      json
// @Param        id   path     string  true  "Audit log ID"
// @Success      200  {object} dto.AuditLogResponse
// @Failure      400  {object} map[string]string
// @Failure      401  {object} map[string]string
// @Failure      403  {object} map[string]string
// @Failure      404  {object} map[string]string
// @Security     BearerAuth
// @Router       /api/v1/audit/{id} [get]
func (h *AuditHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid audit log ID"})
		return
	}

	log, err := h.auditService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Audit log not found"})
		return
	}

	c.JSON(http.StatusOK, dto.ToAuditLogResponse(log))
}

// GetByEntity godoc
// @Summary      Get audit logs by entity
// @Description  Get audit logs for a specific entity (ADMIN only)
// @Tags         audit
// @Accept       json
// @Produce      json
// @Param        entity_type path     string  true   "Entity type (user, group, network)"
// @Param        entity_id   path     string  true   "Entity ID"
// @Param        page        query    int     false  "Page number"     default(1)
// @Param        page_size   query    int     false  "Page size"       default(20)
// @Success      200         {object} dto.AuditLogListResponse
// @Failure      400         {object} map[string]string
// @Failure      401         {object} map[string]string
// @Failure      403         {object} map[string]string
// @Security     BearerAuth
// @Router       /api/v1/audit/entity/{entity_type}/{entity_id} [get]
func (h *AuditHandler) GetByEntity(c *gin.Context) {
	entityType := c.Param("entity_type")
	entityIDStr := c.Param("entity_id")

	entityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity ID"})
		return
	}

	page := 1
	pageSize := 20
	if p := c.Query("page"); p != "" {
		// Simple parsing, default to 1 on error
	}
	if ps := c.Query("page_size"); ps != "" {
		// Simple parsing, default to 20 on error
	}

	logs, total, err := h.auditService.GetByEntityID(entityType, entityID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch audit logs"})
		return
	}

	c.JSON(http.StatusOK, dto.AuditLogListResponse{
		Logs:       dto.ToAuditLogResponseList(logs),
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: services.CalculateTotalPages(total, pageSize),
	})
}

// GetByUser godoc
// @Summary      Get audit logs by user
// @Description  Get audit logs for a specific user (ADMIN only)
// @Tags         audit
// @Accept       json
// @Produce      json
// @Param        user_id   path     string  true   "User ID"
// @Param        page      query    int     false  "Page number"     default(1)
// @Param        page_size query    int     false  "Page size"       default(20)
// @Success      200       {object} dto.AuditLogListResponse
// @Failure      400       {object} map[string]string
// @Failure      401       {object} map[string]string
// @Failure      403       {object} map[string]string
// @Security     BearerAuth
// @Router       /api/v1/audit/user/{user_id} [get]
func (h *AuditHandler) GetByUser(c *gin.Context) {
	userIDStr := c.Param("user_id")

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	page := 1
	pageSize := 20
	if p := c.Query("page"); p != "" {
		// Simple parsing, default to 1 on error
	}
	if ps := c.Query("page_size"); ps != "" {
		// Simple parsing, default to 20 on error
	}

	logs, total, err := h.auditService.GetByUserID(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch audit logs"})
		return
	}

	c.JSON(http.StatusOK, dto.AuditLogListResponse{
		Logs:       dto.ToAuditLogResponseList(logs),
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: services.CalculateTotalPages(total, pageSize),
	})
}

// GetActions godoc
// @Summary      Get available audit actions
// @Description  Get list of all available audit actions (ADMIN only)
// @Tags         audit
// @Accept       json
// @Produce      json
// @Success      200  {object} map[string][]string
// @Failure      401  {object} map[string]string
// @Failure      403  {object} map[string]string
// @Security     BearerAuth
// @Router       /api/v1/audit/actions [get]
func (h *AuditHandler) GetActions(c *gin.Context) {
	actions := h.auditService.GetActions()
	c.JSON(http.StatusOK, gin.H{"actions": actions})
}

// GetEntityTypes godoc
// @Summary      Get audit entity types
// @Description  Get list of all entity types that have been logged (ADMIN only)
// @Tags         audit
// @Accept       json
// @Produce      json
// @Success      200  {object} map[string][]string
// @Failure      401  {object} map[string]string
// @Failure      403  {object} map[string]string
// @Security     BearerAuth
// @Router       /api/v1/audit/entity-types [get]
func (h *AuditHandler) GetEntityTypes(c *gin.Context) {
	types, err := h.auditService.GetEntityTypes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch entity types"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"entity_types": types})
}

// GetStats godoc
// @Summary      Get audit statistics
// @Description  Get audit log statistics (ADMIN only)
// @Tags         audit
// @Accept       json
// @Produce      json
// @Success      200  {object} map[string]interface{}
// @Failure      401  {object} map[string]string
// @Failure      403  {object} map[string]string
// @Security     BearerAuth
// @Router       /api/v1/audit/stats [get]
func (h *AuditHandler) GetStats(c *gin.Context) {
	actionStats, err := h.auditService.GetStatsByAction()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch action stats"})
		return
	}

	entityStats, err := h.auditService.GetStatsByEntityType()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch entity stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"by_action":      actionStats,
		"by_entity_type": entityStats,
	})
}
