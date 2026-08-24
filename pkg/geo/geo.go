package geo

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/alexgorbatchev/goshazam/pkg/client"
	"github.com/alexgorbatchev/goshazam/pkg/models"
)

// GeoService resolves Shazam chart playlist IDs for countries, cities, and genres.
type GeoService struct {
	client       *client.HTTPClient
	locationsURL string
	cacheLock    sync.RWMutex
	cached       *models.LocationsResponse
	cachedAt     time.Time
	cacheTTL     time.Duration
}

// NewGeoService creates a new GeoService with the given HTTPClient.
func NewGeoService(httpClient *client.HTTPClient) *GeoService {
	return &GeoService{
		client:       httpClient,
		locationsURL: client.LocationsURL,
		cacheTTL:     1 * time.Hour,
	}
}

// SetLocationsURL overrides the default locations endpoint URL (useful for testing).
func (g *GeoService) SetLocationsURL(targetURL string) {
	g.cacheLock.Lock()
	defer g.cacheLock.Unlock()
	g.locationsURL = targetURL
	g.cached = nil
}

// FetchLocations loads and caches locations metadata from Shazam's charts API.
func (g *GeoService) FetchLocations(ctx context.Context) (*models.LocationsResponse, error) {
	g.cacheLock.RLock()
	if g.cached != nil && time.Since(g.cachedAt) < g.cacheTTL {
		cached := g.cached
		g.cacheLock.RUnlock()
		return cached, nil
	}
	g.cacheLock.RUnlock()

	g.cacheLock.Lock()
	defer g.cacheLock.Unlock()

	// Double-check after acquiring write lock
	if g.cached != nil && time.Since(g.cachedAt) < g.cacheTTL {
		return g.cached, nil
	}

	targetURL := g.locationsURL
	if targetURL == "" {
		targetURL = client.LocationsURL
	}

	var resp models.LocationsResponse
	err := g.client.RequestJSON(ctx, "GET", targetURL, nil, nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("fetching shazam locations: %w", err)
	}

	g.cached = &resp
	g.cachedAt = time.Now()
	return g.cached, nil
}

// GetTop returns the global top chart playlist ID.
func (g *GeoService) GetTop(ctx context.Context) (string, error) {
	loc, err := g.FetchLocations(ctx)
	if err != nil {
		return "", err
	}

	if loc.Global.Top.ListID == "" {
		return "", fmt.Errorf("%w: top playlist id not found in locations", models.ErrBadParseData)
	}

	return loc.Global.Top.ListID, nil
}

// GetCountryPlaylist returns the playlist ID for a given country code (ISO 3166-1 alpha-2, e.g. "US", "GB", "NL").
func (g *GeoService) GetCountryPlaylist(ctx context.Context, countryCode string) (string, error) {
	loc, err := g.FetchLocations(ctx)
	if err != nil {
		return "", err
	}

	countryCode = strings.ToUpper(countryCode)
	for _, c := range loc.Countries {
		if strings.EqualFold(c.ID, countryCode) {
			if c.ListID != "" {
				return c.ListID, nil
			}
		}
	}

	return "", fmt.Errorf("%w: country code %q", models.ErrBadCountryName, countryCode)
}

// GetCityPlaylist returns the playlist ID for a specific city within a country.
func (g *GeoService) GetCityPlaylist(ctx context.Context, countryCode, cityName string) (string, error) {
	loc, err := g.FetchLocations(ctx)
	if err != nil {
		return "", err
	}

	countryCode = strings.ToUpper(countryCode)
	for _, c := range loc.Countries {
		if strings.EqualFold(c.ID, countryCode) {
			for _, city := range c.Cities {
				if strings.EqualFold(city.Name, cityName) {
					if city.ListID != "" {
						return city.ListID, nil
					}
				}
			}
			return "", fmt.Errorf("%w: city %q in country %q", models.ErrBadCityName, cityName, countryCode)
		}
	}

	return "", fmt.Errorf("%w: country code %q", models.ErrBadCountryName, countryCode)
}

// GetGenre returns the global playlist ID for a specific music genre.
func (g *GeoService) GetGenre(ctx context.Context, genre models.GenreMusic) (string, error) {
	loc, err := g.FetchLocations(ctx)
	if err != nil {
		return "", err
	}

	for _, gLoc := range loc.Global.Genres {
		if strings.EqualFold(gLoc.URLName, string(genre)) || strings.EqualFold(gLoc.Name, string(genre)) {
			if gLoc.ListID != "" {
				return gLoc.ListID, nil
			}
		}
	}

	return "", fmt.Errorf("%w: genre %q", models.ErrBadGenreName, genre)
}

// GetGenreFromCountry returns the country-specific playlist ID for a music genre.
func (g *GeoService) GetGenreFromCountry(ctx context.Context, countryCode string, genre models.GenreMusic) (string, error) {
	loc, err := g.FetchLocations(ctx)
	if err != nil {
		return "", err
	}

	countryCode = strings.ToUpper(countryCode)
	for _, c := range loc.Countries {
		if strings.EqualFold(c.ID, countryCode) {
			for _, gLoc := range c.Genres {
				if strings.EqualFold(gLoc.URLName, string(genre)) || strings.EqualFold(gLoc.Name, string(genre)) {
					if gLoc.ListID != "" {
						return gLoc.ListID, nil
					}
				}
			}
			return "", fmt.Errorf("%w: genre %q in country %q", models.ErrBadGenreName, genre, countryCode)
		}
	}

	return "", fmt.Errorf("%w: country code %q", models.ErrBadCountryName, countryCode)
}
