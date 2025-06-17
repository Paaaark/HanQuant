package main

import (
	"log"
	"net/http"

	"github.com/Paaaark/hanquant/internal/server"
)

func main() {
	// kisClient := data.NewKISClient()
	// stockService, _ := service.NewStockService()
	// stockHandler := handler.NewStockHandler(stockService)

	// // Create WebSocket service and handler
	// wsService := service.NewWebSocketService(kisClient)
	// wsHandler := handler.NewWebSocketHandler(wsService)
	// wsService.Start()

	// // Register HTTP routes
	// http.HandleFunc("/api/stocks/recent/", stockHandler.GetRecentPrice)
	// http.HandleFunc("/api/stocks/historical/", stockHandler.GetHistoricalPrice)
	// http.HandleFunc("/api/stocks/top-fluctuation", stockHandler.GetTopFluctuationStocks)
	// http.HandleFunc("/api/stocks/most-traded", stockHandler.GetMostTradedStocks)
	// http.HandleFunc("/api/stocks/top-market-cap", stockHandler.GetTopMarketCapStocks)
	// http.HandleFunc("/api/stocks/snapshot", stockHandler.GetMultipleStockSnapshot)
	// http.HandleFunc("/api/index/", stockHandler.GetIndexPrice)

	// // Register WebSocket route
	// http.HandleFunc("/ws/stocks", wsHandler.HandleWebSocket)

	srv := server.NewHTTPServer()

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", srv); err != nil {
		log.Fatal(err)
	}
}
