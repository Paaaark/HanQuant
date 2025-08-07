package service

import (
	"fmt"
	"os"
	"time"

	"github.com/Paaaark/hanquant/internal/data"
)

type StockService struct {
	store     *data.StockStore
	kis       *data.KISClient // default, but not used for user-specific calls
	s3Storage *data.S3Storage
}

func NewStockService() (*StockService, error) {
	// Initialize S3 storage
	s3Config := data.S3Config{
		BucketName: os.Getenv("S3_BUCKET_NAME"),
		Region:     os.Getenv("AWS_REGION"),
		AccessKey:  os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretKey:  os.Getenv("AWS_SECRET_ACCESS_KEY"),
	}

	fmt.Printf("DEBUG: S3 Config - Bucket: %s, Region: %s, AccessKey: %s, SecretKey: %s\n", 
		s3Config.BucketName, s3Config.Region, 
		func() string { if s3Config.AccessKey != "" { return "SET" } else { return "NOT SET" } }(),
		func() string { if s3Config.SecretKey != "" { return "SET" } else { return "NOT SET" } }())

	var s3Storage *data.S3Storage
	var err error
	if s3Config.BucketName != "" && s3Config.Region != "" && s3Config.AccessKey != "" && s3Config.SecretKey != "" {
		fmt.Printf("DEBUG: All S3 environment variables are set, initializing S3 storage...\n")
		s3Storage, err = data.NewS3Storage(s3Config)
		if err != nil {
			fmt.Printf("DEBUG: Failed to initialize S3 storage: %v\n", err)
			return nil, fmt.Errorf("failed to initialize S3 storage: %w", err)
		}
		fmt.Printf("DEBUG: S3 storage initialized successfully\n")
	} else {
		fmt.Printf("DEBUG: S3 environment variables missing, S3 storage will not be available\n")
	}

	return &StockService{
		store:     nil,
		kis:       data.NewKISClient(),
		s3Storage: s3Storage,
	}, nil
}

func (s *StockService) GetRecentPrice(symbol string) (interface{}, error) {
	return s.kis.GetRecentDailyPrice(symbol)
}

// GetHistoricalPrice prioritizes S3 data over KIS API calls
func (s *StockService) GetHistoricalPrice(symbol, from, to, duration string) (interface{}, error) {
	// Set default date range to 3 months if not specified
	if from == "" || to == "" {
		now := time.Now()
		to = now.Format("20060102")
		from = now.AddDate(0, -3, 0).Format("20060102") // 3 months ago
	}

	// Validate duration
	if duration == "" {
		duration = "D" // Default to daily
	}

	fmt.Printf("DEBUG: GetHistoricalPrice called for %s from %s to %s duration %s\n", symbol, from, to, duration)
	fmt.Printf("DEBUG: S3 storage available: %v\n", s.s3Storage != nil)

	// Try to get data from S3 first
	if s.s3Storage != nil {
		fmt.Printf("DEBUG: Attempting to get data from S3...\n")
		data, err := s.getHistoricalDataFromS3(symbol, from, to, duration)
		if err != nil {
			fmt.Printf("DEBUG: S3 error: %v\n", err)
		} else if data != nil {
			fmt.Printf("DEBUG: S3 data retrieved successfully\n")
			// Check if we have sufficient data coverage
			if s.hasSufficientDataCoverage(data, from, to) {
				fmt.Printf("DEBUG: Using S3 data (sufficient coverage)\n")
				return data, nil
			} else {
				fmt.Printf("DEBUG: S3 data insufficient coverage, falling back to KIS API\n")
			}
		} else {
			fmt.Printf("DEBUG: No S3 data found\n")
		}
	} else {
		fmt.Printf("DEBUG: S3 storage not available\n")
	}

	// If S3 data is not available or insufficient, fetch from KIS API
	fmt.Printf("DEBUG: Fetching from KIS API...\n")
	return s.getHistoricalDataFromKIS(symbol, from, to, duration)
}

// getHistoricalDataFromS3 retrieves historical data from S3 storage
func (s *StockService) getHistoricalDataFromS3(symbol, from, to, duration string) (interface{}, error) {
	fmt.Printf("DEBUG: getHistoricalDataFromS3 called for %s duration %s\n", symbol, duration)
	
	if duration == "D" {
		// Load all daily data from S3
		fmt.Printf("DEBUG: Loading daily data from S3 for %s\n", symbol)
		allData, err := s.s3Storage.LoadDailyData(symbol)
		if err != nil {
			fmt.Printf("DEBUG: LoadDailyData error: %v\n", err)
			return nil, err
		}

		fmt.Printf("DEBUG: Loaded %d records from S3\n", len(allData))

		// Filter data to requested date range
		filteredData := s.filterDataByDateRange(allData, from, to)
		if sliceData, ok := filteredData.([]data.PriceStruct); ok {
			fmt.Printf("DEBUG: Filtered to %d records in date range %s to %s\n", len(sliceData), from, to)
		} else {
			fmt.Printf("DEBUG: Filtered data type: %T\n", filteredData)
		}
		
		return filteredData, nil
	}

	fmt.Printf("DEBUG: Unsupported duration: %s\n", duration)
	return nil, fmt.Errorf("unsupported duration: %s", duration)
}

// hasSufficientDataCoverage checks if the data has sufficient coverage for the requested date range
func (s *StockService) hasSufficientDataCoverage(data interface{}, from, to string) bool {
	// For now, just return true if data is not nil
	return data != nil
}

// filterDataByDateRange filters data to the specified date range
func (s *StockService) filterDataByDateRange(input interface{}, from, to string) interface{} {
	fmt.Printf("DEBUG: filterDataByDateRange called with data type: %T\n", input)
	
	// Handle SlicePriceStruct type (which is []PriceStruct)
	if sliceData, ok := input.(data.SlicePriceStruct); ok {
		fmt.Printf("DEBUG: Filtering slice with %d records\n", len(sliceData))
		
		var filteredData data.SlicePriceStruct

		for _, record := range sliceData {
			if record.Date >= from && record.Date <= to {
				filteredData = append(filteredData, record)
			}
		}
		
		fmt.Printf("DEBUG: Filtered from %d to %d records in date range %s to %s\n", 
			len(sliceData), len(filteredData), from, to)
		return filteredData
	}
	
	// If we can't handle the type, return as-is with a warning
	fmt.Printf("DEBUG: Unknown data type %T, returning as-is\n", input)
	return input
}

// getHistoricalDataFromKIS fetches historical data from KIS API with chunking for large ranges
func (s *StockService) getHistoricalDataFromKIS(symbol, from, to, duration string) (interface{}, error) {
	// For now, just use the original method
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
	return s.kis.GetMultipleStockSnapshot(tickers)
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