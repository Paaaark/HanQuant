package service

import (
	"fmt"
	"sort"
	"strings"
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

// rateLimiter ensures we don't exceed 20 API calls per second
type rateLimiter struct {
	lastCall time.Time
}

func (rl *rateLimiter) wait() {
	// Ensure at least 50ms between calls (20 calls per second = 50ms per call)
	elapsed := time.Since(rl.lastCall)
	if elapsed < 50*time.Millisecond {
		time.Sleep(50*time.Millisecond - elapsed)
	}
	rl.lastCall = time.Now()
}

// isRateLimitError checks if an error is a rate limiting error
func isRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "RATE LIMIT EXCEEDED") ||
		   strings.Contains(errStr, "TOO_MANY_REQUESTS") ||
		   strings.Contains(errStr, "429") ||
		   strings.Contains(errStr, "503")
}

// handleRateLimitError handles rate limiting with exponential backoff
func handleRateLimitError(err error, attempt int) error {
	if !isRateLimitError(err) {
		return err
	}
	
	// Exponential backoff: 1s, 2s, 4s, 8s, 16s, 32s
	backoffDuration := time.Duration(1<<uint(attempt)) * time.Second
	if backoffDuration > 32*time.Second {
		backoffDuration = 32 * time.Second
	}
	
	fmt.Printf("Rate limit detected! Waiting %v before retry (attempt %d)...\n", backoffDuration, attempt+1)
	time.Sleep(backoffDuration)
	
	return fmt.Errorf("rate limit retry attempt %d: %w", attempt+1, err)
}

// parseDate parses date string in YYYYMMDD format
func parseDate(dateStr string) (time.Time, error) {
	return time.Parse("20060102", dateStr)
}

// formatDate formats time.Time to YYYYMMDD format
func formatDate(t time.Time) string {
	return t.Format("20060102")
}

// getLatestDateFromS3 gets the latest date from existing S3 data
func (s *HistoricalService) getLatestDateFromS3(symbol, dataType string) (string, error) {
	switch dataType {
	case "daily":
		existingData, err := s.s3Storage.LoadDailyData(symbol)
		if err != nil {
			if strings.Contains(err.Error(), "NoSuchKey") {
				return "", nil // No existing data
			}
			return "", err
		}
		if len(existingData) == 0 {
			return "", nil
		}
		// Sort by date and return the latest
		sort.Slice(existingData, func(i, j int) bool {
			return existingData[i].Date > existingData[j].Date
		})
		return existingData[0].Date, nil
	case "minute":
		// For minute data, we need to check all months
		// This is a simplified approach - in practice, you might want to list S3 objects
		// For now, we'll return empty string to fetch all data
		return "", nil
	default:
		return "", fmt.Errorf("unsupported data type: %s", dataType)
	}
}

// splitDateRange splits a large date range into smaller chunks
// Daily data: ~100 days per API call (with overlap)
// Minute data: ~120 minutes per API call (with overlap)
func splitDateRange(fromDate, toDate, dataType string) ([][]string, error) {
	from, err := parseDate(fromDate)
	if err != nil {
		return nil, fmt.Errorf("invalid from date: %w", err)
	}
	
	to, err := parseDate(toDate)
	if err != nil {
		return nil, fmt.Errorf("invalid to date: %w", err)
	}

	var chunks [][]string
	var chunkSize int
	var overlap int

	switch dataType {
	case "daily":
		chunkSize = 90  // Conservative: 90 days per call (API limit is ~100)
		overlap = 5     // 5 days overlap to ensure no gaps
	case "minute":
		chunkSize = 100 // Conservative: 100 minutes per call (API limit is ~120)
		overlap = 10    // 10 minutes overlap
	default:
		return nil, fmt.Errorf("unsupported data type: %s", dataType)
	}

	current := from
	for current.Before(to) || current.Equal(to) {
		chunkEnd := current.AddDate(0, 0, chunkSize)
		if chunkEnd.After(to) {
			chunkEnd = to
		}
		
		chunks = append(chunks, []string{
			formatDate(current),
			formatDate(chunkEnd),
		})
		
		// Move to next chunk with overlap
		current = current.AddDate(0, 0, chunkSize-overlap)
	}

	return chunks, nil
}

