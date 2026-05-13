package usecase

import (
	"context"
	"errors"
	"strings"

	"auth-service/internal/domain/entity"
	"auth-service/internal/domain/port/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UpdateMeRequest captures fields a user can self-update.
type UpdateMeRequest struct {
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Password  *string `json:"password,omitempty"`
}

// AdminUpdateRequest captures fields an admin can update on another user.
type AdminUpdateRequest struct {
	FirstName  *string   `json:"first_name,omitempty"`
	LastName   *string   `json:"last_name,omitempty"`
	Permission *[]string `json:"permission,omitempty"`
}

// PaginatedUsers is the standard list-response shape (CLAUDE.md G1 #10).
type PaginatedUsers struct {
	Items []entity.User `json:"items"`
	Total int           `json:"total"`
	Page  int           `json:"page"`
	Limit int           `json:"limit"`
}

type UserUseCase interface {
	GetByID(ctx context.Context, id string) (*entity.User, error)
	UpdateMe(ctx context.Context, id string, req UpdateMeRequest) (*entity.User, error)
	List(ctx context.Context, page, limit int) (PaginatedUsers, error)
	AdminUpdate(ctx context.Context, id string, req AdminUpdateRequest) (*entity.User, error)
}

type userUseCase struct {
	userRepo repository.UserRepository
}

func NewUserUseCase(repo repository.UserRepository) UserUseCase {
	return &userUseCase{userRepo: repo}
}

func parseUUID(id string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, errors.New("invalid user id")
	}
	return parsed, nil
}

func (uc *userUseCase) GetByID(ctx context.Context, id string) (*entity.User, error) {
	parsed, err := parseUUID(id)
	if err != nil {
		return nil, err
	}
	return uc.userRepo.GetByID(ctx, parsed)
}

func (uc *userUseCase) UpdateMe(ctx context.Context, id string, req UpdateMeRequest) (*entity.User, error) {
	parsed, err := parseUUID(id)
	if err != nil {
		return nil, err
	}
	user, err := uc.userRepo.GetByID(ctx, parsed)
	if err != nil {
		return nil, err
	}

	changed := false
	if req.FirstName != nil {
		v := strings.TrimSpace(*req.FirstName)
		if v == "" {
			return nil, errors.New("first_name cannot be empty")
		}
		user.FirstName = v
		changed = true
	}
	if req.LastName != nil {
		v := strings.TrimSpace(*req.LastName)
		if v == "" {
			return nil, errors.New("last_name cannot be empty")
		}
		user.LastName = v
		changed = true
	}

	if req.Password != nil {
		if len(*req.Password) < 8 {
			return nil, errors.New("password must be at least 8 characters")
		}
		hashed, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		if err := uc.userRepo.UpdatePassword(ctx, parsed, string(hashed)); err != nil {
			return nil, err
		}
		changed = true
	}

	if !changed {
		return nil, errors.New("no fields to update")
	}

	if req.FirstName != nil || req.LastName != nil {
		if err := uc.userRepo.Update(ctx, user); err != nil {
			return nil, err
		}
	}

	return uc.userRepo.GetByID(ctx, parsed)
}

func (uc *userUseCase) List(ctx context.Context, page, limit int) (PaginatedUsers, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	users, total, err := uc.userRepo.List(ctx, (page-1)*limit, limit)
	if err != nil {
		return PaginatedUsers{}, err
	}
	return PaginatedUsers{Items: users, Total: total, Page: page, Limit: limit}, nil
}

func (uc *userUseCase) AdminUpdate(ctx context.Context, id string, req AdminUpdateRequest) (*entity.User, error) {
	parsed, err := parseUUID(id)
	if err != nil {
		return nil, err
	}
	user, err := uc.userRepo.GetByID(ctx, parsed)
	if err != nil {
		return nil, err
	}

	changed := false
	if req.FirstName != nil {
		v := strings.TrimSpace(*req.FirstName)
		if v == "" {
			return nil, errors.New("first_name cannot be empty")
		}
		user.FirstName = v
		changed = true
	}
	if req.LastName != nil {
		v := strings.TrimSpace(*req.LastName)
		if v == "" {
			return nil, errors.New("last_name cannot be empty")
		}
		user.LastName = v
		changed = true
	}
	if req.Permission != nil {
		user.Permission = *req.Permission
		changed = true
	}

	if !changed {
		return nil, errors.New("no fields to update")
	}

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}
	return uc.userRepo.GetByID(ctx, parsed)
}
