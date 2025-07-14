package handler

import (
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

	positions, summary, err := h.svc.GetAccountPortfolio(accNo, false /*mock*/ )
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

	positions, summary, err := h.svc.GetAccountPortfolio(accNo, true /*mock*/ )
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
