package models

// PlayListRelationshipDTO represents playlist associations to music videos and artists.
type PlayListRelationshipDTO struct {
	MusicVideos BaseHrefData[[]BaseIdTypeHref] `json:"music-videos,omitempty"`
	Artists     BaseHrefData[[]BaseIdTypeHref] `json:"artists,omitempty"`
}

// PlayListAttributes contains track attributes in a playlist.
type PlayListAttributes struct {
	HasTimeSyncedLyrics       bool        `json:"hasTimeSyncedLyrics,omitempty"`
	AlbumName                 string      `json:"albumName,omitempty"`
	GenreNames                []string    `json:"genreNames,omitempty"`
	TrackNumber               int         `json:"trackNumber,omitempty"`
	ReleaseDate               *string     `json:"releaseDate,omitempty"`
	DurationInMillis          *int        `json:"durationInMillis,omitempty"`
	IsVocalAttenuationAllowed bool        `json:"isVocalAttenuationAllowed,omitempty"`
	IsMasteredForItunes       bool        `json:"isMasteredForItunes,omitempty"`
	ISRC                      string      `json:"isrc,omitempty"`
	Artwork                   ImageModel  `json:"artwork"`
	AudioLocale               string      `json:"audioLocale,omitempty"`
	URL                       string      `json:"url,omitempty"`
	PlayParams                *PlayParams `json:"playParams,omitempty"`
	DiscNumber                int         `json:"discNumber,omitempty"`
	HasCredits                *bool       `json:"hasCredits,omitempty"`
	IsAppleDigitalMaster      bool        `json:"isAppleDigitalMaster,omitempty"`
	HasLyrics                 bool        `json:"hasLyrics,omitempty"`
	AudioTraits               []string    `json:"audioTraits,omitempty"`
	Name                      string      `json:"name,omitempty"`
	Previews                  []UrlDTO    `json:"previews,omitempty"`
	ContentRating             *string     `json:"contentRating,omitempty"`
	ArtistName                string      `json:"artistName,omitempty"`
}

// PlayList represents an individual song item in a Shazam playlist.
type PlayList struct {
	BaseIdTypeHref
	Attributes    PlayListAttributes      `json:"attributes"`
	Relationships PlayListRelationshipDTO `json:"relationships"`
	Meta          *MetaContentVersion     `json:"meta,omitempty"`
}

// PlaylistResponse represents a collection of playlist tracks.
type PlaylistResponse struct {
	Data []PlayList `json:"data,omitempty"`
}
