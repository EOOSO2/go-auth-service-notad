package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"auth-service/internal/constant"
	"auth-service/internal/domain/entity"
	"auth-service/internal/domain/port/repository"
	"auth-service/internal/domain/port/service"
	"auth-service/internal/domain/port/session"

	"github.com/google/uuid"
)

// AuthUseCase defines the application's authentication contract.
type AuthUseCase interface {
	Login(ctx context.Context, req service.AuthRequest) (*entity.AuthClaims, string, error)
	Authenticate(ctx context.Context, tokenString string) (*entity.AuthClaims, error)
	Signup(ctx context.Context, firstName, lastName, email, password string) (uuid.UUID, error)
}

type authUseCase struct {
	userRepo        repository.UserRepository
	tokenSvc        service.TokenService
	sessionStore    session.SessionStore
	providerFactory service.AuthProviderFactory
}

func NewAuthUseCase(
	repo repository.UserRepository,
	token service.TokenService,
	sessionStore session.SessionStore,
	factory service.AuthProviderFactory,
) AuthUseCase {
	return &authUseCase{
		userRepo:        repo,
		tokenSvc:        token,
		sessionStore:    sessionStore,
		providerFactory: factory,
	}
}

// Authenticate validates JWT and checks Redis session (single-session enforcement).
// tokenString must be the raw JWT without the "Bearer " prefix.
func (uc *authUseCase) Authenticate(ctx context.Context, tokenString string) (*entity.AuthClaims, error) {
	claims, err := uc.tokenSvc.Validate(tokenString)
	if err != nil {
		return nil, err
	}

	// Session key uses the UUID primary key, consistent with Login.
	sessionKey := "user_token:" + claims.UserID
	storedToken, err := uc.sessionStore.Get(ctx, sessionKey)
	if err != nil {
		return nil, errors.New("session expired or invalid")
	}
	if storedToken != tokenString {
		return nil, errors.New("session superseded by another login")
	}

	return claims, nil
}

func (uc *authUseCase) Login(ctx context.Context, req service.AuthRequest) (*entity.AuthClaims, string, error) {
	provider, err := uc.providerFactory.Get(req.Type)
	if err != nil {
		return nil, "", err
	}

	result, err := provider.Authenticate(ctx, req)
	if err != nil {
		return nil, "", err
	}

	token, err := uc.tokenSvc.Create(result.User)
	if err != nil {
		return nil, "", err
	}

	// Session key uses UUID primary key — must match Authenticate's lookup key.
	sessionKey := fmt.Sprintf("user_token:%s", result.User.ID.String())
	if err := uc.sessionStore.Set(ctx, sessionKey, token, 24*time.Hour); err != nil {
		slog.Error("failed to persist session", "user_id", result.User.ID, "error", err)
	}

	return &entity.AuthClaims{
		UserID:       result.User.ID.String(),
		EmailID:      result.User.UserID,
		EmailAddress: result.User.EmailAddress,
		Permission:   result.User.Permission,
	}, token, nil
}

func (uc *authUseCase) Signup(ctx context.Context, firstName, lastName, email, password string) (uuid.UUID, error) {
	user, err := entity.NewUser("", firstName, lastName, email, password, []string{constant.PermStudent})
	if err != nil {
		return uuid.Nil, err
	}
	return uc.userRepo.Create(ctx, user)
}
