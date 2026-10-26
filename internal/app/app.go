// Package app represents application singleton
// which handles initialization and web setup
package app

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/sawthorn/merch-wallet/internal/auth"
	"github.com/sawthorn/merch-wallet/internal/repository"
	"github.com/sawthorn/merch-wallet/internal/repository/transactions"
)

type App struct {
	AuthService *auth.AuthService
}

func NewApp(db *sql.DB) *App {
	executor := transactions.NewDBExecutor(db)
	trManager := transactions.NewTransactionManager(db)
	userRepo := repository.NewUserRepository(executor)

	return &App{
		AuthService: auth.NewAuthService(userRepo, trManager),
	}
}

func (a *App) SetupRouter() *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc("GET /api/info", func(w http.ResponseWriter, r *http.Request) {})
	router.HandleFunc("POST /api/sendCoin", func(w http.ResponseWriter, r *http.Request) {})
	router.HandleFunc("GET /api/buy/{item}", func(w http.ResponseWriter, r *http.Request) {})
	router.HandleFunc("POST /api/auth", a.AuthService.GetUserAuth)

	return router
}

func RunServer(router *http.ServeMux) {
	serv := http.Server{
		Addr:    "localhost:8080",
		Handler: router,
	}

	log.Fatal(serv.ListenAndServe())
}
