package provider

import (
	"errors"

	"auth-service/internal/domain/port/repository"
	"auth-service/internal/domain/port/service"
	"auth-service/internal/infrastructure/auth"
)

type authProviderFactory struct {
	passwordProvider *PasswordProvider
	googleProvider   *GoogleProvider
}

func NewAuthProviderFactory(
	userRepo repository.UserRepository,
	firebaseSvc *auth.FirebaseService,
) service.AuthProviderFactory {

	return &authProviderFactory{
		passwordProvider: NewPasswordProvider(userRepo),
		googleProvider:   NewGoogleProvider(userRepo, firebaseSvc),
	}
}

func (f *authProviderFactory) Get(authType service.AuthType) (service.AuthProvider, error) {
	switch authType {
	case service.AuthPassword:
		return f.passwordProvider, nil
	case service.AuthGoogle:
		return f.googleProvider, nil
	default:
		return nil, errors.New("unsupported auth type")
	}
}
