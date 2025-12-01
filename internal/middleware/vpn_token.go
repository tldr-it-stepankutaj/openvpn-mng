package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/dto"
)

const (
	VpnTokenHeader = "X-VPN-Token"
)

// VpnTokenAuth creates middleware for VPN token authentication
// This is used by the OpenVPN server to authenticate API requests
// instead of using a service account with JWT
func VpnTokenAuth(vpnToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if VPN token is configured
		if vpnToken == "" {
			c.JSON(http.StatusServiceUnavailable, dto.ErrorResponse{
				Error:   "Service Unavailable",
				Message: "VPN token authentication is not configured",
				Code:    http.StatusServiceUnavailable,
			})
			c.Abort()
			return
		}

		// Get token from header
		token := c.GetHeader(VpnTokenHeader)
		if token == "" {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Missing VPN token",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// Validate token
		if token != vpnToken {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Invalid VPN token",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
