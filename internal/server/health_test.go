package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"git-bridge/internal/mirror"
	"time"

	"git-bridge/internal/consumer"
)

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	healthHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	var body map[string]string
	json.NewDecoder(resp.Body).Decode(&body)

	if body["status"] != "ok" {
		t.Errorf("status = %q, want ok", body["status"])
	}
	if body["service"] != "git-bridge" {
		t.Errorf("service = %q, want git-bridge", body["service"])
	}
}

func TestHealthHandler_ContentType(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	healthHandler(w, req)

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("content-type = %q, want application/json", ct)
	}
}

func TestNewMux_NilWebhook(t *testing.T) {
	mux := NewMux(nil)

	// /health should work
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("/health status = %d, want 200", w.Code)
	}

	// /ready should work
	req = httptest.NewRequest(http.MethodGet, "/ready", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("/ready status = %d, want 200", w.Code)
	}
}

func TestNewMux_APIDocsEndpoint(t *testing.T) {
	mux := NewMux(nil)

	req := httptest.NewRequest(http.MethodGet, "/api-docs", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("/api-docs status = %d, want 200", w.Code)
	}
	ct := w.Header().Get("Content-Type")
	if ct != "text/html; charset=utf-8" {
		t.Errorf("content-type = %q, want text/html; charset=utf-8", ct)
	}
	if w.Body.Len() == 0 {
		t.Error("body should not be empty")
	}
}

func TestNewMux_UnknownPathReturns404(t *testing.T) {
	mux := NewMux(nil)

	req := httptest.NewRequest(http.MethodGet, "/unknown/path", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("/unknown/path status = %d, want 404", w.Code)
	}
}

// mockMirrorer for webhook testing within server package
type mockMirrorer struct{}

func (m *mockMirrorer) SyncByTarget(_ context.Context, providerName, repoPath string, meta mirror.EventMeta) error {
	return nil
}

func TestNewMux_WithWebhook(t *testing.T) {
	wh := consumer.NewWebhook(context.Background(), &mockMirrorer{}, "", "")
	mux := NewMux(wh)

	// /health
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("/health status = %d, want 200", w.Code)
	}

	// /webhook/gitlab
	payload := `{"event_name":"push","ref":"refs/heads/main","project":{"path_with_namespace":"team/repo"},"repository":{"name":"repo"}}`
	req = httptest.NewRequest(http.MethodPost, "/webhook/gitlab", bytes.NewReader([]byte(payload)))
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("/webhook/gitlab status = %d, want 200", w.Code)
	}

	// /webhook/github
	payload = `{"ref":"refs/heads/main","repository":{"name":"repo","full_name":"org/repo"}}`
	req = httptest.NewRequest(http.MethodPost, "/webhook/github", bytes.NewReader([]byte(payload)))
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("/webhook/github status = %d, want 200", w.Code)
	}
}

func TestRunServer_GracefulShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		RunServer(ctx, 0, nil)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("server did not shut down in time")
	}
}

func TestRunServer_PortInUse(t *testing.T) {
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("failed to bind: %v", err)
	}
	defer ln.Close()

	port := ln.Addr().(*net.TCPAddr).Port

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		RunServer(ctx, port, nil)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("server should have exited due to port conflict")
	}
}
