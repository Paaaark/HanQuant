package service

import (
	"fmt"
	"time"

	"github.com/Paaaark/hanquant/internal/data"
)

// HistoricalService manages historical stock data operations
type HistoricalService struct {
	kisClient *data.KISClient
	s3Storage *data.S3Storage
}

// NewHistoricalService creates a new HistoricalService instance
func NewHistoricalService(kisClient *data.KISClient, s3Storage *data.S3Storage) *HistoricalService {
	return &HistoricalService{
		kisClient: kisClient,
		s3Storage: s3Storage,
	}
}

// FetchAndStoreDailyData fetches daily stock data from KIS API and stores it in S3
func (s *HistoricalService) FetchAndStoreDailyData(symbol, fromDate, toDate string) error {
	// Fetch data from KIS API
	dailyData, err := s.kisClient.GetDailyStockData(symbol, fromDate, toDate)
	if err != nil {
		return fmt.Errorf("failed to fetch daily data for %s: %w", symbol, err)
	}

	// Merge with existing data and store in S3
	err = s.s3Storage.MergeAndStoreData(symbol, dailyData, "daily")
	if err != nil {
		return fmt.Errorf("failed to store daily data for %s: %w", symbol, err)
	}

	return nil
}

// FetchAndStoreMinuteData fetches minute-by-minute stock data from KIS API and stores it in S3
func (s *HistoricalService) FetchAndStoreMinuteData(symbol, fromDate, toDate string) error {
	// Fetch data from KIS API
	minuteData, err := s.kisClient.GetMinuteStockData(symbol, fromDate, toDate)
	if err != nil {
		return fmt.Errorf("failed to fetch minute data for %s: %w", symbol, err)
	}

	// Merge with existing data and store in S3
	err = s.s3Storage.MergeAndStoreData(symbol, minuteData, "minute")
	if err != nil {
		return fmt.Errorf("failed to store minute data for %s: %w", symbol, err)
	}

	return nil
}

// LoadDailyData loads daily stock data from S3
func (s *HistoricalService) LoadDailyData(symbol string) (data.SlicePriceStruct, error) {
	return s.s3Storage.LoadDailyData(symbol)
}

// LoadMinuteData loads minute-by-minute stock data from S3 for a specific month
func (s *HistoricalService) LoadMinuteData(symbol, yearMonth string) (data.SliceMinutePriceStruct, error) {
	return s.s3Storage.LoadMinuteData(symbol, yearMonth)
}

// FetchHistoricalData fetches historical data for a given period
func (s *HistoricalService) FetchHistoricalData(req data.HistoricalDataRequest) error {
	switch req.Duration {
	case "D":
		return s.FetchAndStoreDailyData(req.Symbol, req.FromDate, req.ToDate)
	case "M":
		return s.FetchAndStoreMinuteData(req.Symbol, req.FromDate, req.ToDate)
	default:
		return fmt.Errorf("unsupported duration: %s", req.Duration)
	}
}

// GetDateRangeForPeriod returns the date range for fetching historical data
func (s *HistoricalService) GetDateRangeForPeriod(period string) (string, string, error) {
	now := time.Now()
	
	switch period {
	case "5years":
		// 5 years of daily data
		fromDate := now.AddDate(-5, 0, 0)
		return fromDate.Format("20060102"), now.Format("20060102"), nil
	case "1year":
		// 1 year of minute data
		fromDate := now.AddDate(-1, 0, 0)
		return fromDate.Format("20060102"), now.Format("20060102"), nil
	default:
		return "", "", fmt.Errorf("unsupported period: %s", period)
	}
}

// BulkFetchDailyData fetches daily data for multiple symbols
func (s *HistoricalService) BulkFetchDailyData(symbols []string) error {
	fromDate, toDate, err := s.GetDateRangeForPeriod("5years")
	if err != nil {
		return err
	}

	for _, symbol := range symbols {
		fmt.Printf("Fetching daily data for %s...\n", symbol)
		err := s.FetchAndStoreDailyData(symbol, fromDate, toDate)
		if err != nil {
			fmt.Printf("Error fetching daily data for %s: %v\n", symbol, err)
			continue
		}
		fmt.Printf("Successfully fetched daily data for %s\n", symbol)
	}

	return nil
}

// BulkFetchMinuteData fetches minute data for multiple symbols
func (s *HistoricalService) BulkFetchMinuteData(symbols []string) error {
	fromDate, toDate, err := s.GetDateRangeForPeriod("1year")
	if err != nil {
		return err
	}

	for _, symbol := range symbols {
		fmt.Printf("Fetching minute data for %s...\n", symbol)
		err := s.FetchAndStoreMinuteData(symbol, fromDate, toDate)
		if err != nil {
			fmt.Printf("Error fetching minute data for %s: %v\n", symbol, err)
			continue
		}
		fmt.Printf("Successfully fetched minute data for %s\n", symbol)
	}

	return nil
} 