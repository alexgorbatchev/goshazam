package geo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexgorbatchev/goshazam/pkg/client"
	"github.com/alexgorbatchev/goshazam/pkg/models"
)

const sampleLocationsJSON = `{
	"countries": [
		{
			"id": "US",
			"listid": "pl.us-top-200",
			"name": "United States",
			"cities": [
				{
					"id": "1",
					"listid": "pl.us-nyc-50",
					"name": "New York"
				}
			],
			"genres": [
				{
					"id": "1",
					"listid": "pl.us-pop",
					"name": "Pop",
					"urlName": "pop"
				}
			]
		}
	],
	"global": {
		"top": {
			"listid": "pl.global-top-200"
		},
		"genres": [
			{
				"id": "1",
				"listid": "pl.global-rock",
				"name": "Rock",
				"urlName": "rock"
			}
		]
	}
}`

func TestGeoService(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sampleLocationsJSON))
	}))
	defer server.Close()

	c := client.NewHTTPClient(client.WithHTTPClient(server.Client()))
	geo := NewGeoService(c)
	geo.SetLocationsURL(server.URL)
	ctx := context.Background()

	// Test GetTop
	topID, err := geo.GetTop(ctx)
	if err != nil || topID != "pl.global-top-200" {
		t.Fatalf("GetTop expected pl.global-top-200, got %s, err: %v", topID, err)
	}

	// Test GetCountryPlaylist
	countryID, err := geo.GetCountryPlaylist(ctx, "US")
	if err != nil || countryID != "pl.us-top-200" {
		t.Fatalf("GetCountryPlaylist expected pl.us-top-200, got %s, err: %v", countryID, err)
	}

	// Test GetCityPlaylist
	cityID, err := geo.GetCityPlaylist(ctx, "US", "New York")
	if err != nil || cityID != "pl.us-nyc-50" {
		t.Fatalf("GetCityPlaylist expected pl.us-nyc-50, got %s, err: %v", cityID, err)
	}

	// Test GetGenre
	genreID, err := geo.GetGenre(ctx, models.GenreRock)
	if err != nil || genreID != "pl.global-rock" {
		t.Fatalf("GetGenre expected pl.global-rock, got %s, err: %v", genreID, err)
	}

	// Test GetGenreFromCountry
	countryGenreID, err := geo.GetGenreFromCountry(ctx, "US", models.GenrePop)
	if err != nil || countryGenreID != "pl.us-pop" {
		t.Fatalf("GetGenreFromCountry expected pl.us-pop, got %s, err: %v", countryGenreID, err)
	}

	// Test NotFound errors
	if _, err := geo.GetCountryPlaylist(ctx, "ZZ"); err == nil {
		t.Errorf("expected error for non-existent country")
	}
	if _, err := geo.GetCityPlaylist(ctx, "US", "Atlantis"); err == nil {
		t.Errorf("expected error for non-existent city")
	}
	if _, err := geo.GetCityPlaylist(ctx, "ZZ", "Atlantis"); err == nil {
		t.Errorf("expected error for non-existent country in GetCityPlaylist")
	}
	if _, err := geo.GetGenre(ctx, models.GenreMusic("nonexistent")); err == nil {
		t.Errorf("expected error for non-existent genre")
	}
	if _, err := geo.GetGenreFromCountry(ctx, "ZZ", models.GenrePop); err == nil {
		t.Errorf("expected error for non-existent country in GetGenreFromCountry")
	}
	if _, err := geo.GetGenreFromCountry(ctx, "US", models.GenreMusic("nonexistent")); err == nil {
		t.Errorf("expected error for non-existent genre in country")
	}

	// Test missing top in global
	emptyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"countries": [], "global": {}}`))
	}))
	defer emptyServer.Close()

	geoEmpty := NewGeoService(client.NewHTTPClient(client.WithHTTPClient(emptyServer.Client())))
	geoEmpty.SetLocationsURL(emptyServer.URL)
	if _, err := geoEmpty.GetTop(ctx); err == nil {
		t.Errorf("expected error when top playlist id is missing")
	}

	// Test FetchLocations error
	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer errorServer.Close()

	geoError := NewGeoService(client.NewHTTPClient(
		client.WithHTTPClient(errorServer.Client()),
		client.WithRetryConfig(client.RetryConfig{MaxAttempts: 1}),
	))
	geoError.SetLocationsURL(errorServer.URL)
	if _, err := geoError.GetTop(ctx); err == nil {
		t.Errorf("expected error on failed network fetch in GetTop")
	}
	if _, err := geoError.GetCountryPlaylist(ctx, "US"); err == nil {
		t.Errorf("expected error on failed network fetch in GetCountryPlaylist")
	}
	if _, err := geoError.GetCityPlaylist(ctx, "US", "NYC"); err == nil {
		t.Errorf("expected error on failed network fetch in GetCityPlaylist")
	}
	if _, err := geoError.GetGenre(ctx, models.GenreRock); err == nil {
		t.Errorf("expected error on failed network fetch in GetGenre")
	}
	if _, err := geoError.GetGenreFromCountry(ctx, "US", models.GenrePop); err == nil {
		t.Errorf("expected error on failed network fetch in GetGenreFromCountry")
	}
}
