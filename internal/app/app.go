// Package app represents application singleton
// which handles initialization and web setup
package app

import (
	"database/sql"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/sawthorn/merch-wallet/internal/auth"
	"github.com/sawthorn/merch-wallet/internal/middleware"
	"github.com/sawthorn/merch-wallet/internal/repository"
	"github.com/sawthorn/merch-wallet/internal/repository/transactions"
	"github.com/sawthorn/merch-wallet/internal/transfer"
)

type App struct {
	AuthService     *auth.AuthService
	TransferService *transfer.TransferService
}

func NewApp(db *sql.DB) *App {
	// db infrastructure setup
	executor := transactions.NewDBExecutor(db)
	trManager := transactions.NewTransactionManager(db)

	// repos setup
	userRepo := repository.NewUserRepository(executor)
	transferRepo := repository.NewTransferRepository(*executor)

	// services setup
	authService := auth.NewAuthService(userRepo, trManager)
	transferService := transfer.NewTransferService(userRepo, transferRepo, trManager)

	return &App{
		AuthService:     authService,
		TransferService: transferService,
	}
}

func (a *App) SetupRouter() *http.ServeMux {
	rootMux := http.NewServeMux()

	rootMux.HandleFunc("POST /api/auth", a.AuthService.GetUserAuth)

	rootMux.HandleFunc("POST /api/sendCoin",
		middleware.WithJWTAuth(a.TransferService.TransferCoins),
	)
	rootMux.HandleFunc("GET /api/info", func(w http.ResponseWriter, r *http.Request) {})
	rootMux.HandleFunc("GET /api/buy/{item}", func(w http.ResponseWriter, r *http.Request) {})

	return rootMux
}

func RunServer(router *http.ServeMux) {
	serv := http.Server{
		Addr:    "localhost:8080",
		Handler: middleware.WithLogging(router),
	}
	log.Debugf("starting server...")
	log.Fatal(serv.ListenAndServe())
}
