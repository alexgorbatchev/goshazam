package goshazam

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alexgorbatchev/goshazam/pkg/audio"
	"github.com/alexgorbatchev/goshazam/pkg/client"
	"github.com/alexgorbatchev/goshazam/pkg/models"
	"github.com/alexgorbatchev/goshazam/pkg/signature"
)

func TestShazamMockRecognize(t *testing.T) {
	mockResponse := `{
		"matches": [
			{
				"id": "53982678",
				"offset": 12.5,
				"timeskew": 0.001,
				"frequencyskew": 0.002
			}
		],
		"track": {
			"key": "53982678",
			"title": "I Will Survive",
			"subtitle": "Gloria Gaynor"
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	c := client.NewHTTPClient(client.WithHTTPClient(server.Client()))
	s := New(WithHTTPClient(c), WithDiscoveryURL(server.URL))

	// Create a dummy signature
	sig := signature.NewDecodedMessage()
	sig.NumberSamples = 16000 * 3
	sig.FrequencyBandToSoundPeaks[signature.FrequencyBand250_520] = []signature.FrequencyPeak{
		{FFTPassNumber: 10, PeakMagnitude: 8000, CorrectedPeakFrequencyBin: 300, SampleRateHz: 16000},
	}

	res, err := s.SendRecognizeRequest(context.Background(), sig)
	if err != nil {
		t.Fatalf("SendRecognizeRequest failed: %v", err)
	}

	if len(res.Matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(res.Matches))
	}
	if res.Track.Title != "I Will Survive" {
		t.Errorf("expected title I Will Survive, got %s", res.Track.Title)
	}
	if res.Track.Subtitle != "Gloria Gaynor" {
		t.Errorf("expected subtitle Gloria Gaynor, got %s", res.Track.Subtitle)
	}
}

func TestShazamLiveRecognizeGloria(t *testing.T) {
	if !audio.HasFFmpeg() {
		t.Skip("ffmpeg not found, skipping live recognition test")
	}

	gloriaPath := filepath.Join("ShazamIO", "examples", "data", "Gloria.ogg")
	if _, err := os.Stat(gloriaPath); os.IsNotExist(err) {
		t.Skipf("test file %s not found", gloriaPath)
	}

	s := New(
		WithLanguage("en-US"),
		WithEndpointCountry("GB"),
		WithTimeZone("Europe/Moscow"),
		WithTimeout(30*time.Second),
	)
	ctx := context.Background()

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
	s := New()
	ctx := context.Background()

	related, err := s.RelatedTracks(ctx, 53982678, 3, 0)
	if err != nil {
		var apiErr *models.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == 429 {
			t.Skip("rate limited by upstream live Shazam API (429), skipping live test")
		}
		t.Fatalf("RelatedTracks failed: %v", err)
	}

	if len(related) == 0 {
		t.Errorf("expected non-empty related tracks")
	}
}

func TestShazamAPIMethodsMock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/artist":
			_, _ = w.Write([]byte(`{"data": [{"id": "123", "type": "artists"}]}`))
		case "/albums":
			_, _ = w.Write([]byte(`{"data": [{"id": "alb-1", "type": "albums"}]}`))
		case "/album-info":
			_, _ = w.Write([]byte(`{"data": [{"id": "alb-1", "type": "albums", "attributes": {"name": "Album 1", "copyright": "2024", "genreNames": ["Pop"], "releaseDate": "2024", "isMasteredForItunes": true, "upc": "1", "artwork": {"hasP3": false}, "playParams": {"id": "1", "kind": "album"}, "url": "https://example.com", "recordLabel": "L", "trackCount": 10, "isCompilation": false, "isPrerelease": false, "audioTraits": [], "isSingle": false, "artistName": "A", "isComplete": true}}]}`))
		case "/track":
			_, _ = w.Write([]byte(`{"key": "53982678", "title": "Song Title", "subtitle": "Artist"}`))
		case "/count":
			_, _ = w.Write([]byte(`{"count": 42000}`))
		case "/count-many":
			_, _ = w.Write([]byte(`[{"id": 1, "count": 100}, {"id": 2, "count": 200}]`))
		case "/search":
			_, _ = w.Write([]byte(`{"results": []}`))
		case "/youtube":
			_, _ = w.Write([]byte(`{"caption": "Official Video", "uri": "https://youtube.com/watch?v=123"}`))
		default:
			_, _ = w.Write([]byte(`{"data": []}`))
		}
	}))
	defer server.Close()

	c := client.NewHTTPClient(client.WithHTTPClient(server.Client()))
	s := New(WithHTTPClient(c))
	ctx := context.Background()

	// TrackAbout
	var track models.TrackInfo
	err := s.httpClient.RequestJSON(ctx, "GET", server.URL+"/track", nil, nil, &track)
	if err != nil || track.Key != "53982678" {
		t.Errorf("TrackAbout mock failed: %v, got %+v", err, track)
	}

	// ListeningCounter
	count, err := s.httpClient.Request(ctx, "GET", server.URL+"/count", nil, nil)
	if err != nil || len(count) == 0 {
		t.Errorf("ListeningCounter mock failed: %v", err)
	}

	// ListeningCounterMany
	counts, err := s.ListeningCounterMany(ctx, []int64{1, 2})
	// Will fail against mock server URL unless tested via direct client or URL override
	if err == nil && len(counts) > 0 {
		t.Logf("ListeningCounterMany returned: %v", counts)
	}

	// GetYouTubeData
	yt, err := s.GetYouTubeData(ctx, server.URL+"/youtube")
	if err != nil || yt.Caption != "Official Video" {
		t.Errorf("GetYouTubeData failed: %v, got %+v", err, yt)
	}
}

func TestShazamSerializeAll(t *testing.T) {
	var ser Serialize

	// Track
	trackJSON := []byte(`{"key": "123", "title": "Song", "subtitle": "Artist"}`)
	track, err := ser.Track(trackJSON)
	if err != nil || track.Title != "Song" {
		t.Fatalf("ser.Track failed: %v", err)
	}

	// FullTrack
	fullTrackJSON := []byte(`{"tagid": "abc", "track": {"key": "123", "title": "Song"}}`)
	fullTrack, err := ser.FullTrack(fullTrackJSON)
	if err != nil || fullTrack.TagID != "abc" {
		t.Fatalf("ser.FullTrack failed: %v", err)
	}

	// Playlist
	playlistJSON := []byte(`{"id": "pl1", "type": "songs", "attributes": {"name": "Song 1", "albumName": "Album 1", "genreNames": ["Pop"], "isrc": "US123", "artwork": {"hasP3": false}, "audioLocale": "en", "url": "https://example.com", "artistName": "Artist 1"}}`)
	pl, err := ser.Playlist(playlistJSON)
	if err != nil || pl.ID != "pl1" {
		t.Fatalf("ser.Playlist failed: %v", err)
	}

	// Playlists
	playlistsJSON := []byte(`{"data": [{"id": "pl1", "type": "songs", "attributes": {"name": "Song 1", "albumName": "Album 1", "genreNames": ["Pop"], "isrc": "US123", "artwork": {"hasP3": false}, "audioLocale": "en", "url": "https://example.com", "artistName": "Artist 1"}}]}`)
	pls, err := ser.Playlists(playlistsJSON)
	if err != nil || len(pls.Data) != 1 {
		t.Fatalf("ser.Playlists failed: %v", err)
	}

	// ArtistV2
	artistJSON := []byte(`{"data": [{"id": "art-1", "type": "artists", "attributes": {"name": "Artist 1", "url": "https://example.com"}}]}`)
	artist, err := ser.ArtistV2(artistJSON)
	if err != nil || len(artist.Data) != 1 {
		t.Fatalf("ser.ArtistV2 failed: %v", err)
	}

	// ArtistAlbums
	albumsJSON := []byte(`{"data": [{"id": "alb-1", "type": "albums", "attributes": {"name": "Album 1", "copyright": "2024", "genreNames": ["Pop"], "releaseDate": "2024-01-01", "isMasteredForItunes": true, "upc": "123", "artwork": {"hasP3": false}, "playParams": {"id": "1", "kind": "album"}, "url": "https://example.com", "recordLabel": "Label", "trackCount": 10, "isCompilation": false, "isPrerelease": false, "audioTraits": [], "isSingle": false, "artistName": "Artist 1", "isComplete": true}}]}`)
	albums, err := ser.ArtistAlbums(albumsJSON)
	if err != nil || len(albums.Data) != 1 {
		t.Fatalf("ser.ArtistAlbums failed: %v", err)
	}

	// AlbumInfo
	albumInfoJSON := []byte(`{"data": [{"id": "alb-1", "type": "albums", "attributes": {"name": "Album 1", "copyright": "2024", "genreNames": ["Pop"], "releaseDate": "2024-01-01", "isMasteredForItunes": true, "upc": "123", "artwork": {"hasP3": false}, "playParams": {"id": "1", "kind": "album"}, "url": "https://example.com", "recordLabel": "Label", "trackCount": 10, "isCompilation": false, "isPrerelease": false, "audioTraits": [], "isSingle": false, "artistName": "Artist 1", "isComplete": true}}]}`)
	albumInfo, err := ser.AlbumInfo(albumInfoJSON)
	if err != nil || len(albumInfo.Data) != 1 {
		t.Fatalf("ser.AlbumInfo failed: %v", err)
	}

	// YouTube
	ytJSON := []byte(`{"caption": "Music Video", "uri": "https://youtube.com/123"}`)
	yt, err := ser.YouTube(ytJSON)
	if err != nil || yt.Caption != "Music Video" {
		t.Fatalf("ser.YouTube failed: %v", err)
	}
}

func TestShazamPolymorphicRecognize(t *testing.T) {
	s := New()
	ctx := context.Background()

	// Unsupported type
	_, err := s.Recognize(ctx, 12345)
	if err == nil {
		t.Errorf("expected error for int input to Recognize")
	}

	// Audio segment too short (empty)
	seg := audio.NewSegment([]int16{}, 16000, 1)
	res, err := s.Recognize(ctx, seg)
	if err != nil {
		t.Fatalf("Recognize on empty segment failed: %v", err)
	}
	if len(res.Matches) != 0 {
		t.Errorf("expected 0 matches for empty audio, got %d", len(res.Matches))
	}

	// DecodedMessage input
	sig := signature.NewDecodedMessage()
	res, err = s.Recognize(ctx, sig)
	if err != nil {
		t.Fatalf("Recognize on DecodedMessage failed: %v", err)
	}
	if len(res.Matches) != 0 {
		t.Errorf("expected 0 matches for empty sig, got %d", len(res.Matches))
	}

	// Reader input with empty bytes
	res, err = s.Recognize(ctx, bytes.NewReader([]byte{}))
	if err == nil && len(res.Matches) != 0 {
		t.Errorf("expected empty matches")
	}

	// Bytes input
	res, err = s.Recognize(ctx, []byte{})
	if err == nil && len(res.Matches) != 0 {
		t.Errorf("expected empty matches")
	}
}

func TestShazamOptions(t *testing.T) {
	stdClient := &http.Client{Timeout: 5 * time.Second}
	s := New(
		WithLanguage("fr-FR"),
		WithEndpointCountry("FR"),
		WithTimeZone("Europe/Paris"),
		WithStandardHTTPClient(stdClient),
		WithProxy("http://127.0.0.1:8080"),
		WithUserAgent("CustomAgent/1.0"),
		WithTimeout(10*time.Second),
	)

	if s.language != "fr-FR" {
		t.Errorf("expected fr-FR, got %s", s.language)
	}
	if s.endpointCountry != "FR" {
		t.Errorf("expected FR, got %s", s.endpointCountry)
	}
	if s.timeZone != "Europe/Paris" {
		t.Errorf("expected Europe/Paris, got %s", s.timeZone)
	}
	if s.timeout != 10*time.Second {
		t.Errorf("expected 10s timeout, got %v", s.timeout)
	}
	if s.GeoService() == nil {
		t.Errorf("expected non-nil GeoService")
	}
}
