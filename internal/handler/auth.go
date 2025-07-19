package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

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

type refreshTokenRecord struct {
	UserID    int64
	Token     string
	ExpiresAt time.Time
}

var refreshTokens = make(map[string]refreshTokenRecord) // in-memory for now

func generateRefreshToken(userID int64) (string, time.Time, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", time.Time{}, err
	}
	token := base64.URLEncoding.EncodeToString(b)
	expires := time.Now().Add(7 * 24 * time.Hour)
	refreshTokens[token] = refreshTokenRecord{UserID: userID, Token: token, ExpiresAt: expires}
	return token, expires, nil
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
	token, err := auth.GenerateJWT(user.ID, user.Username)
	if err != nil {
		http.Error(w, "failed to generate authentication token", http.StatusInternalServerError)
		return
	}
	refreshToken, refreshExpires, err := generateRefreshToken(user.ID)
	if err != nil {
		http.Error(w, "failed to generate refresh token", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(struct {
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
		ExpiresIn    int       `json:"expires_in"`
		RefreshExpiresAt time.Time `json:"refresh_expires_at"`
	}{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresIn:    3600,
		RefreshExpiresAt: refreshExpires,
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	rec, ok := refreshTokens[req.RefreshToken]
	if !ok || time.Now().After(rec.ExpiresAt) {
		http.Error(w, "invalid or expired refresh token", http.StatusUnauthorized)
		return
	}
	// Optionally: rotate refresh token here
	user, err := data.GetUserByID(h.DB, rec.UserID)
	if err != nil || user == nil {
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}
	token, err := auth.GenerateJWT(user.ID, user.Username)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(struct {
		Token string `json:"token"`
		ExpiresIn int `json:"expires_in"`
	}{
		Token: token,
		ExpiresIn: 3600,
	})
}

// Handler for POST /accounts (link KIS account)
func (h *AuthHandler) LinkAccount(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := requireJWT(w, r)
	if !ok {
		return
	}
	var req struct {
		AccountID  string `json:"account_id"`
		AppKey     string `json:"app_key"`
		AppSecret  string `json:"app_secret"`
		CANO       string `json:"cano"`
		IsMock     bool   `json:"is_mock"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":{"code":"VALIDATION","message":"invalid request"}}`, http.StatusBadRequest)
		return
	}
	if req.AccountID == "" || req.AppKey == "" || req.AppSecret == "" || req.CANO == "" {
		http.Error(w, `{"error":{"code":"VALIDATION","message":"missing required fields"}}`, http.StatusBadRequest)
		return
	}
	encCANO, err := utils.Encrypt(req.CANO)
	if err != nil {
		http.Error(w, `{"error":{"code":"VALIDATION","message":"encryption error"}}`, http.StatusInternalServerError)
		return
	}
	encAppKey, err := utils.Encrypt(req.AppKey)
	if err != nil {
		http.Error(w, `{"error":{"code":"VALIDATION","message":"encryption error"}}`, http.StatusInternalServerError)
		return
	}
	encAppSecret, err := utils.Encrypt(req.AppSecret)
	if err != nil {
		http.Error(w, `{"error":{"code":"VALIDATION","message":"encryption error"}}`, http.StatusInternalServerError)
		return
	}
	ua := &data.UserAccount{
		UserID:      userID,
		AccountID:   req.AccountID,
		EncCANO:     []byte(encCANO),
		EncAppKey:   []byte(encAppKey),
		EncAppSecret: []byte(encAppSecret),
		IsMock:      req.IsMock,
	}
	err = data.CreateUserAccount(h.DB, ua)
	if err != nil {
		if err.Error() == "pq: duplicate key value violates unique constraint \"user_accounts_user_id_account_id_key\"" {
			http.Error(w, `{"error":{"code":"CONFLICT","message":"account already linked"}}`, http.StatusConflict)
			return
		}
		http.Error(w, `{"error":{"code":"DB","message":"`+err.Error()+`"}}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ua)
}

// Handler for GET /accounts (list linked accounts)
func (h *AuthHandler) ListAccounts(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := requireJWT(w, r)
	if !ok {
		return
	}
	accounts, err := data.GetUserAccountsByUserID(h.DB, userID)
	if err != nil {
		http.Error(w, `{"error":{"code":"DB","message":"`+err.Error()+`"}}`, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(accounts)
}

// Handler for DELETE /accounts/{id} (unlink account)
func (h *AuthHandler) UnlinkAccount(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := requireJWT(w, r)
	if !ok {
		return
	}
	idStr := r.URL.Path[len("/accounts/"):] // crude extraction
	var accountID int64
	_, err := fmt.Sscanf(idStr, "%d", &accountID)
	if err != nil {
		http.Error(w, `{"error":{"code":"VALIDATION","message":"invalid account id"}}`, http.StatusBadRequest)
		return
	}
	err = data.DeleteUserAccount(h.DB, userID, accountID)
	if err != nil {
		http.Error(w, `{"error":{"code":"DB","message":"`+err.Error()+`"}}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
} 