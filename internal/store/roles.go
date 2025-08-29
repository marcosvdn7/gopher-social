package store

import (
	"context"
	"database/sql"
)

type RoleStore struct {
	db *sql.DB
}

type Role struct {
	Id          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Level       string `json:"level"`
}

func (r *RoleStore) GetByName(ctx context.Context, name string) (Role, error) {
	query := `
		SELECT id, description, level FROM roles WHERE name = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var role Role

	if err := r.db.QueryRowContext(
		ctx,
		query,
		name,
	).Scan(
		&role.Id,
		&role.Description,
		&role.Level,
	); err != nil {
		switch err {
		case sql.ErrNoRows:
			return Role{}, ErrNotFound
		default:
			return Role{}, err
		}
	}

	return role, nil
}
