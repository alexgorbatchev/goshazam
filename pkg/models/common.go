package models

// BaseHref represents an object with an API href URL.
type BaseHref struct {
	Href string `json:"href,omitempty"`
}

// BaseIdTypeHref represents a resource with ID, Type, and Href.
type BaseIdTypeHref struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
	Href string `json:"href,omitempty"`
}

// BaseHrefNext represents a paginated list with an optional next URL.
type BaseHrefNext struct {
	Href string  `json:"href,omitempty"`
	Next *string `json:"next,omitempty"`
}

// BaseDataModel is a generic wrapper around a data payload.
type BaseDataModel[T any] struct {
	Data T `json:"data"`
}

// BaseAttributesModel is a generic wrapper around attributes.
type BaseAttributesModel[T any] struct {
	Attributes T `json:"attributes"`
}

// BaseIdTypeHrefAttributesModel is a generic resource with ID, Type, Href, and typed Attributes.
type BaseIdTypeHrefAttributesModel[T any] struct {
	ID         string `json:"id,omitempty"`
	Type       string `json:"type,omitempty"`
	Href       string `json:"href,omitempty"`
	Attributes T      `json:"attributes"`
}

// BaseHrefData represents an object containing an href and data array.
type BaseHrefData[T any] struct {
	Href string `json:"href,omitempty"`
	Data T      `json:"data"`
}

// BaseHrefNextData represents a paginated collection containing href, next, and data.
type BaseHrefNextData[T any] struct {
	Href string  `json:"href,omitempty"`
	Next *string `json:"next,omitempty"`
	Data T       `json:"data"`
}

// ImageModel represents image metadata and URLs.
type ImageModel struct {
	Width      int     `json:"width,omitempty"`
	Height     int     `json:"height,omitempty"`
	URL        string  `json:"url,omitempty"`
	TextColor1 *string `json:"textColor1,omitempty"`
	TextColor2 *string `json:"textColor2,omitempty"`
	TextColor3 *string `json:"textColor3,omitempty"`
	TextColor4 *string `json:"textColor4,omitempty"`
	BgColor    *string `json:"bgColor,omitempty"`
	HasP3      bool    `json:"hasP3,omitempty"`
}

// PlayParams holds Apple Music playback parameter IDs.
type PlayParams struct {
	ID   string `json:"id,omitempty"`
	Kind string `json:"kind,omitempty"`
}

// UrlDTO represents a simple URL container.
type UrlDTO struct {
	URL string `json:"url,omitempty"`
}

// ContentVersion represents backend content versioning information.
type ContentVersion struct {
	RTCI      int `json:"RTCI,omitempty"`
	MZIndexer int `json:"MZ_INDEXER,omitempty"`
}

// MetaContentVersion wraps ContentVersion in a metadata container.
type MetaContentVersion struct {
	ContentVersion ContentVersion `json:"contentVersion,omitempty"`
}

// AttributeName represents a title attribute container.
type AttributeName struct {
	Title string `json:"title,omitempty"`
}

// ErrorModel represents an API error entry.
type ErrorModel struct {
	ID     string `json:"id,omitempty"`
	Title  string `json:"title,omitempty"`
	Detail string `json:"detail,omitempty"`
	Status string `json:"status,omitempty"`
	Code   string `json:"code,omitempty"`
}
