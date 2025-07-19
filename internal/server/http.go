package server

import (
	"net/http"

	"log"
	"strings"

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

	mux.HandleFunc("/accounts/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/buy") && r.Method == http.MethodPost {
			apiHandler.BuyStock(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/sell") && r.Method == http.MethodPost {
			apiHandler.SellStock(w, r)
			return
		}
		// fallback to portfolio handler
		apiHandler.GetAccountPortfolio(w, r)
	})
	mux.HandleFunc("/accounts_mock/", apiHandler.GetAccountPortfolioMock)

	mux.HandleFunc("/ws/stocks", wsHandler.HandleWebSocket)

	db, err := data.OpenDB()
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	authHandler := handler.NewAuthHandler(db)
	mux.HandleFunc("/register", authHandler.Register)
	mux.HandleFunc("/login", authHandler.Login)

	return mux
}