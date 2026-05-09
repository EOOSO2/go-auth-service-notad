package router

import (
	"net/http"
	"os"
	"strings"

	"auth-service/internal/constant"
	"auth-service/internal/presentation/handler"
	"auth-service/internal/presentation/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewRouter(authHandler *handler.AuthHandler, authMiddleware *middleware.AuthMiddleware) http.Handler {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins(),
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Public routes
	auth := r.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/signup", authHandler.Signup)
		auth.GET("/verify", authHandler.Verify)
	}

	// Protected routes
	protected := r.Group("/api")
	protected.Use(authMiddleware.RequireAuth())
	{
		protected.GET("/profile", func(c *gin.Context) {
			userID, _ := c.Get(string(constant.CtxUserID))
			c.JSON(http.StatusOK, gin.H{"user_id": userID})
		})

		admin := protected.Group("/admin")
		admin.Use(authMiddleware.RequirePermission(constant.PermAdmin))
		{
			admin.GET("/users", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "admin only"})
			})
		}
	}

	return r
}

// allowedOrigins reads CORS_ALLOWED_ORIGINS env (comma-separated).
// Falls back to localhost dev origins when unset.
func allowedOrigins() []string {
	if v := os.Getenv("CORS_ALLOWED_ORIGINS"); v != "" {
		return strings.Split(v, ",")
	}
	return []string{"http://localhost:5173", "http://localhost:5174"}
}
