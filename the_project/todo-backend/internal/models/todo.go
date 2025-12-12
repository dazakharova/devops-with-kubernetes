package models

import (
	"context"
	"database/sql"
)

type TodoStore struct {
	db *sql.DB
}

func NewTodoStore(db *sql.DB) *TodoStore {
	return &TodoStore{db: db}
}

func (s *TodoStore) EnsureSchema(ctx context.Context) error {
	const q = `
	CREATE TABLE IF NOT EXISTS todos (
		id         BIGSERIAL PRIMARY KEY,
		title      TEXT NOT NULL
	);`
	_, err := s.db.ExecContext(ctx, q)
	return err
}

func (s *TodoStore) Create(ctx context.Context, title string) error {
	const q = `INSERT INTO todos (title) VALUES ($1);`
	_, err := s.db.ExecContext(ctx, q, title)
	return err
}

func (s *TodoStore) ListAllTitles(ctx context.Context) ([]string, error) {
	const q = `
		SELECT title
		FROM todos;
	`

	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	titles := make([]string, 0, 16)
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		titles = append(titles, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return titles, nil
}
