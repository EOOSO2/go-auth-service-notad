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

const userColumns = `id, email_id, email_address, first_name, last_name, password, permission, created_at, updated_at`

func scanUser(row interface{ Scan(...any) error }, u *entity.User) error {
	return row.Scan(
		&u.ID, &u.UserID, &u.EmailAddress, &u.FirstName, &u.LastName,
		&u.Password, pq.Array(&u.Permission), &u.CreatedAt, &u.UpdatedAt,
	)
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

func (r *userPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	const query = `SELECT ` + userColumns + ` FROM users WHERE id = $1`
	user := &entity.User{}
	err := scanUser(r.db.QueryRowContext(ctx, query, id), user)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *userPostgresRepository) GetByEmailID(ctx context.Context, userID string) (*entity.User, error) {
	const query = `SELECT ` + userColumns + ` FROM users WHERE email_id = $1`
	user := &entity.User{}
	err := scanUser(r.db.QueryRowContext(ctx, query, userID), user)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *userPostgresRepository) GetByEmailAddress(ctx context.Context, email string) (*entity.User, error) {
	const query = `SELECT ` + userColumns + ` FROM users WHERE email_address = $1`
	user := &entity.User{}
	err := scanUser(r.db.QueryRowContext(ctx, query, email), user)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *userPostgresRepository) List(ctx context.Context, offset, limit int) ([]entity.User, int, error) {
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&total); err != nil {
		return nil, 0, err
	}

	const query = `SELECT ` + userColumns + ` FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	users := make([]entity.User, 0, limit)
	for rows.Next() {
		var u entity.User
		if err := scanUser(rows, &u); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *userPostgresRepository) Update(ctx context.Context, user *entity.User) error {
	user.UpdatedAt = time.Now()
	const query = `UPDATE users SET email_address=$1, first_name=$2, last_name=$3, permission=$4, updated_at=$5 WHERE id=$6`
	res, err := r.db.ExecContext(ctx, query,
		user.EmailAddress, user.FirstName, user.LastName,
		pq.Array(user.Permission), user.UpdatedAt, user.ID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return repository.ErrUserNotFound
	}
	return nil
}

func (r *userPostgresRepository) UpdatePassword(ctx context.Context, id uuid.UUID, hashed string) error {
	const query = `UPDATE users SET password=$1, updated_at=$2 WHERE id=$3`
	res, err := r.db.ExecContext(ctx, query, hashed, time.Now(), id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return repository.ErrUserNotFound
	}
	return nil
}

func (r *userPostgresRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)
	return err
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
