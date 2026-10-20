package goshazam

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/alexgorbatchev/goshazam/pkg/client"
	"github.com/alexgorbatchev/goshazam/pkg/models"
)

func TestRelatedTracksHTTPResponses(t *testing.T) {
	for _, tc := range []struct {
		name            string
		status          int
		body            string
		wantDecodeError bool
		wantAPIError    bool
	}{
		{name: "no content", status: http.StatusNoContent},
		{name: "recommendations", status: http.StatusOK, body: `{"tracks":[{"key":"123","title":"Related"}]}`},
		{name: "empty JSON list", status: http.StatusOK, body: `{"tracks":[]}`},
		{name: "empty OK is malformed", status: http.StatusOK, wantDecodeError: true},
		{name: "malformed JSON", status: http.StatusOK, body: `{`, wantDecodeError: true},
		{name: "unauthorized", status: http.StatusUnauthorized, body: `{"message":"Unauthorized"}`, wantAPIError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Path != "/shazam/v3/fr-FR/GB/web/-/tracks/track-similarities-id-53982678" || r.URL.Query().Get("startFrom") != "2" || r.URL.Query().Get("pageSize") != "3" {
					t.Errorf("unexpected related request: %s %s", r.Method, r.URL)
				}
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()
			target, err := url.Parse(server.URL)
			if err != nil {
				t.Fatal(err)
			}
			s := New(WithLanguage("fr-FR"), WithEndpointCountry("GB"), WithHTTPClient(client.NewHTTPClient(client.WithHTTPClient(&http.Client{Transport: &rewriteTransport{target: target, transport: server.Client().Transport}}))))
			got, err := s.RelatedTracks(context.Background(), 53982678, 3, 2)
			if tc.wantDecodeError {
				if !errors.Is(err, models.ErrFailedDecodeJSON) {
					t.Fatalf("expected JSON error, got %v", err)
				}
				return
			}
			if tc.wantAPIError {
				var apiErr *models.APIError
				if !errors.As(err, &apiErr) || apiErr.StatusCode != tc.status {
					t.Fatalf("expected API status %d, got %v", tc.status, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			tracks, ok := got["tracks"].([]any)
			if !ok {
				t.Fatalf("expected tracks list, got %#v", got)
			}
			if tc.name == "recommendations" {
				if len(tracks) != 1 || tracks[0].(map[string]any)["key"] != "123" {
					t.Fatalf("lost native recommendations: %#v", tracks)
				}
			} else if len(tracks) != 0 {
				t.Fatalf("expected empty related tracks, got %#v", got)
			}
		})
	}
}
