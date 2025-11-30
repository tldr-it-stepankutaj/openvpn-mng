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
					users.GET("", middleware.RequireManagerOrAdmin(), userHandler.List)
					users.POST("", middleware.RequireAdmin(), userHandler.Create)
					users.GET("/:id", userHandler.Get)
					users.PUT("/:id", middleware.RequireManagerOrAdmin(), userHandler.Update)
					users.DELETE("/:id", middleware.RequireAdmin(), userHandler.Delete)
					users.PUT("/profile", userHandler.UpdateProfile)
					users.PUT("/password", userHandler.UpdatePassword)
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
