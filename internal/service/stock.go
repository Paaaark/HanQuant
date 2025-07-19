package service

import (
	"github.com/Paaaark/hanquant/internal/data"
)

type StockService struct {
	store *data.StockStore
    kis   *data.KISClient // default, but not used for user-specific calls
}

func NewStockService() (*StockService, error) {
	// store, err := data.Load("stock_listings.csv")
	// if err != nil {
	// 	return nil, err
	// }

    return &StockService{
		store: nil,
        kis:   data.NewKISClient(),
    }, nil
}

func (s *StockService) GetRecentPrice(symbol string) (data.SlicePriceStruct, error) {
    return s.kis.GetRecentDailyPrice(symbol)
}

func (s *StockService) GetHistoricalPrice(symbol, from, to, duration string) (data.SlicePriceStruct, error) {
    return s.kis.GetDailyPrice(symbol, from, to, duration)
}

func (s *StockService) GetTopFluctuationStocks() (data.SliceRankingStock, error) {
	return s.kis.GetTopFluctuationStocks()
}

func (s *StockService) GetMostTradedStocks() (data.SliceRankingStock, error) {
	return s.kis.GetMostTradedStocks()
}

func (s *StockService) GetTopMarketCapStocks() (data.SliceRankingStock, error) {
	return s.kis.GetTopMarketCapStocks()
}

func (s *StockService) GetMultipleStockSnapshot(tickers []string) (data.SliceStockSnapshot, error) {
	return s.kis.GetMultipleStockSnapshot(tickers);
}

func (s *StockService) GetIndexPrice(code string) (*data.IndexStruct, error) {
	return s.kis.GetIndexPrice(code)
}

// Accepts a KISClient instance with user credentials
func (s *StockService) GetAccountPortfolio(kis *data.KISClient, accNo string, mock bool) (data.SlicePortfolioPosition, *data.AccountSummary, error) {
	return kis.GetAccountPortfolio(accNo, mock)
}

// Accepts a KISClient instance with user credentials
func (s *StockService) PlaceOrder(kis *data.KISClient, accNo string, req data.OrderRequest) (*data.OrderResponse, error) {
	return kis.PlaceOrder(accNo, req)
}