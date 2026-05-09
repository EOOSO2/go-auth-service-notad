package middleware

import (
	"net/http"
	"strings"

	"auth-service/internal/application/usecase"
	"auth-service/internal/constant"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	authUseCase usecase.AuthUseCase
}

func NewAuthMiddleware(authUseCase usecase.AuthUseCase) *AuthMiddleware {
	return &AuthMiddleware{authUseCase: authUseCase}
}

// RequireAuth validates JWT + Redis session (single-session enforcement).
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			c.Abort()
			return
		}
		tokenString := strings.TrimSpace(parts[1])
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token missing"})
			c.Abort()
			return
		}

		claims, err := m.authUseCase.Authenticate(c.Request.Context(), tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication failed"})
			c.Abort()
			return
		}

		c.Set(string(constant.CtxUserID), claims.UserID)
		c.Set(string(constant.CtxEmailID), claims.EmailID)
		c.Set(string(constant.CtxEmail), claims.EmailAddress)
		c.Set(string(constant.CtxPermission), claims.Permission)

		c.Next()
	}
}

// RequirePermission checks for a specific role (ADMIN, TEACHER, STUDENT).
// Must be used after RequireAuth.
func (m *AuthMiddleware) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		permsInterface, exists := c.Get(string(constant.CtxPermission))
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
			c.Abort()
			return
		}

		perms, ok := permsInterface.([]string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
			c.Abort()
			return
		}

		if !hasPermission(perms, permission) {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permission"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func hasPermission(perms []string, target string) bool {
	for _, p := range perms {
		if p == target {
			return true
		}
	}
	return false
}
