// Package service holds business logic. It depends on repository interfaces,
// never on concrete storage, so it is unit-testable without a database.
package service

import (
	"context"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/naphat1243/go-service-kit/internal/pkg/apperror"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserRepository is implemented by the postgres repository (and by fakes in tests).
type UserRepository interface {
	Create(ctx context.Context, u User) error
	GetByID(ctx context.Context, id uuid.UUID) (User, error)
	List(ctx context.Context, limit, offset int) ([]User, error)
	Update(ctx context.Context, u User) error
	Delete(ctx context.Context, id uuid.UUID) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

type UserService struct {
	repo UserRepository
	now  func() time.Time
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo, now: time.Now}
}

type CreateUserInput struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (s *UserService) Create(ctx context.Context, in CreateUserInput) (User, error) {
	email, name, err := normalize(in.Email, in.Name)
	if err != nil {
		return User{}, err
	}

	taken, err := s.repo.ExistsByEmail(ctx, email)
	if err != nil {
		return User{}, apperror.Internal("checking email uniqueness", err)
	}
	if taken {
		return User{}, apperror.Conflict("email already registered")
	}

	now := s.now().UTC()
	u := User{ID: uuid.New(), Email: email, Name: name, CreatedAt: now, UpdatedAt: now}
	if err := s.repo.Create(ctx, u); err != nil {
		return User{}, apperror.Internal("creating user", err)
	}
	return u, nil
}

func (s *UserService) Get(ctx context.Context, id uuid.UUID) (User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *UserService) List(ctx context.Context, limit, offset int) ([]User, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	users, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, apperror.Internal("listing users", err)
	}
	return users, nil
}

type UpdateUserInput struct {
	Name string `json:"name"`
}

func (s *UserService) Update(ctx context.Context, id uuid.UUID, in UpdateUserInput) (User, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" || len(name) > 120 {
		return User{}, apperror.Invalid("name must be 1-120 characters")
	}

	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return User{}, err
	}

	u.Name = name
	u.UpdatedAt = s.now().UTC()
	if err := s.repo.Update(ctx, u); err != nil {
		return User{}, apperror.Internal("updating user", err)
	}
	return u, nil
}

func (s *UserService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func normalize(email, name string) (string, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	name = strings.TrimSpace(name)

	if _, err := mail.ParseAddress(email); err != nil {
		return "", "", apperror.Invalid("invalid email address")
	}
	if name == "" || len(name) > 120 {
		return "", "", apperror.Invalid("name must be 1-120 characters")
	}
	return email, name, nil
}
