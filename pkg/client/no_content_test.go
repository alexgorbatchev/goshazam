package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/alexgorbatchev/goshazam/pkg/models"
)

func TestHTTPNoContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	c := NewHTTPClient(WithHTTPClient(server.Client()))
	data, err := c.Request(context.Background(), http.MethodGet, server.URL, nil, nil)
	if err != nil || len(data) != 0 {
		t.Fatalf("raw 204: data=%q error=%v", data, err)
	}
	target := map[string]string{"existing": "preserved"}
	err = c.RequestJSON(context.Background(), http.MethodGet, server.URL, nil, nil, &target)
	if !errors.Is(err, models.ErrNoContent) || target["existing"] != "preserved" {
		t.Fatalf("JSON 204: target=%#v error=%v", target, err)
	}
}

func TestHTTPClientFailureRecovery(t *testing.T) {
	for _, kind := range []string{"gzip", "deflate", "truncated"} {
		for _, attempts := range []int{1, 2} {
			t.Run(kind+strconv.Itoa(attempts), func(t *testing.T) {
				var calls atomic.Int32
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if calls.Add(1) == 1 {
						if kind == "truncated" {
							w.Header().Set("Content-Length", "100")
						} else {
							w.Header().Set("Content-Encoding", kind)
						}
						_, _ = w.Write([]byte("broken"))
						return
					}
					_, _ = w.Write([]byte(`{"recovered":true}`))
				}))
				defer server.Close()
				c := NewHTTPClient(WithHTTPClient(server.Client()), WithRetryConfig(RetryConfig{MaxAttempts: attempts}))
				var got map[string]bool
				err := c.RequestJSON(context.Background(), http.MethodGet, server.URL, nil, nil, &got)
				if calls.Load() != int32(attempts) {
					t.Fatalf("attempts: %d", calls.Load())
				}
				if attempts == 1 && err == nil {
					t.Fatal("expected damaged response error")
				}
				if attempts == 2 && (err != nil || !got["recovered"]) {
					t.Fatalf("retry lost: %#v %v", got, err)
				}
			})
		}
	}
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	server.Close()
	c := NewHTTPClient(WithRetryConfig(RetryConfig{MaxAttempts: 2}))
	if _, err := c.Request(context.Background(), http.MethodGet, server.URL, nil, nil); err == nil || !strings.Contains(err.Error(), "request failed") {
		t.Fatalf("expected transport failure: %v", err)
	}
	if _, err := c.Request(context.Background(), http.MethodGet, "://invalid", nil, nil); err == nil || !strings.Contains(err.Error(), "creating request") {
		t.Fatalf("expected invalid URL failure: %v", err)
	}
}
