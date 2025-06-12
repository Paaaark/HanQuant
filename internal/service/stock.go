package service

import (
	"github.com/Paaaark/hanquant/internal/data"
)

type StockService struct {
	store *data.StockStore
    kis   *data.KISClient
}

func NewStockService() (*StockService, error) {
	store, err := data.Load("stock_listings.csv")
	if err != nil {
		return nil, err
	}

    return &StockService{
		store: store,
        kis:   data.NewKISClient(),
    }, nil
}

func (s *StockService) SearchStocks(query string) []data.StockIdentity {
    return s.store.SearchStocks(query)
}

func (s *StockService) GetRecentPrice(symbol string) ([]data.PriceStruct, error) {
    return s.kis.GetRecentDailyPrice(symbol)
}

func (s *StockService) GetHistoricalPrice(symbol, from, to, duration string) ([]data.PriceStruct, error) {
    return s.kis.GetDailyPrice(symbol, from, to, duration)
}

func (s *StockService) GetTopFluctuationStocks() ([]data.RankingStock, error) {
	return s.kis.GetTopFluctuationStocks()
}

func (s *StockService) GetIndexPrice(code string) (*data.IndexStruct, error) {
	return s.kis.GetIndexPrice(code)
}
