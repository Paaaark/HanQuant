package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Paaaark/hanquant/internal/auth"
	"github.com/Paaaark/hanquant/internal/data"
	"github.com/Paaaark/hanquant/internal/service"
	"github.com/Paaaark/hanquant/pkg/utils"
)

type StockHandler struct {
	svc *service.StockService
	DB  *sql.DB
}

func NewStockHandler(svc *service.StockService, db *sql.DB) *StockHandler {
	return &StockHandler{svc: svc, DB: db}
}

func (h *StockHandler) GetRecentPrice(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		http.Error(w, "invalid path", http.StatusBadRequest)
	}

    symbol := parts[3]

    result, err := h.svc.GetRecentPrice(symbol)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.Write(result.EncodeJSON())
}

func (h *StockHandler) GetHistoricalPrice(w http.ResponseWriter, r *http.Request) {
    // Split the path to extract symbol
    path := r.URL.Path
    parts := strings.Split(path, "/")
    if len(parts) < 4 {
        http.Error(w, "missing stock symbol in path", http.StatusBadRequest)
        return
    }
    symbol := parts[3]

    from := r.URL.Query().Get("from")
    to := r.URL.Query().Get("to")
    duration := r.URL.Query().Get("duration")

    if symbol == "" || from == "" || to == "" || duration == "" {
        http.Error(w, "missing query parameters", http.StatusBadRequest)
        return
    }

    result, err := h.svc.GetHistoricalPrice(symbol, from, to, duration)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.Write(result.EncodeJSON())
}

func (h *StockHandler) GetTopFluctuationStocks(w http.ResponseWriter, r *http.Request) {
    result, err := h.svc.GetTopFluctuationStocks()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.Write(result.EncodeJSON())
}

func (h *StockHandler) GetMostTradedStocks(w http.ResponseWriter, r *http.Request) {
    result, err := h.svc.GetMostTradedStocks()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.Write(result.EncodeJSON())
}

func (h *StockHandler) GetTopMarketCapStocks(w http.ResponseWriter, r *http.Request) {
    result, err := h.svc.GetTopMarketCapStocks()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.Write(result.EncodeJSON())
}

func (h *StockHandler) GetMultipleStockSnapshot(w http.ResponseWriter, r *http.Request) {
    tickerParam := r.URL.Query().Get("tickers")
    if tickerParam == "" {
        http.Error(w, "missing tickers parameter", http.StatusBadRequest)
    }
    tickers := strings.Split(tickerParam, ",")
    if len(tickers) > 30 {
        http.Error(w, "cannot request more than 30 tickers", http.StatusBadRequest)
        return
    }
 
    result, err := h.svc.GetMultipleStockSnapshot(tickers)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.Write(result.EncodeJSON())
}

func (h *StockHandler) GetIndexPrice(w http.ResponseWriter, r *http.Request) {
    path := r.URL.Path
    parts := strings.Split(path, "/")
    if len(parts) < 3 || parts[2] == "" {
        http.Error(w, "missing index code in path", http.StatusBadRequest)
        return
    }

    code := parts[2]
    result, err := h.svc.GetIndexPrice(code)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.Write(result.EncodeJSON())
}

