package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

var (
	DuplicatedKeyErrorMessage = `pq: duplicate key value violates unique constraint "followers_pkey`
	ErrDuplicatedKey          = errors.New("resource already exists")
)

type Follower struct {
	UserID     int64  `json:"user_id"`
	FollowerID int64  `json:"follower_id"`
	CreatedAt  string `json:"created_at"`
}

type FollowerStore struct {
	db *sql.DB
}

func (s *FollowerStore) Follow(ctx context.Context, followerId, userId int64) error {
	query := `
		INSERT INTO followers(user_id, follower_id)
		VALUES ($1, $2);
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, userId, followerId)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrNotFound
		case strings.Contains(err.Error(), DuplicatedKeyErrorMessage):
			return ErrDuplicatedKey
		default:
			return err
		}
	}

	return nil
}

func (s *FollowerStore) Unfollow(ctx context.Context, followerId, userId int64) error {
	query := `
		DELETE FROM followers
		WHERE user_id = $1 AND follower_id = $2;
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	result, err := s.db.ExecContext(ctx, query, userId, followerId)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return ErrNotFound
		default:
			return err
		}
	}

	rows, resultErr := result.RowsAffected()
	if rows != 1 || resultErr != nil {
		if rows == 0 || errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	return nil
}
