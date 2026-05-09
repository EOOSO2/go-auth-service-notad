package repository

import (
	"context"
	"errors"

	"auth-service/internal/domain/entity"

	"github.com/google/uuid"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository interface {
	Create(ctx context.Context, user entity.User) (uuid.UUID, error)
	GetByEmailID(ctx context.Context, userID string) (*entity.User, error)
	GetByEmailAddress(ctx context.Context, email string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id string) error
}
