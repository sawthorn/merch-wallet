package info

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/sawthorn/merch-wallet/internal/middleware"
)

type InfoResponse struct {
	Balance         int                     `json:"coins"`
	Inventory       []InventoryItemResponse `json:"inventory"`
	TransferHistory TransferHistoryResponse `json:"coinHistory"`
}

type InventoryItemResponse struct {
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
}

type TransferHistoryResponse struct {
	Sent     []TransferSentResponse     `json:"sent"`
	Received []TransferReceivedResponse `json:"received"`
}

type TransferSentResponse struct {
	ToUser string `json:"toUser"`
	Amount int    `json:"amount"`
}

type TransferReceivedResponse struct {
	FromUser string `json:"fromUser"`
	Amount   int    `json:"amount"`
}

func NewUserInfoResponse(userInfo *UserInfo) *InfoResponse {
	response := InfoResponse{
		Balance: userInfo.Balance,
	}
	for _, item := range userInfo.Inventory {
		response.Inventory = append(response.Inventory, InventoryItemResponse(item))
	}

	for _, tr := range userInfo.TransferSent {
		response.TransferHistory.Sent = append(
			response.TransferHistory.Sent,
			TransferSentResponse{
				ToUser: tr.Username,
				Amount: tr.Amount,
			})
	}

	for _, tr := range userInfo.TransferReceived {
		response.TransferHistory.Received = append(
			response.TransferHistory.Received,
			TransferReceivedResponse{
				FromUser: tr.Username,
				Amount:   tr.Amount,
			})
	}

	return &response
}

func (is *InfoService) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	username, _ := middleware.GetUsernameFromCtx(r.Context())
	userInfo, err := is.CollectUserInformation(r.Context(), username)
	if err != nil {
		log.Errorf("info: collect for %s: %v", username, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := NewUserInfoResponse(userInfo)

	w.Header().Set("Content-Type", "application/json")
	var bytes bytes.Buffer
	err = json.NewEncoder(&bytes).Encode(response)
	if err != nil {
		log.Errorf("info: response encoding failed: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Write(bytes.Bytes())
}
