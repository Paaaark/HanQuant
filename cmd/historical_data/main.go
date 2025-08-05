package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Paaaark/hanquant/internal/data"
	"github.com/Paaaark/hanquant/internal/service"
)

func main() {
	// Define command line flags
	var (
		action     = flag.String("action", "", "Action to perform: fetch-daily, fetch-minute, load-daily, load-minute, bulk-daily, bulk-minute")
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

	default:
		log.Fatalf("Unknown action: %s", *action)
	}
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