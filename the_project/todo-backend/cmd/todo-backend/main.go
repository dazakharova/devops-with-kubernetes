package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dazakharova/the_project/todo-backend/internal/models"
	_ "github.com/lib/pq"
)

const maxTodoLen = 140

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func requestLogger(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(sw, r)

		logger.Info("http request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", sw.status),
			slog.Int64("duration_ms", time.Since(start).Milliseconds()),
		)
	})
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

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

func createTodo(store *models.TodoStore, logger *slog.Logger) http.HandlerFunc {
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
			logger.Warn("rejected todo",
				slog.String("reason", "invalid_json"),
				slog.String("error", err.Error()),
			)

			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		todo := strings.TrimSpace(data.Todo)
		if todo == "" {
			logger.Warn("rejected todo",
				slog.String("reason", "empty"),
			)

			http.Error(w, "todo cannot be empty", http.StatusBadRequest)
			return
		}

		if len(todo) > maxTodoLen {
			preview := todo[:maxTodoLen] + "…"

			logger.Warn("rejected todo",
				slog.String("reason", "too_long"),
				slog.Int("length", len(todo)),
				slog.String("preview", preview),
			)

			http.Error(w, "todo too long (max 140 characters)", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		if err := store.Create(ctx, todo); err != nil {
			logger.Error("failed to create todo",
				slog.String("error", err.Error()),
			)

			http.Error(w, "failed to create todo", http.StatusInternalServerError)
			return
		}

		logger.Info("created todo",
			slog.Int("length", len(todo)),
			slog.String("todo", todo),
		)

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

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

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
		logger.Error("db ping failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pingCancel()

	store := models.NewTodoStore(db)

	schemaCtx, schemaCancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := store.EnsureSchema(schemaCtx); err != nil {
		logger.Error("failed to ensure schema", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer schemaCancel()

	logger.Info("starting server", slog.String("port", port))

	mux := http.NewServeMux()
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("GET /todos", getTodos(store))
	mux.HandleFunc("POST /todos", createTodo(store, logger))

	handler := requestLogger(logger, mux)

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		logger.Error("server failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
