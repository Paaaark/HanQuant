package data

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// S3Storage handles S3 operations for historical stock data
type S3Storage struct {
	s3Client   *s3.S3
	uploader   *s3manager.Uploader
	bucketName string
}

// NewS3Storage creates a new S3Storage instance
func NewS3Storage(config S3Config) (*S3Storage, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(config.Region),
		Credentials: credentials.NewStaticCredentials(config.AccessKey, config.SecretKey, ""),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	s3Client := s3.New(sess)
	uploader := s3manager.NewUploader(sess)

	return &S3Storage{
		s3Client:   s3Client,
		uploader:   uploader,
		bucketName: config.BucketName,
	}, nil
}

// StoreDailyData stores daily stock data to S3 as CSV
func (s *S3Storage) StoreDailyData(symbol string, data SlicePriceStruct) error {
	// Sort data by date
	sort.Slice(data, func(i, j int) bool {
		return data[i].Date < data[j].Date
	})

	// Create CSV content
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	
	// Write header
	writer.Write([]string{"Date", "Open", "High", "Low", "Close", "Volume", "Duration"})
	
	// Write data
	for _, row := range data {
		writer.Write([]string{
			row.Date,
			row.Open,
			row.High,
			row.Low,
			row.Close,
			row.Volume,
			row.Duration,
		})
	}
	writer.Flush()

	// Upload to S3
	key := fmt.Sprintf("daily/%s.csv", symbol)
	_, err := s.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
		Body:   bytes.NewReader(buf.Bytes()),
	})
	if err != nil {
		return fmt.Errorf("failed to upload daily data for %s: %w", symbol, err)
	}

	return nil
}

// StoreMinuteData stores minute-by-minute stock data to S3 as CSV
func (s *S3Storage) StoreMinuteData(symbol string, data SliceMinutePriceStruct) error {
	// Group data by month
	monthlyData := make(map[string]SliceMinutePriceStruct)
	
	for _, row := range data {
		// Extract year and month from DateTime (YYYYMMDDHHMMSS format)
		if len(row.DateTime) >= 6 {
			yearMonth := row.DateTime[:6] // YYYYMM
			monthlyData[yearMonth] = append(monthlyData[yearMonth], row)
		}
	}

	// Store each month's data separately
	for yearMonth, monthData := range monthlyData {
		// Sort data by datetime
		sort.Slice(monthData, func(i, j int) bool {
			return monthData[i].DateTime < monthData[j].DateTime
		})

		// Create CSV content
		var buf bytes.Buffer
		writer := csv.NewWriter(&buf)
		
		// Write header
		writer.Write([]string{"DateTime", "Open", "High", "Low", "Close", "Volume", "Duration"})
		
		// Write data
		for _, row := range monthData {
			writer.Write([]string{
				row.DateTime,
				row.Open,
				row.High,
				row.Low,
				row.Close,
				row.Volume,
				row.Duration,
			})
		}
		writer.Flush()

		// Upload to S3
		key := fmt.Sprintf("minute/%s/%s.csv", symbol, yearMonth)
		_, err := s.uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(s.bucketName),
			Key:    aws.String(key),
			Body:   bytes.NewReader(buf.Bytes()),
		})
		if err != nil {
			return fmt.Errorf("failed to upload minute data for %s/%s: %w", symbol, yearMonth, err)
		}
	}

	return nil
}

// LoadDailyData loads daily stock data from S3 CSV
func (s *S3Storage) LoadDailyData(symbol string) (SlicePriceStruct, error) {
	key := fmt.Sprintf("daily/%s.csv", symbol)
	
	// Get object from S3
	result, err := s.s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download daily data for %s: %w", symbol, err)
	}
	defer result.Body.Close()

	// Read all data
	body, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read daily data for %s: %w", symbol, err)
	}

	// Parse CSV
	reader := csv.NewReader(bytes.NewReader(body))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV for %s: %w", symbol, err)
	}

	var data SlicePriceStruct
	// Skip header row
	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) >= 7 {
			data = append(data, PriceStruct{
				Date:     record[0],
				Open:     record[1],
				High:     record[2],
				Low:      record[3],
				Close:    record[4],
				Volume:   record[5],
				Duration: record[6],
			})
		}
	}

	return data, nil
}

