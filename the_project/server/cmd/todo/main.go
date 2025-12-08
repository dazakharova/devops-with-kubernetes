package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	cacheDir  = "/usr/src/app/cache"
	fileName  = "image.jpg"
	remoteURL = "https://picsum.photos/1200"
	ttl       = 10 * time.Minute
)

func downloadAndCacheImage(url, localPath string) error {
    if err := os.MkdirAll(cacheDir, 0o755); err != nil {
        return fmt.Errorf("failed to prepare cache dir: %w", err)
    }

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	f, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		os.Remove(localPath)
		return err
	}

	return nil
}

func imageHandler(w http.ResponseWriter, r *http.Request) {
	localPath := filepath.Join(cacheDir, fileName)

	needNew := false

	info, err := os.Stat(localPath)
	if os.IsNotExist(err) {
		needNew = true
	} else if err != nil {
		http.Error(w, "failed to stat cache file", http.StatusInternalServerError)
		return
	} else {
		if time.Since(info.ModTime()) > ttl {
			_ = os.Remove(localPath)
			needNew = true
		}
	}

	if needNew {
		if err := downloadAndCacheImage(remoteURL, localPath); err != nil {
			http.Error(w, "failed to fetch image", http.StatusBadGateway)
			return
		}
	}

	w.Header().Set("Content-Type", "image/jpeg")
	http.ServeFile(w, r, localPath)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.Handle("GET /", http.FileServer(http.Dir("public")))
	mux.Handle("GET /image", http.HandlerFunc(imageHandler))

	fmt.Printf("Server started in port %s\n", port)

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
