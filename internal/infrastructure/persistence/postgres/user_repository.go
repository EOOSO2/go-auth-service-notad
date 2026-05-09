package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"auth-service/internal/domain/entity"
	"auth-service/internal/domain/port/repository"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type userPostgresRepository struct {
	db *sql.DB
}

func NewUserPostgresRepository(db *sql.DB) repository.UserRepository {
	return &userPostgresRepository{db: db}
}

func (r *userPostgresRepository) Create(ctx context.Context, user entity.User) (uuid.UUID, error) {
	if user.Password != "" {
		hashed, err := hashPassword(user.Password)
		if err != nil {
			return uuid.Nil, err
		}
		user.Password = hashed
	}

	const query = `
		INSERT INTO users (id, email_id, email_address, first_name, last_name, password, permission, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`

	var id uuid.UUID
	err := r.db.QueryRowContext(ctx, query,
		user.ID, user.UserID, user.EmailAddress, user.FirstName, user.LastName,
		user.Password, pq.Array(user.Permission), user.CreatedAt, user.UpdatedAt,
	).Scan(&id)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (r *userPostgresRepository) GetByEmailID(ctx context.Context, userID string) (*entity.User, error) {
	const query = `
		SELECT id, email_id, email_address, first_name, last_name, password, permission, created_at, updated_at
		FROM users WHERE email_id = $1`

	user := &entity.User{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID, &user.UserID, &user.EmailAddress, &user.FirstName, &user.LastName,
		&user.Password, pq.Array(&user.Permission), &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *userPostgresRepository) GetByEmailAddress(ctx context.Context, email string) (*entity.User, error) {
	const query = `
		SELECT id, email_id, email_address, first_name, last_name, password, permission, created_at, updated_at
		FROM users WHERE email_address = $1`

	user := &entity.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.UserID, &user.EmailAddress, &user.FirstName, &user.LastName,
		&user.Password, pq.Array(&user.Permission), &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *userPostgresRepository) Update(ctx context.Context, user *entity.User) error {
	user.UpdatedAt = time.Now()
	const query = `UPDATE users SET email_address=$1, first_name=$2, last_name=$3, permission=$4, updated_at=$5 WHERE id=$6`
	_, err := r.db.ExecContext(ctx, query,
		user.EmailAddress, user.FirstName, user.LastName,
		pq.Array(user.Permission), user.UpdatedAt, user.ID,
	)
	return err
}

func (r *userPostgresRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)
	return err
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
