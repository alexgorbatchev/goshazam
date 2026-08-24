package models

import (
	"encoding/json"
	"net/url"
	"strings"
)

// ResponseTrack represents the response from Shazam's song recognition API.
type ResponseTrack struct {
	TagID     string         `json:"tagid,omitempty"`
	RetryMS   *int           `json:"retryms,omitempty"`
	Location  *LocationModel `json:"location,omitempty"`
	Matches   []MatchModel   `json:"matches,omitempty"`
	Timestamp *int64         `json:"timestamp,omitempty"`
	Timezone  *string        `json:"timezone,omitempty"`
	Track     *TrackInfo     `json:"track,omitempty"`
}

// LocationModel represents geographic accuracy data from the API.
type LocationModel struct {
	Accuracy float64 `json:"accuracy,omitempty"`
}

// MatchModel represents a fingerprint match with audio offset and skew.
type MatchModel struct {
	ID            string  `json:"id,omitempty"`
	Offset        float64 `json:"offset,omitempty"`
	TimeSkew      float64 `json:"timeskew,omitempty"`
	FrequencySkew float64 `json:"frequencyskew,omitempty"`
	Channel       *string `json:"channel,omitempty"`
}

// ShareModel contains URLs and text for sharing the recognized track.
type ShareModel struct {
	Subject  string `json:"subject,omitempty"`
	Text     string `json:"text,omitempty"`
	Href     string `json:"href,omitempty"`
	Image    string `json:"image,omitempty"`
	Twitter  string `json:"twitter,omitempty"`
	HTML     string `json:"html,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
	Snapchat string `json:"snapchat,omitempty"`
}

// ActionModel represents an actionable button/link in Shazam responses.
type ActionModel struct {
	Name  string      `json:"name,omitempty"`
	Type  string      `json:"type,omitempty"`
	ID    *string     `json:"id,omitempty"`
	URI   *string     `json:"uri,omitempty"`
	Share *ShareModel `json:"share,omitempty"`
}

// HubOption represents hub open/buy options.
type HubOption struct {
	Caption      string        `json:"caption,omitempty"`
	Actions      []ActionModel `json:"actions,omitempty"`
	BeaconData   map[string]any `json:"beacondata,omitempty"`
	Image        string        `json:"image,omitempty"`
	Type         string        `json:"type,omitempty"`
	ListCaption  string        `json:"listcaption,omitempty"`
	ProviderName string        `json:"providername,omitempty"`
}

// HubProvider represents music service providers (e.g. Spotify, Apple Music, Deezer).
type HubProvider struct {
	Caption string            `json:"caption,omitempty"`
	Images  map[string]string `json:"images,omitempty"`
	Actions []ActionModel     `json:"actions,omitempty"`
	Type    string            `json:"type,omitempty"`
}

// HubModel represents streaming service links and actions for a track.
type HubModel struct {
	Type        string        `json:"type,omitempty"`
	Image       string        `json:"image,omitempty"`
	Actions     []ActionModel `json:"actions,omitempty"`
	Options     []HubOption   `json:"options,omitempty"`
	Providers   []HubProvider `json:"providers,omitempty"`
	Explicit    bool          `json:"explicit,omitempty"`
	DisplayName string        `json:"displayname,omitempty"`
}

// SongMetaPages represents album art / metadata page.
type SongMetaPages struct {
	Image   string `json:"image,omitempty"`
	Caption string `json:"caption,omitempty"`
}

// SongMetadata represents key-value metadata (e.g., Album, Label, Released).
type SongMetadata struct {
	Title string `json:"title,omitempty"`
	Text  string `json:"text,omitempty"`
}

// BeaconDataLyricsSection holds provider data for lyrics.
type BeaconDataLyricsSection struct {
	LyricsID      string `json:"lyricsid,omitempty"`
	ProviderName  string `json:"providername,omitempty"`
	CommonTrackID string `json:"commontrackid,omitempty"`
}

// TopTracksModel holds URL for an artist's top tracks.
type TopTracksModel struct {
	URL string `json:"url,omitempty"`
}

// TrackSection represents a section in the track details (Song, Video, Lyrics, Artist, Related).
type TrackSection struct {
	Type        string                   `json:"type,omitempty"`
	TabName     string                   `json:"tabname,omitempty"`
	MetaPages   []SongMetaPages          `json:"metapages,omitempty"`
	Metadata    []SongMetadata           `json:"metadata,omitempty"`
	YouTubeURL  string                   `json:"youtubeurl,omitempty"`
	Text        []string                 `json:"text,omitempty"`
	Footer      string                   `json:"footer,omitempty"`
	BeaconData  *BeaconDataLyricsSection `json:"beacondata,omitempty"`
	ID          string                   `json:"id,omitempty"`
	Name        string                   `json:"name,omitempty"`
	Verified    bool                     `json:"verified,omitempty"`
	Avatar      string                   `json:"avatar,omitempty"`
	Actions     []ActionModel            `json:"actions,omitempty"`
	TopTracks   *TopTracksModel          `json:"toptracks,omitempty"`
	URL         string                   `json:"url,omitempty"`
}

// TrackArtist represents basic artist reference within a track.
type TrackArtist struct {
	ID     string `json:"id,omitempty"`
	AdamID string `json:"adamid,omitempty"`
}

// TrackInfo represents comprehensive metadata for a recognized or queried song.
type TrackInfo struct {
	Key             string            `json:"key,omitempty"`
	Title           string            `json:"title,omitempty"`
	Subtitle        string            `json:"subtitle,omitempty"`
	ArtistID        string            `json:"artist_id,omitempty"`
	ShazamURL       string            `json:"shazam_url,omitempty"`
	PhotoURL        string            `json:"photo_url,omitempty"`
	SpotifyURIQuery string            `json:"spotify_uri_query,omitempty"`
	AppleMusicURL   string            `json:"apple_music_url,omitempty"`
	Ringtone        string            `json:"ringtone,omitempty"`
	SpotifyURL      string            `json:"spotify_url,omitempty"`
	SpotifyURI      string            `json:"spotify_uri,omitempty"`
	YouTubeLink     string            `json:"youtube_link,omitempty"`
	Sections        []TrackSection    `json:"sections,omitempty"`
	Images          map[string]string `json:"images,omitempty"`
	Share           ShareModel        `json:"share,omitempty"`
	Hub             HubModel          `json:"hub,omitempty"`
	Artists         []TrackArtist     `json:"artists,omitempty"`
	ISRC            string            `json:"isrc,omitempty"`
	Genres          map[string]string `json:"genres,omitempty"`
	URL             string            `json:"url,omitempty"`
	Type            string            `json:"type,omitempty"`
	Layout          string            `json:"layout,omitempty"`
}

// UnmarshalJSON customizes unmarshaling to automatically populate derived convenience fields.
func (t *TrackInfo) UnmarshalJSON(data []byte) error {
	type Alias TrackInfo
	aux := (*Alias)(t)
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// Populate ArtistID if not set
	if t.ArtistID == "" && len(t.Artists) > 0 {
		t.ArtistID = t.Artists[0].ID
	}

	// Populate ShazamURL
	if t.ShazamURL == "" && t.Key != "" {
		t.ShazamURL = "https://www.shazam.com/track/" + t.Key
	}

	// Populate PhotoURL from images
	if t.PhotoURL == "" && t.Images != nil {
		if hq, ok := t.Images["coverarthq"]; ok && hq != "" {
			t.PhotoURL = hq
		} else if cover, ok := t.Images["coverart"]; ok {
			t.PhotoURL = cover
		}
	}

	// Extract Hub actions (Apple Music, Spotify, Ringtone)
	for _, opt := range t.Hub.Options {
		for _, act := range opt.Actions {
			if act.URI != nil && *act.URI != "" {
				if strings.Contains(*act.URI, "music.apple.com") && t.AppleMusicURL == "" {
					t.AppleMusicURL = cleanQueryURL(*act.URI)
				}
			}
		}
	}

	for _, act := range t.Hub.Actions {
		if act.URI != nil && *act.URI != "" {
			if strings.EqualFold(act.Name, "ringtone") || strings.EqualFold(act.Type, "ringtone") || strings.Contains(*act.URI, "m4r") {
				t.Ringtone = *act.URI
				break
			}
		}
	}
	if t.Ringtone == "" && len(t.Hub.Actions) > 1 && t.Hub.Actions[1].URI != nil {
		t.Ringtone = *t.Hub.Actions[1].URI
	}

	for _, prov := range t.Hub.Providers {
		for _, act := range prov.Actions {
			if act.URI != nil && *act.URI != "" {
				uri := *act.URI
				if strings.HasPrefix(uri, "spotify:search:") {
					t.SpotifyURI = uri
					t.SpotifyURIQuery = strings.TrimPrefix(uri, "spotify:search:")
				} else if strings.Contains(uri, "spotify.com") && t.SpotifyURL == "" {
					t.SpotifyURL = uri
				}
			}
		}
	}

	// Extract YouTube link from sections
	for _, sec := range t.Sections {
		if strings.EqualFold(sec.Type, "VIDEO") && sec.YouTubeURL != "" && t.YouTubeLink == "" {
			t.YouTubeLink = sec.YouTubeURL
		}
	}

	return nil
}

func cleanQueryURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	// Preserve Apple Music track identifier (?i=...) if present
	trackID := u.Query().Get("i")
	if trackID != "" {
		v := url.Values{}
		v.Set("i", trackID)
		u.RawQuery = v.Encode()
	} else {
		u.RawQuery = ""
	}
	return u.String()
}

// YoutubeData represents YouTube video details and action metadata.
type YoutubeData struct {
	Caption string        `json:"caption,omitempty"`
	Image   ImageModel    `json:"image,omitempty"`
	Actions []ActionModel `json:"actions,omitempty"`
	URI     string        `json:"uri,omitempty"`
}
