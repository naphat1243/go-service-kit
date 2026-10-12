// Package repository implements storage interfaces declared by the service
// layer, using pgx against PostgreSQL.
package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/naphat1243/go-service-kit/internal/pkg/apperror"
	"github.com/naphat1243/go-service-kit/internal/service"
)

const uniqueViolation = "23505"

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

var _ service.UserRepository = (*UserRepository)(nil)

func (r *UserRepository) Create(ctx context.Context, u service.User) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO users (id, email, name, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`,
		u.ID, u.Email, u.Name, u.CreatedAt, u.UpdatedAt)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
		return apperror.Conflict("email already registered")
	}
	return err
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (service.User, error) {
	var u service.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1`, id).
		Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return service.User{}, apperror.NotFound("user not found")
	}
	if err != nil {
		return service.User{}, apperror.Internal("querying user", err)
	}
	return u, nil
}

func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]service.User, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, email, name, created_at, updated_at FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []service.User{}
	for rows.Next() {
		var u service.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *UserRepository) Update(ctx context.Context, u service.User) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE users SET name = $2, updated_at = $3 WHERE id = $1`,
		u.ID, u.Name, u.UpdatedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperror.NotFound("user not found")
	}
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperror.NotFound("user not found")
	}
	return nil
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM users WHERE email = $1)`, email).Scan(&exists)
	return exists, err
}
