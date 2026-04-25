package main

import (
	"os"
	"regexp"
	"strings"

	"github.com/agnivade/levenshtein"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

const similarityThreshold = 0.85
const durationDeltaMs = 2000

var normRe = regexp.MustCompile(`(?i)\s*[-–(]\s*(single(?: version)?|album version|remaster(ed)?.*|radio edit|explicit(?: version)?|clean(?: version)?|feat\..*|bonus track.*|live.*|anniversary.*|deluxe.*)\s*[-)]*\s*$`)

type songRow struct {
	SpotifyID     string `db:"spotify_id"`
	Name          string `db:"name"`
	Duration      int    `db:"duration"`
	ArchiveCount  int    `db:"archive_count"`
	CreatedAtUnix int64  `db:"created_at_unix"`
}

func normalizeName(name string) string {
	n := strings.ToLower(strings.TrimSpace(name))
	return strings.TrimSpace(normRe.ReplaceAllString(n, ""))
}

func similarity(a, b string) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 1.0
	}
	maxLen := len([]rune(a))
	if lb := len([]rune(b)); lb > maxLen {
		maxLen = lb
	}
	if maxLen == 0 {
		return 1.0
	}
	dist := levenshtein.ComputeDistance(a, b)
	return 1.0 - float64(dist)/float64(maxLen)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func main() {
	log.SetFormatter(&log.JSONFormatter{})

	host, hostExists := os.LookupEnv("POSTGRES_HOST")
	port, portExists := os.LookupEnv("POSTGRES_PORT")
	user, userExists := os.LookupEnv("POSTGRES_USER")
	password, passwordExists := os.LookupEnv("POSTGRES_PASSWORD")
	dbName, dbNameExists := os.LookupEnv("POSTGRES_DB")

	if !hostExists || !portExists || !userExists || !passwordExists || !dbNameExists {
		log.Fatal("POSTGRES_HOST, POSTGRES_PORT, POSTGRES_USER, POSTGRES_PASSWORD, and POSTGRES_DB must be set")
	}

	log.Infof("Running song deduplication on %s@%s", user, host)

	dsn := "host=" + host + " port=" + port + " user=" + user + " password=" + password + " dbname=" + dbName + " sslmode=require channel_binding=require"

	db := sqlx.MustConnect("pgx", dsn)
	defer db.Close()

	var songs []songRow
	err := db.Select(&songs, `
		SELECT
			s.spotify_id,
			s.name,
			s.duration,
			COALESCE(COUNT(sa.id), 0)::int AS archive_count,
			EXTRACT(EPOCH FROM s.created_at)::bigint AS created_at_unix
		FROM songs s
		LEFT JOIN song_archive sa ON sa.song_id = s.id
		GROUP BY s.spotify_id, s.name, s.duration, s.created_at
		ORDER BY s.name
	`)
	if err != nil {
		log.Fatalf("Failed to fetch songs: %v", err)
	}

	log.Infof("Loaded %d songs for deduplication analysis", len(songs))

	type pair struct {
		dupID       string
		canonicalID string
	}
	var candidates []pair

	for i := 0; i < len(songs); i++ {
		for j := i + 1; j < len(songs); j++ {
			a, b := songs[i], songs[j]

			if abs(a.Duration-b.Duration) > durationDeltaMs {
				continue
			}

			normA := normalizeName(a.Name)
			normB := normalizeName(b.Name)
			sim := similarity(normA, normB)
			if sim < similarityThreshold {
				continue
			}

			// canonical = more archive entries; tiebreak = earlier created_at
			canonical, dup := a, b
			if b.ArchiveCount > a.ArchiveCount ||
				(b.ArchiveCount == a.ArchiveCount && b.CreatedAtUnix < a.CreatedAtUnix) {
				canonical, dup = b, a
			}

			log.Infof(
				"Duplicate candidate: %q (%s, %d archives) → canonical %q (%s, %d archives), similarity=%.3f, duration_delta=%dms",
				dup.Name, dup.SpotifyID, dup.ArchiveCount,
				canonical.Name, canonical.SpotifyID, canonical.ArchiveCount,
				sim, abs(dup.Duration-canonical.Duration),
			)

			candidates = append(candidates, pair{dupID: dup.SpotifyID, canonicalID: canonical.SpotifyID})
		}
	}

	if len(candidates) == 0 {
		log.Info("No duplicate candidates found")
		return
	}

	log.Infof("Found %d duplicate pair(s), updating duplicate_song_lookup", len(candidates))

	tx := db.MustBegin()
	updated := 0
	for _, c := range candidates {
		res, err := tx.Exec(`
			UPDATE duplicate_song_lookup
			SET target_song_spotify_id = $2
			WHERE source_song_spotify_id = $1
			  AND target_song_spotify_id = $1
		`, c.dupID, c.canonicalID)
		if err != nil {
			tx.Rollback()
			log.Fatalf("Failed to update duplicate_song_lookup: %v", err)
		}
		n, _ := res.RowsAffected()
		updated += int(n)
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}

	log.Infof("Deduplication complete: %d row(s) updated in duplicate_song_lookup", updated)
}
