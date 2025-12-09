package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

var todos = []string{
	"Learn JavaScript",
	"Learn React",
	"Build a project",
}

func getTodos(w http.ResponseWriter, r *http.Request) {
	jsonTodos, err := json.Marshal(&todos)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonTodos)
}

func createTodo(w http.ResponseWriter, r *http.Request) {
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

	todos = append(todos, todo)

	jsonTodos, err := json.Marshal(todos)
	if err != nil {
		http.Error(w, "failed to encode todos", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonTodos)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "4001"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /todos", getTodos)
	mux.HandleFunc("POST /todos", createTodo)

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
