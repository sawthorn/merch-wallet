package transfer

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/sawthorn/merch-wallet/internal/middleware"
)

const TransferMaxBytes = 128

type TransferRequest struct {
	ToUser *string `json:"toUser"`
	Amount *int    `json:"amount"`
}

func (ts *TransferService) TransferCoins(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Merhod not allowed", http.StatusMethodNotAllowed)
		return
	}

	reader := http.MaxBytesReader(w, r.Body, TransferMaxBytes)
	defer reader.Close()
	var req TransferRequest
	if err := json.NewDecoder(reader).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if req.ToUser == nil || req.Amount == nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	fromUsername, ok := middleware.GetUsernameFromCtx(r.Context())
	// usually won't not be the case, because
	// middleware will handle bad auth
	// this will fire only if context gets corrupted somehow
	if !ok {
		log.Errorf("transfer: cant get authenticated user from context")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	err := ts.TransferCoinsToUser(
		r.Context(),
		fromUsername,
		*req.ToUser,
		*req.Amount,
	)

	if errors.Is(err, ErrRecipientNotFound) {
		http.Error(w, ErrRecipientNotFound.Error(), http.StatusBadRequest)
		return
	}
	if errors.Is(err, ErrInsufficientBalance) {
		http.Error(w, ErrInsufficientBalance.Error(), http.StatusBadRequest)
		return
	}
	if errors.Is(err, ErrTransferToSelfProhibited) {
		http.Error(w, ErrTransferToSelfProhibited.Error(), http.StatusBadRequest)
		return
	}
	if err != nil {
		log.Errorf("transfer: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
