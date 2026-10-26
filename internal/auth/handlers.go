package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/charmbracelet/log"
)

const AuthMaxBytes = 128

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

func (as *AuthService) GetUserAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	reader := http.MaxBytesReader(w, r.Body, AuthMaxBytes)
	defer reader.Close()
	var credentials AuthRequest
	if err := json.NewDecoder(reader).Decode(&credentials); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	if credentials.Username == "" || credentials.Password == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	// bcrypt len check
	if len(credentials.Password) > 72 {
		http.Error(w, "Password should be less than 72 bytes", http.StatusBadRequest)
		return
	}

	token, err := as.LoginOrRegisterUser(
		r.Context(),
		credentials.Username,
		credentials.Password,
	)
	if errors.Is(err, ErrPasswordMismatch) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if err != nil {
		log.Errorf("auth: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(AuthResponse{Token: token}); err != nil {
		log.Errorf("auth: JSON encoding failed: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(buf.Bytes()); err != nil {
		log.Errorf("auth: response write failed: %v", err)
	}
}
