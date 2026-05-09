package handler

import (
	"log/slog"
	"net/http"
	"strings"

	"auth-service/internal/application/usecase"
	"auth-service/internal/domain/port/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authUseCase usecase.AuthUseCase
}

func NewAuthHandler(uc usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{authUseCase: uc}
}

type LoginRequest struct {
	Type     string `json:"type" binding:"required"` // "password" or "google"
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
	Token    string `json:"token,omitempty"` // Firebase ID token
}

type SignupRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
}

type LoginResponse struct {
	UserID     string   `json:"user_id"`
	Email      string   `json:"email"`
	Permission []string `json:"permission"`
	Token      string   `json:"token"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authType := service.AuthPassword
	if req.Type == "google" {
		authType = service.AuthGoogle
	}

	claims, token, err := h.authUseCase.Login(c.Request.Context(), service.AuthRequest{
		Type:     authType,
		Email:    req.Email,
		Password: req.Password,
		Token:    req.Token,
	})
	if err != nil {
		slog.Warn("login failed", "type", req.Type, "email", req.Email, "error", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		UserID:     claims.UserID,
		Email:      claims.EmailAddress,
		Permission: claims.Permission,
		Token:      token,
	})
}

func (h *AuthHandler) Verify(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
		return
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
		return
	}
	tokenString := strings.TrimSpace(parts[1])
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token missing"})
		return
	}

	claims, err := h.authUseCase.Authenticate(c.Request.Context(), tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":    claims.UserID,
		"email_id":   claims.EmailID,
		"email":      claims.EmailAddress,
		"permission": claims.Permission,
	})
}

func (h *AuthHandler) Signup(c *gin.Context) {
	var req SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := h.authUseCase.Signup(c.Request.Context(), req.FirstName, req.LastName, req.Email, req.Password)
	if err != nil {
		slog.Warn("signup failed", "email", req.Email, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "user created successfully",
		"user_id": userID,
	})
}
