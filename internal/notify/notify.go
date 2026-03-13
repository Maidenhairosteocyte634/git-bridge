package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"git-bridge/internal/config"
)

const httpTimeout = 10 * time.Second

// Message represents a notification message.
type Message struct {
	Level string // success, error, warning
	Title string
	Body  string
}

// Notifier sends notifications.
type Notifier interface {
	Send(msg Message)
}

// Slack sends notifications to Slack via webhook.
type Slack struct {
	webhookURL string
	channel    string
	client     *http.Client
}

func NewSlack(cfg config.SlackConfig) *Slack {
	return &Slack{
		webhookURL: cfg.WebhookURL,
		channel:    cfg.Channel,
		client:     &http.Client{Timeout: httpTimeout},
	}
}

func (s *Slack) Send(msg Message) {
	prefix := "✅"
	if msg.Level == "error" {
		prefix = "❌"
	} else if msg.Level == "warning" {
		prefix = "⚠️"
	}

	payload := map[string]interface{}{
		"text": fmt.Sprintf("%s *%s*\n%s", prefix, msg.Title, msg.Body),
	}
	if s.channel != "" {
		payload["channel"] = s.channel
	}

	body, err := json.Marshal(payload)
	if err != nil {
		slog.Error("slack notification marshal failed", "error", err)
		return
	}
	resp, err := s.client.Post(s.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		slog.Error("slack notification failed", "error", err)
		return
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("slack notification failed", "status", resp.StatusCode)
	}
}

// Noop is a no-op notifier (when Slack is not configured).
type Noop struct{}

func NewNoop() *Noop       { return &Noop{} }
func (n *Noop) Send(_ Message) {}
