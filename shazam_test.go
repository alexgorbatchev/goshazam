package goshazam

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alexgorbatchev/goshazam/pkg/audio"
	"github.com/alexgorbatchev/goshazam/pkg/client"
	"github.com/alexgorbatchev/goshazam/pkg/models"
	"github.com/alexgorbatchev/goshazam/pkg/signature"
)

func createSimpleWAV(numSamples int) []byte {
	var buf bytes.Buffer
	dataSize := numSamples * 2
	chunkSize := 36 + dataSize

	buf.WriteString("RIFF")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(chunkSize))
	buf.WriteString("WAVE")

	buf.WriteString("fmt ")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(16))
	_ = binary.Write(&buf, binary.LittleEndian, uint16(1))
	_ = binary.Write(&buf, binary.LittleEndian, uint16(1))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(16000))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(32000))
	_ = binary.Write(&buf, binary.LittleEndian, uint16(2))
	_ = binary.Write(&buf, binary.LittleEndian, uint16(16))

	buf.WriteString("data")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(dataSize))
	for range numSamples {
		_ = binary.Write(&buf, binary.LittleEndian, int16(1000))
	}
	return buf.Bytes()
}

type rewriteTransport struct {
	target    *url.URL
	transport http.RoundTripper
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = t.target.Scheme
	req.URL.Host = t.target.Host
	return t.transport.RoundTrip(req)
}

const mockLocationsJSON = `{
	"countries": [
		{
			"id": "US",
			"listid": "pl.us-top-200",
			"name": "United States",
			"cities": [
				{
					"id": "1",
					"listid": "pl.us-nyc-50",
					"name": "New York"
				}
			],
			"genres": [
				{
					"id": "1",
					"listid": "pl.us-pop",
					"name": "Pop",
					"urlName": "pop"
				}
			]
		}
	],
	"global": {
		"top": {
			"listid": "pl.global-top-200"
		},
		"genres": [
			{
				"id": "1",
				"listid": "pl.global-rock",
				"name": "Rock",
				"urlName": "rock"
			}
		]
	}
}`

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

