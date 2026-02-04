package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/dto"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/middleware"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/models"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/services"
)

// VpnClientConfigHandler handles VPN client configuration requests
type VpnClientConfigHandler struct {
	vpnClientConfigService *services.VpnClientConfigService
	auditLogger            *middleware.AuditLogger
}

// NewVpnClientConfigHandler creates a new VPN client config handler
func NewVpnClientConfigHandler() *VpnClientConfigHandler {
	return &VpnClientConfigHandler{
		vpnClientConfigService: services.NewVpnClientConfigService(),
		auditLogger:            middleware.NewAuditLogger(),
	}
}

// Get godoc
// @Summary Get VPN client configuration
// @Description Get the current VPN client configuration (Admin only)
// @Tags vpn-client-config
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.VpnClientConfigResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/vpn/client-config [get]
func (h *VpnClientConfigHandler) Get(c *gin.Context) {
	config, err := h.vpnClientConfigService.Get()
	if err != nil {
		if errors.Is(err, services.ErrVpnClientConfigNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "Not Found",
				Message: "VPN client configuration not found. Please configure it first.",
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

	c.JSON(http.StatusOK, dto.ToVpnClientConfigResponse(config))
}

// Update godoc
// @Summary Create or update VPN client configuration
// @Description Create or update the VPN client configuration (Admin only)
// @Tags vpn-client-config
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.VpnClientConfigRequest true "VPN client configuration"
// @Success 200 {object} dto.VpnClientConfigResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /api/v1/vpn/client-config [put]
func (h *VpnClientConfigHandler) Update(c *gin.Context) {
	var req dto.VpnClientConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Get old config for audit logging
	oldConfig, _ := h.vpnClientConfigService.Get()

	updatedBy := middleware.GetAuthUserID(c)
	config, err := h.vpnClientConfigService.CreateOrUpdate(&req, updatedBy)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCACert) || errors.Is(err, services.ErrInvalidTLSKey) {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "Bad Request",
				Message: err.Error(),
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

	// Audit logging
	if oldConfig != nil {
		h.auditLogger.LogUpdate(c, "vpn_client_config", config.ID, oldConfig, config)
	} else {
		h.auditLogger.LogCreate(c, "vpn_client_config", config.ID, config)
	}

	c.JSON(http.StatusOK, dto.ToVpnClientConfigResponse(config))
}

// Preview godoc
// @Summary Preview generated .ovpn configuration
// @Description Preview the generated .ovpn file content (Admin only)
// @Tags vpn-client-config
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.VpnClientConfigPreviewResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/vpn/client-config/preview [get]
func (h *VpnClientConfigHandler) Preview(c *gin.Context) {
	content, filename, err := h.vpnClientConfigService.GenerateOvpnConfig()
	if err != nil {
		if errors.Is(err, services.ErrVpnClientConfigNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "Not Found",
				Message: "VPN client configuration not found. Please configure it first.",
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

	c.JSON(http.StatusOK, dto.VpnClientConfigPreviewResponse{
		Content:  content,
		Filename: filename,
	})
}

// Download godoc
// @Summary Download .ovpn configuration file
// @Description Download the generated .ovpn file (Any authenticated user)
// @Tags vpn-client-config
// @Produce application/x-openvpn-profile
// @Security BearerAuth
// @Success 200 {file} file "OpenVPN configuration file"
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/vpn/client-config/download [get]
func (h *VpnClientConfigHandler) Download(c *gin.Context) {
	content, filename, err := h.vpnClientConfigService.GenerateOvpnConfig()
	if err != nil {
		if errors.Is(err, services.ErrVpnClientConfigNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "Not Found",
				Message: "VPN client configuration not available. Please contact your administrator.",
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

	// Log download action
	authUser := middleware.GetAuthUser(c)
	h.auditLogger.Log(c, models.AuditActionRead, "vpn_client_config", &models.WellKnownVpnClientConfigID, nil, nil, "Downloaded .ovpn file by user: "+authUser.Username)

	// Set headers for file download
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/x-openvpn-profile")
	c.Header("Content-Length", strconv.Itoa(len(content)))

	c.String(http.StatusOK, content)
}

// GetDefaultTemplate godoc
// @Summary Get default OpenVPN client template
// @Description Get the default template for OpenVPN client configuration (Admin only)
// @Tags vpn-client-config
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.DefaultTemplateResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /api/v1/vpn/client-config/default-template [get]
func (h *VpnClientConfigHandler) GetDefaultTemplate(c *gin.Context) {
	template := h.vpnClientConfigService.GetDefaultTemplate()

	c.JSON(http.StatusOK, dto.DefaultTemplateResponse{
		Template: template,
	})
}
