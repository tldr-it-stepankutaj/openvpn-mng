package routes

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/config"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/handlers"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/middleware"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/models"
)

// SetupRoutes sets up all routes
func SetupRoutes(r *gin.Engine, cfg *config.Config, rateLimiter *middleware.RateLimiter, blacklist *middleware.TokenBlacklist) {
	// Static files
	r.Static("/static", "./web/static")

	// HTML templates
	r.LoadHTMLGlob("web/templates/*")

	// Handlers
	authHandler := handlers.NewAuthHandler(&cfg.Auth, blacklist, &cfg.Security)
	userHandler := handlers.NewUserHandler(&cfg.VPN)
	groupHandler := handlers.NewGroupHandler()
	networkHandler := handlers.NewNetworkHandler()
	vpnSessionHandler := handlers.NewVpnSessionHandler()
	vpnAuthHandler := handlers.NewVpnAuthHandler()
	vpnIPHandler := handlers.NewVPNIPHandler(&cfg.VPN)
	vpnClientConfigHandler := handlers.NewVpnClientConfigHandler()
	auditHandler := handlers.NewAuditHandler()
	webHandler := handlers.NewWebHandler()

	// Web routes (HTML pages)
	webRoutes := r.Group("/")
	{
		// Public routes
		webRoutes.GET("/", webHandler.IndexPage)
		webRoutes.GET("/login", webHandler.LoginPage)

		// Protected web routes
		protected := webRoutes.Group("/")
		protected.Use(middleware.AuthMiddleware(&cfg.Auth, blacklist))
		{
			protected.GET("/dashboard", webHandler.DashboardPage)
			protected.GET("/users", webHandler.UsersPage)
			protected.GET("/users/:id", webHandler.UserDetailPage)
			protected.GET("/groups", webHandler.GroupsPage)
			protected.GET("/networks", webHandler.NetworksPage)
			protected.GET("/audit", webHandler.AuditPage)
			protected.GET("/sessions", webHandler.SessionsPage)
			protected.GET("/profile", webHandler.ProfilePage)
			protected.GET("/vpn-settings", webHandler.VpnSettingsPage)
		}
	}

	// API routes (if enabled)
	if cfg.API.Enabled {
		api := r.Group("/api/v1")
		{
			// Auth routes (public)
			auth := api.Group("/auth")
			{
				if rateLimiter != nil {
					auth.POST("/login", rateLimiter.Middleware(), authHandler.Login)
				} else {
					auth.POST("/login", authHandler.Login)
				}
			}

			// Protected API routes
			protected := api.Group("/")
			protected.Use(middleware.AuthMiddleware(&cfg.Auth, blacklist))
			{
				// Auth
				protected.POST("/auth/logout", authHandler.Logout)
				protected.GET("/auth/me", authHandler.Me)

				// Users
				users := protected.Group("/users")
				{
					// Profile endpoints (must be before /:id to avoid conflict)
					users.PUT("/profile", userHandler.UpdateProfile)
					users.PUT("/password", userHandler.UpdatePassword)

					// CRUD endpoints with role-based access
					users.GET("", userHandler.List)
					users.POST("", middleware.RequireManagerOrAdmin(), userHandler.Create)
					users.GET("/:id", userHandler.Get)
					users.PUT("/:id", userHandler.Update)
					users.DELETE("/:id", middleware.RequireAdmin(), userHandler.Delete)

					// User groups management
					users.GET("/:id/groups", userHandler.GetGroups)
					users.POST("/:id/groups", middleware.RequireRole(models.RoleAdmin, models.RoleManager), userHandler.AddGroup)
					users.DELETE("/:id/groups/:group_id", middleware.RequireRole(models.RoleAdmin, models.RoleManager), userHandler.RemoveGroup)
				}

				// Groups
				groups := protected.Group("/groups")
				{
					groups.GET("", groupHandler.List)
					groups.POST("", middleware.RequireAdmin(), groupHandler.Create)
					groups.GET("/:id", groupHandler.Get)
					groups.PUT("/:id", middleware.RequireAdmin(), groupHandler.Update)
					groups.DELETE("/:id", middleware.RequireAdmin(), groupHandler.Delete)
					groups.GET("/:id/users", groupHandler.GetUsers)
					groups.POST("/:id/users", middleware.RequireRole(models.RoleAdmin, models.RoleManager), groupHandler.AddUser)
					groups.DELETE("/:id/users/:user_id", middleware.RequireRole(models.RoleAdmin, models.RoleManager), groupHandler.RemoveUser)
					groups.GET("/:id/networks", groupHandler.GetNetworks)
					groups.POST("/:id/networks", middleware.RequireAdmin(), groupHandler.AddNetwork)
					groups.DELETE("/:id/networks/:network_id", middleware.RequireAdmin(), groupHandler.RemoveNetwork)
				}

				// Networks (Admin only)
				networks := protected.Group("/networks")
				networks.Use(middleware.RequireAdmin())
				{
					networks.GET("", networkHandler.List)
					networks.POST("", networkHandler.Create)
					networks.GET("/:id", networkHandler.Get)
					networks.PUT("/:id", networkHandler.Update)
					networks.DELETE("/:id", networkHandler.Delete)
					networks.GET("/:id/groups", networkHandler.GetGroups)
					networks.POST("/:id/groups", networkHandler.AddGroup)
					networks.DELETE("/:id/groups/:group_id", networkHandler.RemoveGroup)
				}

				// VPN Sessions
				vpn := protected.Group("/vpn")
				{
					// VPN IP allocation endpoints
					vpn.GET("/next-ip", vpnIPHandler.GetNextAvailableIP)
					vpn.GET("/network-info", vpnIPHandler.GetNetworkInfo)
					vpn.POST("/validate-ip", vpnIPHandler.ValidateIP)
					vpn.GET("/used-ips", vpnIPHandler.GetUsedIPs)

					// VPN Client Config - Download available to any authenticated user
					vpn.GET("/client-config/download", vpnClientConfigHandler.Download)

					// Write endpoints - any authenticated user (VPN server calls these)
					vpn.POST("/sessions", vpnSessionHandler.Create)
					vpn.PUT("/sessions/:id/disconnect", vpnSessionHandler.Disconnect)
					vpn.POST("/traffic-stats", vpnSessionHandler.CreateTrafficStats)

					// Read endpoints - Admin only
					vpnAdmin := vpn.Group("")
					vpnAdmin.Use(middleware.RequireAdmin())
					{
						vpnAdmin.GET("/sessions", vpnSessionHandler.List)
						vpnAdmin.GET("/sessions/active", vpnSessionHandler.GetActive)
						vpnAdmin.GET("/sessions/:id", vpnSessionHandler.Get)
						vpnAdmin.GET("/stats", vpnSessionHandler.GetStats)
						vpnAdmin.GET("/stats/users", vpnSessionHandler.GetUserStats)
						vpnAdmin.GET("/traffic-stats", vpnSessionHandler.ListTrafficStats)

						// VPN Client Config management - Admin only
						vpnAdmin.GET("/client-config", vpnClientConfigHandler.Get)
						vpnAdmin.PUT("/client-config", vpnClientConfigHandler.Update)
						vpnAdmin.GET("/client-config/preview", vpnClientConfigHandler.Preview)
						vpnAdmin.GET("/client-config/default-template", vpnClientConfigHandler.GetDefaultTemplate)
					}
				}

				// Audit logs (Admin only)
				audit := protected.Group("/audit")
				audit.Use(middleware.RequireAdmin())
				{
					audit.GET("", auditHandler.List)
					audit.GET("/actions", auditHandler.GetActions)
					audit.GET("/entity-types", auditHandler.GetEntityTypes)
					audit.GET("/stats", auditHandler.GetStats)
					audit.GET("/entity/:entity_type/:entity_id", auditHandler.GetByEntity)
					audit.GET("/user/:user_id", auditHandler.GetByUser)
					audit.GET("/:id", auditHandler.Get)
				}
			}
		}

		// VPN Auth routes (token-based authentication for OpenVPN server)
		if cfg.API.VpnToken != "" {
			vpnAuth := api.Group("/vpn-auth")
			vpnAuth.Use(middleware.VpnTokenAuth(cfg.API.VpnToken))
			{
				if rateLimiter != nil {
					vpnAuth.POST("/authenticate", rateLimiter.Middleware(), vpnAuthHandler.Authenticate)
				} else {
					vpnAuth.POST("/authenticate", vpnAuthHandler.Authenticate)
				}
				vpnAuth.GET("/users", vpnAuthHandler.ListAllUsers)
				vpnAuth.GET("/users/:id", vpnAuthHandler.GetUserByID)
				vpnAuth.GET("/users/:id/routes", vpnAuthHandler.GetUserRoutes)
				vpnAuth.GET("/users/by-username/:username", vpnAuthHandler.GetUserByUsername)
				vpnAuth.POST("/sessions", vpnAuthHandler.CreateSession)
				vpnAuth.PUT("/sessions/:id/disconnect", vpnAuthHandler.DisconnectSession)
			}
		}

		// Swagger documentation (if enabled)
		if cfg.API.SwaggerEnabled {
			swagger := r.Group("/swagger")
			swagger.Use(middleware.IPFilter(cfg.API.SwaggerAllowedIPs))
			swagger.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		}
	}
}
