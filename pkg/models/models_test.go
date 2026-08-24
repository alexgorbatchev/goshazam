package models

import (
	"encoding/json"
	"testing"
)

func TestTrackInfoUnmarshaling(t *testing.T) {
	rawJSON := `{
		"matches": [
			{
				"id": "230272433",
				"offset": 187.4215,
				"timeskew": -0.0001565814,
				"frequencyskew": -0.000080525875
			}
		],
		"location": {"accuracy": 0.01},
		"timestamp": 1652380596486,
		"timezone": "Europe/Moscow",
		"track": {
			"layout": "5",
			"type": "MUSIC",
			"key": "47440537",
			"title": "Arrival To Earth",
			"subtitle": "Steve Jablonsky",
			"images": {
				"coverart": "https://is2-ssl.mzstatic.com/cover.jpg",
				"coverarthq": "https://is2-ssl.mzstatic.com/coverhq.jpg"
			},
			"artists": [{"id": "10194644", "adamid": "21402948"}],
			"sections": [
				{
					"type": "SONG",
					"tabname": "Song",
					"metadata": [
						{"title": "Album", "text": "Transformers"}
					]
				},
				{
					"type": "VIDEO",
					"tabname": "Video",
					"youtubeurl": "https://cdn.shazam.com/video/youtube"
				}
			],
			"hub": {
				"type": "APPLEMUSIC",
				"providers": [
					{
						"actions": [
							{
								"name": "hub:spotify:searchdeeplink",
								"type": "uri",
								"uri": "spotify:search:Arrival%20To%20Earth%20Steve%20Jablonsky"
							}
						],
						"type": "SPOTIFY"
					}
				]
			}
		},
		"tagid": "89A4C33B-58C6-4A50-8475-94032FC34D06"
	}`

	var resp ResponseTrack
	if err := json.Unmarshal([]byte(rawJSON), &resp); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if resp.TagID != "89A4C33B-58C6-4A50-8475-94032FC34D06" {
		t.Errorf("expected TagID match, got %s", resp.TagID)
	}
	if len(resp.Matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(resp.Matches))
	}
	if resp.Matches[0].Channel != nil {
		t.Errorf("expected nil channel, got %v", *resp.Matches[0].Channel)
	}

	track := resp.Track
	if track == nil {
		t.Fatalf("expected non-nil track")
	}
	if track.Title != "Arrival To Earth" {
		t.Errorf("expected title Arrival To Earth, got %s", track.Title)
	}
	if track.ArtistID != "10194644" {
		t.Errorf("expected artist ID 10194644, got %s", track.ArtistID)
	}
	if track.PhotoURL != "https://is2-ssl.mzstatic.com/coverhq.jpg" {
		t.Errorf("expected photo URL coverhq, got %s", track.PhotoURL)
	}
	if track.SpotifyURIQuery != "Arrival%20To%20Earth%20Steve%20Jablonsky" {
		t.Errorf("expected spotify query Arrival%%20To%%20Earth%%20Steve%%20Jablonsky, got %s", track.SpotifyURIQuery)
	}
	if track.YouTubeLink != "https://cdn.shazam.com/video/youtube" {
		t.Errorf("expected youtube link https://cdn.shazam.com/video/youtube, got %s", track.YouTubeLink)
	}
}

func TestArtistQueryParams(t *testing.T) {
	q := &ArtistQuery{
		Views: []ArtistView{
			ArtistViewFullAlbums,
			ArtistViewLatestRelease,
		},
		Extend: []ArtistExtend{
			ArtistExtendBio,
			ArtistExtendBornOrFormed,
		},
	}

	params := q.Params()
	if params["views"] != "full-albums,latest-release" {
		t.Errorf("expected views full-albums,latest-release, got %s", params["views"])
	}
	if params["extend"] != "artistBio,bornOrFormed" {
		t.Errorf("expected extend artistBio,bornOrFormed, got %s", params["extend"])
	}
}
