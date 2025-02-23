package main

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"github.com/vaultbotx/vaultbot-lite/internal/database/postgres/migrations"
	"os"
)

var (
	migrationsList = []*migrations.Migration{
		migrations.Migration_001,
	}
)

// could eventually configure this as a CLI tool, but for now just runs all `up` migrations
func main() {
	log.SetFormatter(&log.JSONFormatter{})

	host := os.Getenv("POSTGRES_HOST")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")

	for _, envVar := range []string{"POSTGRES_HOST", "POSTGRES_USER", "POSTGRES_PASSWORD"} {
		if os.Getenv(envVar) == "" {
			panic("missing required environment variable: " + envVar)
		}
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
		// locate the migration in the migrations table
		exists := false
		rows, err := db.Queryx("SELECT EXISTS(SELECT 1 FROM migrations WHERE name = $1)", migration.Name)
		if err != nil {
			panic(err)
		}

		for rows.Next() {
			err := rows.Scan(&exists)
			if err != nil {
				panic(err)
			}
		}
		err = rows.Close()
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
