package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestHTTPClientSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Shazam-Platform") != "IPHONE" {
			t.Errorf("missing platform header")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	c := NewHTTPClient(WithHTTPClient(server.Client()))
	var result map[string]string
	err := c.RequestJSON(context.Background(), "GET", server.URL, nil, nil, &result)
	if err != nil {
		t.Fatalf("RequestJSON failed: %v", err)
	}

	if result["status"] != "ok" {
		t.Errorf("expected status ok, got %s", result["status"])
	}
}

func TestHTTPClientRetryOn500(t *testing.T) {
	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if attempts.Add(1) < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`internal error`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"recovered": true}`))
	}))
	defer server.Close()

	cfg := RetryConfig{
		MaxAttempts:     3,
		InitialInterval: 10 * time.Millisecond,
		MaxInterval:     50 * time.Millisecond,
		RetryStatuses: map[int]bool{
			http.StatusInternalServerError: true,
		},
	}

	c := NewHTTPClient(WithHTTPClient(server.Client()), WithRetryConfig(cfg))
	var result map[string]bool
	err := c.RequestJSON(context.Background(), "GET", server.URL, nil, nil, &result)
	if err != nil {
		t.Fatalf("RequestJSON should recover after retry: %v", err)
	}

	if !result["recovered"] {
		t.Errorf("expected recovered true")
	}

	if attempts.Load() != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts.Load())
	}
}
