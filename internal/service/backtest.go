package service

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"time"

	"github.com/Paaaark/hanquant/internal/data"
)

type BacktestService struct {
	stockService *StockService
	universe     []string
}

func NewBacktestService(stockService *StockService) *BacktestService {
	return &BacktestService{
		stockService: stockService,
		universe: []string{
			"005930", // Samsung
			"035420", // Naver
			"000660", // SK Hynix
			"086790", // Hana
			"071050", // Korea Investment Holdings
		},
	}
}

// RunSMABacktest executes the SMA crossover strategy backtest
func (s *BacktestService) RunSMABacktest(params data.BacktestParams) (*data.BacktestResult, error) {
	// Set default parameters
	if params.From == "" {
		params.From = "20170801"
	}
	if params.To == "" {
		params.To = "20250801"
	}
	if params.SMA_short == 0 {
		params.SMA_short = 20
	}
	if params.SMA_long == 0 {
		params.SMA_long = 50
	}

	// Validate parameters
	if err := s.validateParams(params); err != nil {
		return nil, err
	}

	// Parse dates
	fromDate, err := time.Parse("20060102", params.From)
	if err != nil {
		return nil, fmt.Errorf("invalid from date: %w", err)
	}

	toDate, err := time.Parse("20060102", params.To)
	if err != nil {
		return nil, fmt.Errorf("invalid to date: %w", err)
	}

	// Initialize portfolio
	portfolio := &data.Portfolio{
		Cash:      10000000, // 10M KRW starting capital
		Positions: make(map[string]int),
		Values:    make(map[string]float64),
		Total:     10000000,
	}

	// Fetch historical data for all stocks in universe
	stockData, err := s.fetchHistoricalData(params.From, params.To)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch historical data: %w", err)
	}

	// Calculate SMAs for all stocks
	smaData := s.calculateSMAs(stockData, params.SMA_short, params.SMA_long)

	// Run backtest
	trades, portfolioHistory := s.executeStrategy(stockData, smaData, portfolio, fromDate, toDate)

	// Calculate metrics
	metrics := s.calculateMetrics(portfolioHistory, trades)

	result := &data.BacktestResult{
		Params:           params,
		PortfolioHistory: portfolioHistory,
		Trades:           trades,
		Metrics:          metrics,
		Universe:         s.universe,
	}

	return result, nil
}

func (s *BacktestService) validateParams(params data.BacktestParams) error {
	if params.SMA_short >= params.SMA_long {
		return fmt.Errorf("short SMA window must be less than long SMA window")
	}

	if params.SMA_short < 1 || params.SMA_long < 1 {
		return fmt.Errorf("SMA windows must be positive integers")
	}

	today := time.Now()
	if params.To != "" {
		toDate, err := time.Parse("20060102", params.To)
		if err == nil && toDate.After(today) {
			return fmt.Errorf("end date cannot be in the future")
		}
	}

	return nil
}

func (s *BacktestService) fetchHistoricalData(from, to string) (map[string][]data.StockData, error) {
	stockData := make(map[string][]data.StockData)

	for _, symbol := range s.universe {
		data, err := s.stockService.GetHistoricalPrice(symbol, from, to, "D")
		if err != nil {
			return nil, fmt.Errorf("failed to fetch data for %s: %w", symbol, err)
		}

		// Convert to StockData format
		convertedData, err := s.convertToStockData(symbol, data)
		if err != nil {
			return nil, fmt.Errorf("failed to convert data for %s: %w", symbol, err)
		}

		stockData[symbol] = convertedData
	}

	return stockData, nil
}

func (s *BacktestService) convertToStockData(symbol string, rawData interface{}) ([]data.StockData, error) {
	var result []data.StockData

	// Try to convert from SlicePriceStruct
	if sliceData, ok := rawData.(data.SlicePriceStruct); ok {
		for _, price := range sliceData {
			// Parse date
			date, err := time.Parse("20060102", price.Date)
			if err != nil {
				continue // Skip invalid dates
			}

			// Parse numeric values
			open, _ := strconv.ParseFloat(price.Open, 64)
			high, _ := strconv.ParseFloat(price.High, 64)
			low, _ := strconv.ParseFloat(price.Low, 64)
			close, _ := strconv.ParseFloat(price.Close, 64)
			volume, _ := strconv.ParseInt(price.Volume, 10, 64)

			stockData := data.StockData{
				Symbol: symbol,
				Date:   date,
				Open:   open,
				High:   high,
				Low:    low,
				Close:  close,
				Volume: volume,
			}

			result = append(result, stockData)
		}
		return result, nil
	}

	// If not SlicePriceStruct, return empty result
	return result, nil
}

