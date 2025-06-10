package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/Paaaark/hanquant/internal/data"
)

func main() {
    inputPath := filepath.Join("kospi_code.txt")
    outputPath := "stock_listings.csv"

    stocks, err := data.ParseStockListingFile(inputPath)
    if err != nil {
        log.Fatalf("Failed to parse: %v", err)
    }

    if err := data.WriteToCSV(stocks, outputPath); err != nil {
        log.Fatalf("Failed to write CSV: %v", err)
    }

    fmt.Printf("Parsed %d stocks to %s\n", len(stocks), outputPath)
}
