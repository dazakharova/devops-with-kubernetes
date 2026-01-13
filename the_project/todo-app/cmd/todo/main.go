package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var (
	cacheDir  = getEnv("CACHE_DIR")
	fileName  = getEnv("IMAGE_FILE_NAME")
	remoteURL = getEnv("IMAGE_URL")
	ttl       = getTTL("IMAGE_TTL_SECONDS")
)

func getEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("environment variable %s must be set", key)
	}
	return v
}

func getTTL(key string) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("environment variable %s must be set", key)
	}

	n, err := strconv.Atoi(v)
	if err != nil {
		log.Fatalf("invalid TTL value for %s: %v", key, err)
	}

	return time.Duration(n) * time.Second
}

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

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func readyzHandler(w http.ResponseWriter, r *http.Request) {
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		http.Error(w, "cache dir not writable", http.StatusServiceUnavailable)
		return
	}

	f, err := os.CreateTemp(cacheDir, ".readyz-*")
	if err != nil {
		http.Error(w, "cache dir not writable", http.StatusServiceUnavailable)
		return
	}
	name := f.Name()
	_ = f.Close()

	if err := os.Remove(name); err != nil && !errors.Is(err, os.ErrNotExist) {
		http.Error(w, "cache dir cleanup failed", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthzHandler)
	mux.HandleFunc("GET /readyz", readyzHandler)
	mux.Handle("GET /", http.FileServer(http.Dir("public")))
	mux.Handle("GET /image", http.HandlerFunc(imageHandler))

	fmt.Printf("Server started in port %s\n", port)

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
