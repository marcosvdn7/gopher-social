package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
)

type Post struct {
	ID        int64    `json:"id"`
	Content   string   `json:"content"`
	Title     string   `json:"title"`
	UserID    int64    `json:"user_id"`
	Tags      []string `json:"tags,omitempty"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
	// TODO implementar lock no database
	Version  int       `json:"version"`
	Comments []Comment `json:"comments"`
	Username string    `json:"username"`
}

type PostWithMetadata struct {
	Post
	CommentCount int64 `json:"comment_count" `
}

type PostStore struct {
	db *sql.DB
}

func (s *PostStore) Create(ctx context.Context, post *Post) error {
	query := `
		INSERT INTO posts (content, title, user_id, tags, version)
		VALUES ($1, $2, $3, $4, 0) RETURNING id, created_at, updated_at
		`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	if err := s.db.QueryRowContext(
		ctx,
		query,
		post.Content,
		post.Title,
		post.UserID,
		pq.Array(post.Tags),
	).Scan(
		&post.ID,
		&post.CreatedAt,
		&post.UpdatedAt,
	); err != nil {
		return err
	}

	return nil
}

func (s *PostStore) GetById(ctx context.Context, postID int64) (Post, error) {
	query := fmt.Sprintf("SELECT content, title, user_id, tags, created_at, updated_at, version FROM posts WHERE id = '%d'", postID)
	var p Post

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	if err := s.db.QueryRowContext(
		ctx,
		query,
	).Scan(
		&p.Content,
		&p.Title,
		&p.UserID,
		pq.Array(&p.Tags),
		&p.CreatedAt,
		&p.UpdatedAt,
		&p.Version,
	); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return Post{}, ErrNotFound
		default:
			return Post{}, err
		}
	}
	p.ID = postID

	return p, nil
}

func (s *PostStore) Delete(ctx context.Context, id int64) (int64, error) {
	query := "DELETE FROM posts WHERE id = $1"

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (s *PostStore) Update(ctx context.Context, newPost *Post) error {
	query := `
		UPDATE posts 
		SET title = $2,
		content = $3,
		tags = $4,
		version = version + 1
		WHERE id = $1 AND version = $5
		RETURNING title, content, tags, version;
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	if err := s.db.QueryRowContext(
		ctx,
		query,
		newPost.ID,
		newPost.Title,
		newPost.Content,
		pq.Array(newPost.Tags),
		newPost.Version,
	).Scan(
		&newPost.Title,
		&newPost.Content,
		pq.Array(&newPost.Tags),
		&newPost.Version,
	); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrNotFound
		default:
			return err
		}
	}

	return nil
}

func (s *PostStore) GetUserFeed(ctx context.Context, userId int64, fq PaginatedFeedQuery) ([]PostWithMetadata, error) {
	query := `
		select p.id, p.user_id, p.title, p.content, p.created_at, p.tags, 
		COUNT(c.id) as comments_count, u.username
		from posts p
		left join comments c on c.post_id = p.id
		left join users u on u.id = p.user_id
		join followers f on f.follower_id = p.user_id or p.user_id = $1
		where 
			f.follower_id = $1 or p.user_id = $1 AND
			(p.title ILIKE '%' || $4 || '%' OR p.content ILIKE '%' || $4 || '%') AND
			(p.tags @> $5 OR $5 = '{}')
		group by p.id, u.id
		order by p.created_at ` + fq.Sort + `
		limit $2 offset $3;
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(
		ctx,
		query,
		userId,
		fq.Limit,
		fq.Offset,
		fq.Search,
		pq.Array(fq.Tags),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var (
		feed = make([]PostWithMetadata, 0)
		p    = PostWithMetadata{}
	)

	for rows.Next() {
		if err := rows.Scan(
			&p.ID,
			&p.UserID,
			&p.Title,
			&p.Content,
			&p.CreatedAt,
			pq.Array(&p.Tags),
			&p.CommentCount,
			&p.Username,
		); err != nil {
			return nil, err
		}

		feed = append(feed, p)
	}

	return feed, nil
}
