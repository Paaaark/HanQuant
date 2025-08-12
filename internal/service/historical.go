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

	// Check if the date range is valid
	if from.After(to) {
		return nil, fmt.Errorf("from date (%s) is after to date (%s)", fromDate, toDate)
	}

	// Check if we're trying to fetch future data
	now := time.Now()
	if from.After(now) {
		return nil, fmt.Errorf("from date (%s) is in the future", fromDate)
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

	// For very small date ranges, we need to be more careful
	daysDiff := int(to.Sub(from).Hours() / 24)
	if daysDiff <= chunkSize {
		// For small ranges, we still want to ensure we have enough days to potentially contain trading data
		// A 2-day range might not contain any trading days, so we need to be more conservative
		if daysDiff < 7 {
			// For very small ranges (less than a week), expand to at least 7 days to ensure we have trading days
			expandedFrom := from.AddDate(0, 0, -3) // Go back 3 days
			expandedTo := to.AddDate(0, 0, 3)      // Go forward 3 days
			
			// But don't go into the future
			now := time.Now()
			if expandedTo.After(now) {
				expandedTo = now
			}
			
			chunks = append(chunks, []string{
				formatDate(expandedFrom),
				formatDate(expandedTo),
			})
			return chunks, nil
		} else {
			// For ranges 7+ days, use as-is
			chunks = append(chunks, []string{
				formatDate(from),
				formatDate(to),
			})
			return chunks, nil
		}
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

// FetchAndStoreDailyData uses intelligent heuristics to determine what data to fetch
func (s *HistoricalService) FetchAndStoreDailyData(symbol, fromDate, toDate string) error {
	// Check existing data completeness
	isComplete, recordCount, err := s.CheckDataCompleteness(symbol, fromDate, toDate)
	if err != nil {
		return fmt.Errorf("failed to check data completeness: %w", err)
	}

	if isComplete {
		fmt.Printf("Data for %s is already complete (%d records), skipping fetch\n", symbol, recordCount)
		return nil
	}

	// Load existing data to determine what we need to fetch
	existingData, err := s.s3Storage.LoadDailyData(symbol)
	if err != nil && !strings.Contains(err.Error(), "NoSuchKey") {
		return fmt.Errorf("failed to load existing data: %w", err)
	}

	// Parse date range
	from, err := parseDate(fromDate)
	if err != nil {
		return fmt.Errorf("invalid from date: %w", err)
	}
	
	to, err := parseDate(toDate)
	if err != nil {
		return fmt.Errorf("invalid to date: %w", err)
	}

	// Check for future dates and adjust if necessary
	now := time.Now()
	if from.After(now) {
		fmt.Printf("Warning: From date %s is in the future, adjusting to today\n", fromDate)
		from = now
	}
	if to.After(now) {
		fmt.Printf("Warning: To date %s is in the future, adjusting to today\n", toDate)
		to = now
	}

	// Determine what date ranges we need to fetch
	neededRanges := s.determineNeededDateRanges(existingData, from, to)
	
	if len(neededRanges) == 0 {
		fmt.Printf("No new data needed for %s (from: %s, to: %s)\n", symbol, fromDate, toDate)
		return nil
	}

	fmt.Printf("\tFetching data for %s in %d ranges: %d existing records\n", symbol, len(neededRanges), len(existingData))

	// Fetch data for each needed range (from newest to oldest)
	rateLimiter := &rateLimiter{}
	var allNewData data.SlicePriceStruct
	reachedEndOfData := false

	for i, dateRange := range neededRanges {
		if reachedEndOfData {
			fmt.Printf("\tSkipping remaining ranges for %s (reached end of historical data)\n", symbol)
			break
		}

		rateLimiter.wait()
		
		fmt.Printf("\tRange %d/%d: %s to %s\n", i+1, len(neededRanges), 
			formatDate(dateRange.start), formatDate(dateRange.end))
		
		rateLimiter.wait()
		
		// Fetch data from KIS API with retry logic for rate limiting
		var dailyData data.SlicePriceStruct
		maxRetries := 3
		
		for attempt := 0; attempt < maxRetries; attempt++ {
			dailyData, err = s.kisClient.GetDailyStockData(symbol, formatDate(dateRange.start), formatDate(dateRange.end))
			if err == nil {
				break // Success, exit retry loop
			}
			
			if isRateLimitError(err) {
				fmt.Printf("\tRate limit error in range %d/%d: %v\n", i+1, len(neededRanges), err)
				if attempt < maxRetries-1 {
					handleRateLimitError(err, attempt)
					continue // Retry
				}
			}
			
			// Non-rate-limit error or max retries reached
			return fmt.Errorf("failed to fetch daily data for %s (%s-%s): %w", symbol, formatDate(dateRange.start), formatDate(dateRange.end), err)
		}

		allNewData = append(allNewData, dailyData...)
		fmt.Printf("\t\tReceived %d daily records for range %d\n", len(dailyData), i+1)

		// If the chunk returned 0 records, check if this is likely the end of historical data
		if len(dailyData) == 0 {
			reachedEndOfData = true
		}
	}

	if len(allNewData) == 0 {
		fmt.Printf("No new daily data received for %s\n", symbol)
		return nil
	}

	// Merge with existing data and store in S3
	err = s.s3Storage.MergeAndStoreData(symbol, allNewData, "daily")
	if err != nil {
		return fmt.Errorf("failed to store daily data for %s: %w", symbol, err)
	}

	// fmt.Printf("Successfully stored %d new daily records for %s\n", len(allNewData), symbol)
	return nil
}

// dateRange represents a date range for fetching
type dateRange struct {
	start time.Time
	end   time.Time
}

// determineNeededDateRanges analyzes existing data and determines what date ranges need to be fetched
// Returns ranges from newest to oldest to optimize for early termination when no more data exists
func (s *HistoricalService) determineNeededDateRanges(existingData data.SlicePriceStruct, from, to time.Time) []dateRange {
	// Don't fetch future data
	now := time.Now()
	if from.After(now) {
		fmt.Printf("Adjusting from date from %s to %s (future date)\n", formatDate(from), formatDate(now))
		from = now
	}
	if to.After(now) {
		fmt.Printf("Adjusting to date from %s to %s (future date)\n", formatDate(to), formatDate(now))
		to = now
	}
	
	// If the adjusted range is invalid, return empty
	if from.After(to) {
		fmt.Printf("Adjusted date range is invalid (%s to %s), returning empty ranges\n", formatDate(from), formatDate(to))
		return []dateRange{}
	}

	// Create a map of existing dates for quick lookup
	existingDates := make(map[string]bool)
	for _, record := range existingData {
		existingDates[record.Date] = true
	}

	var neededRanges []dateRange
	var start time.Time
	var end time.Time
	current := from

	// Parse through all dates from "from" to "to"
	for current.Before(to) || current.Equal(to) {
		currentDateStr := formatDate(current)
		
		// Check if current date is a trading day
		if data.IsTradingDay(currentDateStr) {
			// If it's a trading day, check if we have data for it
			if !existingDates[currentDateStr] {
				// We don't have data for this trading day
				if start.IsZero() {
					start = current
				} else {
					// If "start" is not empty, check if current is within 100 trading days of "start"
					// Calculate trading days between start and current (inclusive)
					startDateStr := formatDate(start)
					tradingDaysCount := data.CountInRange(startDateStr, currentDateStr)
					
					if tradingDaysCount <= 100 {
						// Current is within 100 trading days of start, set end = current
						end = current
					} else {
						// Current is not within 100 trading days of start
						// Push start and end as a range, then clear and set start = current
						if !end.IsZero() {
							neededRanges = append(neededRanges, dateRange{
								start: start,
								end:   end,
							})
						}
						start = current
						end = time.Time{}
					}
				}
			}
		}
		
		current = current.AddDate(0, 0, 1)
	}
	
	// Don't forget to add the last range if we have one
	if !start.IsZero() && !end.IsZero() {
		neededRanges = append(neededRanges, dateRange{
			start: start,
			end:   end,
		})
	} else if !start.IsZero() {
		// If we have a start but no end, use the start date as both start and end
		neededRanges = append(neededRanges, dateRange{
			start: start,
			end:   start,
		})
	}

	// Reverse the order to fetch from newest to oldest
	for i, j := 0, len(neededRanges)-1; i < j; i, j = i+1, j-1 {
		neededRanges[i], neededRanges[j] = neededRanges[j], neededRanges[i]
	}

	return neededRanges
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

// CheckDataCompleteness checks if a symbol has sufficient data coverage
func (s *HistoricalService) CheckDataCompleteness(symbol, fromDate, toDate string) (bool, int, error) {
	// Load existing data
	existingData, err := s.s3Storage.LoadDailyData(symbol)
	if err != nil {
		if strings.Contains(err.Error(), "NoSuchKey") {
			return false, 0, nil // No data exists
		}
		return false, 0, fmt.Errorf("failed to load data for %s: %w", symbol, err)
	}

	if len(existingData) == 0 {
		return false, 0, nil
	}

	// Parse date range
	from, err := parseDate(fromDate)
	if err != nil {
		return false, 0, fmt.Errorf("invalid from date: %w", err)
	}
	
	to, err := parseDate(toDate)
	if err != nil {
		return false, 0, fmt.Errorf("invalid to date: %w", err)
	}

	// Count records within the date range
	count := 0
	for _, record := range existingData {
		recordTime, err := parseDate(record.Date)
		if err != nil {
			continue
		}
		
		if (recordTime.After(from) || recordTime.Equal(from)) && 
		   (recordTime.Before(to) || recordTime.Equal(to)) {
			count++
		}
	}

	// Calculate the number of trading days
	tradingDays := data.CountInRange(fromDate, toDate)
	return tradingDays == count, count, nil
}

// BulkFetchDailyData fetches daily data for multiple symbols
func (s *HistoricalService) BulkFetchDailyData(symbols []string) error {
	fromDate, toDate, err := s.GetDateRangeForPeriod("5years")
	if err != nil {
		return err
	}

	rateLimiter := &rateLimiter{}
	
	successCount := 0
	errorCount := 0
	incompleteCount := 0

	for i, symbol := range symbols {
		rateLimiter.wait()
		
		fmt.Printf("Fetching daily data for %s (%d/%d)...\n", symbol, i+1, len(symbols))
		
		// Check if data is already complete
		isComplete, recordCount, err := s.CheckDataCompleteness(symbol, fromDate, toDate)
		if err != nil {
			fmt.Printf("Warning: Could not check data completeness for %s: %v\n", symbol, err)
		} else if isComplete {
			fmt.Printf("Data for %s is already complete (%d records), skipping\n", symbol, recordCount)
			successCount++
			continue
		}
		
		err = s.FetchAndStoreDailyData(symbol, fromDate, toDate)
		if err != nil {
			fmt.Printf("Error fetching daily data for %s: %v\n", symbol, err)
			errorCount++
			continue
		}
		
		// Check if data is now complete
		isComplete, recordCount, err = s.CheckDataCompleteness(symbol, fromDate, toDate)
		if err != nil {
			fmt.Printf("Warning: Could not verify data completeness for %s: %v\n", symbol, err)
		} else if !isComplete {
			fmt.Printf("Warning: Data for %s may be incomplete (%d records)\n", symbol, recordCount)
			incompleteCount++
		}
		
		successCount++
		fmt.Printf("Successfully fetched daily data for %s\n", symbol)
	}

	fmt.Printf("Completed bulk daily data fetch. Success: %d, Errors: %d, Incomplete: %d\n", 
		successCount, errorCount, incompleteCount)
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

// FetchAndStoreTodayData fetches today's data for multiple symbols using GetMultipleStockSnapshot
// This is more efficient than individual API calls as it can fetch 30 symbols at once
func (s *HistoricalService) FetchAndStoreTodayData(symbols []string, targetDate string) error {
	rateLimiter := &rateLimiter{}
	
	// Process symbols in batches of 30 (KIS API limit for GetMultipleStockSnapshot)
	batchSize := 30
	totalBatches := (len(symbols) + batchSize - 1) / batchSize
	
	fmt.Printf("Processing %d symbols in %d batches of %d symbols each\n", len(symbols), totalBatches, batchSize)
	
	successCount := 0
	errorCount := 0
	skippedCount := 0
	
	for batchIndex := 0; batchIndex < totalBatches; batchIndex++ {
		start := batchIndex * batchSize
		end := start + batchSize
		if end > len(symbols) {
			end = len(symbols)
		}
		
		batchSymbols := symbols[start:end]
		
		rateLimiter.wait()
		
		fmt.Printf("Processing batch %d/%d: symbols %d-%d (%d symbols)\n", 
			batchIndex+1, totalBatches, start+1, end, len(batchSymbols))
		
		// Fetch snapshot data for this batch
		snapshots, err := s.kisClient.GetMultipleStockSnapshot(batchSymbols)
		if err != nil {
			fmt.Printf("Error fetching snapshot for batch %d: %v\n", batchIndex+1, err)
			errorCount += len(batchSymbols)
			continue
		}
		
		fmt.Printf("Received %d snapshots for batch %d\n", len(snapshots), batchIndex+1)
		
		// Convert snapshots to daily data format and store each symbol
		for _, snapshot := range snapshots {
			// Check if we already have data for this date
			existingData, err := s.s3Storage.LoadDailyData(snapshot.Code)
			if err == nil {
				// Check if we already have data for the target date
				hasDataForDate := false
				for _, record := range existingData {
					if record.Date == targetDate {
						hasDataForDate = true
						break
					}
				}
				
				if hasDataForDate {
					fmt.Printf("Data for %s on %s already exists, skipping\n", snapshot.Code, targetDate)
					skippedCount++
					continue
				}
			}
			
			// Convert StockSnapshot to PriceStruct
			dailyData := data.PriceStruct{
				Date:     targetDate,
				Open:     snapshot.Open,
				High:     snapshot.High,
				Low:      snapshot.Low,
				Close:    snapshot.Price, // Current price is the close price for today
				Volume:   snapshot.Volume,
				Duration: "D",
			}
			
			// Create a slice with single daily record
			dailySlice := data.SlicePriceStruct{dailyData}
			
			// Store to S3
			err = s.s3Storage.MergeAndStoreData(snapshot.Code, dailySlice, "daily")
			if err != nil {
				fmt.Printf("Error storing data for %s: %v\n", snapshot.Code, err)
				errorCount++
				continue
			}
			
			successCount++
			fmt.Printf("Successfully stored today's data for %s (%s)\n", snapshot.Code, snapshot.Name)
		}
	}
	
	fmt.Printf("Completed FetchAndStoreTodayData. Success: %d, Skipped: %d, Errors: %d\n", 
		successCount, skippedCount, errorCount)
	return nil
} 