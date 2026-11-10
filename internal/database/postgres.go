package database

import (
	"database/sql"
	"os"

	"github.com/charmbracelet/log"

	"github.com/lib/pq"
)

func NewPostgresConnection() *sql.DB {
	cfg, err := parseDBConfig()
	if err != nil {
		log.Fatal(err)
	}

	c, err := pq.NewConnectorConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}

	db := sql.OpenDB(c)

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func parseDBConfig() (pq.Config, error) {
	if v, ok := os.LookupEnv("DATABASE_URL"); ok {
		return pq.NewConfig(v)
	}

	return pq.NewConfig("host=localhost dbname=mwall sslmode=disable")
}
