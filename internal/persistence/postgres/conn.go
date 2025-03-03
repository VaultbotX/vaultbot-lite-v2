package postgres

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"os"
)

func NewPostgresConnection() (*sqlx.DB, error) {
	host := os.Getenv("POSTGRES_HOST")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")

	for _, envVar := range []string{"POSTGRES_HOST", "POSTGRES_USER", "POSTGRES_PASSWORD"} {
		if os.Getenv(envVar) == "" {
			panic("missing required environment variable: " + envVar)
		}
	}

	dsn := "host=" + host + " user=" + user + " password=" + password + " dbname=vaultbot sslmode=disable"

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	return db, nil
}
