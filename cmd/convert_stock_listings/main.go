package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Paaaark/hanquant/internal/data"
)

type marketConfig struct {
	path   string
	markID string // "1" = KOSPI, "2" = KOSDAQ
}

func main() {
	outPath := filepath.Join(".kis_data", "stock_listings_2.csv")

	// add more markets here later
	markets := []marketConfig{
		{path: filepath.Join(".kis_data", "kospi_code.mst"), markID: "1"},
		{path: filepath.Join(".kis_data", "kosdaq_code.mst"), markID: "2"},
	}

	// open writer once
	outFile, err := os.Create(outPath)
	if err != nil {
		log.Fatalf("create CSV: %v", err)
	}
	defer outFile.Close()

	w := csv.NewWriter(outFile)
	defer w.Flush()

	// header (existing fields + Market)
	if err := w.Write([]string{
		"Code", "ISIN", "Name", "SecurityType",
		"CapSize", "IndLarge", "IndMedium", "IndSmall",
		"Market",
	}); err != nil {
		log.Fatalf("write header: %v", err)
	}

	var total int
	for _, m := range markets {
		stocks, err := data.ParseStockListingFile(m.path)
		if err != nil {
			log.Fatalf("parse %s: %v", m.path, err)
		}

		for _, s := range stocks {
			if err := w.Write([]string{
				s.Code, s.ISIN, s.Name, s.SecurityType,
				s.CapSize, s.IndLarge, s.IndMedium, s.IndSmall,
				m.markID,
			}); err != nil {
				log.Fatalf("write row: %v", err)
			}
		}
		total += len(stocks)
	}

	fmt.Printf("Parsed %d stocks (KOSPI+KOSDAQ) to %s\n", total, outPath)
}
