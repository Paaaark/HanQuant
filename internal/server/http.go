package server

import (
	"database/sql"
	"net/http"
	"os"

	"log"

	"github.com/Paaaark/hanquant/internal/data"
	"github.com/Paaaark/hanquant/internal/handler"
	"github.com/Paaaark/hanquant/internal/service"
)

func NewHTTPServer() http.Handler {
	kisClient := data.NewKISClient()
	mux := http.NewServeMux()

	stockService, err := service.NewStockService()
	if err != nil {
		log.Printf("Warning: failed to create stock service: %v", err)
	}
	
	var db *sql.DB
	var authHandler *handler.AuthHandler
	if os.Getenv("POSTGRES_DSN") != "" {
		db, err = data.OpenDB()
		if err != nil {
			log.Printf("Warning: failed to connect to db: %v", err)
		} else {
			authHandler = handler.NewAuthHandler(db)
		}
	} else {
		log.Printf("Warning: POSTGRES_DSN not set, database features will be disabled")
	}
	
	apiHandler := handler.NewStockHandler(stockService, db)

	// Initialize backtesting service and handler
	backtestService := service.NewBacktestService(stockService)
	backtestHandler := handler.NewBacktestHandler(backtestService)

	wsService := service.NewWebSocketService(kisClient)
	wsHandler := handler.NewWebSocketHandler(wsService)
	wsService.Start()

	// --- New REST API endpoints ---
	if authHandler != nil {
		mux.HandleFunc("/auth/register", authHandler.Register)
		mux.HandleFunc("/auth/login", authHandler.Login)
		mux.HandleFunc("/auth/refresh", authHandler.Refresh)
	}
	if authHandler != nil {
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
	}

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

	// --- Backtesting endpoints ---
	mux.HandleFunc("/backtest/sma", backtestHandler.RunSMABacktest)

	return mux
}