func (s *BacktestService) calculateSMAs(stockData map[string][]data.StockData, shortWindow, longWindow int) map[string]data.SMACalculation {
	smaData := make(map[string]data.SMACalculation)

	for symbol, stockDataSlice := range stockData {
		if len(stockDataSlice) < longWindow {
			continue
		}

		shortSMA := make([]float64, 0)
		longSMA := make([]float64, 0)
		dates := make([]string, 0)

		// Calculate short SMA
		for i := shortWindow - 1; i < len(stockDataSlice); i++ {
			sum := 0.0
			for j := 0; j < shortWindow; j++ {
				sum += stockDataSlice[i-j].Close
			}
			shortSMA = append(shortSMA, sum/float64(shortWindow))
		}

		// Calculate long SMA
		for i := longWindow - 1; i < len(stockDataSlice); i++ {
			sum := 0.0
			for j := 0; j < longWindow; j++ {
				sum += stockDataSlice[i-j].Close
			}
			longSMA = append(longSMA, sum/float64(longWindow))
		}

		// Align dates
		for i := longWindow - 1; i < len(stockDataSlice); i++ {
			dates = append(dates, stockDataSlice[i].Date.Format("2006-01-02"))
		}

		smaData[symbol] = data.SMACalculation{
			ShortSMA: shortSMA,
			LongSMA:  longSMA,
			Dates:    dates,
		}
	}

	return smaData
}

func (s *BacktestService) executeStrategy(stockData map[string][]data.StockData, smaData map[string]data.SMACalculation, portfolio *data.Portfolio, fromDate, toDate time.Time) ([]data.Trade, []data.PortfolioSnapshot) {
	var trades []data.Trade
	var portfolioHistory []data.PortfolioSnapshot

	// Get all trading dates
	tradingDates := s.getTradingDates(stockData, fromDate, toDate)

	for _, date := range tradingDates {
		// Update portfolio values
		s.updatePortfolioValues(portfolio, stockData, date)

		// Check for trading signals
		for _, symbol := range s.universe {
			if signal := s.checkSMASignal(smaData, symbol, date); signal != "" {
				trade := s.executeTrade(portfolio, symbol, signal, stockData, date)
				if trade != nil {
					trades = append(trades, *trade)
				}
			}
		}

		// Record portfolio snapshot
		snapshot := data.PortfolioSnapshot{
			Date:      date.Format("2006-01-02"),
			Portfolio: *portfolio,
		}
		portfolioHistory = append(portfolioHistory, snapshot)
	}

	return trades, portfolioHistory
}

func (s *BacktestService) getTradingDates(stockData map[string][]data.StockData, fromDate, toDate time.Time) []time.Time {
	// Get all unique dates from stock data
	dateSet := make(map[string]bool)
	var dates []time.Time

	for _, data := range stockData {
		for _, stock := range data {
			if stock.Date.After(fromDate) && stock.Date.Before(toDate) {
				dateStr := stock.Date.Format("2006-01-02")
				if !dateSet[dateStr] {
					dateSet[dateStr] = true
					dates = append(dates, stock.Date)
				}
			}
		}
	}

	// Sort dates
	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Before(dates[j])
	})

	return dates
}

func (s *BacktestService) updatePortfolioValues(portfolio *data.Portfolio, stockData map[string][]data.StockData, date time.Time) {
	portfolio.Total = portfolio.Cash

	for symbol, quantity := range portfolio.Positions {
		if quantity == 0 {
			continue
		}

		// Find price for this date
		if data, exists := stockData[symbol]; exists {
			for _, stock := range data {
				if stock.Date.Equal(date) {
					value := float64(quantity) * stock.Close
					portfolio.Values[symbol] = value
					portfolio.Total += value
					break
				}
			}
		}
	}
}

func (s *BacktestService) checkSMASignal(smaData map[string]data.SMACalculation, symbol string, date time.Time) string {
	calculation, exists := smaData[symbol]
	if !exists {
		return ""
	}

	dateStr := date.Format("2006-01-02")
	
	// Find index for this date
	var idx int = -1
	for i, d := range calculation.Dates {
		if d == dateStr {
			idx = i
			break
		}
	}

	if idx < 1 || idx >= len(calculation.ShortSMA) || idx >= len(calculation.LongSMA) {
		return ""
	}

	// Check for crossover
	prevShort := calculation.ShortSMA[idx-1]
	prevLong := calculation.LongSMA[idx-1]
	currShort := calculation.ShortSMA[idx]
	currLong := calculation.LongSMA[idx]

	// Golden cross: short SMA crosses above long SMA
	if prevShort <= prevLong && currShort > currLong {
		return "BUY"
	}

	// Death cross: short SMA crosses below long SMA
	if prevShort >= prevLong && currShort < currLong {
		return "SELL"
	}

	return ""
}

