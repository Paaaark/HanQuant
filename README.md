# HanQuant

[![Frontend Repository](https://img.shields.io/badge/Frontend_HanQuant-Open%20Repo-blue?style=for-the-badge&logo=github)](https://github.com/Paaaark/hanquant_web)

[![Go](https://img.shields.io/badge/Go-1.24.3+-blue?style=flat-square&logo=go)](https://golang.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-13+-blue?style=flat-square&logo=postgresql)](https://www.postgresql.org/)
[![TimescaleDB](https://img.shields.io/badge/TimescaleDB-2.0+-blue?style=flat-square&logo=timescaledb)](https://www.timescale.com/)
[![AWS S3](https://img.shields.io/badge/AWS_S3-Cloud_Storage-orange?style=flat-square&logo=amazon-aws)](https://aws.amazon.com/s3/)
[![AWS EC2](https://img.shields.io/badge/AWS_EC2-Cloud_Deployment-orange?style=flat-square&logo=amazon-ec2)](https://aws.amazon.com/ec2/)
[![WebSocket](https://img.shields.io/badge/WebSocket-Real--time_Data-green?style=flat-square&logo=websocket)](https://websocket.org/)
[![JWT](https://img.shields.io/badge/JWT-Authentication-red?style=flat-square&logo=json-web-tokens)](https://jwt.io/)
[![KIS API](https://img.shields.io/badge/KIS%20API-Trading%20Integration-purple?style=flat-square)](https://securities.koreainvestment.com/)

> **HanQuant** is a sophisticated quantitative trading platform that provides real-time Korean stock market data, historical data management, and algorithmic trading capabilities. Built with Go, it integrates with Korea Investment & Securities (KIS) API, TimescaleDB for time-series data, and AWS S3 for scalable historical data storage.

## Key Features

### Real-Time Market Data

- **Live Stock Prices**: Real-time streaming of Korean stock market data via WebSocket
- **Market Rankings**: Top gainers, most traded, and highest market cap stocks
- **Index Tracking**: KOSPI, KOSDAQ, KOSPI200, and KRX100 indices
- **Multi-Symbol Snapshots**: Batch retrieval of multiple stock prices

### Historical Data Management

- **5 Years Daily Data**: Comprehensive daily price history for all Korean stocks
- **1 Year Minute Data**: High-frequency minute-by-minute price data
- **S3 Cloud Storage**: Scalable CSV-based storage with automatic deduplication
- **TimescaleDB Integration**: Time-series optimized database for real-time trading
- **Bulk Operations**: Efficient processing of multiple symbols simultaneously

### Secure Trading Infrastructure

- **JWT Authentication**: Secure user authentication and session management
- **Account Management**: Support for multiple KIS trading accounts per user
- **Encrypted Credentials**: AES-encrypted storage of sensitive API keys
- **Paper Trading**: Mock trading environment for strategy testing

### Portfolio & Order Management

- **Portfolio Tracking**: Real-time portfolio positions and P&L
- **Order Execution**: Market and limit order placement
- **Order History**: Comprehensive order tracking and status management
- **Account Summary**: Detailed account balance and asset information

### WebSocket Real-Time Updates

- **Live Price Feeds**: Sub-second updates during market hours
- **Subscription Management**: Dynamic symbol subscription/unsubscription
- **Broadcast System**: Efficient multi-client data distribution
- **Market Hours Detection**: Automatic trading session awareness

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   KIS API       │    │   AWS S3        │    │  TimescaleDB    │
│                 │    │                 │    │                 │
│ • Real-time     │───▶│ • Historical    │───▶│ • Time-series   │
│ • Historical    │    │ • CSV Storage   │    │ • Fast Access   │
│ • Trading       │    │ • Organized     │    │ • Real-time     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   WebSocket     │    │   PostgreSQL    │    │   REST API      │
│                 │    │                 │    │                 │
│ • Live Feeds    │    │ • User Data     │    │ • HTTP Endpoints│
│ • Subscriptions │    │ • Orders        │    │ • Authentication│
│ • Broadcasting  │    │ • Accounts      │    │ • Portfolio     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Quick Start

### Prerequisites

```bash
# Required environment variables
export KIS_APP_KEY="your_kis_app_key"
export KIS_APP_SECRET="your_kis_app_secret"
export KIS_ACCESS_TOKEN="your_kis_access_token"
export S3_BUCKET_NAME="your_s3_bucket"
export AWS_REGION="ap-northeast-2"
export AWS_ACCESS_KEY_ID="your_aws_key"
export AWS_SECRET_ACCESS_KEY="your_aws_secret"
```

### Installation

```bash
# Clone the repository
git clone https://github.com/Paaaark/hanquant.git
cd hanquant

# Install dependencies
go mod download

# Build the server
go build -o server cmd/server/main.go

# Build historical data tool
go build -o historical_data cmd/historical_data/main.go

# Run database migrations
go run cmd/server/main.go
```

### Running the Server

```bash
# Start the main server
./server

# The server will be available at http://localhost:8080
```

## API Endpoints

### Authentication

- `POST /auth/register` - User registration
- `POST /auth/login` - User login
- `POST /auth/refresh` - Token refresh

### Market Data

- `GET /prices/recent/{symbol}` - Recent stock price
- `GET /prices/historical/{symbol}?from=20240101&to=20241231&duration=D` - Historical data
- `GET /ranking/fluctuation` - Top gainers/losers
- `GET /ranking/volume` - Most traded stocks
- `GET /ranking/market-cap` - Highest market cap stocks
- `GET /snapshot/multstock` - Multiple stock snapshots
- `GET /index/{code}` - Market index data

### Trading

- `GET /portfolio` - User portfolio
- `POST /orders` - Place new order
- `GET /orders/{id}` - Get order details
- `POST /accounts` - Link KIS account
- `GET /accounts` - List user accounts

### WebSocket

- `WS /ws/stocks` - Real-time stock data feed

## Historical Data Management

### Fetch Historical Data

```bash
# Fetch 5 years of daily data
./historical_data -action fetch-daily -symbol 005930 -from 20200101 -to 20241231

# Fetch 1 year of minute data
./historical_data -action fetch-minute -symbol 005930 -from 20240101 -to 20241231

# Bulk fetch for multiple symbols
./historical_data -action bulk-daily -symbols symbols.txt
```

### Load Data from S3

```bash
# Load daily data
./historical_data -action load-daily -symbol 005930 -output daily_005930.json

# Load minute data for specific month
./historical_data -action load-minute -symbol 005930 -month 202401 -output minute_005930_202401.json
```

## Configuration

### Database Setup

The platform uses PostgreSQL with TimescaleDB extension for time-series data:

```sql
-- Enable TimescaleDB extension
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- Create hypertable for time-series data
SELECT create_hypertable('stock_prices', 'timestamp');
```

### S3 Storage Structure

```
s3://your-bucket/
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

## Security Features

- **JWT Token Authentication**: Secure stateless authentication
- **Password Hashing**: bcrypt-based password security
- **Encrypted Storage**: AES-256 encryption for sensitive data
- **Rate Limiting**: API rate limit compliance with KIS
- **Input Validation**: Comprehensive request validation

## Performance Optimizations

- **Connection Pooling**: Efficient database connection management
- **Caching**: In-memory caching for frequently accessed data
- **Batch Processing**: Bulk operations for multiple symbols
- **Compression**: S3 data compression for storage efficiency
- **Indexing**: Optimized database indexes for time-series queries

## Real-Time Features

### WebSocket Implementation

```javascript
// Connect to real-time feed
const ws = new WebSocket("ws://localhost:8080/ws/stocks");

// Subscribe to symbols
ws.send(
  JSON.stringify({
    type: "subscribe",
    tickers: ["005930", "000660", "035420"],
  })
);

// Receive real-time updates
ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log("Real-time update:", data);
};
```

## Testing & Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/handler
go test ./internal/service
```

### Development Tools

- **Historical Data Tool**: Command-line tool for data management
- **Stock Listings Converter**: Convert and process stock listings
- **KIS API Testing**: Test KIS API integration
- **Database Migrations**: Automated schema management

## Documentation

- [TimescaleDB Implementation](./TIMESCALEDB_IMPLEMENTATION.md) - Detailed database setup
- [Historical Data Management](./cmd/historical_data/README.md) - Data management guide
- [API Documentation](./docs/api.md) - Complete API reference

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

---
