package goshazam

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/alexgorbatchev/goshazam/pkg/audio"
	"github.com/alexgorbatchev/goshazam/pkg/client"
	"github.com/alexgorbatchev/goshazam/pkg/geo"
	"github.com/alexgorbatchev/goshazam/pkg/models"
	"github.com/alexgorbatchev/goshazam/pkg/signature"
)

// Shazam is the main client for Shazam music recognition and catalog exploration.
type Shazam struct {
	language         string
	endpointCountry  string
	timeZone         string
	proxyURL         string
	customUA         string
	timeout          time.Duration
	discoveryURL     string
	customHTTPClient *http.Client

	httpClient *client.HTTPClient
	geoService *geo.GeoService
}

// New creates a new Shazam client with the provided functional options.
func New(opts ...Option) *Shazam {
	s := &Shazam{
		language:        "en-US",
		endpointCountry: "GB",
		timeZone:        "Europe/Moscow",
		timeout:         30 * time.Second,
	}

	for _, opt := range opts {
		opt(s)
	}

	if s.httpClient == nil {
		var clientOpts []client.ClientOption
		clientOpts = append(clientOpts, client.WithLanguage(s.language))
		if s.customHTTPClient != nil {
			clientOpts = append(clientOpts, client.WithHTTPClient(s.customHTTPClient))
		}
		if s.timeout > 0 {
			clientOpts = append(clientOpts, client.WithTimeout(s.timeout))
		}
		if s.proxyURL != "" {
			clientOpts = append(clientOpts, client.WithProxy(s.proxyURL))
		}
		if s.customUA != "" {
			clientOpts = append(clientOpts, client.WithUserAgent(s.customUA))
		}
		s.httpClient = client.NewHTTPClient(clientOpts...)
	}

	s.geoService = geo.NewGeoService(s.httpClient)

	return s
}

// GeoService returns the associated GeoService for resolving chart playlist IDs.
func (s *Shazam) GeoService() *geo.GeoService {
	return s.geoService
}

// Recognize recognizes a track from audio file path (string), raw audio bytes ([]byte),
// io.Reader stream, *audio.Segment, or precomputed *signature.DecodedMessage.
func (s *Shazam) Recognize(ctx context.Context, input any) (*models.ResponseTrack, error) {
	switch v := input.(type) {
	case string:
		return s.RecognizeFile(ctx, v)
	case []byte:
		return s.RecognizeBytes(ctx, v)
	case io.Reader:
		return s.RecognizeReader(ctx, v)
	case *audio.Segment:
		return s.RecognizeAudio(ctx, v)
	case *signature.DecodedMessage:
		return s.RecognizeSignature(ctx, v)
	default:
		return nil, fmt.Errorf("unsupported recognize input type: %T (expected string file path, []byte, io.Reader, *audio.Segment, or *signature.DecodedMessage)", input)
	}
}

// RecognizeFile decodes an audio file at filePath and queries Shazam for a match.
func (s *Shazam) RecognizeFile(ctx context.Context, filePath string) (*models.ResponseTrack, error) {
	seg, err := audio.DecodeAudioFile(ctx, filePath)
	if err != nil {
		return nil, fmt.Errorf("decoding audio file: %w", err)
	}
	return s.RecognizeAudio(ctx, seg)
}

// RecognizeBytes decodes audio data from an in-memory byte slice and queries Shazam.
func (s *Shazam) RecognizeBytes(ctx context.Context, data []byte) (*models.ResponseTrack, error) {
	seg, err := audio.DecodeAudioBytes(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("decoding audio bytes: %w", err)
	}
	return s.RecognizeAudio(ctx, seg)
}

// RecognizeReader decodes audio from an io.Reader stream and queries Shazam.
func (s *Shazam) RecognizeReader(ctx context.Context, r io.Reader) (*models.ResponseTrack, error) {
	seg, err := audio.DecodeAudioReader(ctx, r)
	if err != nil {
		return nil, fmt.Errorf("decoding audio stream: %w", err)
	}
	return s.RecognizeAudio(ctx, seg)
}

// RecognizeAudio generates a Shazam fingerprint from an audio segment and queries the recognition API.
func (s *Shazam) RecognizeAudio(ctx context.Context, seg *audio.Segment) (*models.ResponseTrack, error) {
	sg := audio.CreateSignatureGenerator(seg)
	sig := sg.GetNextSignature()
	if sig == nil {
		return &models.ResponseTrack{Matches: []models.MatchModel{}}, nil
	}
	return s.SendRecognizeRequest(ctx, sig)
}

// RecognizeSignature queries the Shazam recognition API with a pre-computed DecodedMessage signature.
func (s *Shazam) RecognizeSignature(ctx context.Context, sig *signature.DecodedMessage) (*models.ResponseTrack, error) {
	if sig == nil {
		return &models.ResponseTrack{Matches: []models.MatchModel{}}, nil
	}
	return s.SendRecognizeRequest(ctx, sig)
}

