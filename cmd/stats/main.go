package main

import (
	"encoding/json"
	"os"
	"time"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	postgres "github.com/vaultbotx/vaultbot-lite/internal/persistence/postgres"
)

type MonthlyCount struct {
	Month string `json:"month" db:"month"`
	Count int    `json:"count" db:"count"`
}

type ArtistCount struct {
	Name      string `json:"name" db:"name"`
	SongCount int    `json:"song_count" db:"song_count"`
}

type GenreCount struct {
	Name      string `json:"name" db:"name"`
	SongCount int    `json:"song_count" db:"song_count"`
}

type Summary struct {
	TotalSongs          int `json:"total_songs"`
	TotalArchiveEntries int `json:"total_archive_entries"`
	TotalArtists        int `json:"total_artists"`
	TotalGenres         int `json:"total_genres"`
}

type Stats struct {
	GeneratedAt       time.Time      `json:"generated_at"`
	Summary           Summary        `json:"summary"`
	SongsOverTime     []MonthlyCount `json:"songs_over_time"`
	TopArtists        []ArtistCount  `json:"top_artists"`
	GenreDistribution []GenreCount   `json:"genre_distribution"`
}

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	// Write logs to stderr so they don't pollute the JSON written to stdout
	log.SetOutput(os.Stderr)
	_ = godotenv.Load()

	db, err := postgres.NewPostgresConnection()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	var stats Stats
	stats.GeneratedAt = time.Now().UTC()

	// Summary counts
	err = db.QueryRowx(`
		SELECT
			(SELECT COUNT(*) FROM songs)        AS total_songs,
			(SELECT COUNT(*) FROM song_archive) AS total_archive_entries,
			(SELECT COUNT(*) FROM artists)      AS total_artists,
			(SELECT COUNT(*) FROM genres)       AS total_genres
	`).Scan(
		&stats.Summary.TotalSongs,
		&stats.Summary.TotalArchiveEntries,
		&stats.Summary.TotalArtists,
		&stats.Summary.TotalGenres,
	)
	if err != nil {
		log.Fatalf("Failed to query summary: %v", err)
	}

	// Archive entries bucketed by month
	err = db.Select(&stats.SongsOverTime, `
		SELECT
			TO_CHAR(DATE_TRUNC('month', created_at), 'YYYY-MM') AS month,
			COUNT(*)                                             AS count
		FROM song_archive
		GROUP BY DATE_TRUNC('month', created_at)
		ORDER BY DATE_TRUNC('month', created_at)
	`)
	if err != nil {
		log.Fatalf("Failed to query songs over time: %v", err)
	}

	// Top 15 artists by unique song count
	err = db.Select(&stats.TopArtists, `
		SELECT a.name, COUNT(DISTINCT lsa.song_id) AS song_count
		FROM artists a
		JOIN link_song_artists lsa ON a.id = lsa.artist_id
		GROUP BY a.id, a.name
		ORDER BY song_count DESC
		LIMIT 15
	`)
	if err != nil {
		log.Fatalf("Failed to query top artists: %v", err)
	}

	// Top 30 genres by unique song count
	err = db.Select(&stats.GenreDistribution, `
		SELECT g.name, COUNT(DISTINCT lsg.song_id) AS song_count
		FROM genres g
		JOIN link_song_genres lsg ON g.id = lsg.genre_id
		GROUP BY g.id, g.name
		ORDER BY song_count DESC
		LIMIT 30
	`)
	if err != nil {
		log.Fatalf("Failed to query genre distribution: %v", err)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(stats); err != nil {
		log.Fatalf("Failed to encode stats JSON: %v", err)
	}
}
