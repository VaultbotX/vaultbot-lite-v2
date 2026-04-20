package main

import (
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

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

	log.Infof("Refreshing genre graph materialized views on %s@%s", user, host)

	dsn := "host=" + host + " port=" + port + " user=" + user + " password=" + password + " dbname=" + dbName + " sslmode=require channel_binding=require"

	db := sqlx.MustConnect("pgx", dsn)
	defer db.Close()

	log.Info("Refreshing genre_graph_vertices")
	db.MustExec("REFRESH MATERIALIZED VIEW CONCURRENTLY genre_graph_vertices")

	log.Info("Refreshing genre_graph_edges")
	db.MustExec("REFRESH MATERIALIZED VIEW CONCURRENTLY genre_graph_edges")

	log.Info("Successfully refreshed genre graph materialized views")
}
