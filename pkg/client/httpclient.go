package client

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/alexgorbatchev/goshazam/pkg/models"
)

// RetryConfig configures retry behavior for HTTP requests.
type RetryConfig struct {
	MaxAttempts     int
	InitialInterval time.Duration
	MaxInterval     time.Duration
	RetryStatuses   map[int]bool
}

// DefaultRetryConfig provides sensible defaults for retrying requests.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:     5,
		InitialInterval: 500 * time.Millisecond,
		MaxInterval:     10 * time.Second,
		RetryStatuses: map[int]bool{
			http.StatusTooManyRequests:     true,
			http.StatusInternalServerError: true,
			http.StatusBadGateway:          true,
			http.StatusServiceUnavailable:  true,
			http.StatusGatewayTimeout:      true,
		},
	}
}

// HTTPClient wraps an http.Client with Shazam headers, retries, and JSON decoding.
type HTTPClient struct {
	client      *http.Client
	language    string
	retryConfig RetryConfig
	customUA    string
}

// ClientOption configures an HTTPClient.
type ClientOption func(*HTTPClient)

// WithLanguage sets the Accept-Language header (default: "en-US").
func WithLanguage(lang string) ClientOption {
	return func(c *HTTPClient) {
		c.language = lang
	}
}

// WithRetryConfig customizes retry options.
func WithRetryConfig(cfg RetryConfig) ClientOption {
	return func(c *HTTPClient) {
		c.retryConfig = cfg
	}
}

// WithUserAgent overrides randomized User-Agent with a fixed string.
func WithUserAgent(ua string) ClientOption {
	return func(c *HTTPClient) {
		c.customUA = ua
	}
}

// WithProxy configures a proxy URL (HTTP/HTTPS/SOCKS5).
func WithProxy(proxyURL string) ClientOption {
	return func(c *HTTPClient) {
		if proxyURL == "" {
			return
		}
		u, err := url.Parse(proxyURL)
		if err != nil {
			return
		}
		transport := &http.Transport{
			Proxy:           http.ProxyURL(u),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: false}, //nolint:gosec
		}
		c.client.Transport = transport
	}
}

// WithHTTPClient replaces the underlying standard library http.Client.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *HTTPClient) {
		if httpClient != nil {
			c.client = httpClient
		}
	}
}

// NewHTTPClient creates a new HTTPClient with options.
func NewHTTPClient(opts ...ClientOption) *HTTPClient {
	c := &HTTPClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		language:    "en-US",
		retryConfig: DefaultRetryConfig(),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// DefaultHeaders builds standard headers expected by Shazam endpoints.
func (c *HTTPClient) DefaultHeaders() http.Header {
	h := make(http.Header)
	h.Set("X-Shazam-Platform", "IPHONE")
	h.Set("X-Shazam-AppVersion", "14.1.0")
	h.Set("Accept", "*/*")
	h.Set("Accept-Language", c.language)
	h.Set("Accept-Encoding", "gzip, deflate")

	if c.customUA != "" {
		h.Set("User-Agent", c.customUA)
	} else {
		h.Set("User-Agent", RandomUserAgent())
	}

	return h
}

// Request executes an HTTP request with automatic retries and returns the raw response body.
func (c *HTTPClient) Request(ctx context.Context, method, targetURL string, body []byte, headers http.Header) ([]byte, error) {
	method = strings.ToUpper(method)
	if method != http.MethodGet && method != http.MethodPost {
		return nil, models.ErrBadMethod
	}

	var lastErr error
	maxAttempts := c.retryConfig.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 1
	}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		var bodyReader io.Reader
		if len(body) > 0 {
			bodyReader = bytes.NewReader(body)
		}

		req, err := http.NewRequestWithContext(ctx, method, targetURL, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}

		// Apply default headers
		for k, v := range c.DefaultHeaders() {
			req.Header[k] = v
		}
		// If retrying, pick a fresh user agent
		if attempt > 1 && c.customUA == "" {
			req.Header.Set("User-Agent", RandomUserAgent())
		}
		// Apply custom headers override
		for k, v := range headers {
			req.Header[k] = v
		}

		if len(body) > 0 && req.Header.Get("Content-Type") == "" {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = err
			if attempt < maxAttempts {
				c.backoff(ctx, attempt)
				continue
			}
			return nil, fmt.Errorf("request failed: %w", err)
		}

		var reader io.Reader = resp.Body
		contentEncoding := strings.ToLower(resp.Header.Get("Content-Encoding"))
		if strings.Contains(contentEncoding, "gzip") {
			gz, err := gzip.NewReader(resp.Body)
			if err == nil {
				defer gz.Close()
				reader = gz
			}
		} else if strings.Contains(contentEncoding, "deflate") {
			zl, err := zlib.NewReader(resp.Body)
			if err == nil {
				defer zl.Close()
				reader = zl
			}
		}

		respBody, err := io.ReadAll(reader)
		_ = resp.Body.Close()
		if err != nil {
			lastErr = err
			if attempt < maxAttempts {
				c.backoff(ctx, attempt)
				continue
			}
			return nil, fmt.Errorf("reading response: %w", err)
		}

		// Check if status should trigger a retry
		if c.retryConfig.RetryStatuses[resp.StatusCode] && attempt < maxAttempts {
			lastErr = &models.APIError{
				StatusCode: resp.StatusCode,
				Status:     resp.Status,
				Body:       string(respBody),
			}
			c.backoff(ctx, attempt)
			continue
		}

		if resp.StatusCode >= 400 {
			return nil, &models.APIError{
				StatusCode: resp.StatusCode,
				Status:     resp.Status,
				Body:       string(respBody),
			}
		}

		return respBody, nil
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("request failed after %d attempts", maxAttempts)
}

// RequestJSON executes an HTTP request and decodes the response body into target.
func (c *HTTPClient) RequestJSON(ctx context.Context, method, targetURL string, body []byte, headers http.Header, target any) error {
	data, err := c.Request(ctx, method, targetURL, body, headers)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("%w: %w (body: %s)", models.ErrFailedDecodeJSON, err, string(data))
	}

	return nil
}

func (c *HTTPClient) backoff(ctx context.Context, attempt int) {
	interval := float64(c.retryConfig.InitialInterval) * math.Pow(2, float64(attempt-1))
	jitter := interval * (0.8 + 0.4*rand.Float64())
	d := time.Duration(jitter)
	if d > c.retryConfig.MaxInterval {
		d = c.retryConfig.MaxInterval
	}

	select {
	case <-time.After(d):
	case <-ctx.Done():
	}
}