type recognizePayload struct {
	Timezone    string             `json:"timezone"`
	Signature   signaturePayload   `json:"signature"`
	Timestamp   int64              `json:"timestamp"`
	Context     map[string]any     `json:"context"`
	Geolocation map[string]any     `json:"geolocation"`
}

type signaturePayload struct {
	URI      string `json:"uri"`
	SampleMS int    `json:"samplems"`
}

// SendRecognizeRequest sends the encoded signature to Shazam's discovery API.
func (s *Shazam) SendRecognizeRequest(ctx context.Context, sig *signature.DecodedMessage) (*models.ResponseTrack, error) {
	sampleRate := sig.SampleRateHz
	if sampleRate == 0 {
		sampleRate = 16000
	}
	sampleMS := int(float64(sig.NumberSamples) / float64(sampleRate) * 1000.0)
	payload := recognizePayload{
		Timezone: s.timeZone,
		Signature: signaturePayload{
			URI:      sig.EncodeToURI(),
			SampleMS: sampleMS,
		},
		Timestamp:   time.Now().UnixMilli(),
		Context:     make(map[string]any),
		Geolocation: make(map[string]any),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshaling search payload: %w", err)
	}

	targetURL := s.discoveryURL
	if targetURL == "" {
		device := models.DeviceIPhone
		uuid1 := randomUUID()
		uuid2 := randomUUID()
		targetURL = client.FormatSearchFromFile(s.language, s.endpointCountry, string(device), uuid1, uuid2)
	}

	var resp models.ResponseTrack
	if err := s.httpClient.RequestJSON(ctx, http.MethodPost, targetURL, body, nil, &resp); err != nil {
		return nil, fmt.Errorf("shazam recognize request: %w", err)
	}

	return &resp, nil
}

