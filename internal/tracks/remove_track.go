package tracks

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	"github.com/zmb3/spotify/v2"
)

func RemoveTracks(ctx context.Context, trackIds []spotify.ID, spotifyPlaylistService *domain.SpotifyPlaylistService) error {
	log.Debug("Removing ", len(trackIds), " tracks from playlist")
	return spotifyPlaylistService.Repo.RemoveTracksFromPlaylist(ctx, trackIds)
}
