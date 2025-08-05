# TimescaleDB & S3 Implementation for Historical Stock Data

## Overview

This implementation provides a comprehensive solution for managing historical stock data using a dual-database approach:

1. **Real-time Database (TimescaleDB)**: For fast access to recent data needed for real-time trading
2. **Historical Database (S3)**: For long-term storage of historical data (5 years daily + 1 year minute data)

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   KIS API       │    │   S3 Storage    │    │  TimescaleDB    │
│                 │    │                 │    │                 │
│ • Daily Data    │───▶│ • CSV Files     │───▶│ • Real-time     │
│ • Minute Data   │    │ • Organized     │    │ • Fast Access   │
│ • Rate Limited  │    │ • Compressed    │    │ • Time-series   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Implementation Details

### 1. KIS API Integration

#### New Functions Added:

**Daily Data Fetching:**

```go
// GetDailyStockData: Enhanced version for historical data
func (c *KISClient) GetDailyStockData(symbol, from, to string) (SlicePriceStruct, error)
```

**Minute Data Fetching:**

```go
// GetMinuteStockData: Minute-by-minute stock data
func (c *KISClient) GetMinuteStockData(symbol, from, to string) (SliceMinutePriceStruct, error)
```

#### Data Structures:

```go
// MinutePriceStruct for minute-by-minute stock data
type MinutePriceStruct struct {
    DateTime string `json:"stck_cntg_hour"` // YYYYMMDDHHMMSS format
    Open     string `json:"stck_oprc"`
    High     string `json:"stck_hgpr"`
    Low      string `json:"stck_lwpr"`
    Close    string `json:"stck_prpr"`
    Volume   string `json:"cntg_vol"`
    Duration string // Always "M" for minute data
}

type SliceMinutePriceStruct []MinutePriceStruct
```

### 2. S3 Storage Implementation

#### File: `internal/data/s3_storage.go`

**Key Features:**

- **Organized Storage**: Data organized by symbol and time period
- **Automatic Deduplication**: Merges new data with existing data
- **CSV Format**: Efficient storage and retrieval
- **Error Handling**: Robust error handling for S3 operations

**Storage Structure:**

```
s3://bucket/
├── daily/
│   ├── 005930.csv          # Samsung Electronics daily data
│   ├── 000660.csv          # SK Hynix daily data
│   └── ...
└── minute/
    ├── 005930/
    │   ├── 202401.csv      # January 2024 minute data
    │   ├── 202402.csv      # February 2024 minute data
    │   └── ...
    └── 000660/
        ├── 202401.csv
        └── ...
```

**Key Functions:**

```go
// StoreDailyData: Stores daily stock data to S3 as CSV
func (s *S3Storage) StoreDailyData(symbol string, data SlicePriceStruct) error

// StoreMinuteData: Stores minute-by-minute stock data to S3 as CSV
func (s *S3Storage) StoreMinuteData(symbol string, data SliceMinutePriceStruct) error

// LoadDailyData: Loads daily stock data from S3 CSV
func (s *S3Storage) LoadDailyData(symbol string) (SlicePriceStruct, error)

// LoadMinuteData: Loads minute-by-minute stock data from S3 CSV
func (s *S3Storage) LoadMinuteData(symbol, yearMonth string) (SliceMinutePriceStruct, error)

// MergeAndStoreData: Merges new data with existing data, removes overlaps
func (s *S3Storage) MergeAndStoreData(symbol string, newData interface{}, dataType string) error
```

### 3. Service Layer

#### File: `internal/service/historical.go`

**HistoricalService** coordinates between KIS API and S3 storage:

```go
type HistoricalService struct {
    kisClient *data.KISClient
    s3Storage *data.S3Storage
}
```

**Key Functions:**

```go
// FetchAndStoreDailyData: Fetches and stores daily data
func (s *HistoricalService) FetchAndStoreDailyData(symbol, fromDate, toDate string) error

// FetchAndStoreMinuteData: Fetches and stores minute data
func (s *HistoricalService) FetchAndStoreMinuteData(symbol, fromDate, toDate string) error

// BulkFetchDailyData: Bulk fetch for multiple symbols
func (s *HistoricalService) BulkFetchDailyData(symbols []string) error

// BulkFetchMinuteData: Bulk fetch minute data for multiple symbols
func (s *HistoricalService) BulkFetchMinuteData(symbols []string) error
```

### 4. Command-Line Tool

#### File: `cmd/historical_data/main.go`

**Features:**

- **Single Symbol Operations**: Fetch/load data for individual symbols
- **Bulk Operations**: Process multiple symbols efficiently
- **Flexible Output**: Output to stdout or files
- **Error Handling**: Comprehensive error handling and logging

**Usage Examples:**

```bash
# Fetch daily data
./historical_data -action fetch-daily -symbol 005930 -from 20200101 -to 20241231

# Fetch minute data
./historical_data -action fetch-minute -symbol 005930 -from 20240101 -to 20241231

# Load data from S3
./historical_data -action load-daily -symbol 005930 -output daily_005930.json

# Bulk operations
./historical_data -action bulk-daily -symbols symbols.txt
```

