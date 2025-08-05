package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Paaaark/hanquant/internal/data"
	"github.com/Paaaark/hanquant/internal/service"
)

// rateLimiter ensures we don't exceed 20 API calls per second across multiple service calls
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

func main() {
	// Define command line flags
	var (
		action     = flag.String("action", "", "Action to perform: fetch-daily, fetch-minute, load-daily, load-minute, bulk-daily, bulk-minute, fetch-all-daily-data")
		symbol     = flag.String("symbol", "", "Stock symbol (e.g., 005930)")
		fromDate   = flag.String("from", "", "Start date (YYYYMMDD format)")
		toDate     = flag.String("to", "", "End date (YYYYMMDD format)")
		yearMonth  = flag.String("month", "", "Year and month for minute data (YYYYMM format)")
		symbolsFile = flag.String("symbols", "", "File containing list of symbols (one per line)")
		outputFile = flag.String("output", "", "Output file for data (JSON format)")
	)
	flag.Parse()

	// Validate required flags
	if *action == "" {
		log.Fatal("Action is required. Use -action flag.")
	}

	// Initialize KIS client
	kisClient := data.NewKISClient()

	// Initialize S3 storage
	s3Config := data.S3Config{
		BucketName: os.Getenv("S3_BUCKET_NAME"),
		Region:     os.Getenv("AWS_REGION"),
		AccessKey:  os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretKey:  os.Getenv("AWS_SECRET_ACCESS_KEY"),
	}

	if s3Config.BucketName == "" || s3Config.Region == "" || s3Config.AccessKey == "" || s3Config.SecretKey == "" {
		log.Fatal("S3 configuration environment variables are required: S3_BUCKET_NAME, AWS_REGION, AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY")
	}

	s3Storage, err := data.NewS3Storage(s3Config)
	if err != nil {
		log.Fatalf("Failed to initialize S3 storage: %v", err)
	}

	// Initialize historical service
	historicalService := service.NewHistoricalService(kisClient, s3Storage)

	// Execute action
	switch *action {
	case "fetch-daily":
		if *symbol == "" || *fromDate == "" || *toDate == "" {
			log.Fatal("Symbol, from date, and to date are required for fetch-daily action")
		}
		err := historicalService.FetchAndStoreDailyData(*symbol, *fromDate, *toDate)
		if err != nil {
			log.Fatalf("Failed to fetch daily data: %v", err)
		}
		fmt.Printf("Successfully fetched and stored daily data for %s\n", *symbol)

	case "fetch-minute":
		if *symbol == "" || *fromDate == "" || *toDate == "" {
			log.Fatal("Symbol, from date, and to date are required for fetch-minute action")
		}
		err := historicalService.FetchAndStoreMinuteData(*symbol, *fromDate, *toDate)
		if err != nil {
			log.Fatalf("Failed to fetch minute data: %v", err)
		}
		fmt.Printf("Successfully fetched and stored minute data for %s\n", *symbol)

	case "load-daily":
		if *symbol == "" {
			log.Fatal("Symbol is required for load-daily action")
		}
		data, err := historicalService.LoadDailyData(*symbol)
		if err != nil {
			log.Fatalf("Failed to load daily data: %v", err)
		}
		outputData(data, *outputFile)

	case "load-minute":
		if *symbol == "" || *yearMonth == "" {
			log.Fatal("Symbol and month are required for load-minute action")
		}
		data, err := historicalService.LoadMinuteData(*symbol, *yearMonth)
		if err != nil {
			log.Fatalf("Failed to load minute data: %v", err)
		}
		outputData(data, *outputFile)

	case "bulk-daily":
		symbols, err := loadSymbols(*symbolsFile)
		if err != nil {
			log.Fatalf("Failed to load symbols: %v", err)
		}
		err = historicalService.BulkFetchDailyData(symbols)
		if err != nil {
			log.Fatalf("Failed to bulk fetch daily data: %v", err)
		}
		fmt.Println("Successfully completed bulk daily data fetch")

	case "bulk-minute":
		symbols, err := loadSymbols(*symbolsFile)
		if err != nil {
			log.Fatalf("Failed to load symbols: %v", err)
		}
		err = historicalService.BulkFetchMinuteData(symbols)
		if err != nil {
			log.Fatalf("Failed to bulk fetch minute data: %v", err)
		}
		fmt.Println("Successfully completed bulk minute data fetch")

	case "fetch-all-daily-data":
		if *toDate == "" {
			log.Fatal("To date is required for fetch-all-daily-data action")
		}
		err := fetchAllDailyData(historicalService, *toDate)
		if err != nil {
			log.Fatalf("Failed to fetch all daily data: %v", err)
		}
		fmt.Println("Successfully completed fetch-all-daily-data")

	default:
		log.Fatalf("Unknown action: %s", *action)
	}
}

