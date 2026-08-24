package models

import "strings"

// ArtistQuery holds views and extension options when querying artist profile.
type ArtistQuery struct {
	Views  []ArtistView   `json:"views,omitempty"`
	Extend []ArtistExtend `json:"extend,omitempty"`
}

// Params converts ArtistQuery into URL query parameters map.
func (q *ArtistQuery) Params() map[string]string {
	params := make(map[string]string)
	if q == nil {
		return params
	}

	if len(q.Extend) > 0 {
		var strExtends []string
		for _, e := range q.Extend {
			strExtends = append(strExtends, string(e))
		}
		params["extend"] = strings.Join(strExtends, ",")
	}

	if len(q.Views) > 0 {
		var strViews []string
		for _, v := range q.Views {
			strViews = append(strViews, string(v))
		}
		params["views"] = strings.Join(strViews, ",")
	}

	return params
}

// ArtistInfo represents artist profile details.
type ArtistInfo struct {
	Name          string   `json:"name,omitempty"`
	Verified      *bool    `json:"verified,omitempty"`
	Genres        []string `json:"genres,omitempty"`
	Alias         *string  `json:"alias,omitempty"`
	GenresPrimary *string  `json:"genres_primary,omitempty"`
	Avatar        *string  `json:"avatar,omitempty"`
	AdamID        *int64   `json:"adam_id,omitempty"`
	URL           string   `json:"url,omitempty"`
}

// ArtistV2 wraps ArtistInfo in an artist object.
type ArtistV2 struct {
	Artist ArtistInfo `json:"artist"`
}

// AlbumRelationship represents artist's album relationships.
type AlbumRelationship struct {
	Href string           `json:"href,omitempty"`
	Next *string          `json:"next,omitempty"`
	Data []BaseIdTypeHref `json:"data,omitempty"`
}

// ArtistRelationships represents artist relationship links.
type ArtistRelationships struct {
	Albums AlbumRelationship `json:"albums"`
}

// ArtistViews contains expanded views for an artist (albums, top songs, videos, similar artists).
type ArtistViews struct {
	TopMusicVideos *TopMusicVideosView `json:"top-music-videos,omitempty"`
	SimilarArtists *SimularArtist      `json:"similar-artists,omitempty"`
	LatestRelease  *LastReleaseModel   `json:"latest-release,omitempty"`
	FullAlbums     *FullAlbumsModel    `json:"full-albums,omitempty"`
	TopSongs       *TopSong            `json:"top-songs,omitempty"`
}

// ArtistAttribute contains artist metadata attributes.
type ArtistAttribute struct {
	GenreNames []string    `json:"genreNames,omitempty"`
	Name       string      `json:"name,omitempty"`
	URL        string      `json:"url,omitempty"`
	ArtistBio  *string     `json:"artistBio,omitempty"`
	Origin     *string     `json:"origin,omitempty"`
	Artwork    *ImageModel `json:"artwork,omitempty"`
}

// ArtistV3 represents the v3 artist schema.
type ArtistV3 struct {
	ID            string              `json:"id,omitempty"`
	Type          string              `json:"type,omitempty"`
	Attributes    ArtistAttribute     `json:"attributes"`
	Relationships ArtistRelationships `json:"relationships"`
	Views         ArtistViews         `json:"views"`
}

// ArtistResponse represents the top-level response for artist API queries.
type ArtistResponse struct {
	Errors []ErrorModel `json:"errors,omitempty"`
	Data   []ArtistV3   `json:"data,omitempty"`
}

// SimularArtistDatum represents a single similar artist entry.
type SimularArtistDatum struct {
	BaseIdTypeHref
	Attributes    SimilarArtistAttributes   `json:"attributes"`
	Relationships SimilarArtistRelationship `json:"relationships"`
}

// SimilarArtistAttributes represents attributes of a similar artist.
type SimilarArtistAttributes struct {
	GenreNames        []string          `json:"genreNames,omitempty"`
	EditorialArtwork  *EditorialArtwork `json:"editorialArtwork,omitempty"`
	Name              string            `json:"name,omitempty"`
	Artwork           ImageModel        `json:"artwork"`
	URL               string            `json:"url,omitempty"`
	Origin            *string           `json:"origin,omitempty"`
	ArtistBio         *string           `json:"artistBio,omitempty"`
}

