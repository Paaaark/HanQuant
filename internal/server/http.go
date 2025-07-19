package server

import (
	"net/http"

	"log"

	"github.com/Paaaark/hanquant/internal/data"
	"github.com/Paaaark/hanquant/internal/handler"
	"github.com/Paaaark/hanquant/internal/service"
)

func NewHTTPServer() http.Handler {
	kisClient := data.NewKISClient()
	mux := http.NewServeMux()

	stockService, _ := service.NewStockService()
	db, err := data.OpenDB()
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	authHandler := handler.NewAuthHandler(db)
	apiHandler := handler.NewStockHandler(stockService, db)

	wsService := service.NewWebSocketService(kisClient)
	wsHandler := handler.NewWebSocketHandler(wsService)
	wsService.Start()

	// --- New REST API endpoints ---
	mux.HandleFunc("/auth/register", authHandler.Register)
	mux.HandleFunc("/auth/login", authHandler.Login)
	mux.HandleFunc("/auth/refresh", authHandler.Refresh)
	mux.HandleFunc("/accounts", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			authHandler.LinkAccount(w, r)
		case http.MethodGet:
			authHandler.ListAccounts(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/accounts/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			authHandler.UnlinkAccount(w, r)
			return
		}
		// legacy fallback
		apiHandler.GetAccountPortfolio(w, r)
	})
	mux.HandleFunc("/accounts_mock/", apiHandler.GetAccountPortfolioMock)
	mux.HandleFunc("/portfolio", apiHandler.GetPortfolio)
	mux.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			apiHandler.PlaceOrder(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/orders/", apiHandler.GetOrder)

	// --- WebSocket ---
	mux.HandleFunc("/ws/stocks", wsHandler.HandleWebSocket)

	// --- Restore all ranking and price endpoints ---
	mux.HandleFunc("/prices/recent/", apiHandler.GetRecentPrice)
	mux.HandleFunc("/prices/historical/", apiHandler.GetHistoricalPrice)
	mux.HandleFunc("/ranking/fluctuation", apiHandler.GetTopFluctuationStocks)
	mux.HandleFunc("/ranking/volume", apiHandler.GetMostTradedStocks)
	mux.HandleFunc("/ranking/market-cap", apiHandler.GetTopMarketCapStocks)
	mux.HandleFunc("/snapshot/multstock", apiHandler.GetMultipleStockSnapshot)
	mux.HandleFunc("/index/", apiHandler.GetIndexPrice)

	return mux
}