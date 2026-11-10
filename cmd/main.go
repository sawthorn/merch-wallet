package main

import (
	"os"

	"github.com/charmbracelet/log"
	_ "github.com/lib/pq"
	"github.com/sawthorn/merch-wallet/internal/app"
	"github.com/sawthorn/merch-wallet/internal/database"
)

func main() {
	log.SetLevel(getLogLevel())

	pool := database.NewPostgresConnection()
	application := app.NewApp(pool)

	router := application.SetupRouter()
	app.RunServer(router)
}

func getLogLevel() log.Level {
	loglevelStr, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		return log.Default().GetLevel()
	}
	var ll log.Level
	switch loglevelStr {
	case log.DebugLevel.String():
		ll = log.DebugLevel
	case log.InfoLevel.String():
		ll = log.InfoLevel
	case log.WarnLevel.String():
		ll = log.WarnLevel
	case log.ErrorLevel.String():
		ll = log.ErrorLevel
	case log.FatalLevel.String():
		ll = log.FatalLevel
	}
	return ll
}
