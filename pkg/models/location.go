package models

// LocationsResponse represents the response from Shazam charts locations endpoint.
type LocationsResponse struct {
	Countries []CountryLocation `json:"countries,omitempty"`
	Global    GlobalLocation    `json:"global,omitempty"`
}

// CountryLocation represents country metadata and chart playlists.
type CountryLocation struct {
	ID     string          `json:"id,omitempty"`
	ListID string          `json:"listid,omitempty"`
	Name   string          `json:"name,omitempty"`
	Cities []CityLocation  `json:"cities,omitempty"`
	Genres []GenreLocation `json:"genres,omitempty"`
}

// CityLocation represents city chart playlist metadata.
type CityLocation struct {
	ID        string `json:"id,omitempty"`
	ListID    string `json:"listid,omitempty"`
	Name      string `json:"name,omitempty"`
	CountryID string `json:"countryid,omitempty"`
}

// GenreLocation represents genre chart playlist metadata.
type GenreLocation struct {
	ID      string `json:"id,omitempty"`
	ListID  string `json:"listid,omitempty"`
	Name    string `json:"name,omitempty"`
	URLName string `json:"urlName,omitempty"`
}

// GlobalLocation represents world chart playlists.
type GlobalLocation struct {
	Top    TopLocation     `json:"top,omitempty"`
	Genres []GenreLocation `json:"genres,omitempty"`
}

// TopLocation represents the global top chart playlist ID.
type TopLocation struct {
	ListID string `json:"listid,omitempty"`
}
