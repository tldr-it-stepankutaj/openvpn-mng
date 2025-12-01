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
func SetupRoutes(r *gin.Engine, cfg *config.Config) {
	// Static files
	r.Static("/static", "./web/static")

	// HTML templates
	r.LoadHTMLGlob("web/templates/*")

	// Handlers
	authHandler := handlers.NewAuthHandler(&cfg.Auth)
	userHandler := handlers.NewUserHandler()
	groupHandler := handlers.NewGroupHandler()
	networkHandler := handlers.NewNetworkHandler()
	webHandler := handlers.NewWebHandler()

	// Web routes (HTML pages)
	webRoutes := r.Group("/")
	{
		// Public routes
		webRoutes.GET("/", webHandler.IndexPage)
		webRoutes.GET("/login", webHandler.LoginPage)

		// Protected web routes
		protected := webRoutes.Group("/")
		protected.Use(middleware.AuthMiddleware(&cfg.Auth))
		{
			protected.GET("/dashboard", webHandler.DashboardPage)
			protected.GET("/users", webHandler.UsersPage)
			protected.GET("/users/:id", webHandler.UserDetailPage)
			protected.GET("/groups", webHandler.GroupsPage)
			protected.GET("/profile", webHandler.ProfilePage)
		}
	}

	// API routes (if enabled)
	if cfg.API.Enabled {
		api := r.Group("/api/v1")
		{
			// Auth routes (public)
			auth := api.Group("/auth")
			{
				auth.POST("/login", authHandler.Login)
			}

			// Protected API routes
			protected := api.Group("/")
			protected.Use(middleware.AuthMiddleware(&cfg.Auth))
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
					users.GET("", userHandler.List)                                        // Access control in handler
					users.POST("", middleware.RequireManagerOrAdmin(), userHandler.Create) // MANAGER can create subordinates
					users.GET("/:id", userHandler.Get)                                     // Access control in handler
					users.PUT("/:id", userHandler.Update)                                  // Access control in handler
					users.DELETE("/:id", middleware.RequireAdmin(), userHandler.Delete)    // Only ADMIN can delete
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
