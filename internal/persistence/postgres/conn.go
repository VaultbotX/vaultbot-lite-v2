package postgres

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"os"
	"sync"
)

var (
	// db is the global database connection. It is perfectly fine for a sqlx.DB to be global since it
	// simply represents a pool of connections rather than a single connection.
	db *sqlx.DB
	mu sync.Mutex
)

func NewPostgresConnection() (*sqlx.DB, error) {
	mu.Lock()
	defer mu.Unlock()

	if db != nil {
		return db, nil
	}

	host := os.Getenv("POSTGRES_HOST")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")

	for _, envVar := range []string{"POSTGRES_HOST", "POSTGRES_USER", "POSTGRES_PASSWORD"} {
		if os.Getenv(envVar) == "" {
			panic("missing required environment variable: " + envVar)
		}
	}

	dsn := "host=" + host + " user=" + user + " password=" + password + " dbname=vaultbot sslmode=disable"

	newDb, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db = newDb

	return db, nil
}
