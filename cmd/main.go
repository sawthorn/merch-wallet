package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"github.com/sawthorn/merch-wallet/internal/app"
)

func main() {
	pool := getDBConn()
	application := app.NewApp(pool)

	router := application.SetupRouter()
	app.RunServer(router)
}

func getDBConn() *sql.DB {
	db, err := sql.Open("postgres", "host=localhost dbname=mwall sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	return db
}
