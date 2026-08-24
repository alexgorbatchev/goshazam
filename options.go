package goshazam

import (
	"net/http"
	"time"

	"github.com/alexgorbatchev/goshazam/pkg/client"
)

// Option configures a Shazam instance.
type Option func(*Shazam)

// WithLanguage sets the preferred language (e.g., "en-US", "es-ES", "fr-FR"). Default is "en-US".
func WithLanguage(lang string) Option {
	return func(s *Shazam) {
		s.language = lang
	}
}

// WithEndpointCountry sets the catalog country endpoint (e.g., "GB", "US"). Default is "GB".
func WithEndpointCountry(country string) Option {
	return func(s *Shazam) {
		s.endpointCountry = country
	}
}

// WithTimeZone sets the client timezone string in search payload. Default is "Europe/Moscow".
func WithTimeZone(tz string) Option {
	return func(s *Shazam) {
		s.timeZone = tz
	}
}

// WithHTTPClient provides a custom client.HTTPClient.
func WithHTTPClient(httpClient *client.HTTPClient) Option {
	return func(s *Shazam) {
		s.httpClient = httpClient
	}
}

// WithStandardHTTPClient provides a custom standard library *http.Client.
func WithStandardHTTPClient(httpClient *http.Client) Option {
	return func(s *Shazam) {
		s.customHTTPClient = httpClient
	}
}

// WithProxy configures an HTTP/HTTPS/SOCKS5 proxy.
func WithProxy(proxyURL string) Option {
	return func(s *Shazam) {
		s.proxyURL = proxyURL
	}
}

// WithUserAgent configures a custom User-Agent string.
func WithUserAgent(ua string) Option {
	return func(s *Shazam) {
		s.customUA = ua
	}
}

// WithTimeout configures the default HTTP client request timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(s *Shazam) {
		s.timeout = timeout
	}
}

// WithDiscoveryURL overrides the default discovery base URL (useful for testing or private mirrors).
func WithDiscoveryURL(discoveryURL string) Option {
	return func(s *Shazam) {
		s.discoveryURL = discoveryURL
	}
}
