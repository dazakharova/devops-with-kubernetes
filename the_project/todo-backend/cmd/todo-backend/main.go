package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dazakharova/the_project/todo-backend/internal/models"
	_ "github.com/lib/pq"
)

const maxTodoLen = 140

func getTodos(store *models.TodoStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		todos, err := store.ListAllTitles(ctx)
		if err != nil {
			http.Error(w, "failed to load todos", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(todos)
	}
}

func createTodo(store *models.TodoStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusBadRequest)
			return
		}

		var data struct {
			Todo string `json:"todo"`
		}

		if err := json.Unmarshal(bodyBytes, &data); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		todo := strings.TrimSpace(data.Todo)
		if todo == "" {
			http.Error(w, "todo cannot be empty", http.StatusBadRequest)
			return
		}

		if len(todo) > maxTodoLen {
			http.Error(w, "todo too long (max 140 characters)", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		if err := store.Create(ctx, todo); err != nil {
			http.Error(w, "failed to create todo", http.StatusInternalServerError)
			return
		}

		// Keep your old behavior: return all todos after insert
		todos, err := store.ListAllTitles(ctx)
		if err != nil {
			http.Error(w, "failed to load todos", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(todos)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT must be set")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	// Verify DB connection
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 3*time.Second)
	if err := db.PingContext(pingCtx); err != nil {
		log.Fatalf("db ping failed: %v", err)
	}
	defer pingCancel()

	store := models.NewTodoStore(db)

	schemaCtx, schemaCancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := store.EnsureSchema(schemaCtx); err != nil {
		log.Fatalf("failed to ensure schema: %v", err)
	}
	defer schemaCancel()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /todos", getTodos(store))
	mux.HandleFunc("POST /todos", createTodo(store))

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
