package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/middleware"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/models"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/services"
)

// WebHandler handles web page requests
type WebHandler struct {
	userService            *services.UserService
	groupService           *services.GroupService
	networkService         *services.NetworkService
	dashboardService       *services.DashboardService
	vpnClientConfigService *services.VpnClientConfigService
}

// NewWebHandler creates a new web handler
func NewWebHandler() *WebHandler {
	return &WebHandler{
		userService:            services.NewUserService(),
		groupService:           services.NewGroupService(),
		networkService:         services.NewNetworkService(),
		dashboardService:       services.NewDashboardService(),
		vpnClientConfigService: services.NewVpnClientConfigService(),
	}
}

// IndexPage redirects to login
func (h *WebHandler) IndexPage(c *gin.Context) {
	c.Redirect(http.StatusFound, "/login")
}

// LoginPage renders the login page
func (h *WebHandler) LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{
		"title": "Login - OpenVPN Manager",
	})
}

// DashboardPage renders the dashboard page
func (h *WebHandler) DashboardPage(c *gin.Context) {
	authUser := middleware.GetAuthUser(c)
	authUserID := middleware.GetAuthUserID(c)

	user, _ := h.userService.GetByID(authUserID)

	data := gin.H{
		"title": "Dashboard - OpenVPN Manager",
		"user":  user,
		"role":  authUser.Role,
	}

	// Add stats for ADMIN role
	if authUser.Role == models.RoleAdmin {
		stats, err := h.dashboardService.GetAdminStats()
		if err == nil {
			data["stats"] = stats
		}
	}

	// Add subordinate count for MANAGER role
	if authUser.Role == models.RoleManager {
		count, err := h.dashboardService.GetManagerSubordinateCount(authUserID.String())
		if err == nil {
			data["subordinateCount"] = count
		}
	}

	c.HTML(http.StatusOK, "dashboard.html", data)
}

// UsersPage renders the users list page
func (h *WebHandler) UsersPage(c *gin.Context) {
	authUser := middleware.GetAuthUser(c)
	authUserID := middleware.GetAuthUserID(c)

	var users []models.User
	var err error

	if authUser.Role == models.RoleAdmin {
		users, _, err = h.userService.List(1, 100, models.RoleAdmin, nil)
	} else if authUser.Role == models.RoleManager {
		users, _, err = h.userService.GetManagedUsers(authUserID, 1, 100)
	} else {
		// Regular users can only see themselves
		user, _ := h.userService.GetByID(authUserID)
		users = []models.User{*user}
	}

	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"title":   "Error - OpenVPN Manager",
			"message": "Failed to load users",
		})
		return
	}

	c.HTML(http.StatusOK, "users.html", gin.H{
		"title": "Users - OpenVPN Manager",
		"users": users,
		"role":  authUser.Role,
	})
}

// UserDetailPage renders the user detail page
func (h *WebHandler) UserDetailPage(c *gin.Context) {
	// Implementation for user detail page
	c.HTML(http.StatusOK, "user_detail.html", gin.H{
		"title": "User Detail - OpenVPN Manager",
	})
}

// GroupsPage renders the groups list page
func (h *WebHandler) GroupsPage(c *gin.Context) {
	authUser := middleware.GetAuthUser(c)

	groups, _, err := h.groupService.List(1, 100)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"title":   "Error - OpenVPN Manager",
			"message": "Failed to load groups",
		})
		return
	}

	c.HTML(http.StatusOK, "groups.html", gin.H{
		"title":  "Groups - OpenVPN Manager",
		"groups": groups,
		"role":   authUser.Role,
	})
}

// ProfilePage renders the profile page
func (h *WebHandler) ProfilePage(c *gin.Context) {
	authUser := middleware.GetAuthUser(c)
	authUserID := middleware.GetAuthUserID(c)

	user, err := h.userService.GetByID(authUserID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"title":   "Error - OpenVPN Manager",
			"message": "Failed to load profile",
		})
		return
	}

	c.HTML(http.StatusOK, "profile.html", gin.H{
		"title": "Profile - OpenVPN Manager",
		"user":  user,
		"role":  authUser.Role,
	})
}

// NetworksPage renders the networks list page (ADMIN only)
func (h *WebHandler) NetworksPage(c *gin.Context) {
	authUser := middleware.GetAuthUser(c)

	// Only ADMIN can access networks
	if authUser.Role != models.RoleAdmin {
		c.HTML(http.StatusForbidden, "error.html", gin.H{
			"title":   "Access Denied - OpenVPN Manager",
			"message": "You don't have permission to access this page",
		})
		return
	}

	networks, _, err := h.networkService.List(1, 100)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"title":   "Error - OpenVPN Manager",
			"message": "Failed to load networks",
		})
		return
	}

	c.HTML(http.StatusOK, "networks.html", gin.H{
		"title":    "Networks - OpenVPN Manager",
		"networks": networks,
		"role":     authUser.Role,
	})
}

// AuditPage renders the audit logs page (ADMIN only)
func (h *WebHandler) AuditPage(c *gin.Context) {
	authUser := middleware.GetAuthUser(c)

	// Only ADMIN can access audit logs
	if authUser.Role != models.RoleAdmin {
		c.HTML(http.StatusForbidden, "error.html", gin.H{
			"title":   "Access Denied - OpenVPN Manager",
			"message": "You don't have permission to access this page",
		})
		return
	}

	c.HTML(http.StatusOK, "audit.html", gin.H{
		"title": "Audit Logs - OpenVPN Manager",
		"role":  authUser.Role,
	})
}

// SessionsPage renders the VPN sessions history page (ADMIN only)
func (h *WebHandler) SessionsPage(c *gin.Context) {
	authUser := middleware.GetAuthUser(c)

	// Only ADMIN can access session history
	if authUser.Role != models.RoleAdmin {
		c.HTML(http.StatusForbidden, "error.html", gin.H{
			"title":   "Access Denied - OpenVPN Manager",
			"message": "You don't have permission to access this page",
		})
		return
	}

	c.HTML(http.StatusOK, "sessions.html", gin.H{
		"title": "VPN Session History - OpenVPN Manager",
		"role":  authUser.Role,
	})
}

// VpnSettingsPage renders the VPN settings page (ADMIN only)
func (h *WebHandler) VpnSettingsPage(c *gin.Context) {
	authUser := middleware.GetAuthUser(c)

	// Only ADMIN can access VPN settings
	if authUser.Role != models.RoleAdmin {
		c.HTML(http.StatusForbidden, "error.html", gin.H{
			"title":   "Access Denied - OpenVPN Manager",
			"message": "You don't have permission to access this page",
		})
		return
	}

	// Get current config if exists
	config, _ := h.vpnClientConfigService.Get()

	// Get default template
	defaultTemplate := h.vpnClientConfigService.GetDefaultTemplate()

	c.HTML(http.StatusOK, "vpn_settings.html", gin.H{
		"title":           "VPN Settings - OpenVPN Manager",
		"role":            authUser.Role,
		"config":          config,
		"defaultTemplate": defaultTemplate,
	})
}
