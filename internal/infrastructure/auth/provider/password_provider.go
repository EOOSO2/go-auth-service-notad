package provider

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"auth-service/internal/domain/port/repository"
	"auth-service/internal/domain/port/service"

	"golang.org/x/crypto/bcrypt"
)

type PasswordProvider struct {
	userRepo repository.UserRepository
}

func NewPasswordProvider(userRepo repository.UserRepository) *PasswordProvider {
	return &PasswordProvider{userRepo: userRepo}
}

func (p *PasswordProvider) Authenticate(ctx context.Context, req service.AuthRequest) (*service.AuthResult, error) {
	if req.Email == "" || req.Password == "" {
		return nil, errors.New("email and password are required")
	}

	user, err := p.userRepo.GetByEmailAddress(ctx, req.Email)
	if err != nil {
		fmt.Println(err)
		slog.Debug("user lookup failed during password auth", "error", err)
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return &service.AuthResult{User: user}, nil
}
