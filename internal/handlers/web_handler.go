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
	userService  *services.UserService
	groupService *services.GroupService
}

// NewWebHandler creates a new web handler
func NewWebHandler() *WebHandler {
	return &WebHandler{
		userService:  services.NewUserService(),
		groupService: services.NewGroupService(),
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

	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"title": "Dashboard - OpenVPN Manager",
		"user":  user,
		"role":  authUser.Role,
	})
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
