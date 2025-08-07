# Historical Data Management Tool

This tool manages historical stock data by fetching data from KIS API and storing it in AWS S3 for efficient retrieval.

## Features

- **Daily Data**: Fetch and store 10 years of daily stock price data with intelligent gap detection
- **Minute Data**: Fetch and store 1 year of minute-by-minute stock price data
- **S3 Storage**: Organized CSV storage in S3 with automatic deduplication
- **Bulk Operations**: Process multiple symbols at once with intelligent data completeness checking
- **Data Loading**: Load and export data from S3
- **Smart Fetching**: Intelligent heuristics to fetch only missing data, avoiding redundant API calls

## Prerequisites

### Environment Variables

Set the following environment variables:

```bash
# KIS API Configuration
export KIS_APP_KEY="your_kis_app_key"
export KIS_APP_SECRET="your_kis_app_secret"
export KIS_ACCESS_TOKEN="your_kis_access_token"

# AWS S3 Configuration
export S3_BUCKET_NAME="your_s3_bucket_name"
export AWS_REGION="ap-northeast-2"  # or your preferred region
export AWS_ACCESS_KEY_ID="your_aws_access_key"
export AWS_SECRET_ACCESS_KEY="your_aws_secret_key"
```

### S3 Bucket Structure

The tool organizes data in the following structure:

```
s3://your-bucket/
├── daily/
│   ├── 005930.csv
│   ├── 000660.csv
│   └── ...
└── minute/
    ├── 005930/
    │   ├── 202401.csv
    │   ├── 202402.csv
    │   └── ...
    └── 000660/
        ├── 202401.csv
        └── ...
```

## Usage

### Build the Tool

```bash
go build -o historical_data cmd/historical_data/main.go
```

### Commands

#### 1. Fetch All Daily Data (10 Years) - Most Used

**How it works:**

- Loads all stock symbols from `stock_listings.csv`
- Calculates a 10-year date range from the specified `-to` date
- **Uses intelligent heuristics** to analyze existing data and identify gaps
- Only fetches missing data ranges, avoiding redundant API calls
- Processes each symbol individually with rate limiting (50ms between API calls)
- Splits large date ranges into 90-day chunks with 5-day overlap to handle API limits
- Merges new data with existing data in S3, removing duplicates
- Provides detailed progress logging and statistics (success, skipped, errors)

**Example:**

```bash
./historical_data -action fetch-all-daily-data -to 20241231
```

**What happens:**

1. If `-to` is `20241231`, it targets data from `20141231` to `20241231` (10 years)
2. **Intelligent gap detection** - analyzes existing data to find missing date ranges
3. **Skips complete symbols** - if a symbol already has sufficient data coverage, it's skipped entirely
4. **Fetches only gaps** - for incomplete symbols, only fetches the specific missing date ranges
5. Each symbol is processed with retry logic for rate limiting (up to 3 retries with exponential backoff)
6. Data is stored in S3 with automatic deduplication

**Heuristics used:**

- **Data completeness check**: Calculates if existing data covers 80% of expected trading days
- **Gap analysis**: Identifies specific date ranges where data is missing
- **Efficient fetching**: Only makes API calls for missing data, not entire ranges
- **Smart skipping**: Completely skips symbols that already have complete coverage
- **Newest-to-oldest fetching**: Fetches data from newest to oldest to optimize for early termination
- **Early termination**: Stops fetching when a chunk returns 0 records for past dates, indicating the end of historical data
- **Future date handling**: Automatically adjusts future dates to current date and skips future chunks
- **Smart chunking**: Optimizes chunk size for small date ranges and handles edge cases

**Benefits over force fetch:**

- **Massive efficiency gains**: Skips symbols with complete data, fetches only gaps for others
- **Early termination**: Stops fetching when no more historical data exists, avoiding unnecessary API calls
- **Optimized order**: Fetches newest data first, which is more likely to exist and helps identify data boundaries quickly
- **Reduced API usage**: Dramatically fewer API calls, especially for symbols with recent data
- **Faster execution**: Much quicker for subsequent runs
- **Cost effective**: Minimizes API costs and processing time
- **Respectful of existing data**: Doesn't waste resources on data you already have

#### 2. Fetch Today's Daily Data for All Symbols - Most Used

**How it works:**

- Loads all stock symbols from `stock_listings.csv`
- Processes symbols in batches of 30 (KIS API limit for `GetMultipleStockSnapshot`)
- **Checks for existing data** for the target date to avoid duplicates
- Uses efficient batch API calls (30 symbols per call instead of individual calls)
- Converts snapshot data to daily format and stores in S3
- Provides detailed statistics (success, skipped, errors)

**Example:**

```bash
./historical_data -action fetch-all-daily-data-today -to 20241231
```

**What happens:**

1. Fetches current day's data for all symbols using efficient batch processing
2. **Skips symbols that already have data for the target date** (prevents overwrites)
3. Uses `GetMultipleStockSnapshot` API for efficiency (30 symbols per call)
4. Converts real-time snapshot data to daily OHLCV format
5. Stores data with automatic merging and deduplication

**Efficiency benefits:**

- Reduces API calls by ~97% (30 symbols per call vs 1 symbol per call)
- Much faster execution for daily updates
- Prevents unnecessary API calls for existing data

#### 3. Fetch Daily Data for a Single Symbol

```bash
./historical_data -action fetch-daily -symbol 005930 -from 20200101 -to 20241231
```

#### 4. Fetch Minute Data for a Single Symbol

```bash
./historical_data -action fetch-minute -symbol 005930 -from 20240101 -to 20241231
```

#### 5. Load Daily Data from S3

