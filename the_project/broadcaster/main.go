package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/nats-io/nats.go"
)

const (
	defaultSubject = "todos.events"
	defaultQueue   = "broadcaster"
	natsTimeout    = 3 * time.Second
)

type TodoEvent struct {
	Event     string `json:"event"`
	Title     string `json:"title,omitempty"`
	TodoID    int64  `json:"todoId,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
	Service   string `json:"service,omitempty"`
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func readyzHandler(nc *nats.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if nc == nil || nc.Status() != nats.CONNECTED {
			http.Error(w, "nats not ready", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}
}

func formatMessage(ev TodoEvent, raw string) string {
	switch ev.Event {
	case "todo_created":
		if ev.Title != "" {
			return "Todo created: " + ev.Title
		}
		return "Todo created"
	case "todo_done":
		if ev.TodoID != 0 {
			return "Todo marked done (id=" + strconv.FormatInt(ev.TodoID, 10) + ")"
		}
		return "Todo marked done"
	default:
		return "Todo event: " + raw
	}
}

func sendTelegram(client *http.Client, token, chatID, text string) error {
	url := "https://api.telegram.org/bot" + token + "/sendMessage"

	body, err := json.Marshal(map[string]any{
		"chat_id": chatID,
		"text":    text,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return errors.New("telegram returned status " + res.Status)
	}
	return nil
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		log.Fatal("NATS_URL must be set")
	}

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN must be set")
	}

	chatID := os.Getenv("TELEGRAM_CHAT_ID")
	if chatID == "" {
		log.Fatal("TELEGRAM_CHAT_ID must be set")
	}

	subject := os.Getenv("NATS_SUBJECT")
	if subject == "" {
		subject = defaultSubject
	}
	queue := os.Getenv("NATS_QUEUE")
	if queue == "" {
		queue = defaultQueue
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	nc, err := nats.Connect(
		natsURL,
		nats.Name("todo-broadcaster"),
		nats.Timeout(natsTimeout),
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(1*time.Second),
	)
	if err != nil {
		logger.Error("failed to connect to nats", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer nc.Close()

	client := &http.Client{Timeout: 10 * time.Second}

	// Queue subscription => no duplicates when scaled
	_, err = nc.QueueSubscribe(subject, queue, func(msg *nats.Msg) {
		raw := string(msg.Data)

		var ev TodoEvent
		if err := json.Unmarshal(msg.Data, &ev); err != nil {
			logger.Warn("invalid event json", slog.String("error", err.Error()))
			return
		}

		text := formatMessage(ev, raw)

		if err := sendTelegram(client, botToken, chatID, text); err != nil {
			logger.Error("failed to send telegram message", slog.String("error", err.Error()))
			return
		}

		logger.Info("sent telegram message", slog.String("event", ev.Event))
	})
	if err != nil {
		logger.Error("failed to subscribe", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Ensure subscription is registered
	if err := nc.FlushTimeout(natsTimeout); err != nil {
		logger.Error("nats flush failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("broadcaster started",
		slog.String("subject", subject),
		slog.String("queue", queue),
		slog.String("port", port),
	)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthzHandler)
	mux.HandleFunc("GET /readyz", readyzHandler(nc))

	// blocks forever like your todo-backend
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		logger.Error("server failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
