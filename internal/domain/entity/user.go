package entity

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	EmailAddress string    `json:"email_address"`
	UserID       string    `json:"user_id"` // Firebase UID or empty for password users
	Password     string    `json:"-"`
	Permission   []string  `json:"permission"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func NewUser(userID, firstName, lastName, email, password string, permission []string) (User, error) {
	if strings.TrimSpace(firstName) == "" || strings.TrimSpace(lastName) == "" {
		return User{}, errors.New("first and last name cannot be empty")
	}
	if len(password) < 8 {
		return User{}, errors.New("password must be at least 8 characters")
	}

	id, err := uuid.NewV7()
	if err != nil {
		return User{}, err
	}

	now := time.Now()
	return User{
		ID:           id,
		UserID:       strings.TrimSpace(userID),
		EmailAddress: strings.ToLower(strings.TrimSpace(email)),
		FirstName:    strings.TrimSpace(firstName),
		LastName:     strings.TrimSpace(lastName),
		Password:     password,
		Permission:   permission,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}