// FetchAndStoreDailyData fetches daily stock data from KIS API and stores it in S3
// Now includes intelligent fetching, rate limiting, and pagination
func (s *HistoricalService) FetchAndStoreDailyData(symbol, fromDate, toDate string) error {
	// Check existing data to avoid redundant fetches
	latestExisting, err := s.getLatestDateFromS3(symbol, "daily")
	if err != nil {
		return fmt.Errorf("failed to check existing data: %w", err)
	}

	// If we have existing data, only fetch from the latest date onwards
	if latestExisting != "" {
		latestTime, err := parseDate(latestExisting)
		if err != nil {
			return fmt.Errorf("failed to parse latest date: %w", err)
		}
		
		requestedFrom, err := parseDate(fromDate)
		if err != nil {
			return fmt.Errorf("failed to parse from date: %w", err)
		}
		
		// If the latest existing data is newer than our requested from date,
		// adjust the from date to avoid redundant fetches
		if latestTime.After(requestedFrom) {
			fromDate = formatDate(latestTime.AddDate(0, 0, 1)) // Start from next day
			fmt.Printf("Adjusted from date to %s (latest existing: %s)\n", fromDate, latestExisting)
		}
	}

	// If fromDate is after toDate, we already have all the data
	from, err := parseDate(fromDate)
	if err != nil {
		return fmt.Errorf("invalid from date: %w", err)
	}
	
	to, err := parseDate(toDate)
	if err != nil {
		return fmt.Errorf("invalid to date: %w", err)
	}
	
	if from.After(to) {
		fmt.Printf("No new data to fetch for %s (from: %s, to: %s)\n", symbol, fromDate, toDate)
		return nil
	}

	// Split date range into chunks
	chunks, err := splitDateRange(fromDate, toDate, "daily")
	if err != nil {
		return fmt.Errorf("failed to split date range: %w", err)
	}

	rateLimiter := &rateLimiter{}
	var allData data.SlicePriceStruct

	fmt.Printf("Fetching daily data for %s in %d chunks...\n", symbol, len(chunks))

	for i, chunk := range chunks {
		rateLimiter.wait()
		
		fmt.Printf("Chunk %d/%d: %s to %s\n", i+1, len(chunks), chunk[0], chunk[1])
		
		// Fetch data from KIS API with retry logic for rate limiting
		var dailyData data.SlicePriceStruct
		var err error
		maxRetries := 3
		
		for attempt := 0; attempt < maxRetries; attempt++ {
			dailyData, err = s.kisClient.GetDailyStockData(symbol, chunk[0], chunk[1])
			if err == nil {
				break // Success, exit retry loop
			}
			
			if isRateLimitError(err) {
				fmt.Printf("Rate limit error in chunk %d/%d: %v\n", i+1, len(chunks), err)
				if attempt < maxRetries-1 {
					handleRateLimitError(err, attempt)
					continue // Retry
				}
			}
			
			// Non-rate-limit error or max retries reached
			return fmt.Errorf("failed to fetch daily data for %s (%s-%s): %w", symbol, chunk[0], chunk[1], err)
		}

		allData = append(allData, dailyData...)
		fmt.Printf("Received %d daily records for chunk %d\n", len(dailyData), i+1)
	}

	if len(allData) == 0 {
		fmt.Printf("No daily data received for %s\n", symbol)
		return nil
	}

	// Merge with existing data and store in S3
	err = s.s3Storage.MergeAndStoreData(symbol, allData, "daily")
	if err != nil {
		return fmt.Errorf("failed to store daily data for %s: %w", symbol, err)
	}

	fmt.Printf("Successfully stored %d daily records for %s\n", len(allData), symbol)
	return nil
}

// FetchAndStoreMinuteData fetches minute-by-minute stock data from KIS API and stores it in S3
// Now includes intelligent fetching, rate limiting, and pagination
func (s *HistoricalService) FetchAndStoreMinuteData(symbol, fromDate, toDate string) error {
	// For minute data, we'll fetch all data and let the merge function handle deduplication
	// This is because minute data is organized by month and checking existing data is complex

	// Split date range into chunks
	chunks, err := splitDateRange(fromDate, toDate, "minute")
	if err != nil {
		return fmt.Errorf("failed to split date range: %w", err)
	}

	rateLimiter := &rateLimiter{}
	var allData data.SliceMinutePriceStruct

	fmt.Printf("Fetching minute data for %s in %d chunks...\n", symbol, len(chunks))

	for i, chunk := range chunks {
		rateLimiter.wait()
		
		fmt.Printf("Chunk %d/%d: %s to %s\n", i+1, len(chunks), chunk[0], chunk[1])
		
		// Fetch data from KIS API with retry logic for rate limiting
		var minuteData data.SliceMinutePriceStruct
		var err error
		maxRetries := 3
		
		for attempt := 0; attempt < maxRetries; attempt++ {
			minuteData, err = s.kisClient.GetMinuteStockData(symbol, chunk[0], chunk[1])
			if err == nil {
				break // Success, exit retry loop
			}
			
			if isRateLimitError(err) {
				fmt.Printf("Rate limit error in chunk %d/%d: %v\n", i+1, len(chunks), err)
				if attempt < maxRetries-1 {
					handleRateLimitError(err, attempt)
					continue // Retry
				}
			}
			
			// Non-rate-limit error or max retries reached
			return fmt.Errorf("failed to fetch minute data for %s (%s-%s): %w", symbol, chunk[0], chunk[1], err)
		}

		allData = append(allData, minuteData...)
		fmt.Printf("Received %d minute records for chunk %d\n", len(minuteData), i+1)
	}

	if len(allData) == 0 {
		fmt.Printf("No minute data received for %s\n", symbol)
		return nil
	}

	// Merge with existing data and store in S3
	err = s.s3Storage.MergeAndStoreData(symbol, allData, "minute")
	if err != nil {
		return fmt.Errorf("failed to store minute data for %s: %w", symbol, err)
	}

	fmt.Printf("Successfully stored %d minute records for %s\n", len(allData), symbol)
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
		return formatDate(fromDate), formatDate(now), nil
	case "1year":
		// 1 year of minute data
		fromDate := now.AddDate(-1, 0, 0)
		return formatDate(fromDate), formatDate(now), nil
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

	rateLimiter := &rateLimiter{}
	
	for i, symbol := range symbols {
		rateLimiter.wait()
		
		fmt.Printf("Fetching daily data for %s (%d/%d)...\n", symbol, i+1, len(symbols))
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

	rateLimiter := &rateLimiter{}
	
	for i, symbol := range symbols {
		rateLimiter.wait()
		
		fmt.Printf("Fetching minute data for %s (%d/%d)...\n", symbol, i+1, len(symbols))
		err := s.FetchAndStoreMinuteData(symbol, fromDate, toDate)
		if err != nil {
			fmt.Printf("Error fetching minute data for %s: %v\n", symbol, err)
			continue
		}
		fmt.Printf("Successfully fetched minute data for %s\n", symbol)
	}

	return nil
} 