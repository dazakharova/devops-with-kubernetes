package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)


func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.Handle("GET /", http.FileServer(http.Dir("public")))

	fmt.Printf("Server started in port %s\n", port)

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