```bash
# Output to stdout
./historical_data -action load-daily -symbol 005930

# Output to file
./historical_data -action load-daily -symbol 005930 -output daily_005930.json
```

#### 6. Load Minute Data from S3

```bash
# Output to stdout
./historical_data -action load-minute -symbol 005930 -month 202401

# Output to file
./historical_data -action load-minute -symbol 005930 -month 202401 -output minute_005930_202401.json
```

#### 7. Bulk Fetch Daily Data (5 Years)

**How it works:**

- Fetches 5 years of daily data for multiple symbols
- **Checks data completeness before fetching** - skips symbols with sufficient data
- Verifies data completeness after fetching and reports incomplete symbols
- Uses intelligent fetching (non-force mode) to avoid redundant API calls

```bash
./historical_data -action bulk-daily -symbols symbols.txt
```

#### 8. Bulk Fetch Minute Data

```bash
./historical_data -action bulk-minute -symbols symbols.txt
```

## Data Completeness Checking

The tool includes intelligent data completeness checking:

- **Expected trading days**: Calculates expected trading days (70% of calendar days)
- **Completeness threshold**: Considers data complete if 80% of expected trading days are covered
- **Pre-fetch checking**: Skips symbols with sufficient data to avoid redundant API calls
- **Post-fetch verification**: Reports symbols with incomplete data for manual review

## Data Format

### Daily Data CSV Format

```csv
Date,Open,High,Low,Close,Volume,Duration
20240101,75000,75500,74800,75200,1234567,D
20240102,75200,75800,75000,75600,2345678,D
```

### Minute Data CSV Format

```csv
DateTime,Open,High,Low,Close,Volume,Duration
20240101090000,75000,75100,74900,75050,12345,M
20240101090100,75050,75200,75000,75150,23456,M
```

### JSON Output Format

The tool outputs data in JSON format with user-friendly field names:

```json
[
  {
    "Date": "20240101",
    "Open": "75000",
    "High": "75500",
    "Low": "74800",
    "Close": "75200",
    "Volume": "1234567",
    "Duration": "D"
  }
]
```

## Error Handling

- **API Rate Limits**: The tool handles KIS API rate limits gracefully with exponential backoff
- **Network Errors**: Automatic retry for transient network issues (up to 3 retries)
- **Data Validation**: Validates data before storing
- **Duplicate Handling**: Automatically merges and deduplicates data
- **Intelligent Fetching**: Analyzes existing data to fetch only missing gaps, avoiding redundant API calls

## Performance Considerations

- **Batch Processing**: Use bulk operations for multiple symbols
- **Efficient Daily Updates**: `fetch-all-daily-data-today` uses batch API calls (30 symbols per call)
- **S3 Optimization**: Data is organized by month for efficient retrieval
- **Memory Usage**: Large datasets are processed in chunks
- **API Limits**: Respects KIS API rate limits (50ms between calls)
- **Data Chunking**: Large date ranges are split into manageable chunks (90 days for daily data)

## Troubleshooting

### Common Issues

1. **S3 Access Denied**: Check AWS credentials and bucket permissions
2. **KIS API Errors**: Verify API keys and access token
3. **Network Timeouts**: Check internet connectivity and API endpoints
4. **Memory Issues**: Process symbols in smaller batches
5. **Incomplete Data**: Check logs for symbols with insufficient data coverage

### Logs

The tool provides detailed logging for debugging:

```bash
./historical_data -action fetch-all-daily-data -to 20241231
# Output: Loaded 100 stock symbols from stock_listings.csv
# Output: Fetching 10 years of daily data from 20141231 to 20241231 for 100 symbols
# Output: Using intelligent heuristics to fetch only missing data gaps
# Output: Processing symbol 1/100: 005930
# Output: Data for 005930 is already complete (2520 records), skipping fetch
# Output: Processing symbol 2/100: 000660
# Output: Fetching data for 000660 in 2 ranges: 1500 existing records
# Output: Range 1/2: 20241231 to 20240630 (newest to oldest)
# Output:   Chunk 1/3: 20241231 to 20241001
# Output:   Received 65 daily records for chunk 1
# Output:   Chunk 2/3: 20240930 to 20240801
# Output:   Received 0 daily records for chunk 2
# Output: Chunk 2/3 returned 0 records for past dates (20240930-20240801), reached end of historical data for 000660
# Output: Skipping remaining ranges for 000660 (reached end of historical data)
# Output: Successfully processed 000660 (2/100)

# Example with future dates:
./historical_data -action fetch-daily -symbol 005930 -from 20250801 -to 20250831
# Output: Warning: From date 20250801 is in the future, adjusting to today
# Output: Warning: To date 20250831 is in the future, adjusting to today
# Output: Fetching data for 005930 in 1 ranges: 0 existing records
# Output: Range 1/1: 20250806 to 20250806 (newest to oldest)
# Output:   Chunk 1/1: 20250806 to 20250806
# Output:   Received 0 daily records for chunk 1
# Output: Chunk 1/1 returned 0 records for future dates (20250806-20250806), skipping
# Output: No new daily data received for 005930
```

## Integration with TimescaleDB

This tool is designed to work with TimescaleDB for real-time trading:

1. **Historical Data**: Store 10 years of daily data and 1 year of minute data in S3
2. **Real-time Data**: Load required data into TimescaleDB for fast access
3. **Data Synchronization**: Keep S3 and TimescaleDB in sync

## Future Enhancements

- **Incremental Updates**: Only fetch new data since last update
- **Data Compression**: Compress CSV files for storage efficiency
- **Parallel Processing**: Process multiple symbols concurrently
- **Data Validation**: Enhanced validation and quality checks
- **Monitoring**: Integration with monitoring and alerting systems
