package provider

import (
	"context"
	"errors"
	"log/slog"

	"auth-service/internal/constant"
	"auth-service/internal/domain/entity"
	"auth-service/internal/domain/port/repository"
	"auth-service/internal/domain/port/service"
	"auth-service/internal/infrastructure/auth"
)

type GoogleProvider struct {
	userRepo        repository.UserRepository
	firebaseService *auth.FirebaseService
}

func NewGoogleProvider(userRepo repository.UserRepository, firebaseSvc *auth.FirebaseService) *GoogleProvider {
	return &GoogleProvider{
		userRepo:        userRepo,
		firebaseService: firebaseSvc,
	}
}

func (p *GoogleProvider) Authenticate(ctx context.Context, req service.AuthRequest) (*service.AuthResult, error) {
	if req.Token == "" {
		return nil, errors.New("firebase id token is required for google login")
	}

	uid, email, name, err := p.firebaseService.VerifyIDToken(ctx, req.Token)
	if err != nil {
		return nil, errors.New("invalid google token")
	}

	user, err := p.userRepo.GetByEmailID(ctx, uid)
	if err != nil {
		if !errors.Is(err, repository.ErrUserNotFound) {
			return nil, err
		}
		// First-time Google login: auto-create user
		slog.Info("creating new user via Google login", "email", email)
		newUser, err := entity.NewUser(uid, name, name, email, uid, []string{constant.PermStudent})
		if err != nil {
			return nil, err
		}
		createdID, err := p.userRepo.Create(ctx, newUser)
		if err != nil {
			return nil, err
		}
		newUser.ID = createdID
		return &service.AuthResult{User: &newUser}, nil
	}

	if email != "" && user.EmailAddress != email {
		return nil, errors.New("email mismatch")
	}

	return &service.AuthResult{User: user}, nil
}