// loadStockSymbolsFromCSV reads stock symbols from the stock_listings.csv file
func loadStockSymbolsFromCSV() ([]string, error) {
	csvPath := filepath.Join(".kis_data", "stock_listings.csv")
	
	file, err := os.Open(csvPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open stock listings CSV: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file is empty or has no data rows")
	}

	var symbols []string
	// Skip header row, start from index 1
	for i := 1; i < len(records); i++ {
		if len(records[i]) > 0 {
			symbol := strings.TrimSpace(records[i][0])
			if symbol != "" {
				symbols = append(symbols, symbol)
			}
		}
	}

	return symbols, nil
}

// fetchAllDailyData fetches 10 years of daily data for all stock symbols
func fetchAllDailyData(historicalService *service.HistoricalService, toDate string) error {
	// Load all stock symbols from CSV
	symbols, err := loadStockSymbolsFromCSV()
	if err != nil {
		return fmt.Errorf("failed to load stock symbols: %w", err)
	}

	fmt.Printf("Loaded %d stock symbols from stock_listings.csv\n", len(symbols))

	// Parse toDate to calculate fromDate (10 years back)
	toTime, err := time.Parse("20060102", toDate)
	if err != nil {
		return fmt.Errorf("invalid to date format: %w", err)
	}

	fromTime := toTime.AddDate(-10, 0, 0)
	fromDate := fromTime.Format("20060102")

	fmt.Printf("Fetching 10 years of daily data from %s to %s for %d symbols\n", fromDate, toDate, len(symbols))

	// Create rate limiter for cross-service call rate limiting
	rateLimiter := &rateLimiter{}
	
	successCount := 0
	errorCount := 0

	for i, symbol := range symbols {
		rateLimiter.wait()
		
		fmt.Printf("Processing symbol %d/%d: %s\n", i+1, len(symbols), symbol)
		
		err := historicalService.FetchAndStoreDailyData(symbol, fromDate, toDate)
		if err != nil {
			fmt.Printf("Error fetching daily data for %s: %v\n", symbol, err)
			errorCount++
			continue
		}
		
		successCount++
		fmt.Printf("Successfully processed %s (%d/%d)\n", symbol, i+1, len(symbols))
	}

	fmt.Printf("Completed fetch-all-daily-data. Success: %d, Errors: %d\n", successCount, errorCount)
	return nil
}

// loadSymbols reads symbols from a file
func loadSymbols(filename string) ([]string, error) {
	if filename == "" {
		return nil, fmt.Errorf("symbols file is required")
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read symbols file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	var symbols []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			symbols = append(symbols, line)
		}
	}

	return symbols, nil
}

// outputData outputs data to file or stdout
func outputData(resultData interface{}, outputFile string) {
	var output []byte
	var err error

	// Use custom JSON encoding if available
	switch d := resultData.(type) {
	case data.SlicePriceStruct:
		output = d.EncodeJSON()
	case data.SliceMinutePriceStruct:
		output = d.EncodeJSON()
	default:
		output, err = json.MarshalIndent(resultData, "", "  ")
		if err != nil {
			log.Fatalf("Failed to marshal data: %v", err)
		}
	}

	if outputFile != "" {
		err = os.WriteFile(outputFile, output, 0644)
		if err != nil {
			log.Fatalf("Failed to write output file: %v", err)
		}
		fmt.Printf("Data written to %s\n", outputFile)
	} else {
		fmt.Println(string(output))
	}
} 