// SimilarArtistRelationship represents albums relationship of a similar artist.
type SimilarArtistRelationship struct {
	Albums BaseHrefNextData[[]BaseIdTypeHref] `json:"albums"`
}

// SimularArtist represents similar artists view response.
type SimularArtist struct {
	Href       *string              `json:"href,omitempty"`
	Next       *string              `json:"next,omitempty"`
	Attributes *AttributeName       `json:"attributes,omitempty"`
	Data       []SimularArtistDatum `json:"data,omitempty"`
}

// TopMusicVideosView represents artist's top music videos view.
type TopMusicVideosView struct {
	Href       *string                                     `json:"href,omitempty"`
	Attributes *AttributeName                              `json:"attributes,omitempty"`
	Data       []BaseAttributesModel[MusicVideoAttributes] `json:"data,omitempty"`
}

// MusicVideoAttributes represents metadata for a music video.
type MusicVideoAttributes struct {
	GenreNames       []string       `json:"genreNames,omitempty"`
	ReleaseDate      string         `json:"releaseDate,omitempty"`
	DurationInMillis int            `json:"durationInMillis,omitempty"`
	ISRC             string         `json:"isrc,omitempty"`
	Artwork          ImageModel     `json:"artwork"`
	PlayParams       PlayParams     `json:"playParams"`
	URL              string         `json:"url,omitempty"`
	Has4K            bool           `json:"has4K,omitempty"`
	HasHDR           bool           `json:"hasHDR,omitempty"`
	Name             string         `json:"name,omitempty"`
	Previews         []VideoPreview `json:"previews,omitempty"`
	ArtistName       string         `json:"artistName,omitempty"`
	ContentRating    *string        `json:"contentRating,omitempty"`
	AlbumName        *string        `json:"albumName,omitempty"`
	TrackNumber      *int           `json:"trackNumber,omitempty"`
}

// VideoPreview represents video preview stream URLs.
type VideoPreview struct {
	URL     string     `json:"url,omitempty"`
	HLSURL  string     `json:"hlsUrl,omitempty"`
	Artwork ImageModel `json:"artwork"`
}

// TopSong represents top songs view.
type TopSong struct {
	BaseHrefNext
	Attributes *AttributeName                              `json:"attributes,omitempty"`
	Data       []BaseAttributesModel[AttributesTopSong]    `json:"data,omitempty"`
}

// AttributesTopSong represents attributes of a top song.
type AttributesTopSong struct {
	HasTimeSyncedLyrics        bool       `json:"hasTimeSyncedLyrics,omitempty"`
	AlbumName                  *string    `json:"albumName,omitempty"`
	GenreNames                 []string   `json:"genreNames,omitempty"`
	TrackNumber                int        `json:"trackNumber,omitempty"`
	ReleaseDate                string     `json:"releaseDate,omitempty"`
	DurationInMillis           int        `json:"durationInMillis,omitempty"`
	IsVocalAttenuationAllowed  bool       `json:"isVocalAttenuationAllowed,omitempty"`
	IsMasteredForItunes        bool       `json:"isMasteredForItunes,omitempty"`
	ISRC                       string     `json:"isrc,omitempty"`
	Artwork                    ImageModel `json:"artwork"`
	ComposerName               string     `json:"composerName,omitempty"`
	AudioLocale                string     `json:"audioLocale,omitempty"`
	URL                        string     `json:"url,omitempty"`
	PlayParams                 PlayParams `json:"playParams"`
	DiscNumber                 int        `json:"discNumber,omitempty"`
	HasLyrics                  bool       `json:"hasLyrics,omitempty"`
	IsAppleDigitalMaster       bool       `json:"isAppleDigitalMaster,omitempty"`
	AudioTraits                []string   `json:"audioTraits,omitempty"`
	Name                       string     `json:"name,omitempty"`
	Previews                   []UrlDTO   `json:"previews,omitempty"`
	ArtistName                 string     `json:"artistName,omitempty"`
	ContentRating              *string    `json:"contentRating,omitempty"`
}
