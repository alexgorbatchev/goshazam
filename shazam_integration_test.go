//go:build integration

package goshazam

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/alexgorbatchev/goshazam/pkg/audio"
	"github.com/alexgorbatchev/goshazam/pkg/models"
)

func TestShazamLiveRecognizeGloria(t *testing.T) {
	if !audio.HasFFmpeg() {
		t.Skip("ffmpeg not found, skipping live recognition test")
	}

	gloriaPath := os.Getenv("SHAZAM_TEST_AUDIO")
	if gloriaPath == "" {
		t.Skip("set SHAZAM_TEST_AUDIO to the ShazamIO Gloria.ogg fixture")
	}

	s := New(
		WithLanguage("en-US"),
		WithEndpointCountry("GB"),
		WithTimeZone("Europe/Moscow"),
		WithTimeout(30*time.Second),
	)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	res, err := s.RecognizeFile(ctx, gloriaPath)
	if err != nil {
		var apiErr *models.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == 429 {
			t.Skip("rate limited by upstream live Shazam API (429), skipping live test")
		}
		t.Fatalf("RecognizeFile failed: %v", err)
	}

	if len(res.Matches) == 0 {
		t.Fatalf("expected matches for Gloria.ogg, got 0")
	}

	if res.Track == nil {
		t.Fatalf("expected track metadata, got nil")
	}

	if res.Track.Key != "53982678" {
		t.Errorf("expected track key 53982678, got %s", res.Track.Key)
	}

	t.Logf("Successfully recognized: %s - %s (key: %s, matches: %d)",
		res.Track.Title, res.Track.Subtitle, res.Track.Key, len(res.Matches))
}

func TestShazamLiveRelatedTracks(t *testing.T) {
	s := New(WithTimeout(30 * time.Second))
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	related, err := s.RelatedTracks(ctx, 53982678, 3, 0)
	if err != nil {
		var apiErr *models.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == 429 {
			t.Skip("rate limited by upstream live Shazam API (429), skipping live test")
		}
		t.Fatalf("RelatedTracks failed: %v", err)
	}

	tracks, ok := related["tracks"].([]any)
	if !ok {
		t.Fatalf("expected related tracks list, got %#v", related)
	}
	for _, value := range tracks {
		track, ok := value.(map[string]any)
		if !ok || track["key"] == nil {
			t.Fatalf("invalid related track: %#v", value)
		}
	}
	t.Logf("Shazam returned %d related tracks; HTTP 204 legitimately returns none", len(tracks))
}
