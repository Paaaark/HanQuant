package main

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Paaaark/hanquant/internal/data"
)

type marketConfig struct {
	url     string
	markID  string // "1" = KOSPI, "2" = KOSDAQ
	filename string
}

func main() {
	// Create .kis_data directory if it doesn't exist
	dataDir := ".kis_data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("create data directory: %v", err)
	}

	outPath := filepath.Join(dataDir, "stock_listings.csv")

	// Market configurations with download URLs
	markets := []marketConfig{
		{
			url:      "https://new.real.download.dws.co.kr/common/master/kospi_code.mst.zip",
			markID:   "1",
			filename: "kospi_code.mst",
		},
		{
			url:      "https://new.real.download.dws.co.kr/common/master/kosdaq_code.mst.zip",
			markID:   "2",
			filename: "kosdaq_code.mst",
		},
	}

	// Download and extract files
	for _, m := range markets {
		if err := downloadAndExtract(m.url, dataDir, m.filename); err != nil {
			log.Fatalf("download/extract %s: %v", m.filename, err)
		}
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
		filePath := filepath.Join(dataDir, m.filename)
		stocks, err := data.ParseStockListingFile(filePath)
		if err != nil {
			log.Fatalf("parse %s: %v", filePath, err)
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

	fmt.Printf("Downloaded, parsed %d stocks (KOSPI+KOSDAQ) to %s\n", total, outPath)
}

// downloadAndExtract downloads a zip file from URL and extracts the specified file
func downloadAndExtract(url, destDir, filename string) error {
	fmt.Printf("Downloading %s...\n", url)
	
	// Download the zip file
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	// Read the entire response body
	zipData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	// Create a zip reader
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return fmt.Errorf("create zip reader: %w", err)
	}

	// Find and extract the target file
	var extracted bool
	for _, file := range zipReader.File {
		if strings.HasSuffix(file.Name, filename) {
			if err := extractFile(file, destDir); err != nil {
				return fmt.Errorf("extract file %s: %w", file.Name, err)
			}
			extracted = true
			break
		}
	}

	if !extracted {
		return fmt.Errorf("file %s not found in zip", filename)
	}

	fmt.Printf("Successfully extracted %s\n", filename)
	return nil
}

// extractFile extracts a single file from zip to destination directory
func extractFile(zipFile *zip.File, destDir string) error {
	// Open the zip file
	rc, err := zipFile.Open()
	if err != nil {
		return fmt.Errorf("open zip file: %w", err)
	}
	defer rc.Close()

	// Create the destination file
	destPath := filepath.Join(destDir, filepath.Base(zipFile.Name))
	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("create destination file: %w", err)
	}
	defer destFile.Close()

	// Copy the file contents
	_, err = io.Copy(destFile, rc)
	if err != nil {
		return fmt.Errorf("copy file contents: %w", err)
	}

	return nil
}
