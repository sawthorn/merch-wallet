package main

import (
	_ "github.com/lib/pq"
	"github.com/sawthorn/merch-wallet/internal/app"
	"github.com/sawthorn/merch-wallet/internal/database"
)

func main() {
	pool := database.NewPostgresConnection()
	application := app.NewApp(pool)

	router := application.SetupRouter()
	app.RunServer(router)
}