// GetAccountPortfolio handles:
//   GET /accounts/{accNo}/portfolio
func (h *StockHandler) GetAccountPortfolio(w http.ResponseWriter, r *http.Request) {
	//   /accounts/12345678-01/portfolio
	trimmed := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(trimmed, "/")
	if len(parts) != 3 || parts[0] != "accounts" || parts[2] != "portfolio" {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	accNo := parts[1]

	userID, _, ok := requireJWT(w, r)
	if !ok {
		return
	}
	// Find user account
	accounts, err := data.GetUserAccountsByUserID(h.DB, userID)
	if err != nil {
		http.Error(w, `{"error":{"code":"DB","message":"`+err.Error()+`"}}`, http.StatusInternalServerError)
		return
	}
	var ua *data.UserAccount
	for i := range accounts {
		if accounts[i].AccountID == accNo {
			ua = &accounts[i]
			break
		}
	}
	if ua == nil {
		http.Error(w, `{"error":{"code":"ACCOUNT_NOT_FOUND","message":"No linked account for user"}}`, http.StatusNotFound)
		return
	}
	cano, _ := utils.Decrypt(string(ua.EncCANO))
	appKey, _ := utils.Decrypt(string(ua.EncAppKey))
	appSecret, _ := utils.Decrypt(string(ua.EncAppSecret))
	kis := &data.KISClient{AppKey: appKey, AppSecret: appSecret}
	isMock := ua.IsMock
	positions, summary, err := h.svc.GetAccountPortfolio(kis, cano, isMock)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	var resp []byte
	resp = append(resp, []byte(`{"positions":`)...)
	resp = append(resp, positions.EncodeJSON()...)
	resp = append(resp, []byte(`,"summary":`)...)
	if summary != nil {
		resp = append(resp, summary.EncodeJSON()...)
	} else {
		resp = append(resp, []byte(`null`)...)
	}
	resp = append(resp, byte('}'))
	w.Write(resp)
}

// GetAccountPortfolioMock handles:
//   GET /accounts_mock/{accNo}/portfolio
func (h *StockHandler) GetAccountPortfolioMock(w http.ResponseWriter, r *http.Request) {
	//   /accounts_mock/12345678-01/portfolio
	trimmed := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(trimmed, "/")
	if len(parts) != 3 || parts[0] != "accounts_mock" || parts[2] != "portfolio" {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	accNo := parts[1]

	// Find user account
	userID, _, ok := requireJWT(w, r)
	if !ok {
		return
	}
	accounts, err := data.GetUserAccountsByUserID(h.DB, userID)
	if err != nil {
		http.Error(w, `{"error":{"code":"DB","message":"`+err.Error()+`"}}`, http.StatusInternalServerError)
		return
	}
	var ua *data.UserAccount
	for i := range accounts {
		if accounts[i].AccountID == accNo {
			ua = &accounts[i]
			break
		}
	}
	if ua == nil {
		http.Error(w, `{"error":{"code":"ACCOUNT_NOT_FOUND","message":"No linked account for user"}}`, http.StatusNotFound)
		return
	}
	cano, _ := utils.Decrypt(string(ua.EncCANO))
	appKey, _ := utils.Decrypt(string(ua.EncAppKey))
	appSecret, _ := utils.Decrypt(string(ua.EncAppSecret))
	kis := &data.KISClient{AppKey: appKey, AppSecret: appSecret}
	isMock := ua.IsMock
	positions, summary, err := h.svc.GetAccountPortfolio(kis, cano, isMock)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	var resp []byte
	resp = append(resp, []byte(`{"positions":`)...)
	resp = append(resp, positions.EncodeJSON()...)
	resp = append(resp, []byte(`,"summary":`)...)
	if summary != nil {
		resp = append(resp, summary.EncodeJSON()...)
	} else {
		resp = append(resp, []byte(`null`)...)
	}
	resp = append(resp, byte('}'))
	w.Write(resp)
}

// Handler for GET /portfolio?account_id=...
func (h *StockHandler) GetPortfolio(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := requireJWT(w, r)
	if !ok {
		return
	}
	accountID := r.URL.Query().Get("account_id")
	if accountID == "" {
		http.Error(w, `{"error":{"code":"VALIDATION","message":"account_id required"}}`, http.StatusBadRequest)
		return
	}
	// Find user account
	accounts, err := data.GetUserAccountsByUserID(h.DB, userID)
	if err != nil {
		http.Error(w, `{"error":{"code":"DB","message":"`+err.Error()+`"}}`, http.StatusInternalServerError)
		return
	}
	var ua *data.UserAccount
	for i := range accounts {
		if accounts[i].AccountID == accountID {
			ua = &accounts[i]
			break
		}
	}
	if ua == nil {
		http.Error(w, `{"error":{"code":"ACCOUNT_NOT_FOUND","message":"No linked account for user"}}`, http.StatusNotFound)
		return
	}
	cano, _ := utils.Decrypt(string(ua.EncCANO))
	appKey, _ := utils.Decrypt(string(ua.EncAppKey))
	appSecret, _ := utils.Decrypt(string(ua.EncAppSecret))
	kis := &data.KISClient{AppKey: appKey, AppSecret: appSecret}
	isMock := ua.IsMock
	positions, summary, err := h.svc.GetAccountPortfolio(kis, cano, isMock)
	if err != nil {
		http.Error(w, `{"error":{"code":"KIS","message":"`+err.Error()+`"}}`, http.StatusInternalServerError)
		return
	}
	resp := struct {
		AsOf      string                      `json:"as_of"`
		Positions data.SlicePortfolioPosition `json:"positions"`
		Summary   *data.AccountSummary        `json:"summary"`
	}{
		AsOf:      time.Now().Format(time.RFC3339),
		Positions: positions,
		Summary:   summary,
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	var buf []byte
	buf = append(buf, []byte(`{"AsOf":`)...)
	buf = append(buf, []byte(fmt.Sprintf("\"%s\"", resp.AsOf))...)
	buf = append(buf, []byte(`,"Positions":`)...)
	buf = append(buf, resp.Positions.EncodeJSON()...)
	buf = append(buf, []byte(`,"Summary":`)...)
	if resp.Summary != nil {
		buf = append(buf, resp.Summary.EncodeJSON()...)
	} else {
		buf = append(buf, []byte(`null`)...)
	}
	buf = append(buf, byte('}'))
	w.Write(buf)
}

// Handler for POST /orders
func (h *StockHandler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := requireJWT(w, r)
	if !ok {
		return
	}
	var req struct {
		AccountID  string  `json:"account_id"`
		Symbol     string  `json:"symbol"`
		Side       string  `json:"side"`
		Qty        float64 `json:"qty"`
		OrderType  string  `json:"order_type"`
		LimitPrice *float64 `json:"limit_price,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":{"code":"VALIDATION","message":"invalid request"}}`, http.StatusBadRequest)
		return
	}
	if req.AccountID == "" || req.Symbol == "" || req.Side == "" || req.Qty <= 0 || req.OrderType == "" {
		http.Error(w, `{"error":{"code":"VALIDATION","message":"missing or invalid fields"}}`, http.StatusBadRequest)
		return
	}
	accounts, err := data.GetUserAccountsByUserID(h.DB, userID)
	if err != nil {
		http.Error(w, `{"error":{"code":"DB","message":"`+err.Error()+`"}}`, http.StatusInternalServerError)
		return
	}
	var ua *data.UserAccount
	for i := range accounts {
		if accounts[i].AccountID == req.AccountID {
			ua = &accounts[i]
			break
		}
	}
	if ua == nil {
		http.Error(w, `{"error":{"code":"ACCOUNT_NOT_FOUND","message":"No linked account for user"}}`, http.StatusNotFound)
		return
	}
	cano, _ := utils.Decrypt(string(ua.EncCANO))
	appKey, _ := utils.Decrypt(string(ua.EncAppKey))
	appSecret, _ := utils.Decrypt(string(ua.EncAppSecret))
	kis := &data.KISClient{AppKey: appKey, AppSecret: appSecret}
	isMock := ua.IsMock
	orderReq := data.OrderRequest{
		Symbol:    req.Symbol,
		Qty:       fmt.Sprintf("%.2f", req.Qty),
		OrderType: req.OrderType,
		Side:      req.Side,
		Mock:      isMock,
	}
	if req.LimitPrice != nil {
		orderReq.Price = fmt.Sprintf("%.2f", *req.LimitPrice)
	}
	orderResp, err := h.svc.PlaceOrder(kis, cano, orderReq)
	if err != nil {
		http.Error(w, `{"error":{"code":"KIS","message":"`+err.Error()+`"}}`, http.StatusInternalServerError)
		return
	}
	// Save order to DB
	ord := &data.Order{
		UserAccountID: ua.ID,
		Symbol:        req.Symbol,
		Side:          req.Side,
		Qty:           req.Qty,
		OrderType:     req.OrderType,
		LimitPrice:    req.LimitPrice,
		Status:        "PENDING",
		KISOrderID:    orderResp.OrderNo,
	}
	err = data.CreateOrder(h.DB, ord)
	if err != nil {
		http.Error(w, `{"error":{"code":"DB","message":"`+err.Error()+`"}}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(ord.EncodeJSON())
}

// Handler for GET /orders/{id}
func (h *StockHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := requireJWT(w, r)
	if !ok {
		return
	}
	idStr := r.URL.Path[len("/orders/"):] // crude extraction
	var orderID int64
	_, err := fmt.Sscanf(idStr, "%d", &orderID)
	if err != nil {
		http.Error(w, `{"error":{"code":"VALIDATION","message":"invalid order id"}}`, http.StatusBadRequest)
		return
	}
	order, err := data.GetOrderByID(h.DB, userID, orderID)
	if err != nil {
		http.Error(w, `{"error":{"code":"DB","message":"`+err.Error()+`"}}`, http.StatusInternalServerError)
		return
	}
	if order == nil {
		http.Error(w, `{"error":{"code":"NOT_FOUND","message":"order not found"}}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(order.EncodeJSON())
}

// Helper to extract and validate JWT from Authorization header
func requireJWT(w http.ResponseWriter, r *http.Request) (int64, string, bool) {
	header := r.Header.Get("Authorization")
	if header == "" || !strings.HasPrefix(header, "Bearer ") {
		http.Error(w, "missing or invalid Authorization header", http.StatusUnauthorized)
		return 0, "", false
	}
	token := strings.TrimPrefix(header, "Bearer ")
	claims, err := auth.ValidateJWT(token)
	if err != nil {
		http.Error(w, "invalid or expired token", http.StatusUnauthorized)
		return 0, "", false
	}
	return claims.UserID, claims.Username, true
}

// POST /accounts/{accNo}/buy
func (h *StockHandler) BuyStock(w http.ResponseWriter, r *http.Request) {
	trimmed := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(trimmed, "/")
	if len(parts) != 3 || parts[0] != "accounts" || parts[2] != "buy" {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	accNo := parts[1]
	var req data.OrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	req.Side = "buy"
	// Find user account and decrypt credentials
	userID, _, ok := requireJWT(w, r)
	if !ok {
		return
	}
	accounts, err := data.GetUserAccountsByUserID(h.DB, userID)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	var ua *data.UserAccount
	for i := range accounts {
		if accounts[i].AccountID == accNo {
			ua = &accounts[i]
			break
		}
	}
	if ua == nil {
		http.Error(w, "No linked account for user", http.StatusNotFound)
		return
	}
	cano, _ := utils.Decrypt(string(ua.EncCANO))
	appKey, _ := utils.Decrypt(string(ua.EncAppKey))
	appSecret, _ := utils.Decrypt(string(ua.EncAppSecret))
	kis := &data.KISClient{AppKey: appKey, AppSecret: appSecret}
	resp, err := h.svc.PlaceOrder(kis, cano, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(resp.EncodeJSON())
}

// POST /accounts/{accNo}/sell
func (h *StockHandler) SellStock(w http.ResponseWriter, r *http.Request) {
	trimmed := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(trimmed, "/")
	if len(parts) != 3 || parts[0] != "accounts" || parts[2] != "sell" {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	accNo := parts[1]
	var req data.OrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	req.Side = "sell"
	// Find user account and decrypt credentials
	userID, _, ok := requireJWT(w, r)
	if !ok {
		return
	}
	accounts, err := data.GetUserAccountsByUserID(h.DB, userID)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	var ua *data.UserAccount
	for i := range accounts {
		if accounts[i].AccountID == accNo {
			ua = &accounts[i]
			break
		}
	}
	if ua == nil {
		http.Error(w, "No linked account for user", http.StatusNotFound)
		return
	}
	cano, _ := utils.Decrypt(string(ua.EncCANO))
	appKey, _ := utils.Decrypt(string(ua.EncAppKey))
	appSecret, _ := utils.Decrypt(string(ua.EncAppSecret))
	kis := &data.KISClient{AppKey: appKey, AppSecret: appSecret}
	resp, err := h.svc.PlaceOrder(kis, cano, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(resp.EncodeJSON())
}
