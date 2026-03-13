package server

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"git-bridge/internal/consumer"
)

const shutdownTimeout = 5 * time.Second

//go:embed api-docs.html
var apiDocsHTML []byte

// NewMux creates the HTTP handler with health checks and webhook endpoints.
func NewMux(webhook *consumer.Webhook) *http.ServeMux {
	mux := http.NewServeMux()

	// Health checks
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/ready", healthHandler)

	// API docs
	mux.HandleFunc("/api-docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(apiDocsHTML)
	})

	// Webhook endpoints
	if webhook != nil {
		mux.HandleFunc("/webhook/gitlab", webhook.GitLabHandler)
		mux.HandleFunc("/webhook/github", webhook.GitHubHandler)
		slog.Info("webhook endpoints registered", "gitlab", "/webhook/gitlab", "github", "/webhook/github")
	}

	return mux
}

// RunServer starts the HTTP server with health checks and webhook endpoints.
func RunServer(ctx context.Context, port int, webhook *consumer.Webhook) {
	mux := NewMux(webhook)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()

	slog.Info("server started", "port", port)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		slog.Error("server error", "error", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "git-bridge",
	})
}
