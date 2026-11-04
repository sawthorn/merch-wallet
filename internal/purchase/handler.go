package purchase

import (
	"errors"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/sawthorn/merch-wallet/internal/middleware"
)

const PurchaseMaxBytes = 128

func (ps *PurchaseService) BuyItemHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	itemName := r.PathValue("item")
	if itemName == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	username, _ := middleware.GetUsernameFromCtx(r.Context())
	err := ps.BuyItem(r.Context(), username, itemName)
	if errors.Is(err, ErrItemNotFound) {
		http.Error(w, ErrItemNotFound.Error(), http.StatusBadRequest)
		return
	}
	if errors.Is(err, ErrInsufficientBalance) {
		http.Error(w, ErrInsufficientBalance.Error(), http.StatusBadRequest)
		return
	}
	if err != nil {
		log.Errorf("purchase: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
