package models

// EditorialArtwork holds promotional artwork images.
type EditorialArtwork struct {
	SubscriptionHero               *ImageModel `json:"subscriptionHero,omitempty"`
	StoreFlowcase                  *ImageModel `json:"storeFlowcase,omitempty"`
	CenteredFullscreenBackground   *ImageModel `json:"centeredFullscreenBackground,omitempty"`
	BannerUber                     *ImageModel `json:"bannerUber,omitempty"`
}

// EditorialNotes holds editorial review notes.
type EditorialNotes struct {
	Standard *string `json:"standard,omitempty"`
	Short    *string `json:"short,omitempty"`
}

// AttributesFullAlbums represents album attributes in full albums view.
type AttributesFullAlbums struct {
	Copyright           string            `json:"copyright,omitempty"`
	GenreNames          []string          `json:"genreNames,omitempty"`
	ReleaseDate         string            `json:"releaseDate,omitempty"`
	IsMasteredForItunes bool              `json:"isMasteredForItunes,omitempty"`
	UPC                 string            `json:"upc,omitempty"`
	Artwork             ImageModel        `json:"artwork"`
	PlayParams          PlayParams        `json:"playParams"`
	URL                 string            `json:"url,omitempty"`
	RecordLabel         string            `json:"recordLabel,omitempty"`
	TrackCount          int               `json:"trackCount,omitempty"`
	IsCompilation       bool              `json:"isCompilation,omitempty"`
	IsPrerelease        bool              `json:"isPrerelease,omitempty"`
	AudioTraits         []string          `json:"audioTraits,omitempty"`
	EditorialArtwork    *EditorialArtwork `json:"editorialArtwork,omitempty"`
	IsSingle            bool              `json:"isSingle,omitempty"`
	Name                string            `json:"name,omitempty"`
	ArtistName          string            `json:"artistName,omitempty"`
	ContentRating       *string           `json:"contentRating,omitempty"`
	IsComplete          bool              `json:"isComplete,omitempty"`
	EditorialNotes      *EditorialNotes   `json:"editorialNotes,omitempty"`
}

// FullAlbumsModel represents full albums view response.
type FullAlbumsModel struct {
	Href       *string                                                  `json:"href,omitempty"`
	Attributes *AttributeName                                           `json:"attributes,omitempty"`
	Data       []BaseIdTypeHrefAttributesModel[AttributesFullAlbums]    `json:"data,omitempty"`
}

// SmallAlbumsModel is an alias for album collections.
type SmallAlbumsModel = FullAlbumsModel

// AttributeLastRelease represents attributes for an artist's latest release.
type AttributeLastRelease = AttributesFullAlbums

// LastReleaseModel represents latest release view response.
type LastReleaseModel struct {
	Href       *string                                            `json:"href,omitempty"`
	Attributes *AttributeName                                     `json:"attributes,omitempty"`
	Data       []BaseAttributesModel[AttributeLastRelease]        `json:"data,omitempty"`
}

// TrackInfoDTO extends AttributesTopSong with credits metadata.
type TrackInfoDTO struct {
	AttributesTopSong
	HasCredits *bool `json:"hasCredits,omitempty"`
}

// TrackInfoWithHref represents an album track with href and attributes.
type TrackInfoWithHref struct {
	BaseIdTypeHref
	Attributes TrackInfoDTO `json:"attributes"`
}

// TrackModel represents an album's tracks relationship.
type TrackModel struct {
	Href string              `json:"href,omitempty"`
	Data []TrackInfoWithHref `json:"data,omitempty"`
}

// ArtistModel represents an album's artists relationship.
type ArtistModel struct {
	BaseHref
	Data []BaseIdTypeHref `json:"data,omitempty"`
}

// AlbumRelationships represents relationships of an album (artists and tracks).
type AlbumRelationships struct {
	Artists ArtistModel `json:"artists"`
	Tracks  TrackModel  `json:"tracks"`
}

// AlbumModel represents detailed album information.
type AlbumModel struct {
	BaseIdTypeHref
	Attributes    AttributesFullAlbums `json:"attributes"`
	Relationships AlbumRelationships   `json:"relationships"`
}
