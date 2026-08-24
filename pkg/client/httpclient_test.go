package client

import (
	"compress/gzip"
	"compress/zlib"
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alexgorbatchev/goshazam/pkg/models"
)

func TestHTTPClientSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Shazam-Platform") != "IPHONE" {
			t.Errorf("missing platform header")
		}
		if r.Header.Get("User-Agent") != "CustomUA/1.0" {
			t.Errorf("expected custom UA, got %s", r.Header.Get("User-Agent"))
		}
		if r.Header.Get("Accept-Language") != "fr-FR" {
			t.Errorf("expected Accept-Language fr-FR, got %s", r.Header.Get("Accept-Language"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	c := NewHTTPClient(
		WithHTTPClient(server.Client()),
		WithLanguage("fr-FR"),
		WithUserAgent("CustomUA/1.0"),
		WithTimeout(10*time.Second),
	)
	var result map[string]string
	err := c.RequestJSON(context.Background(), "GET", server.URL, nil, nil, &result)
	if err != nil {
		t.Fatalf("RequestJSON failed: %v", err)
	}

	if result["status"] != "ok" {
		t.Errorf("expected status ok, got %s", result["status"])
	}
}

func TestHTTPClientWithProxy(t *testing.T) {
	c := NewHTTPClient(WithProxy("http://127.0.0.1:8080"))
	if c.client.Transport == nil {
		t.Errorf("expected custom transport for proxy")
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

func TestHTTPClientErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`not found error`))
	}))
	defer server.Close()

	c := NewHTTPClient(WithHTTPClient(server.Client()))
	ctx := context.Background()

	// Unsupported method
	_, err := c.Request(ctx, "DELETE", server.URL, nil, nil)
	if err == nil {
		t.Errorf("expected error for DELETE method")
	}

	// 404 error
	var res map[string]any
	err = c.RequestJSON(ctx, "GET", server.URL, nil, nil, &res)
	if err == nil {
		t.Errorf("expected error on 404 response")
	}

	apiErr := &models.APIError{StatusCode: 404, Status: "404 Not Found", Body: "resource not found"}
	if apiErr.Error() == "" {
		t.Errorf("expected non-empty APIError string")
	}
}

func TestHTTPClientGzipDeflate(t *testing.T) {
	// Gzip endpoint
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/gzip" {
			w.Header().Set("Content-Encoding", "gzip")
			gz := gzip.NewWriter(w)
			_, _ = gz.Write([]byte(`{"encoding": "gzip"}`))
			_ = gz.Close()
		} else if r.URL.Path == "/deflate" {
			w.Header().Set("Content-Encoding", "deflate")
			zl := zlib.NewWriter(w)
			_, _ = zl.Write([]byte(`{"encoding": "deflate"}`))
			_ = zl.Close()
		}
	}))
	defer server.Close()

	c := NewHTTPClient(WithHTTPClient(server.Client()))
	ctx := context.Background()

	var resGz map[string]string
	if err := c.RequestJSON(ctx, "GET", server.URL+"/gzip", nil, nil, &resGz); err != nil || resGz["encoding"] != "gzip" {
		t.Fatalf("gzip decompression failed: %v", err)
	}

	var resZl map[string]string
	if err := c.RequestJSON(ctx, "GET", server.URL+"/deflate", nil, nil, &resZl); err != nil || resZl["encoding"] != "deflate" {
		t.Fatalf("deflate decompression failed: %v", err)
	}
}

func TestHTTPClientPostAndHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("X-Custom") != "val" {
			t.Errorf("expected X-Custom: val, got %s", r.Header.Get("X-Custom"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"received": true}`))
	}))
	defer server.Close()

	c := NewHTTPClient(WithHTTPClient(server.Client()))
	headers := http.Header{"X-Custom": []string{"val"}}
	var res map[string]bool
	err := c.RequestJSON(context.Background(), "POST", server.URL, []byte(`{"ping": true}`), headers, &res)
	if err != nil || !res["received"] {
		t.Fatalf("POST RequestJSON failed: %v", err)
	}
}

func TestHTTPClientRetryExhaustion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`rate limited`))
	}))
	defer server.Close()

	cfg := RetryConfig{
		MaxAttempts:     2,
		InitialInterval: 5 * time.Millisecond,
		MaxInterval:     10 * time.Millisecond,
		RetryStatuses:   map[int]bool{http.StatusTooManyRequests: true},
	}
	c := NewHTTPClient(WithHTTPClient(server.Client()), WithRetryConfig(cfg))
	_, err := c.Request(context.Background(), "GET", server.URL, nil, nil)
	if err == nil {
		t.Fatalf("expected error on retry exhaustion")
	}
}

func TestHTTPClientContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg := RetryConfig{
		MaxAttempts:     5,
		InitialInterval: 500 * time.Millisecond,
		MaxInterval:     1 * time.Second,
		RetryStatuses:   map[int]bool{http.StatusInternalServerError: true},
	}
	c := NewHTTPClient(WithHTTPClient(server.Client()), WithRetryConfig(cfg))
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := c.Request(ctx, "GET", server.URL, nil, nil)
	if err == nil {
		t.Fatalf("expected context deadline error")
	}
}

func TestHTTPClientInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{not valid json`))
	}))
	defer server.Close()

	c := NewHTTPClient(WithHTTPClient(server.Client()))
	var res map[string]any
	err := c.RequestJSON(context.Background(), "GET", server.URL, nil, nil, &res)
	if err == nil {
		t.Fatalf("expected JSON decode error")
	}
}

func TestHTTPClientCorruptedEncodings(t *testing.T) {
	// Corrupted gzip
	serverGz := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		_, _ = w.Write([]byte("not gzip bytes"))
	}))
	defer serverGz.Close()

	c1 := NewHTTPClient(WithHTTPClient(serverGz.Client()), WithRetryConfig(RetryConfig{MaxAttempts: 1}))
	var resGz map[string]any
	if err := c1.RequestJSON(context.Background(), "GET", serverGz.URL, nil, nil, &resGz); err == nil {
		t.Errorf("expected error for corrupted gzip")
	}

	// Corrupted deflate
	serverZl := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "deflate")
		_, _ = w.Write([]byte("not deflate bytes"))
	}))
	defer serverZl.Close()

	c2 := NewHTTPClient(WithHTTPClient(serverZl.Client()), WithRetryConfig(RetryConfig{MaxAttempts: 1}))
	var resZl map[string]any
	if err := c2.RequestJSON(context.Background(), "GET", serverZl.URL, nil, nil, &resZl); err == nil {
		t.Errorf("expected error for corrupted deflate")
	}
}

func TestRandomUserAgent(t *testing.T) {
	ua := RandomUserAgent()
	if len(ua) == 0 {
		t.Errorf("expected non-empty user agent")
	}
}

func TestURLFormatters(t *testing.T) {
	if u := FormatAboutTrack("en-US", "GB", 123); len(u) == 0 {
		t.Errorf("expected non-empty about track URL")
	}
	if u := FormatTopTracksPlaylist("GB", "pl-1", 10, 0, "en-US"); len(u) == 0 {
		t.Errorf("expected non-empty playlist URL")
	}
	if u := FormatRelatedSongs("en-US", "GB", 123, 10, 0); len(u) == 0 {
		t.Errorf("expected non-empty related songs URL")
	}
	if u := FormatSearchArtist("en-US", "GB", "Artist Name", 10, 0); len(u) == 0 {
		t.Errorf("expected non-empty search artist URL")
	}
	if u := FormatSearchMusic("en-US", "GB", "Song Title", 10, 0); len(u) == 0 {
		t.Errorf("expected non-empty search music URL")
	}
	if u := FormatArtistAlbums("GB", 123, 10, 0); len(u) == 0 {
		t.Errorf("expected non-empty artist albums URL")
	}
	if u := FormatArtistAlbumInfo("GB", 123); len(u) == 0 {
		t.Errorf("expected non-empty album info URL")
	}
}