func (s *BacktestService) executeTrade(portfolio *data.Portfolio, symbol, signal string, stockData map[string][]data.StockData, date time.Time) *data.Trade {
	// Find current price
	var currentPrice float64
	if data, exists := stockData[symbol]; exists {
		for _, stock := range data {
			if stock.Date.Equal(date) {
				currentPrice = stock.Close
				break
			}
		}
	}

	if currentPrice == 0 {
		return nil
	}

	var trade *data.Trade

	if signal == "BUY" && portfolio.Cash > 0 {
		// Calculate position size (10% of portfolio per stock)
		positionValue := portfolio.Total * 0.1
		quantity := int(positionValue / currentPrice)
		
		if quantity > 0 {
			tradeValue := float64(quantity) * currentPrice
			if tradeValue <= portfolio.Cash {
				portfolio.Cash -= tradeValue
				portfolio.Positions[symbol] += quantity
				
				trade = &data.Trade{
					Date:      date.Format("2006-01-02"),
					Symbol:    symbol,
					Side:      "BUY",
					Quantity:  quantity,
					Price:     currentPrice,
					Value:     tradeValue,
					Portfolio: portfolio.Total,
				}
			}
		}
	} else if signal == "SELL" && portfolio.Positions[symbol] > 0 {
		quantity := portfolio.Positions[symbol]
		tradeValue := float64(quantity) * currentPrice
		
		portfolio.Cash += tradeValue
		portfolio.Positions[symbol] = 0
		delete(portfolio.Values, symbol)
		
		trade = &data.Trade{
			Date:      date.Format("2006-01-02"),
			Symbol:    symbol,
			Side:      "SELL",
			Quantity:  quantity,
			Price:     currentPrice,
			Value:     tradeValue,
			Portfolio: portfolio.Total,
		}
	}

	return trade
}

func (s *BacktestService) calculateMetrics(portfolioHistory []data.PortfolioSnapshot, trades []data.Trade) data.BacktestMetrics {
	if len(portfolioHistory) < 2 {
		return data.BacktestMetrics{}
	}

	initialValue := portfolioHistory[0].Portfolio.Total
	finalValue := portfolioHistory[len(portfolioHistory)-1].Portfolio.Total

	// Calculate returns
	totalReturn := (finalValue - initialValue) / initialValue * 100
	totalPnL := finalValue - initialValue

	// Calculate daily returns for Sharpe ratio
	var dailyReturns []float64
	for i := 1; i < len(portfolioHistory); i++ {
		prevValue := portfolioHistory[i-1].Portfolio.Total
		currValue := portfolioHistory[i].Portfolio.Total
		if prevValue > 0 {
			dailyReturn := (currValue - prevValue) / prevValue
			dailyReturns = append(dailyReturns, dailyReturn)
		}
	}

	// Calculate Sharpe ratio (assuming 0% risk-free rate)
	sharpeRatio := 0.0
	if len(dailyReturns) > 0 {
		meanReturn := s.calculateMean(dailyReturns)
		stdDev := s.calculateStdDev(dailyReturns, meanReturn)
		if stdDev > 0 {
			sharpeRatio = meanReturn / stdDev * math.Sqrt(252) // Annualized
		}
	}

	// Calculate max drawdown
	maxDrawdown := s.calculateMaxDrawdown(portfolioHistory)

	// Calculate trade metrics
	winRate, avgTradeReturn := s.calculateTradeMetrics(trades, initialValue)

	return data.BacktestMetrics{
		TotalReturn:    totalReturn,
		TotalPnL:       totalPnL,
		SharpeRatio:    sharpeRatio,
		MaxDrawdown:    maxDrawdown,
		TotalTrades:    len(trades),
		WinRate:        winRate,
		AvgTradeReturn: avgTradeReturn,
	}
}

func (s *BacktestService) calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func (s *BacktestService) calculateStdDev(values []float64, mean float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += math.Pow(v-mean, 2)
	}
	return math.Sqrt(sum / float64(len(values)))
}

func (s *BacktestService) calculateMaxDrawdown(portfolioHistory []data.PortfolioSnapshot) float64 {
	if len(portfolioHistory) == 0 {
		return 0
	}

	maxValue := portfolioHistory[0].Portfolio.Total
	maxDrawdown := 0.0

	for _, snapshot := range portfolioHistory {
		if snapshot.Portfolio.Total > maxValue {
			maxValue = snapshot.Portfolio.Total
		}
		
		drawdown := (maxValue - snapshot.Portfolio.Total) / maxValue * 100
		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
		}
	}

	return maxDrawdown
}

func (s *BacktestService) calculateTradeMetrics(trades []data.Trade, initialValue float64) (float64, float64) {
	if len(trades) == 0 {
		return 0, 0
	}

	wins := 0
	totalReturn := 0.0

	for _, trade := range trades {
		if trade.Side == "SELL" {
			// Calculate return for this trade
			// This is simplified - in practice you'd track entry/exit prices
			totalReturn += trade.Value
			wins++
		}
	}

	winRate := float64(wins) / float64(len(trades)) * 100
	avgTradeReturn := totalReturn / float64(len(trades))

	return winRate, avgTradeReturn
}
