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
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByEmailID(ctx context.Context, userID string) (*entity.User, error)
	GetByEmailAddress(ctx context.Context, email string) (*entity.User, error)
	List(ctx context.Context, offset, limit int) ([]entity.User, int, error)
	Update(ctx context.Context, user *entity.User) error
	UpdatePassword(ctx context.Context, id uuid.UUID, hashed string) error
	Delete(ctx context.Context, id string) error
}
