package service

import (
	"auth-service/internal/domain/entity"
	"context"
)

type AuthType string

const (
	AuthPassword AuthType = "password"
	AuthGoogle   AuthType = "google"
)

type AuthRequest struct {
	Type     AuthType
	Email    string
	Password string
	Token    string // Firebase ID token
}

type AuthResult struct {
	User *entity.User
}

type AuthProvider interface {
	Authenticate(ctx context.Context, req AuthRequest) (*AuthResult, error)
}

type AuthProviderFactory interface {
	Get(t AuthType) (AuthProvider, error)
}
