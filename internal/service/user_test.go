package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/naphat1243/go-service-kit/internal/pkg/apperror"
)

// fakeUserRepo is an in-memory UserRepository. No mocks framework needed —
// hand-written fakes keep tests honest and dependency-free.
type fakeUserRepo struct {
	users    map[uuid.UUID]User
	failWith error
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{users: map[uuid.UUID]User{}}
}

func (f *fakeUserRepo) Create(_ context.Context, u User) error {
	if f.failWith != nil {
		return f.failWith
	}
	f.users[u.ID] = u
	return nil
}

func (f *fakeUserRepo) GetByID(_ context.Context, id uuid.UUID) (User, error) {
	u, ok := f.users[id]
	if !ok {
		return User{}, apperror.NotFound("user not found")
	}
	return u, nil
}

func (f *fakeUserRepo) List(_ context.Context, limit, offset int) ([]User, error) {
	out := make([]User, 0, len(f.users))
	for _, u := range f.users {
		out = append(out, u)
	}
	if offset >= len(out) {
		return []User{}, nil
	}
	end := min(offset+limit, len(out))
	return out[offset:end], nil
}

func (f *fakeUserRepo) Update(_ context.Context, u User) error {
	f.users[u.ID] = u
	return nil
}

func (f *fakeUserRepo) Delete(_ context.Context, id uuid.UUID) error {
	if _, ok := f.users[id]; !ok {
		return apperror.NotFound("user not found")
	}
	delete(f.users, id)
	return nil
}

func (f *fakeUserRepo) ExistsByEmail(_ context.Context, email string) (bool, error) {
	for _, u := range f.users {
		if u.Email == email {
			return true, nil
		}
	}
	return false, nil
}

func newTestService(repo UserRepository) *UserService {
	s := NewUserService(repo)
	s.now = func() time.Time { return time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC) }
	return s
}

func kindOf(t *testing.T, err error) apperror.Kind {
	t.Helper()
	var appErr *apperror.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("expected *apperror.Error, got %T: %v", err, err)
	}
	return appErr.Kind
}

func TestCreate_Succeeds(t *testing.T) {
	svc := newTestService(newFakeUserRepo())

	u, err := svc.Create(context.Background(), CreateUserInput{Email: " Jo@Example.com ", Name: " Jo "})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.Email != "jo@example.com" {
		t.Errorf("email not normalized: %q", u.Email)
	}
	if u.Name != "Jo" {
		t.Errorf("name not trimmed: %q", u.Name)
	}
	if u.ID == uuid.Nil {
		t.Error("expected generated ID")
	}
	if u.CreatedAt.IsZero() || !u.CreatedAt.Equal(u.UpdatedAt) {
		t.Error("expected CreatedAt == UpdatedAt on create")
	}
}

func TestCreate_RejectsInvalidInput(t *testing.T) {
	svc := newTestService(newFakeUserRepo())

	cases := []CreateUserInput{
		{Email: "not-an-email", Name: "Jo"},
		{Email: "jo@example.com", Name: ""},
		{Email: "jo@example.com", Name: string(make([]byte, 121))},
	}
	for _, in := range cases {
		if _, err := svc.Create(context.Background(), in); kindOf(t, err) != apperror.KindInvalid {
			t.Errorf("input %+v: expected invalid_argument, got %v", in, err)
		}
	}
}

func TestCreate_RejectsDuplicateEmail(t *testing.T) {
	svc := newTestService(newFakeUserRepo())

	if _, err := svc.Create(context.Background(), CreateUserInput{Email: "jo@example.com", Name: "Jo"}); err != nil {
		t.Fatalf("seed create failed: %v", err)
	}
	_, err := svc.Create(context.Background(), CreateUserInput{Email: "JO@example.com", Name: "Other"})
	if kindOf(t, err) != apperror.KindConflict {
		t.Errorf("expected conflict, got %v", err)
	}
}

func TestCreate_WrapsRepoFailure(t *testing.T) {
	repo := newFakeUserRepo()
	repo.failWith = errors.New("boom")
	svc := newTestService(repo)

	_, err := svc.Create(context.Background(), CreateUserInput{Email: "jo@example.com", Name: "Jo"})
	if kindOf(t, err) != apperror.KindInternal {
		t.Errorf("expected internal, got %v", err)
	}
}

func TestGet_NotFound(t *testing.T) {
	svc := newTestService(newFakeUserRepo())

	_, err := svc.Get(context.Background(), uuid.New())
	if kindOf(t, err) != apperror.KindNotFound {
		t.Errorf("expected not_found, got %v", err)
	}
}

func TestUpdate_ChangesNameAndTimestamp(t *testing.T) {
	repo := newFakeUserRepo()
	svc := newTestService(repo)

	created, err := svc.Create(context.Background(), CreateUserInput{Email: "jo@example.com", Name: "Jo"})
	if err != nil {
		t.Fatalf("seed create failed: %v", err)
	}

	svc.now = func() time.Time { return time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC) }
	updated, err := svc.Update(context.Background(), created.ID, UpdateUserInput{Name: "Joanna"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != "Joanna" {
		t.Errorf("name not updated: %q", updated.Name)
	}
	if !updated.UpdatedAt.After(created.UpdatedAt) {
		t.Error("UpdatedAt should move forward")
	}
}

func TestUpdate_InvalidName(t *testing.T) {
	svc := newTestService(newFakeUserRepo())

	_, err := svc.Update(context.Background(), uuid.New(), UpdateUserInput{Name: "  "})
	if kindOf(t, err) != apperror.KindInvalid {
		t.Errorf("expected invalid_argument, got %v", err)
	}
}

func TestDelete_ThenGetNotFound(t *testing.T) {
	svc := newTestService(newFakeUserRepo())

	u, err := svc.Create(context.Background(), CreateUserInput{Email: "jo@example.com", Name: "Jo"})
	if err != nil {
		t.Fatalf("seed create failed: %v", err)
	}
	if err := svc.Delete(context.Background(), u.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if _, err := svc.Get(context.Background(), u.ID); kindOf(t, err) != apperror.KindNotFound {
		t.Errorf("expected not_found after delete, got %v", err)
	}
}

func TestList_ClampsPagination(t *testing.T) {
	repo := newFakeUserRepo()
	svc := newTestService(repo)

	for i := 0; i < 3; i++ {
		if _, err := svc.Create(context.Background(), CreateUserInput{
			Email: string(rune('a'+i)) + "@example.com",
			Name:  "U",
		}); err != nil {
			t.Fatalf("seed create failed: %v", err)
		}
	}

	users, err := svc.List(context.Background(), -5, -1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 3 {
		t.Errorf("expected 3 users with clamped pagination, got %d", len(users))
	}
}