// TopWorldTracks retrieves the global top chart tracks.
func (s *Shazam) TopWorldTracks(ctx context.Context, limit, offset int) (*models.PlaylistResponse, error) {
	if limit <= 0 {
		limit = 200
	}
	topPlaylistID, err := s.geoService.GetTop(ctx)
	if err != nil {
		return nil, err
	}

	targetURL := client.FormatTopTracksPlaylist(s.endpointCountry, topPlaylistID, limit, offset, s.language)
	var resp models.PlaylistResponse
	if err := s.httpClient.RequestJSON(ctx, http.MethodGet, targetURL, nil, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// TopCountryTracks retrieves the top chart tracks for a given country code (e.g., "US", "GB", "NL").
func (s *Shazam) TopCountryTracks(ctx context.Context, countryCode string, limit, offset int) (*models.PlaylistResponse, error) {
	if limit <= 0 {
		limit = 200
	}
	playlistID, err := s.geoService.GetCountryPlaylist(ctx, countryCode)
	if err != nil {
		return nil, err
	}

	targetURL := client.FormatTopTracksPlaylist(s.endpointCountry, playlistID, limit, offset, s.language)
	var resp models.PlaylistResponse
	if err := s.httpClient.RequestJSON(ctx, http.MethodGet, targetURL, nil, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// TopCityTracks retrieves the top chart tracks for a specific city within a country.
func (s *Shazam) TopCityTracks(ctx context.Context, countryCode, cityName string, limit, offset int) (*models.PlaylistResponse, error) {
	if limit <= 0 {
		limit = 200
	}
	playlistID, err := s.geoService.GetCityPlaylist(ctx, countryCode, cityName)
	if err != nil {
		return nil, err
	}

	targetURL := client.FormatTopTracksPlaylist(s.endpointCountry, playlistID, limit, offset, s.language)
	var resp models.PlaylistResponse
	if err := s.httpClient.RequestJSON(ctx, http.MethodGet, targetURL, nil, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// TopWorldGenreTracks retrieves global chart tracks for a specific music genre.
func (s *Shazam) TopWorldGenreTracks(ctx context.Context, genre models.GenreMusic, limit, offset int) (*models.PlaylistResponse, error) {
	if limit <= 0 {
		limit = 100
	}
	playlistID, err := s.geoService.GetGenre(ctx, genre)
	if err != nil {
		return nil, err
	}

	targetURL := client.FormatTopTracksPlaylist(s.endpointCountry, playlistID, limit, offset, s.language)
	var resp models.PlaylistResponse
	if err := s.httpClient.RequestJSON(ctx, http.MethodGet, targetURL, nil, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// TopCountryGenreTracks retrieves country chart tracks for a specific music genre.
func (s *Shazam) TopCountryGenreTracks(ctx context.Context, countryCode string, genre models.GenreMusic, limit, offset int) (*models.PlaylistResponse, error) {
	if limit <= 0 {
		limit = 200
	}
	playlistID, err := s.geoService.GetGenreFromCountry(ctx, countryCode, genre)
	if err != nil {
		return nil, err
	}

	targetURL := client.FormatTopTracksPlaylist(s.endpointCountry, playlistID, limit, offset, s.language)
	var resp models.PlaylistResponse
	if err := s.httpClient.RequestJSON(ctx, http.MethodGet, targetURL, nil, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ArtistAbout retrieves detailed information for an artist profile.
func (s *Shazam) ArtistAbout(ctx context.Context, artistID int64, query *models.ArtistQuery) (*models.ArtistResponse, error) {
	targetURL := fmt.Sprintf(client.SearchArtistV2URL, s.endpointCountry, artistID)
	if query != nil {
		params := query.Params()
		if len(params) > 0 {
			values := url.Values{}
			for k, v := range params {
				values.Set(k, v)
			}
			targetURL += "?" + values.Encode()
		}
	}

	var resp models.ArtistResponse
	if err := s.httpClient.RequestJSON(ctx, http.MethodGet, targetURL, nil, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ArtistAlbums retrieves albums belonging to a specific artist.
func (s *Shazam) ArtistAlbums(ctx context.Context, artistID int64, limit, offset int) (*models.FullAlbumsModel, error) {
	if limit <= 0 {
		limit = 10
	}
	targetURL := client.FormatArtistAlbums(s.endpointCountry, artistID, limit, offset)
	var resp models.FullAlbumsModel
	if err := s.httpClient.RequestJSON(ctx, http.MethodGet, targetURL, nil, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SearchAlbum retrieves detailed metadata for an album by its ID.
func (s *Shazam) SearchAlbum(ctx context.Context, albumID int64) (*models.AlbumModel, error) {
	targetURL := client.FormatArtistAlbumInfo(s.endpointCountry, albumID)
	var resp models.BaseDataModel[[]models.AlbumModel]
	if err := s.httpClient.RequestJSON(ctx, http.MethodGet, targetURL, nil, nil, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("album ID %d not found", albumID)
	}
	return &resp.Data[0], nil
}

// TrackAbout retrieves metadata for a track by its ID.
func (s *Shazam) TrackAbout(ctx context.Context, trackID int64) (*models.TrackInfo, error) {
	targetURL := client.FormatAboutTrack(s.language, s.endpointCountry, trackID)
	var resp models.TrackInfo
	if err := s.httpClient.RequestJSON(ctx, http.MethodGet, targetURL, nil, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// RelatedTracks retrieves similar and related songs based on track ID.
func (s *Shazam) RelatedTracks(ctx context.Context, trackID int64, limit, offset int) (map[string]any, error) {
	if limit <= 0 {
		limit = 20
	}
	targetURL := client.FormatRelatedSongs(s.language, s.endpointCountry, trackID, limit, offset)
	var resp map[string]any
	if err := s.httpClient.RequestJSON(ctx, http.MethodGet, targetURL, nil, nil, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// SearchArtist searches artists by prefix or full name.
func (s *Shazam) SearchArtist(ctx context.Context, query string, limit, offset int) (map[string]any, error) {
	if limit <= 0 {
		limit = 10
	}
	targetURL := client.FormatSearchArtist(s.language, s.endpointCountry, query, limit, offset)
	var resp map[string]any
	if err := s.httpClient.RequestJSON(ctx, http.MethodGet, targetURL, nil, nil, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// SearchTrack searches songs by title or prefix.
func (s *Shazam) SearchTrack(ctx context.Context, query string, limit, offset int) (map[string]any, error) {
	if limit <= 0 {
		limit = 10
	}
	targetURL := client.FormatSearchMusic(s.language, s.endpointCountry, query, limit, offset)
	var resp map[string]any
	if err := s.httpClient.RequestJSON(ctx, http.MethodGet, targetURL, nil, nil, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// ListeningCounter retrieves total listener count for a specific track.
func (s *Shazam) ListeningCounter(ctx context.Context, trackID int64) (map[string]any, error) {
	targetURL := fmt.Sprintf(client.ListeningCounterURL, trackID)
	var resp map[string]any
	if err := s.httpClient.RequestJSON(ctx, http.MethodGet, targetURL, nil, nil, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// ListeningCounterMany retrieves listener counts for multiple track IDs.
func (s *Shazam) ListeningCounterMany(ctx context.Context, trackIDs []int64) ([]map[string]any, error) {
	if len(trackIDs) == 0 {
		return nil, errors.New("trackIDs cannot be empty")
	}

	values := url.Values{}
	for _, id := range trackIDs {
		values.Add("id", strconv.FormatInt(id, 10))
	}

	targetURL := client.ListeningCounterManyURL + "?" + values.Encode()
	var resp []map[string]any
	if err := s.httpClient.RequestJSON(ctx, http.MethodGet, targetURL, nil, nil, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetYouTubeData retrieves YouTube video information for a link.
func (s *Shazam) GetYouTubeData(ctx context.Context, link string) (*models.YoutubeData, error) {
	var resp models.YoutubeData
	if err := s.httpClient.RequestJSON(ctx, http.MethodGet, link, nil, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
