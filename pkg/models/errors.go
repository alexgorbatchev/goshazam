package models

import (
	"errors"
	"fmt"
)

var (
	// ErrBadCityName is returned when a city is not found in Shazam locations.
	ErrBadCityName = errors.New("city not found, check city name")

	// ErrBadCountryName is returned when a country is not found in Shazam locations.
	ErrBadCountryName = errors.New("country not found, check country code")

	// ErrBadGenreName is returned when a genre is not found.
	ErrBadGenreName = errors.New("genre not found, check genre name")

	// ErrBadParseData is returned when expected keys are missing from location response.
	ErrBadParseData = errors.New("failed to parse location data")

	// ErrFailedDecodeJSON is returned when JSON response decoding fails.
	ErrFailedDecodeJSON = errors.New("failed to decode JSON response")

	// ErrBadMethod is returned when an unsupported HTTP method is provided.
	ErrBadMethod = errors.New("unsupported HTTP method (accepts GET/POST)")

	// ErrNoMatches is returned when audio recognition produces no matches.
	ErrNoMatches = errors.New("no matching track found for audio sample")
)

// APIError represents an error response returned by Shazam's API.
type APIError struct {
	StatusCode int
	Status     string
	Message    string
	Body       string
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("shazam api error (status %d): %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("shazam api error (status %d): %s", e.StatusCode, e.Body)
}
