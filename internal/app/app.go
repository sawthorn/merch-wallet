// Package app represents application singleton
// which handles initialization and web setup
package app

import (
	"database/sql"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/sawthorn/merch-wallet/internal/auth"
	"github.com/sawthorn/merch-wallet/internal/database/transactions"
	"github.com/sawthorn/merch-wallet/internal/info"
	"github.com/sawthorn/merch-wallet/internal/middleware"
	"github.com/sawthorn/merch-wallet/internal/purchase"
	"github.com/sawthorn/merch-wallet/internal/repository"
	"github.com/sawthorn/merch-wallet/internal/transfer"
)

type App struct {
	AuthService     *auth.AuthService
	TransferService *transfer.TransferService
	PurchaseService *purchase.PurchaseService
	InfoService     *info.InfoService
}

func NewApp(db *sql.DB) *App {
	// db infrastructure setup
	executor := transactions.NewDBExecutor(db)
	trManager := transactions.NewTransactionManager(db)

	// repos setup
	userRepo := repository.NewUserRepository(executor)
	transferRepo := repository.NewTransferRepository(executor)
	itemRepo := repository.NewItemRepository(executor)
	purchaseRepo := repository.NewPurchaseRepository(executor)
	inventoryRepo := repository.NewInventoryRepository(executor)

	// services setup
	authService := auth.NewAuthService(userRepo, trManager)
	transferService := transfer.NewTransferService(userRepo, transferRepo, trManager)
	purchaseService := purchase.NewPurchaseService(
		userRepo,
		itemRepo,
		purchaseRepo,
		inventoryRepo,
		trManager,
	)
	infoService := info.NewInfoService(userRepo, inventoryRepo, transferRepo)

	return &App{
		AuthService:     authService,
		TransferService: transferService,
		PurchaseService: purchaseService,
		InfoService:     infoService,
	}
}

func (a *App) SetupRouter() *http.ServeMux {
	rootMux := http.NewServeMux()

	rootMux.HandleFunc("POST /api/auth", a.AuthService.GetUserAuth)

	rootMux.HandleFunc("POST /api/sendCoin",
		middleware.WithJWTAuth(a.TransferService.TransferCoins),
	)
	rootMux.HandleFunc("GET /api/buy/{item}",
		middleware.WithJWTAuth(a.PurchaseService.BuyItemHandler),
	)
	rootMux.HandleFunc("GET /api/info",
		middleware.WithJWTAuth(a.InfoService.GetUserInfo),
	)

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
