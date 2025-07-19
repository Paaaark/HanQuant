package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/Paaaark/hanquant/internal/auth"
	"github.com/Paaaark/hanquant/internal/data"
	"github.com/Paaaark/hanquant/pkg/utils"
)

type AuthHandler struct {
	DB *sql.DB
}

type authRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string `json:"token,omitempty"`
	Error string `json:"error,omitempty"`
}

type kisAccountRequest struct {
	Account string `json:"account"`
}

type kisAPIKeysRequest struct {
	RealKey    string `json:"real_key"`
	RealSecret string `json:"real_secret"`
	PaperKey   string `json:"paper_key"`
	PaperSecret string `json:"paper_secret"`
}

func NewAuthHandler(db *sql.DB) *AuthHandler {
	return &AuthHandler{DB: db}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON format", http.StatusBadRequest)
		return
	}
	if req.Username == "" || req.Password == "" {
		http.Error(w, "username and password are required", http.StatusBadRequest)
		return
	}
	
	// Validate username and password length
	if len(req.Username) < 3 || len(req.Username) > 50 {
		http.Error(w, "username must be between 3 and 50 characters", http.StatusBadRequest)
		return
	}
	if len(req.Password) < 6 {
		http.Error(w, "password must be at least 6 characters long", http.StatusBadRequest)
		return
	}
	
	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "failed to process password", http.StatusInternalServerError)
		return
	}
	
	err = data.RegisterUser(h.DB, req.Username, hash)
	if err != nil {
		// Check for specific error types
		if err.Error() == "username already exists" {
			http.Error(w, "username already exists", http.StatusConflict)
			return
		}
		if err.Error() == "invalid data provided" {
			http.Error(w, "invalid username or password format", http.StatusBadRequest)
			return
		}
		http.Error(w, "failed to register user: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(authResponse{})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON format", http.StatusBadRequest)
		return
	}
	if req.Username == "" || req.Password == "" {
		http.Error(w, "username and password are required", http.StatusBadRequest)
		return
	}
	
	user, err := data.GetUserByUsername(h.DB, req.Username)
	if err != nil {
		http.Error(w, "authentication failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "invalid username or password", http.StatusUnauthorized)
		return
	}
	
	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		http.Error(w, "invalid username or password", http.StatusUnauthorized)
		return
	}
	
	token, err := auth.GenerateJWT(user.Username)
	if err != nil {
		http.Error(w, "failed to generate authentication token", http.StatusInternalServerError)
		return
	}
	
	json.NewEncoder(w).Encode(authResponse{Token: token})
}

// POST /kis/account
func (h *AuthHandler) SetKISAccount(w http.ResponseWriter, r *http.Request) {
	username, ok := requireJWT(w, r)
	if !ok {
		return
	}
	var req kisAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON format in request body", http.StatusBadRequest)
		return
	}
	if req.Account == "" {
		http.Error(w, "account number is required", http.StatusBadRequest)
		return
	}
	enc, err := utils.Encrypt(req.Account)
	if err != nil {
		http.Error(w, "failed to encrypt account information", http.StatusInternalServerError)
		return
	}
	if err := data.UpdateKISAccountInfo(h.DB, username, enc); err != nil {
		http.Error(w, "failed to update account information: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// POST /kis/apikeys
func (h *AuthHandler) SetKISAPIKeys(w http.ResponseWriter, r *http.Request) {
	username, ok := requireJWT(w, r)
	if !ok {
		return
	}
	var req kisAPIKeysRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON format in request body", http.StatusBadRequest)
		return
	}
	
	// Validate required fields
	if req.RealKey == "" || req.RealSecret == "" || req.PaperKey == "" || req.PaperSecret == "" {
		http.Error(w, "all API keys and secrets are required", http.StatusBadRequest)
		return
	}
	
	realKeyEnc, err := utils.Encrypt(req.RealKey)
	if err != nil {
		http.Error(w, "failed to encrypt real trading API key", http.StatusInternalServerError)
		return
	}
	realSecretEnc, err := utils.Encrypt(req.RealSecret)
	if err != nil {
		http.Error(w, "failed to encrypt real trading API secret", http.StatusInternalServerError)
		return
	}
	paperKeyEnc, err := utils.Encrypt(req.PaperKey)
	if err != nil {
		http.Error(w, "failed to encrypt paper trading API key", http.StatusInternalServerError)
		return
	}
	paperSecretEnc, err := utils.Encrypt(req.PaperSecret)
	if err != nil {
		http.Error(w, "failed to encrypt paper trading API secret", http.StatusInternalServerError)
		return
	}
	if err := data.UpdateKISAPIKeys(h.DB, username, realKeyEnc, realSecretEnc, paperKeyEnc, paperSecretEnc); err != nil {
		http.Error(w, "failed to update API keys: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
} 