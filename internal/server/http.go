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
    mux.HandleFunc("/ranking/fluctuation", h.GetTopFluctuationStocks)
    mux.HandleFunc("/ranking/volume", h.GetMostTradedStocks)
    mux.HandleFunc("/ranking/market-cap", h.GetTopMarketCapStocks)
    mux.HandleFunc("/index/", h.GetIndexPrice)

    return mux
}