// LoadMinuteData loads minute-by-minute stock data from S3 CSV for a specific month
func (s *S3Storage) LoadMinuteData(symbol, yearMonth string) (SliceMinutePriceStruct, error) {
	key := fmt.Sprintf("minute/%s/%s.csv", symbol, yearMonth)
	
	// Get object from S3
	result, err := s.s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download minute data for %s/%s: %w", symbol, yearMonth, err)
	}
	defer result.Body.Close()

	// Read all data
	body, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read minute data for %s/%s: %w", symbol, yearMonth, err)
	}

	// Parse CSV
	reader := csv.NewReader(bytes.NewReader(body))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV for %s/%s: %w", symbol, yearMonth, err)
	}

	var data SliceMinutePriceStruct
	// Skip header row
	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) >= 7 {
			data = append(data, MinutePriceStruct{
				DateTime: record[0],
				Open:     record[1],
				High:     record[2],
				Low:      record[3],
				Close:    record[4],
				Volume:   record[5],
				Duration: record[6],
			})
		}
	}

	return data, nil
}

// MergeAndStoreData merges new data with existing data, removes overlaps, and stores back to S3
func (s *S3Storage) MergeAndStoreData(symbol string, newData interface{}, dataType string) error {
	switch dataType {
	case "daily":
		newDailyData := newData.(SlicePriceStruct)
		existingData, err := s.LoadDailyData(symbol)
		if err != nil && !strings.Contains(err.Error(), "NoSuchKey") {
			return fmt.Errorf("failed to load existing daily data: %w", err)
		}

		// Merge and remove duplicates
		mergedData := s.mergeDailyData(existingData, newDailyData)
		return s.StoreDailyData(symbol, mergedData)

	case "minute":
		newMinuteData := newData.(SliceMinutePriceStruct)
		// Group new data by month
		monthlyNewData := make(map[string]SliceMinutePriceStruct)
		for _, row := range newMinuteData {
			if len(row.DateTime) >= 6 {
				yearMonth := row.DateTime[:6]
				monthlyNewData[yearMonth] = append(monthlyNewData[yearMonth], row)
			}
		}

		// Process each month
		for yearMonth, monthData := range monthlyNewData {
			existingData, err := s.LoadMinuteData(symbol, yearMonth)
			if err != nil && !strings.Contains(err.Error(), "NoSuchKey") {
				return fmt.Errorf("failed to load existing minute data for %s/%s: %w", symbol, yearMonth, err)
			}

			// Merge and remove duplicates
			mergedData := s.mergeMinuteData(existingData, monthData)
			
			// Store merged data
			if err := s.StoreMinuteData(symbol, mergedData); err != nil {
				return fmt.Errorf("failed to store merged minute data for %s/%s: %w", symbol, yearMonth, err)
			}
		}
		return nil

	default:
		return fmt.Errorf("unsupported data type: %s", dataType)
	}
}

// mergeDailyData merges two daily data slices and removes duplicates
func (s *S3Storage) mergeDailyData(existing, new SlicePriceStruct) SlicePriceStruct {
	// Create a map to track existing dates
	dateMap := make(map[string]PriceStruct)
	
	// Add existing data
	for _, row := range existing {
		dateMap[row.Date] = row
	}
	
	// Add/overwrite with new data
	for _, row := range new {
		dateMap[row.Date] = row
	}
	
	// Convert back to slice
	var merged SlicePriceStruct
	for _, row := range dateMap {
		merged = append(merged, row)
	}
	
	// Sort by date
	sort.Slice(merged, func(i, j int) bool {
		return merged[i].Date < merged[j].Date
	})
	
	return merged
}

// mergeMinuteData merges two minute data slices and removes duplicates
func (s *S3Storage) mergeMinuteData(existing, new SliceMinutePriceStruct) SliceMinutePriceStruct {
	// Create a map to track existing datetimes
	datetimeMap := make(map[string]MinutePriceStruct)
	
	// Add existing data
	for _, row := range existing {
		datetimeMap[row.DateTime] = row
	}
	
	// Add/overwrite with new data
	for _, row := range new {
		datetimeMap[row.DateTime] = row
	}
	
	// Convert back to slice
	var merged SliceMinutePriceStruct
	for _, row := range datetimeMap {
		merged = append(merged, row)
	}
	
	// Sort by datetime
	sort.Slice(merged, func(i, j int) bool {
		return merged[i].DateTime < merged[j].DateTime
	})
	
	return merged
} 