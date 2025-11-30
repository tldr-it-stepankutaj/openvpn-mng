package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/dto"
)

// IPFilter creates middleware that restricts access based on IP addresses
func IPFilter(allowedCIDRs []string) gin.HandlerFunc {
	// Parse all CIDR ranges
	var networks []*net.IPNet
	allowAll := false

	for _, cidr := range allowedCIDRs {
		cidr = strings.TrimSpace(cidr)
		if cidr == "" {
			continue
		}

		// Check for "allow all" patterns
		if cidr == "0.0.0.0/0" || cidr == "::/0" {
			allowAll = true
			break
		}

		// Parse CIDR
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			// Try parsing as single IP (add /32 for IPv4 or /128 for IPv6)
			ip := net.ParseIP(cidr)
			if ip != nil {
				if ip.To4() != nil {
					_, network, _ = net.ParseCIDR(cidr + "/32")
				} else {
					_, network, _ = net.ParseCIDR(cidr + "/128")
				}
			}
		}
		if network != nil {
			networks = append(networks, network)
		}
	}

	return func(c *gin.Context) {
		// If allow all or no restrictions configured, allow access
		if allowAll || len(networks) == 0 {
			c.Next()
			return
		}

		// Get client IP
		clientIP := getClientIP(c)
		ip := net.ParseIP(clientIP)
		if ip == nil {
			c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Error:   "Forbidden",
				Message: "Invalid client IP",
				Code:    http.StatusForbidden,
			})
			c.Abort()
			return
		}

		// Check if IP is in any allowed network
		for _, network := range networks {
			if network.Contains(ip) {
				c.Next()
				return
			}
		}

		// IP not allowed
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "Forbidden",
			Message: "Access denied from your IP address",
			Code:    http.StatusForbidden,
		})
		c.Abort()
	}
}

// getClientIP extracts the real client IP from the request
func getClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header (for reverse proxy)
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}
	return ip
}
