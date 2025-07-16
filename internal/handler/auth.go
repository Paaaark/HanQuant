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

func NewAuthHandler(db *sql.DB) *AuthHandler {
	return &AuthHandler{DB: db}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.Username == "" || req.Password == "" {
		http.Error(w, "username and password required", http.StatusBadRequest)
		return
	}
	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "failed to hash password", http.StatusInternalServerError)
		return
	}
	err = data.RegisterUser(h.DB, req.Username, hash)
	if err != nil {
		http.Error(w, "failed to register user", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(authResponse{})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	user, err := data.GetUserByUsername(h.DB, req.Username)
	if err != nil || user == nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	token, err := auth.GenerateJWT(user.Username)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(authResponse{Token: token})
} 