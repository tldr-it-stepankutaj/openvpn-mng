package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/config"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/dto"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/services"
)

// VPNIPHandler handles VPN IP allocation requests
type VPNIPHandler struct {
	vpnIPService *services.VPNIPService
}

// NewVPNIPHandler creates a new VPN IP handler
func NewVPNIPHandler(cfg *config.VPNConfig) *VPNIPHandler {
	return &VPNIPHandler{
		vpnIPService: services.NewVPNIPService(cfg),
	}
}

// GetNextAvailableIP godoc
// @Summary Get next available VPN IP
// @Description Get the next available IP address in the VPN network range
// @Tags vpn
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.NextVPNIPResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/vpn/next-ip [get]
func (h *VPNIPHandler) GetNextAvailableIP(c *gin.Context) {
	ip, err := h.vpnIPService.GetNextAvailableIP()
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == services.ErrVPNNetworkNotConfigured {
			statusCode = http.StatusBadRequest
		} else if err == services.ErrNoAvailableIP {
			statusCode = http.StatusConflict
		}

		c.JSON(statusCode, dto.ErrorResponse{
			Error:   "Error",
			Message: err.Error(),
			Code:    statusCode,
		})
		return
	}

	c.JSON(http.StatusOK, dto.NextVPNIPResponse{
		IP: ip,
	})
}

// GetNetworkInfo godoc
// @Summary Get VPN network information
// @Description Get information about the VPN network configuration and usage
// @Tags vpn
// @Produce json
// @Security BearerAuth
// @Success 200 {object} services.VPNNetworkInfo
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/vpn/network-info [get]
func (h *VPNIPHandler) GetNetworkInfo(c *gin.Context) {
	info, err := h.vpnIPService.GetNetworkInfo()
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == services.ErrVPNNetworkNotConfigured {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, dto.ErrorResponse{
			Error:   "Error",
			Message: err.Error(),
			Code:    statusCode,
		})
		return
	}

	c.JSON(http.StatusOK, info)
}

// ValidateIP godoc
// @Summary Validate VPN IP
// @Description Validate that an IP address is valid for use in the VPN network
// @Tags vpn
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.ValidateVPNIPRequest true "IP to validate"
// @Success 200 {object} dto.ValidateVPNIPResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/v1/vpn/validate-ip [post]
func (h *VPNIPHandler) ValidateIP(c *gin.Context) {
	var req dto.ValidateVPNIPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Bad Request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	err := h.vpnIPService.ValidateIP(req.IP, req.ExcludeUserID)
	if err != nil {
		c.JSON(http.StatusOK, dto.ValidateVPNIPResponse{
			Valid:   false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.ValidateVPNIPResponse{
		Valid:   true,
		Message: "IP address is valid and available",
	})
}

// GetUsedIPs godoc
// @Summary Get used VPN IPs
// @Description Get a list of all VPN IP addresses currently in use
// @Tags vpn
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.UsedVPNIPsResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/vpn/used-ips [get]
func (h *VPNIPHandler) GetUsedIPs(c *gin.Context) {
	ips, err := h.vpnIPService.GetUsedIPs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Internal Server Error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, dto.UsedVPNIPsResponse{
		IPs: ips,
	})
}
