package postgres

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
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

	host, hostExists := os.LookupEnv("POSTGRES_HOST")
	port, portExists := os.LookupEnv("POSTGRES_PORT")
	user, userExists := os.LookupEnv("POSTGRES_USER")
	password, passwordExists := os.LookupEnv("POSTGRES_PASSWORD")

	if !hostExists || !portExists || !userExists || !passwordExists {
		log.Fatal("POSTGRES_HOST, POSTGRES_PORT, POSTGRES_USER, and POSTGRES_PASSWORD must be set")
	}

	log.Infof("Initializing db pool connection to %s@%s", user, host)

	dsn := "host=" + host + " port=" + port + " user=" + user + " password=" + password + " dbname=vaultbot"
	_, envPresent := os.LookupEnv("ENVIRONMENT")
	if envPresent {
		// append sslmode=disable - local dev only
		dsn += " sslmode=disable"
	} else {
		// append sslmode=require - prod
		dsn += " sslmode=require"
	}

	newDb, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db = newDb

	return db, nil
}
