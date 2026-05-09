package service

import (
	"auth-service/internal/domain/entity"
)

type TokenService interface {
	Create(user *entity.User) (string, error)
	Validate(token string) (*entity.AuthClaims, error)
}
