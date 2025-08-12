package data

import (
	"time"
)

// BacktestParams holds the parameters for running a backtest
type BacktestParams struct {
	From      string `json:"from"`      // Start date in YYYYMMDD format
	To        string `json:"to"`        // End date in YYYYMMDD format
	SMA_short int    `json:"sma_short"` // Short SMA window
	SMA_long  int    `json:"sma_long"`  // Long SMA window
}

// Portfolio represents the current state of the portfolio
type Portfolio struct {
	Cash      float64            `json:"cash"`      // Available cash
	Positions map[string]int     `json:"positions"` // Stock code -> quantity
	Values    map[string]float64 `json:"values"`    // Stock code -> current value
	Total     float64            `json:"total"`     // Total portfolio value
}

// Trade represents a single trade execution
type Trade struct {
	Date      string  `json:"date"`      // Trade date in YYYY-MM-DD format
	Symbol    string  `json:"symbol"`    // Stock code
	Side      string  `json:"side"`      // "BUY" or "SELL"
	Quantity  int     `json:"quantity"`  // Number of shares
	Price     float64 `json:"price"`     // Execution price
	Value     float64 `json:"value"`     // Total trade value
	Portfolio float64 `json:"portfolio"` // Portfolio value after trade
}

// BacktestResult contains the complete backtest results
type BacktestResult struct {
	Params           BacktestParams         `json:"params"`
	PortfolioHistory []PortfolioSnapshot    `json:"portfolio_history"`
	Trades           []Trade                `json:"trades"`
	Metrics          BacktestMetrics        `json:"metrics"`
	Universe         []string               `json:"universe"`
}

// PortfolioSnapshot represents portfolio state at a specific date
type PortfolioSnapshot struct {
	Date     string    `json:"date"`
	Portfolio Portfolio `json:"portfolio"`
}

// BacktestMetrics contains performance metrics
type BacktestMetrics struct {
	TotalReturn    float64 `json:"total_return"`    // Total return percentage
	TotalPnL       float64 `json:"total_pnl"`       // Total profit/loss
	SharpeRatio    float64 `json:"sharpe_ratio"`    // Sharpe ratio
	MaxDrawdown    float64 `json:"max_drawdown"`    // Maximum drawdown
	TotalTrades    int     `json:"total_trades"`    // Total number of trades
	WinRate        float64 `json:"win_rate"`        // Percentage of profitable trades
	AvgTradeReturn float64 `json:"avg_trade_return"` // Average return per trade
}

// StockData represents historical stock data for backtesting
type StockData struct {
	Symbol string    `json:"symbol"`
	Date   time.Time `json:"date"`
	Open   float64   `json:"open"`
	High   float64   `json:"high"`
	Low    float64   `json:"low"`
	Close  float64   `json:"close"`
	Volume int64     `json:"volume"`
}

// SMAStrategy represents the SMA crossover strategy state
type SMAStrategy struct {
	ShortWindow int                    `json:"short_window"`
	LongWindow  int                    `json:"long_window"`
	SMAs        map[string]SMACalculation `json:"smas"`
}

// SMACalculation holds SMA values for a stock
type SMACalculation struct {
	ShortSMA []float64 `json:"short_sma"`
	LongSMA  []float64 `json:"long_sma"`
	Dates    []string  `json:"dates"`
}
