package server

import (
	"net/http"

	"github.com/Paaaark/hanquant/internal/handler"
	"github.com/Paaaark/hanquant/internal/service"
)

func NewHTTPServer() http.Handler {
    mux := http.NewServeMux()

    stockService, _ := service.NewStockService()
    h := handler.NewStockHandler(stockService)

    mux.HandleFunc("/search", h.SearchStocks)
    mux.HandleFunc("/prices/recent/", h.GetRecentPrice)
    mux.HandleFunc("/prices/historical/", h.GetHistoricalPrice)

    return mux
}