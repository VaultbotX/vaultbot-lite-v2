package main

import (
	"context"
	"time"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/domain"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence/postgres"
	"github.com/vaultbotx/vaultbot-lite/internal/spotify"
	sp "github.com/vaultbotx/vaultbot-lite/internal/spotify/commands"
	zspotify "github.com/zmb3/spotify/v2"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	_ = godotenv.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	spClient, err := spotify.NewSpotifyClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create Spotify client: %v", err)
	}

	playlistService := domain.NewSpotifyPlaylistService(&sp.SpotifyPlaylistRepo{
		Client:   spClient,
		Playlist: domain.DynamicPlaylist,
	})

	trackRepo := &sp.SpotifyTrackRepo{Client: spClient}
	artistRepo := &sp.SpotifyArtistRepo{Client: spClient}

	pgConn, err := postgres.NewPostgresConnection()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	dbRepo := persistence.NewPostgresTrackRepository(pgConn)

	items, err := playlistService.Repo.GetPlaylistTracks(ctx)
	if err != nil {
		log.Fatalf("Failed to get playlist tracks: %v", err)
	}

	log.Infof("Found %d tracks in playlist", len(items))

	newCount := 0
	for _, item := range items {
		if item.Track.Track == nil || item.Track.Track.ID == "" {
			continue
		}

		addedAt, err := time.Parse(zspotify.TimestampLayout, item.AddedAt)
		if err != nil {
			log.Errorf("Failed to parse added_at for track %s: %v", item.Track.Track.ID, err)
			continue
		}

		exists, err := dbRepo.HasRecentArchiveEntry(ctx, item.Track.Track.ID.String(), addedAt)
		if err != nil {
			log.Errorf("Failed to check archive for track %s: %v", item.Track.Track.ID, err)
			continue
		}
		if exists {
			continue
		}

		trackChan := make(chan *zspotify.FullTrack, 1)
		if err := trackRepo.GetTrack(item.Track.Track.ID, trackChan, ctx); err != nil {
			log.Errorf("Failed to get track %s: %v", item.Track.Track.ID, err)
			continue
		}
		fullTrack := <-trackChan

		artistIds := make([]zspotify.ID, len(fullTrack.Artists))
		for i, a := range fullTrack.Artists {
			artistIds[i] = a.ID
		}

		artistChan := make(chan *zspotify.FullArtist, len(artistIds))
		if err := artistRepo.GetArtists(artistIds, artistChan, ctx); err != nil {
			log.Errorf("Failed to get artists for track %s: %v", item.Track.Track.ID, err)
			continue
		}
		var artists []*zspotify.FullArtist
		for a := range artistChan {
			artists = append(artists, a)
		}

		if err := dbRepo.AddTrackToDatabase(fullTrack, artists); err != nil {
			log.Errorf("Failed to record track %s: %v", item.Track.Track.ID, err)
			continue
		}

		log.Infof("Recorded new track: %s", fullTrack.Name)
		newCount++
	}

	log.Infof("Poll complete. Recorded %d new tracks", newCount)
}
