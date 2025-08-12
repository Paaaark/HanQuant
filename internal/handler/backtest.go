package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Paaaark/hanquant/internal/data"
	"github.com/Paaaark/hanquant/internal/service"
)

type BacktestHandler struct {
	backtestService *service.BacktestService
}

func NewBacktestHandler(backtestService *service.BacktestService) *BacktestHandler {
	return &BacktestHandler{
		backtestService: backtestService,
	}
}

// RunSMABacktest handles the /backtest/sma endpoint
func (h *BacktestHandler) RunSMABacktest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	params := data.BacktestParams{
		From:      r.URL.Query().Get("from"),
		To:        r.URL.Query().Get("to"),
		SMA_short: h.parseIntParam(r.URL.Query().Get("sma_short"), 20),
		SMA_long:  h.parseIntParam(r.URL.Query().Get("sma_long"), 50),
	}

	// Run backtest
	result, err := h.backtestService.RunSMABacktest(params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return results as JSON
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

// parseParam parses an integer parameter with a default value
func (h *BacktestHandler) parseIntParam(value string, defaultValue int) int {
	if value == "" {
		return defaultValue
	}
	
	if parsed, err := strconv.Atoi(value); err == nil {
		return parsed
	}
	
	return defaultValue
}
