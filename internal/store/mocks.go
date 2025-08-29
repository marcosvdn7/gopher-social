package store

import (
	"context"
	"database/sql"
	"time"
)

func NewMockStore() Storage {
	return Storage{
		User: &MockUserStore{},
	}
}

type MockUserStore struct {
}

func (u *MockUserStore) GetById(context.Context, int64) (User, error) {
	return User{}, nil
}

func (u *MockUserStore) GetByEmail(context.Context, string) (User, error) {
	return User{}, nil
}

func (u *MockUserStore) create(context.Context, *User, *sql.Tx) error {
	return nil
}

func (u *MockUserStore) CreateAndInvite(ctx context.Context, user *User, token string, tokenExp time.Duration) error {
	return nil
}

func (u *MockUserStore) Activate(context.Context, string) error {
	return nil
}

func (u *MockUserStore) Delete(context.Context, int64) error {
	return nil
}
