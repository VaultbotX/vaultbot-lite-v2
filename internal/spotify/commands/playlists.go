package commands

import (
	"context"
	"errors"

	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	sp "github.com/vaultbotx/vaultbot-lite/internal/spotify"
	"github.com/zmb3/spotify/v2"
)

type SpotifyPlaylistRepo struct {
	Client   *sp.Client
	Playlist domain.Playlist
}

// GetPlaylistTracks gets all tracks from the dynamic playlist. It returns them as *spotify.PlaylistItems,
// which includes information about when the track was added to the playlist.
func (r *SpotifyPlaylistRepo) GetPlaylistTracks(ctx context.Context) ([]spotify.PlaylistItem, error) {
	r.Client.Mu.Lock()
	defer r.Client.Mu.Unlock()

	playlistId := r.getPlaylistId()
	if playlistId == "" {
		return nil, errors.New("invalid playlist type")
	}

	var playlistItems []spotify.PlaylistItem
	playlistItemsCollection, err := r.Client.Client.GetPlaylistItems(ctx, playlistId)
	if err != nil {
		return nil, err
	}

	for _, playlistItem := range playlistItemsCollection.Items {
		playlistItems = append(playlistItems, playlistItem)
	}

	for page := 1; ; page++ {
		err = r.Client.Client.NextPage(ctx, playlistItemsCollection)
		if errors.Is(err, spotify.ErrNoMorePages) {
			break
		}
		if err != nil {
			return nil, err
		}

		for _, playlistItem := range playlistItemsCollection.Items {
			playlistItems = append(playlistItems, playlistItem)
		}
	}

	return playlistItems, nil
}

func (r *SpotifyPlaylistRepo) AddTracksToPlaylist(ctx context.Context, trackIds []spotify.ID) error {
	r.Client.Mu.Lock()
	defer r.Client.Mu.Unlock()

	playlistId := r.getPlaylistId()
	if playlistId == "" {
		return errors.New("invalid playlist type")
	}
	_, err := r.Client.Client.AddTracksToPlaylist(ctx, playlistId, trackIds...)
	if err != nil {
		return err
	}

	return nil
}

func (r *SpotifyPlaylistRepo) RemoveTracksFromPlaylist(ctx context.Context, trackIds []spotify.ID) error {
	r.Client.Mu.Lock()
	defer r.Client.Mu.Unlock()

	playlistId := r.getPlaylistId()
	if playlistId == "" {
		return errors.New("invalid playlist type")
	}
	_, err := r.Client.Client.RemoveTracksFromPlaylist(ctx, playlistId, trackIds...)
	if err != nil {
		return err
	}

	return nil
}

func (r *SpotifyPlaylistRepo) UpdatePlaylistDescription(ctx context.Context, description string) error {
	r.Client.Mu.Lock()
	defer r.Client.Mu.Unlock()

	playlistId := r.getPlaylistId()
	if playlistId == "" {
		return errors.New("invalid playlist type")
	}
	return r.Client.Client.ChangePlaylistDescription(ctx, playlistId, description)
}

func (r *SpotifyPlaylistRepo) getPlaylistId() spotify.ID {
	switch r.Playlist {
	case domain.DynamicPlaylist:
		return r.Client.DynamicPlaylistId
	case domain.GenrePlaylist:
		return r.Client.GenrePlaylistId
	case domain.HighScoresPlaylist:
		return r.Client.HighScoresPlaylistId
	default:
		return ""
	}
}
