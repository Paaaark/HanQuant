# Historical Data Management Tool

This tool manages historical stock data by fetching data from KIS API and storing it in AWS S3 for efficient retrieval.

## Features

- **Daily Data**: Fetch and store 5 years of daily stock price data
- **Minute Data**: Fetch and store 1 year of minute-by-minute stock price data
- **S3 Storage**: Organized CSV storage in S3 with automatic deduplication
- **Bulk Operations**: Process multiple symbols at once
- **Data Loading**: Load and export data from S3

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

#### 1. Fetch Daily Data for a Single Symbol

```bash
./historical_data -action fetch-daily -symbol 005930 -from 20200101 -to 20241231
```

#### 2. Fetch Minute Data for a Single Symbol

```bash
./historical_data -action fetch-minute -symbol 005930 -from 20240101 -to 20241231
```

#### 3. Load Daily Data from S3

```bash
# Output to stdout
./historical_data -action load-daily -symbol 005930

# Output to file
./historical_data -action load-daily -symbol 005930 -output daily_005930.json
```

#### 4. Load Minute Data from S3

```bash
# Output to stdout
./historical_data -action load-minute -symbol 005930 -month 202401

# Output to file
./historical_data -action load-minute -symbol 005930 -month 202401 -output minute_005930_202401.json
```

#### 5. Bulk Fetch Daily Data

Create a file `symbols.txt` with one symbol per line:

```
005930
000660
035420
```

Then run:

```bash
./historical_data -action bulk-daily -symbols symbols.txt
```

#### 6. Bulk Fetch Minute Data

```bash
./historical_data -action bulk-minute -symbols symbols.txt
```

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

- **API Rate Limits**: The tool handles KIS API rate limits gracefully
- **Network Errors**: Automatic retry for transient network issues
- **Data Validation**: Validates data before storing
- **Duplicate Handling**: Automatically merges and deduplicates data

## Performance Considerations

- **Batch Processing**: Use bulk operations for multiple symbols
- **S3 Optimization**: Data is organized by month for efficient retrieval
- **Memory Usage**: Large datasets are processed in chunks
- **API Limits**: Respects KIS API rate limits

## Troubleshooting

### Common Issues

1. **S3 Access Denied**: Check AWS credentials and bucket permissions
2. **KIS API Errors**: Verify API keys and access token
3. **Network Timeouts**: Check internet connectivity and API endpoints
4. **Memory Issues**: Process symbols in smaller batches

### Logs

The tool provides detailed logging for debugging:

```bash
./historical_data -action fetch-daily -symbol 005930 -from 20240101 -to 20240131
# Output: Fetching daily data for 005930...
# Output: Successfully fetched and stored daily data for 005930
```

## Integration with TimescaleDB

This tool is designed to work with TimescaleDB for real-time trading:

1. **Historical Data**: Store 5 years of daily data and 1 year of minute data in S3
2. **Real-time Data**: Load required data into TimescaleDB for fast access
3. **Data Synchronization**: Keep S3 and TimescaleDB in sync

## Future Enhancements

- **Incremental Updates**: Only fetch new data since last update
- **Data Compression**: Compress CSV files for storage efficiency
- **Parallel Processing**: Process multiple symbols concurrently
- **Data Validation**: Enhanced validation and quality checks
- **Monitoring**: Integration with monitoring and alerting systems
