package data

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
)

type StockIdentity struct {
    Code         string // 단축코드
    ISIN         string // 표준코드
    Name         string // 종목명
    SecurityType string // 증권그룹구분코드
}

type StockStore struct {
	Cache []StockIdentity
}

// ParseStockListingFile reads a fixed-width encoded text file line-by-line and extracts stock identities.
func ParseStockListingFile(filePath string) ([]StockIdentity, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to open file: %w", err)
    }
    defer file.Close()

    var stocks []StockIdentity
    scanner := bufio.NewScanner(file)

    for scanner.Scan() {
        line := scanner.Text()
        if len(line) < 63 { // minimum length required to safely parse
            continue
        }

        si := StockIdentity{
            Code:         strings.TrimSpace(line[0:9]),   // mksc_shrn_iscd
            ISIN:         strings.TrimSpace(line[9:21]),  // stnd_iscd
            Name:         strings.TrimSpace(line[21:61]), // hts_kor_isnm
            SecurityType: strings.TrimSpace(line[61:63]), // scrt_grp_cls_code
        }

        stocks = append(stocks, si)
    }

    if err := scanner.Err(); err != nil {
        return nil, fmt.Errorf("error reading lines: %w", err)
    }

    return stocks, nil
}

// WriteToCSV saves the parsed stock identities to a CSV file
func WriteToCSV(stocks []StockIdentity, outPath string) error {
    file, err := os.Create(outPath)
    if err != nil {
        return fmt.Errorf("failed to create CSV: %w", err)
    }
    defer file.Close()

    writer := csv.NewWriter(file)
    defer writer.Flush()

    // Header
    writer.Write([]string{"Code", "ISIN", "Name", "SecurityType"})

    // Records
    for _, s := range stocks {
        writer.Write([]string{s.Code, s.ISIN, s.Name, s.SecurityType})
    }

    return nil
}

func (s *StockStore) SearchStocks(query string) []StockIdentity {
	query = strings.ToLower(query)
	var results []StockIdentity

	for _, stock := range s.Cache {
		if strings.Contains(strings.ToLower(stock.Code), query) ||
			strings.Contains(strings.ToLower(stock.Name), query) {
				results = append(results, stock)
			}
	}

	return results
}

func Load(path string) (*StockStore, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	_, err = reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	var stocks []StockIdentity
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("failed to reac CSV record: %w", err)
		}

		if len(record) < 4 {
			continue
		}

		stocks = append(stocks, StockIdentity{
			Code: record[0],
			ISIN: record[1],
			Name: record[2],
			SecurityType: record[3],
		})
	}

	return &StockStore{Cache: stocks}, nil
}