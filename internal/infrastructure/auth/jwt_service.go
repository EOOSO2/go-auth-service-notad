package auth

import (
	"errors"
	"time"

	"auth-service/internal/domain/entity"
	"auth-service/internal/domain/port/service"

	"github.com/golang-jwt/jwt/v4"
)

type JWTService struct {
	secret []byte
}

func NewJWTService(secret []byte) service.TokenService {
	return &JWTService{secret: secret}
}

type jwtClaims struct {
	UserID       string   `json:"user_id"`
	EmailID      string   `json:"email_id"`
	EmailAddress string   `json:"email_address"`
	Permission   []string `json:"permission"`
	jwt.RegisteredClaims
}

func (s *JWTService) Create(user *entity.User) (string, error) {
	now := time.Now()
	claims := jwtClaims{
		UserID:       user.ID.String(),
		EmailID:      user.UserID,
		EmailAddress: user.EmailAddress,
		Permission:   user.Permission,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// Validate parses a raw JWT string (without "Bearer " prefix).
func (s *JWTService) Validate(tokenString string) (*entity.AuthClaims, error) {
	claims := &jwtClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return &entity.AuthClaims{
		UserID:       claims.UserID,
		EmailID:      claims.EmailID,
		EmailAddress: claims.EmailAddress,
		Permission:   claims.Permission,
	}, nil
}