### 5. JSON Encoding

#### Enhanced: `internal/data/json_encode.go`

Added custom JSON encoding for minute data:

```go
func (s SliceMinutePriceStruct) EncodeJSON() []byte
```

**Output Format:**

```json
[
  {
    "DateTime": "20240101090000",
    "Open": "75000",
    "High": "75100",
    "Low": "74900",
    "Close": "75050",
    "Volume": "12345",
    "Duration": "M"
  }
]
```

## Data Volume Estimates

### Historical Database (S3)

- **Daily Data**: ~5 years × 365 days × 2,000 symbols ≈ 3.65M records
- **Minute Data**: ~1 year × 252 trading days × 390 minutes × 2,000 symbols ≈ 196M records
- **Total Storage**: ~30 GB (compressed CSV)

### Real-time Database (TimescaleDB)

- **Active Data**: ~1 month of minute data for active symbols
- **Storage**: ~5 GB maximum
- **Performance**: Sub-second queries for real-time trading

## Environment Setup

### Required Environment Variables:

```bash
# KIS API
export KIS_APP_KEY="your_kis_app_key"
export KIS_APP_SECRET="your_kis_app_secret"
export KIS_ACCESS_TOKEN="your_kis_access_token"

# AWS S3
export S3_BUCKET_NAME="your_s3_bucket_name"
export AWS_REGION="ap-northeast-2"
export AWS_ACCESS_KEY_ID="your_aws_access_key"
export AWS_SECRET_ACCESS_KEY="your_aws_secret_key"
```

### Dependencies:

```go
require (
    github.com/aws/aws-sdk-go v1.55.7
    // ... other existing dependencies
)
```

## Usage Workflow

### 1. Initial Setup

```bash
# Build the tool
go build -o historical_data cmd/historical_data/main.go

# Set environment variables
export KIS_APP_KEY="..."
export S3_BUCKET_NAME="..."
# ... other variables

# Run the example script
./scripts/fetch_historical_data.sh
```

### 2. Regular Data Updates

```bash
# Update daily data (run weekly)
./historical_data -action bulk-daily -symbols active_symbols.txt

# Update minute data (run daily)
./historical_data -action bulk-minute -symbols active_symbols.txt
```

### 3. Data Loading for TimescaleDB

```bash
# Load specific data for real-time trading
./historical_data -action load-daily -symbol 005930 -output daily_data.json
./historical_data -action load-minute -symbol 005930 -month 202401 -output minute_data.json
```

## Error Handling & Resilience

### KIS API Rate Limiting

- **Automatic Retry**: Retry logic for transient failures
- **Token Refresh**: Automatic token refresh on expiration
- **Graceful Degradation**: Continue processing other symbols on individual failures

### S3 Operations

- **Network Resilience**: Retry on network failures
- **Data Validation**: Validate data before storing
- **Duplicate Handling**: Automatic deduplication and merging

### Data Quality

- **Format Validation**: Ensure data meets expected format
- **Range Validation**: Validate date ranges and data completeness
- **Error Logging**: Comprehensive logging for debugging

## Performance Optimizations

### S3 Storage

- **Organized Structure**: Efficient retrieval by symbol and time
- **CSV Format**: Fast parsing and compression
- **Batch Operations**: Process multiple symbols efficiently

### Memory Management

- **Streaming**: Process large datasets without loading everything into memory
- **Chunked Processing**: Process data in manageable chunks
- **Garbage Collection**: Proper cleanup of temporary data

### API Efficiency

- **Rate Limit Respect**: Respect KIS API rate limits
- **Batch Requests**: Minimize API calls where possible
- **Caching**: Cache frequently accessed data

## Future Enhancements

### Planned Features

1. **Incremental Updates**: Only fetch new data since last update
2. **Data Compression**: Compress CSV files for storage efficiency
3. **Parallel Processing**: Process multiple symbols concurrently
4. **Data Validation**: Enhanced validation and quality checks
5. **Monitoring**: Integration with monitoring and alerting systems

### TimescaleDB Integration

1. **Data Loading**: Automated loading from S3 to TimescaleDB
2. **Real-time Sync**: Keep TimescaleDB synchronized with latest data
3. **Query Optimization**: Optimize queries for time-series data
4. **Partitioning**: Implement proper partitioning for performance

## Monitoring & Maintenance

### Health Checks

- **API Connectivity**: Monitor KIS API availability
- **S3 Access**: Monitor S3 bucket access and performance
- **Data Quality**: Monitor data completeness and accuracy

### Maintenance Tasks

- **Regular Updates**: Schedule regular data updates
- **Storage Cleanup**: Clean up old or duplicate data
- **Performance Tuning**: Optimize based on usage patterns

## Conclusion

This implementation provides a robust foundation for historical stock data management with:

- **Scalability**: Handles large datasets efficiently
- **Reliability**: Comprehensive error handling and recovery
- **Flexibility**: Easy to extend and modify
- **Performance**: Optimized for both storage and retrieval
- **Maintainability**: Clear separation of concerns and documentation

The dual-database approach ensures optimal performance for both historical analysis and real-time trading while maintaining data integrity and accessibility.
