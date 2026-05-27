package repository

import (
	"context"
	"errors"
	"fmt"

	"backend/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, u *domain.User) (*domain.User, error) {
	query := `
		INSERT INTO users (full_name, email, password, phone, gender, dob, role, avatar, is_lock, is_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(ctx, query,
		u.FullName,
		u.Email,
		u.Password,
		u.Phone,
		u.Gender,
		u.DOB,
		u.Role,
		u.Avatar,
		u.IsLock,
		u.IsVerified,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" { // unique_violation
				return nil, domain.ErrEmailTaken
			}
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return u, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int) (*domain.User, error) {
	u := &domain.User{}
	query := `
		SELECT id, full_name, email, password, phone, gender, dob, role, avatar, is_lock, is_verified, created_at, updated_at
		FROM users
		WHERE id = $1`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.FullName, &u.Email, &u.Password, &u.Phone,
		&u.Gender, &u.DOB, &u.Role, &u.Avatar, &u.IsLock, &u.IsVerified,
		&u.CreatedAt, &u.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return u, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	u := &domain.User{}
	query := `
		SELECT id, full_name, email, password, phone, gender, dob, role, avatar, is_lock, is_verified, created_at, updated_at
		FROM users
		WHERE email = $1`

	err := r.db.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.FullName, &u.Email, &u.Password, &u.Phone,
		&u.Gender, &u.DOB, &u.Role, &u.Avatar, &u.IsLock, &u.IsVerified,
		&u.CreatedAt, &u.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return u, nil
}

func (r *UserRepository) UpdateLockStatus(ctx context.Context, id int, isLock bool) error {
	query := `UPDATE users SET is_lock = $1, updated_at = NOW() WHERE id = $2`
	tag, err := r.db.Exec(ctx, query, isLock, id)
	if err != nil {
		return fmt.Errorf("failed to update user lock status: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}
