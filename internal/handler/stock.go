package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Paaaark/hanquant/internal/service"
)

type StockHandler struct {
    svc *service.StockService
}

func NewStockHandler(svc *service.StockService) *StockHandler {
    return &StockHandler{svc: svc}
}

func (h *StockHandler) SearchStocks(w http.ResponseWriter, r *http.Request) {
    query := r.URL.Query().Get("q")
    result := h.svc.SearchStocks(query)
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    json.NewEncoder(w).Encode(result)
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