package client

import (
	"fmt"
	"net/url"
)

const (
	// SearchFromFileURL is the endpoint for audio fingerprint recognition.
	SearchFromFileURL = "https://amp.shazam.com/discovery/v5/%s/%s/%s/-/tag/%s/%s?sync=true&webv3=true&sampling=true&connected=&shazamapiversion=v3&sharehub=true&hubv5minorversion=v5.1&hidelb=true&video=v3"

	// AboutTrackURL is the endpoint for getting track details.
	AboutTrackURL = "https://www.shazam.com/discovery/v5/%s/%s/web/-/track/%d?shazamapiversion=v3&video=v3"

	// TopTracksPlaylistURL is the endpoint for playlist tracks.
	TopTracksPlaylistURL = "https://www.shazam.com/services/amapi/v1/catalog/%s/playlists/%s/tracks?limit=%d&offset=%d&l=%s&relate[songs]=artists,music-videos"

	// LocationsURL is the endpoint for chart locations metadata.
	LocationsURL = "https://www.shazam.com/services/charts/locations"

	// RelatedSongsURL is the endpoint for similar/related tracks.
	RelatedSongsURL = "https://cdn.shazam.com/shazam/v3/%s/%s/web/-/tracks/track-similarities-id-%d?startFrom=%d&pageSize=%d&connected=&channel="

	// SearchArtistURL is the endpoint for artist search.
	SearchArtistURL = "https://www.shazam.com/services/search/v4/%s/%s/web/search?term=%s&limit=%d&offset=%d&types=artists"

	// SearchMusicURL is the endpoint for track search.
	SearchMusicURL = "https://www.shazam.com/services/search/v3/%s/%s/web/search?query=%s&numResults=%d&offset=%d&types=songs"

	// ListeningCounterURL is the endpoint for single track listen count.
	ListeningCounterURL = "https://www.shazam.com/services/count/v2/web/track/%d"

	// ListeningCounterManyURL is the endpoint for multi-track listen counts.
	ListeningCounterManyURL = "https://www.shazam.com/services/count/v2/web/track"

	// SearchArtistV2URL is the endpoint for v2 artist profile catalog.
	SearchArtistV2URL = "https://www.shazam.com/services/amapi/v1/catalog/%s/artists/%d"

	// ArtistAlbumsURL is the endpoint for artist albums catalog.
	ArtistAlbumsURL = "https://www.shazam.com/services/amapi/v1/catalog/%s/artists/%d/albums?limit=%d&offset=%d"

	// ArtistAlbumInfoURL is the endpoint for album info catalog.
	ArtistAlbumInfoURL = "https://www.shazam.com/services/amapi/v1/catalog/%s/albums/%d"
)

// FormatSearchFromFile formats the audio recognition URL.
func FormatSearchFromFile(language, endpointCountry, device, uuid1, uuid2 string) string {
	return fmt.Sprintf(SearchFromFileURL, language, endpointCountry, device, uuid1, uuid2)
}

// FormatAboutTrack formats the track information URL.
func FormatAboutTrack(language, endpointCountry string, trackID int64) string {
	return fmt.Sprintf(AboutTrackURL, language, endpointCountry, trackID)
}

// FormatTopTracksPlaylist formats the playlist tracks URL.
func FormatTopTracksPlaylist(endpointCountry, playlistID string, limit, offset int, language string) string {
	return fmt.Sprintf(TopTracksPlaylistURL, endpointCountry, playlistID, limit, offset, language)
}

// FormatRelatedSongs formats the similar tracks URL.
func FormatRelatedSongs(language, endpointCountry string, trackID int64, limit, offset int) string {
	return fmt.Sprintf(RelatedSongsURL, language, endpointCountry, trackID, offset, limit)
}

// FormatSearchArtist formats the artist search URL.
func FormatSearchArtist(language, endpointCountry, query string, limit, offset int) string {
	return fmt.Sprintf(SearchArtistURL, language, endpointCountry, url.QueryEscape(query), limit, offset)
}

// FormatSearchMusic formats the music search URL.
func FormatSearchMusic(language, endpointCountry, query string, limit, offset int) string {
	return fmt.Sprintf(SearchMusicURL, language, endpointCountry, url.QueryEscape(query), limit, offset)
}

// FormatArtistAlbums formats the artist albums URL.
func FormatArtistAlbums(endpointCountry string, artistID int64, limit, offset int) string {
	return fmt.Sprintf(ArtistAlbumsURL, endpointCountry, artistID, limit, offset)
}

// FormatArtistAlbumInfo formats the album info URL.
func FormatArtistAlbumInfo(endpointCountry string, albumID int64) string {
	return fmt.Sprintf(ArtistAlbumInfoURL, endpointCountry, albumID)
}
