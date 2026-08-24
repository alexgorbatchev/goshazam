package goshazam

import (
	"encoding/json"
	"fmt"

	"github.com/alexgorbatchev/goshazam/pkg/models"
)

// Serialize provides helper functions to decode raw JSON bytes or maps into typed models.
type Serialize struct{}

// Track decodes raw JSON into a TrackInfo model.
func (Serialize) Track(data []byte) (*models.TrackInfo, error) {
	var track models.TrackInfo
	if err := json.Unmarshal(data, &track); err != nil {
		return nil, fmt.Errorf("decoding track info: %w", err)
	}
	return &track, nil
}

// FullTrack decodes raw JSON into a full ResponseTrack model.
func (Serialize) FullTrack(data []byte) (*models.ResponseTrack, error) {
	var resp models.ResponseTrack
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("decoding full track response: %w", err)
	}
	return &resp, nil
}

// Playlist decodes raw JSON into an individual PlayList item.
func (Serialize) Playlist(data []byte) (*models.PlayList, error) {
	var pl models.PlayList
	if err := json.Unmarshal(data, &pl); err != nil {
		return nil, fmt.Errorf("decoding playlist item: %w", err)
	}
	return &pl, nil
}

// Playlists decodes raw JSON into a PlaylistResponse.
func (Serialize) Playlists(data []byte) (*models.PlaylistResponse, error) {
	var resp models.PlaylistResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("decoding playlists response: %w", err)
	}
	return &resp, nil
}

// ArtistV2 decodes raw JSON into an ArtistResponse.
func (Serialize) ArtistV2(data []byte) (*models.ArtistResponse, error) {
	var resp models.ArtistResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("decoding artist response: %w", err)
	}
	return &resp, nil
}

// ArtistAlbums decodes raw JSON into FullAlbumsModel.
func (Serialize) ArtistAlbums(data []byte) (*models.FullAlbumsModel, error) {
	var resp models.FullAlbumsModel
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("decoding artist albums: %w", err)
	}
	return &resp, nil
}

// AlbumInfo decodes raw JSON into BaseDataModel[[]AlbumModel].
func (Serialize) AlbumInfo(data []byte) (*models.BaseDataModel[[]models.AlbumModel], error) {
	var resp models.BaseDataModel[[]models.AlbumModel]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("decoding album info: %w", err)
	}
	return &resp, nil
}

// YouTube decodes raw JSON into YoutubeData.
func (Serialize) YouTube(data []byte) (*models.YoutubeData, error) {
	var yt models.YoutubeData
	if err := json.Unmarshal(data, &yt); err != nil {
		return nil, fmt.Errorf("decoding youtube data: %w", err)
	}
	return &yt, nil
}
