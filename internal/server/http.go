package server

import (
	"net/http"

	"github.com/Paaaark/hanquant/internal/data"
	"github.com/Paaaark/hanquant/internal/handler"
	"github.com/Paaaark/hanquant/internal/service"
)

func NewHTTPServer() http.Handler {
    kisClient := data.NewKISClient()
    mux := http.NewServeMux()

    stockService, _ := service.NewStockService()
    apiHandler := handler.NewStockHandler(stockService)
    
    wsService := service.NewWebSocketService(kisClient)
    wsHandler := handler.NewWebSocketHandler(wsService)
    wsService.Start()

    mux.HandleFunc("/prices/recent/", apiHandler.GetRecentPrice)
    mux.HandleFunc("/prices/historical/", apiHandler.GetHistoricalPrice)
    mux.HandleFunc("/ranking/fluctuation", apiHandler.GetTopFluctuationStocks)
    mux.HandleFunc("/ranking/volume", apiHandler.GetMostTradedStocks)
    mux.HandleFunc("/ranking/market-cap", apiHandler.GetTopMarketCapStocks)
    mux.HandleFunc("/snapshot/multstock", apiHandler.GetMultipleStockSnapshot)
    mux.HandleFunc("/index/", apiHandler.GetIndexPrice)

    mux.HandleFunc("/ws/stocks", wsHandler.HandleWebSocket)

    return mux
}