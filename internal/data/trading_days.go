package data

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"
)

type TradingCalendar struct {
	days []string
	set map[string]bool
}

var (
	instance *TradingCalendar
	once sync.Once
)

// LoadTradingCalendar loads the trading calendar from a CSV file.
func LoadTradingCalendar(csvPath string) (*TradingCalendar, error) {
	var err error
	once.Do(func() {
		var dates []string
		daySet := make(map[string]bool)

		file, e := os.Open(csvPath)
		if e != nil {
			err = e
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		scanner.Scan()
		for scanner.Scan() {
			line := scanner.Text()
			dates = append(dates, line)
			daySet[line] = true
		}

		if scannerErr := scanner.Err(); scannerErr != nil {
			err = scannerErr
			return
		}

		instance = &TradingCalendar{
			days: dates,
			set: daySet,
		}
	})
	if err != nil {
		return nil, err
	}
	return instance, nil
}

func (tc *TradingCalendar) IsTradingDay(date string) bool {
    return tc.set[date]
}

func IsTradingDay(date string) bool {
	path := os.Getenv("TRADING_DAYS_CSV")
	if path == "" {
		path = "weekdays.csv"
	}
	_, err := LoadTradingCalendar(path)
	if err != nil {
		return false
	}
	return instance.IsTradingDay(date)
}

func (tc *TradingCalendar) CountInRange(from, to string) int {
    start := sort.SearchStrings(tc.days, from)
    end := sort.SearchStrings(tc.days, to)

    if end < len(tc.days) && tc.days[end] == to {
        end++
    }

    if start >= len(tc.days) || tc.days[start] > to {
        return 0
    }

    return end - start
}

func CountInRange(from, to string) int {
	path := os.Getenv("TRADING_DAYS_CSV")
	if path == "" {
		path = "weekdays.csv"
	}
	_, err := LoadTradingCalendar(path)
    if err != nil {
        return 0
    }
    return instance.CountInRange(from, to)
}

// AddTradingDays returns the date that is n trading days after the given date.
func (tc *TradingCalendar) AddTradingDays(date string, n int) (string, error) {
    idx := sort.SearchStrings(tc.days, date)
    if idx < len(tc.days) && tc.days[idx] != date {
    } else if idx == len(tc.days) {
        return "", fmt.Errorf("date %q is after last calendar entry", date)
    }
    target := idx + n
    if target >= len(tc.days) {
        return "", fmt.Errorf("not enough trading days after %q", date)
    }
    return tc.days[target], nil
}

// AddTradingDays returns the date that is n trading days after the given date.
func AddTradingDays(date string, n int) (string, error) {
	path := os.Getenv("TRADING_DAYS_CSV")
	if path == "" {
		path = "weekdays.csv"
	}
	_, err := LoadTradingCalendar(path)
	if err != nil {
        return "", fmt.Errorf("failed to load trading calendar: %w", err)
    }

	return instance.AddTradingDays(date, n)
}

// NextTradingDay returns the next trading day after or equal the given date.
func NextTradingDay(date string) (string, error) {
	day, err := time.Parse("20060102", date)
	if err != nil {
		return "", fmt.Errorf("invalid date: %w", err)
	}

	for i := 0; i < 100; i++ {
		if IsTradingDay(day.Format("20060102")) {
			return day.Format("20060102"), nil
		}
		day = day.AddDate(0, 0, 1)
	}
	return "", fmt.Errorf("no trading day found after %q", date)
}

// PreviousTradingDay returns the next trading day before or equal the given date.
func PreviousTradingDay(date string) (string, error) {
	day, err := time.Parse("20060102", date)
	if err != nil {
		return "", fmt.Errorf("invalid date: %w", err)
	}

	for i := 0; i < 100; i++ {
		if IsTradingDay(day.Format("20060102")) {
			return day.Format("20060102"), nil
		}
		day = day.AddDate(0, 0, -1)
	}
	return "", fmt.Errorf("no trading day found after %q", date)
}