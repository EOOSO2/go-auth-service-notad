package handler

import (
	"errors"
	"net/http"
	"strconv"

	"auth-service/internal/application/usecase"
	"auth-service/internal/constant"
	"auth-service/internal/domain/port/repository"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userUseCase usecase.UserUseCase
}

func NewUserHandler(uc usecase.UserUseCase) *UserHandler {
	return &UserHandler{userUseCase: uc}
}

// GET /auth/me — current user profile (resolved from JWT).
func (h *UserHandler) GetMe(c *gin.Context) {
	userID, _ := c.Get(string(constant.CtxUserID))
	id, ok := userID.(string)
	if !ok || id == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, err := h.userUseCase.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

// PATCH /auth/me — self update (first_name, last_name, password).
func (h *UserHandler) UpdateMe(c *gin.Context) {
	userID, _ := c.Get(string(constant.CtxUserID))
	id, ok := userID.(string)
	if !ok || id == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req usecase.UpdateMeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userUseCase.UpdateMe(c.Request.Context(), id, req)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, repository.ErrUserNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

// GET /admin/users — paginated list.
func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	result, err := h.userUseCase.List(c.Request.Context(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// PATCH /admin/users/:id — admin update (first_name, last_name, permission).
func (h *UserHandler) AdminUpdate(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id required"})
		return
	}

	var req usecase.AdminUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userUseCase.AdminUpdate(c.Request.Context(), id, req)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, repository.ErrUserNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}
