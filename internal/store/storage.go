package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNotFound          = errors.New("resource not found")
	QueryTimeoutDuration = time.Second * 5
)

type Storage struct {
	Post interface {
		Create(context.Context, *Post) error
		GetById(context.Context, int64) (Post, error)
		Delete(ctx context.Context, id int64) (int64, error)
		Update(context.Context, *Post) error
		GetUserFeed(context.Context, int64, PaginatedFeedQuery) ([]PostWithMetadata, error)
	}
	User interface {
		GetById(context.Context, int64) (User, error)
		GetByEmail(context.Context, string) (User, error)
		create(context.Context, *User, *sql.Tx) error
		CreateAndInvite(ctx context.Context, user *User, token string, tokenExp time.Duration) error
		Activate(context.Context, string) error
		Delete(context.Context, int64) error
	}
	Comment interface {
		Create(context.Context, *Comment) error
		GetByPostId(context.Context, int64) ([]Comment, error)
		DeleteByPostId(context.Context, int64) (int64, error)
	}
	Follower interface {
		Follow(ctx context.Context, followerId, UserId int64) error
		Unfollow(ctx context.Context, followerId, UserId int64) error
	}
	Role interface {
		GetByName(context.Context, string) (Role, error)
	}
}

func NewPostgresStorage(db *sql.DB) *Storage {
	return &Storage{
		Post:     &PostStore{db: db},
		User:     &UserStore{db: db},
		Comment:  &CommentStore{db: db},
		Follower: &FollowerStore{db: db},
		Role:     &RoleStore{db: db},
	}
}

func withTx(db *sql.DB, ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
