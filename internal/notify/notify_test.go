package notify

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"git-bridge/internal/config"
)

func TestNoop_Send(t *testing.T) {
	n := NewNoop()
	// Should not panic
	n.Send(Message{Level: "success", Title: "test", Body: "body"})
	n.Send(Message{Level: "error", Title: "err", Body: "failed"})
}

func TestSlack_Send_Success(t *testing.T) {
	var received map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	s := NewSlack(config.SlackConfig{
		WebhookURL: server.URL,
		Channel:    "#test",
	})

	s.Send(Message{
		Level: "success",
		Title: "Mirror Complete",
		Body:  "codecommit/repo → gitlab/repo",
	})

	text, ok := received["text"].(string)
	if !ok {
		t.Fatal("text field missing")
	}
	if len(text) == 0 {
		t.Error("text should not be empty")
	}

	ch, ok := received["channel"].(string)
	if !ok || ch != "#test" {
		t.Errorf("channel = %q, want #test", ch)
	}
}

func TestSlack_Send_Error(t *testing.T) {
	var received map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	s := NewSlack(config.SlackConfig{WebhookURL: server.URL})

	s.Send(Message{
		Level: "error",
		Title: "Mirror Failed",
		Body:  "clone failed",
	})

	text := received["text"].(string)
	if len(text) == 0 {
		t.Error("text should not be empty")
	}
}

func TestSlack_Send_NoChannel(t *testing.T) {
	var received map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	s := NewSlack(config.SlackConfig{WebhookURL: server.URL})
	s.Send(Message{Level: "success", Title: "test", Body: "ok"})

	if _, ok := received["channel"]; ok {
		t.Error("channel should not be set when empty")
	}
}

func TestSlack_Send_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	s := NewSlack(config.SlackConfig{WebhookURL: server.URL})
	// Should not panic on server error
	s.Send(Message{Level: "error", Title: "test", Body: "body"})
}

func TestSlack_Send_InvalidURL(t *testing.T) {
	s := NewSlack(config.SlackConfig{WebhookURL: "http://invalid.invalid.invalid:99999"})
	// Should not panic on connection error
	s.Send(Message{Level: "error", Title: "test", Body: "body"})
}

func TestSlack_Send_Warning(t *testing.T) {
	var received map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	s := NewSlack(config.SlackConfig{WebhookURL: server.URL, Channel: "#alerts"})
	s.Send(Message{Level: "warning", Title: "Slow Sync", Body: "took 5m"})

	text, ok := received["text"].(string)
	if !ok || len(text) == 0 {
		t.Fatal("text field missing or empty")
	}
}

func TestNewSlack(t *testing.T) {
	s := NewSlack(config.SlackConfig{
		WebhookURL: "https://hooks.slack.com/test",
		Channel:    "#ch",
	})
	if s.webhookURL != "https://hooks.slack.com/test" {
		t.Errorf("unexpected webhookURL: %q", s.webhookURL)
	}
	if s.channel != "#ch" {
		t.Errorf("unexpected channel: %q", s.channel)
	}
	if s.client == nil {
		t.Error("client should not be nil")
	}
}
