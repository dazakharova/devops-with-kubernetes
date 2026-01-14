package models

import (
	"context"
	"database/sql"
)

type TodoStore struct {
	db *sql.DB
}

type Todo struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

func NewTodoStore(db *sql.DB) *TodoStore {
	return &TodoStore{db: db}
}

func (s *TodoStore) EnsureSchema(ctx context.Context) error {
	const create = `
	CREATE TABLE IF NOT EXISTS todos (
		id    BIGSERIAL PRIMARY KEY,
		title TEXT NOT NULL
	);`

	const addDone = `
	ALTER TABLE todos
	ADD COLUMN IF NOT EXISTS done BOOLEAN NOT NULL DEFAULT false;`

	if _, err := s.db.ExecContext(ctx, create); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, addDone)
	return err
}

func (s *TodoStore) Create(ctx context.Context, title string) error {
	const q = `INSERT INTO todos (title, done) VALUES ($1, false);`
	_, err := s.db.ExecContext(ctx, q, title)
	return err
}

func (s *TodoStore) ListAllTodos(ctx context.Context) ([]Todo, error) {
	const q = `
		SELECT id, title, done
		FROM todos;
	`

	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	todos := make([]Todo, 0, 16)
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.Title, &t.Done); err != nil {
			return nil, err
		}
		todos = append(todos, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return todos, nil
}

func (s *TodoStore) MarkDone(ctx context.Context, id int64) error {
	const q = `UPDATE todos SET done = true WHERE id = $1;`

	res, err := s.db.ExecContext(ctx, q, id)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
