package main

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/persistence/postgres/migrations"
	"os"
)

var (
	migrationsList = [3]*migrations.Migration{
		migrations.Migration001,
		migrations.Migration002,
		migrations.Migration003,
	}
)

// could eventually configure this as a CLI tool, but for now just runs all `up` migrations
func main() {
	log.SetFormatter(&log.JSONFormatter{})

	host, hostExists := os.LookupEnv("POSTGRES_HOST")
	user, userExists := os.LookupEnv("POSTGRES_USER")
	password, passwordExists := os.LookupEnv("POSTGRES_PASSWORD")

	if !hostExists || !userExists || !passwordExists {
		log.Fatal("POSTGRES_HOST, POSTGRES_USER, and POSTGRES_PASSWORD must be set")
	}

	log.Infof("Running migrations on %s@%s", user, host)

	db := sqlx.MustConnect("postgres", "host="+host+" user="+user+" password="+password+" dbname=vaultbot sslmode=disable")

	db.MustExec(`
		CREATE TABLE IF NOT EXISTS migrations (
			id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
			name VARCHAR(255) NOT NULL UNIQUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)

	numMigrations := len(migrationsList)
	log.Infof("Found %d migrations", numMigrations)

	migrationsExecuted := 0
	for _, migration := range migrationsList {
		exists := false
		err := db.QueryRowx("SELECT EXISTS(SELECT 1 FROM migrations WHERE name = $1)", migration.Name).Scan(&exists)
		if err != nil {
			panic(err)
		}

		if exists {
			log.Infof("Migration %s already exists, skipping", migration.Name)
			continue
		}

		log.Infof("Running migration %s", migration.Name)
		tx := db.MustBegin()
		tx.MustExec(migration.Up)
		tx.MustExec("INSERT INTO migrations (name) VALUES ($1)", migration.Name)
		err = tx.Commit()
		if err != nil {
			log.Errorf("Error running migration %s, rolling back transaction: %v", migration.Name, err)
			panic(err)
		}

		migrationsExecuted++
	}

	log.Infof("Executed %d migrations", migrationsExecuted)
}