func TestShazamAllMethodsMock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		path := r.URL.Path

		switch {
		case strings.Contains(path, "locations"):
			_, _ = w.Write([]byte(mockLocationsJSON))
		case strings.Contains(path, "playlists"):
			_, _ = w.Write([]byte(`{"data": [{"id": "track-1", "type": "songs", "attributes": {"name": "Hit Song", "albumName": "Hit Album", "genreNames": ["Pop"], "isrc": "US123", "artwork": {"hasP3": false}, "audioLocale": "en", "url": "https://example.com", "artistName": "Star"}}]}`))
		case strings.Contains(path, "artists") && strings.Contains(path, "albums"):
			_, _ = w.Write([]byte(`{"data": [{"id": "alb-1", "type": "albums", "attributes": {"name": "Album 1", "copyright": "2024", "genreNames": ["Pop"], "releaseDate": "2024-01-01", "isMasteredForItunes": true, "upc": "123", "artwork": {"hasP3": false}, "playParams": {"id": "1", "kind": "album"}, "url": "https://example.com", "recordLabel": "Label", "trackCount": 10, "isCompilation": false, "isPrerelease": false, "audioTraits": [], "isSingle": false, "artistName": "Artist", "isComplete": true}}]}`))
		case strings.Contains(path, "artists"):
			_, _ = w.Write([]byte(`{"data": [{"id": "art-1", "type": "artists", "attributes": {"name": "Artist Name", "url": "https://example.com"}}]}`))
		case strings.Contains(path, "albums/9999"):
			_, _ = w.Write([]byte(`{"data": []}`))
		case strings.Contains(path, "albums"):
			_, _ = w.Write([]byte(`{"data": [{"id": "alb-1", "type": "albums", "attributes": {"name": "Album Info", "copyright": "2024", "genreNames": ["Rock"], "releaseDate": "2024-01-01", "isMasteredForItunes": true, "upc": "123", "artwork": {"hasP3": false}, "playParams": {"id": "1", "kind": "album"}, "url": "https://example.com", "recordLabel": "Label", "trackCount": 10, "isCompilation": false, "isPrerelease": false, "audioTraits": [], "isSingle": false, "artistName": "Artist", "isComplete": true}}]}`))
		case strings.Contains(path, "track-similarities"):
			_, _ = w.Write([]byte(`{"tracks": [{"key": "100", "title": "Similar Track"}]}`))
		case strings.Contains(path, "count/v2/web/track/"):
			_, _ = w.Write([]byte(`{"count": 50000}`))
		case strings.Contains(path, "count/v2/web/track"):
			_, _ = w.Write([]byte(`[{"id": 123, "count": 1000}, {"id": 456, "count": 2000}]`))
		case strings.Contains(path, "track/"):
			_, _ = w.Write([]byte(`{"key": "12345", "title": "Track Title", "subtitle": "Artist"}`))
		case strings.Contains(path, "search"):
			_, _ = w.Write([]byte(`{"results": [{"name": "Found"}]}`))
		case strings.Contains(path, "youtube"):
			_, _ = w.Write([]byte(`{"caption": "Music Video", "uri": "https://youtube.com/watch?v=abc"}`))
		default:
			_, _ = w.Write([]byte(`{"status": "ok"}`))
		}
	}))
	defer server.Close()

	serverURL, _ := url.Parse(server.URL)
	httpClient := &http.Client{
		Transport: &rewriteTransport{
			target:    serverURL,
			transport: server.Client().Transport,
		},
	}

	c := client.NewHTTPClient(client.WithHTTPClient(httpClient))
	s := New(WithHTTPClient(c))
	s.GeoService().SetLocationsURL(server.URL + "/locations")
	ctx := context.Background()

	// TopWorldTracks
	topWorld, err := s.TopWorldTracks(ctx, 10, 0)
	if err != nil || len(topWorld.Data) == 0 {
		t.Fatalf("TopWorldTracks failed: %v", err)
	}

	// TopCountryTracks
	topCountry, err := s.TopCountryTracks(ctx, "US", 10, 0)
	if err != nil || len(topCountry.Data) == 0 {
		t.Fatalf("TopCountryTracks failed: %v", err)
	}

	// TopCityTracks
	topCity, err := s.TopCityTracks(ctx, "US", "New York", 10, 0)
	if err != nil || len(topCity.Data) == 0 {
		t.Fatalf("TopCityTracks failed: %v", err)
	}

	// TopWorldGenreTracks
	topGenre, err := s.TopWorldGenreTracks(ctx, models.GenreRock, 10, 0)
	if err != nil || len(topGenre.Data) == 0 {
		t.Fatalf("TopWorldGenreTracks failed: %v", err)
	}

	// TopCountryGenreTracks
	topCountryGenre, err := s.TopCountryGenreTracks(ctx, "US", models.GenrePop, 10, 0)
	if err != nil || len(topCountryGenre.Data) == 0 {
		t.Fatalf("TopCountryGenreTracks failed: %v", err)
	}

	// ArtistAbout
	artistResp, err := s.ArtistAbout(ctx, 123, &models.ArtistQuery{
		Views:  []models.ArtistView{models.ArtistViewFullAlbums},
		Extend: []models.ArtistExtend{models.ArtistExtendBio},
	})
	if err != nil || len(artistResp.Data) == 0 {
		t.Fatalf("ArtistAbout failed: %v", err)
	}

	// ArtistAlbums
	albumsResp, err := s.ArtistAlbums(ctx, 123, 10, 0)
	if err != nil || len(albumsResp.Data) == 0 {
		t.Fatalf("ArtistAlbums failed: %v", err)
	}

	// SearchAlbum
	albumModel, err := s.SearchAlbum(ctx, 123)
	if err != nil || albumModel.ID != "alb-1" {
		t.Fatalf("SearchAlbum failed: %v", err)
	}

	// SearchAlbum not found error
	if _, err := s.SearchAlbum(ctx, 9999); err == nil {
		t.Errorf("expected error for empty album data")
	}

	// TrackAbout
	trackInfo, err := s.TrackAbout(ctx, 12345)
	if err != nil || trackInfo.Key != "12345" {
		t.Fatalf("TrackAbout failed: %v", err)
	}

	// RelatedTracks
	related, err := s.RelatedTracks(ctx, 12345, 10, 0)
	if err != nil || len(related) == 0 {
		t.Fatalf("RelatedTracks failed: %v", err)
	}

	// SearchArtist
	searchArt, err := s.SearchArtist(ctx, "Artist", 10, 0)
	if err != nil || len(searchArt) == 0 {
		t.Fatalf("SearchArtist failed: %v", err)
	}

	// SearchTrack
	searchTrk, err := s.SearchTrack(ctx, "Track", 10, 0)
	if err != nil || len(searchTrk) == 0 {
		t.Fatalf("SearchTrack failed: %v", err)
	}

	// ListeningCounter
	counter, err := s.ListeningCounter(ctx, 12345)
	if err != nil || counter["count"] == nil {
		t.Fatalf("ListeningCounter failed: %v", err)
	}

	// ListeningCounterMany
	counterMany, err := s.ListeningCounterMany(ctx, []int64{123, 456})
	if err != nil || len(counterMany) != 2 {
		t.Fatalf("ListeningCounterMany failed: %v", err)
	}

	// Test with default (zero) limits
	_, _ = s.TopWorldTracks(ctx, 0, 0)
	_, _ = s.TopCountryTracks(ctx, "US", 0, 0)
	_, _ = s.TopCityTracks(ctx, "US", "New York", 0, 0)
	_, _ = s.TopWorldGenreTracks(ctx, models.GenreRock, 0, 0)
	_, _ = s.TopCountryGenreTracks(ctx, "US", models.GenrePop, 0, 0)
	_, _ = s.ArtistAbout(ctx, 123, nil)
	_, _ = s.ArtistAlbums(ctx, 123, 0, 0)
	_, _ = s.RelatedTracks(ctx, 12345, 0, 0)
	_, _ = s.SearchArtist(ctx, "Artist", 0, 0)
	_, _ = s.SearchTrack(ctx, "Track", 0, 0)

	// Valid WAV bytes and reader recognition
	wavBytes := createSimpleWAV(2048)
	_, _ = s.RecognizeBytes(ctx, wavBytes)
	_, _ = s.RecognizeReader(ctx, bytes.NewReader(wavBytes))
	_, _ = s.Recognize(ctx, wavBytes)
	_, _ = s.Recognize(ctx, bytes.NewReader(wavBytes))

	// GetYouTubeData
	yt, err := s.GetYouTubeData(ctx, server.URL+"/youtube")
	if err != nil || yt.Caption != "Music Video" {
		t.Fatalf("GetYouTubeData failed: %v", err)
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
	if _, err := ser.Track([]byte(`{invalid}`)); err == nil {
		t.Errorf("expected error for invalid track JSON")
	}

	// FullTrack
	fullTrackJSON := []byte(`{"tagid": "abc", "track": {"key": "123", "title": "Song"}}`)
	fullTrack, err := ser.FullTrack(fullTrackJSON)
	if err != nil || fullTrack.TagID != "abc" {
		t.Fatalf("ser.FullTrack failed: %v", err)
	}
	if _, err := ser.FullTrack([]byte(`{invalid}`)); err == nil {
		t.Errorf("expected error for invalid full track JSON")
	}

	// Playlist
	playlistJSON := []byte(`{"id": "pl1", "type": "songs", "attributes": {"name": "Song 1", "albumName": "Album 1", "genreNames": ["Pop"], "isrc": "US123", "artwork": {"hasP3": false}, "audioLocale": "en", "url": "https://example.com", "artistName": "Artist 1"}}`)
	pl, err := ser.Playlist(playlistJSON)
	if err != nil || pl.ID != "pl1" {
		t.Fatalf("ser.Playlist failed: %v", err)
	}
	if _, err := ser.Playlist([]byte(`{invalid}`)); err == nil {
		t.Errorf("expected error for invalid playlist JSON")
	}

	// Playlists
	playlistsJSON := []byte(`{"data": [{"id": "pl1", "type": "songs", "attributes": {"name": "Song 1", "albumName": "Album 1", "genreNames": ["Pop"], "isrc": "US123", "artwork": {"hasP3": false}, "audioLocale": "en", "url": "https://example.com", "artistName": "Artist 1"}}]}`)
	pls, err := ser.Playlists(playlistsJSON)
	if err != nil || len(pls.Data) != 1 {
		t.Fatalf("ser.Playlists failed: %v", err)
	}
	if _, err := ser.Playlists([]byte(`{invalid}`)); err == nil {
		t.Errorf("expected error for invalid playlists JSON")
	}

	// ArtistV2
	artistJSON := []byte(`{"data": [{"id": "art-1", "type": "artists", "attributes": {"name": "Artist 1", "url": "https://example.com"}}]}`)
	artist, err := ser.ArtistV2(artistJSON)
	if err != nil || len(artist.Data) != 1 {
		t.Fatalf("ser.ArtistV2 failed: %v", err)
	}
	if _, err := ser.ArtistV2([]byte(`{invalid}`)); err == nil {
		t.Errorf("expected error for invalid artist JSON")
	}

	// ArtistAlbums
	albumsJSON := []byte(`{"data": [{"id": "alb-1", "type": "albums", "attributes": {"name": "Album 1", "copyright": "2024", "genreNames": ["Pop"], "releaseDate": "2024-01-01", "isMasteredForItunes": true, "upc": "123", "artwork": {"hasP3": false}, "playParams": {"id": "1", "kind": "album"}, "url": "https://example.com", "recordLabel": "Label", "trackCount": 10, "isCompilation": false, "isPrerelease": false, "audioTraits": [], "isSingle": false, "artistName": "Artist 1", "isComplete": true}}]}`)
	albums, err := ser.ArtistAlbums(albumsJSON)
	if err != nil || len(albums.Data) != 1 {
		t.Fatalf("ser.ArtistAlbums failed: %v", err)
	}
	if _, err := ser.ArtistAlbums([]byte(`{invalid}`)); err == nil {
		t.Errorf("expected error for invalid artist albums JSON")
	}

	// AlbumInfo
	albumInfoJSON := []byte(`{"data": [{"id": "alb-1", "type": "albums", "attributes": {"name": "Album 1", "copyright": "2024", "genreNames": ["Pop"], "releaseDate": "2024-01-01", "isMasteredForItunes": true, "upc": "123", "artwork": {"hasP3": false}, "playParams": {"id": "1", "kind": "album"}, "url": "https://example.com", "recordLabel": "Label", "trackCount": 10, "isCompilation": false, "isPrerelease": false, "audioTraits": [], "isSingle": false, "artistName": "Artist 1", "isComplete": true}}]}`)
	albumInfo, err := ser.AlbumInfo(albumInfoJSON)
	if err != nil || len(albumInfo.Data) != 1 {
		t.Fatalf("ser.AlbumInfo failed: %v", err)
	}
	if _, err := ser.AlbumInfo([]byte(`{invalid}`)); err == nil {
		t.Errorf("expected error for invalid album info JSON")
	}

	// YouTube
	ytJSON := []byte(`{"caption": "Music Video", "uri": "https://youtube.com/123"}`)
	yt, err := ser.YouTube(ytJSON)
	if err != nil || yt.Caption != "Music Video" {
		t.Fatalf("ser.YouTube failed: %v", err)
	}
	if _, err := ser.YouTube([]byte(`{invalid}`)); err == nil {
		t.Errorf("expected error for invalid youtube JSON")
	}
}

func TestShazamPolymorphicRecognize(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"matches": []}`))
	}))
	defer server.Close()

	c := client.NewHTTPClient(client.WithHTTPClient(server.Client()))
	s := New(WithHTTPClient(c), WithDiscoveryURL(server.URL))
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

	// Nil DecodedMessage
	res, err = s.RecognizeSignature(ctx, nil)
	if err != nil || len(res.Matches) != 0 {
		t.Errorf("expected empty matches for nil signature")
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

	// Polymorphic Recognize on file path string
	_, err = s.Recognize(ctx, "non_existent_file.mp3")
	if err == nil {
		t.Errorf("expected error for non-existent file path")
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
