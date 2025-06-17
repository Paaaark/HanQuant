package data

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"golang.org/x/text/encoding/korean"
)

// ParseStockListingFile reads the fixed-width master file and returns meta rows.
func ParseStockListingFile(path string) ([]StockMeta, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	var out []StockMeta
	decoder := korean.EUCKR.NewDecoder().Reader(f)
	sc := bufio.NewScanner(decoder)

	for sc.Scan() {
		line := sc.Text()
		if len(line) < 76 { // need at least up to smal_div_code
			continue
		}
		out = append(out, StockMeta{
			Code:         strings.TrimSpace(line[0:9]),
			ISIN:         strings.TrimSpace(line[9:21]),
			Name:         strings.TrimSpace(line[21:61]),
			SecurityType: strings.TrimSpace(line[61:63]),
			CapSize:      strings.TrimSpace(line[63:64]),
			IndLarge:     strings.TrimSpace(line[64:68]),
			IndMedium:    strings.TrimSpace(line[68:72]),
			IndSmall:     strings.TrimSpace(line[72:76]),
		})
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}
	return out, nil
}


// WriteToCSV saves the parsed meta list.
func WriteToCSV(rows []StockMeta, outPath string) error {
	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("create csv: %w", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	_ = w.Write([]string{
		"Code", "ISIN", "Name", "SecurityType",
		"CapSize", "IndLarge", "IndMedium", "IndSmall",
	})
	for _, r := range rows {
		_ = w.Write([]string{
			r.Code, r.ISIN, r.Name, r.SecurityType,
			r.CapSize, r.IndLarge, r.IndMedium, r.IndSmall,
		})
	}
	return nil
}


// SearchStocks returns all stocks that *contain* the query in code or name
// func (s *StockStore) SearchStocks(query string) StockStore {
// 	query = strings.ToLower(query)
// 	var results StockStore

// 	for _, stock := range s.Cache {
// 		if strings.Contains(strings.ToLower(stock.Code), query) ||
// 			strings.Contains(strings.ToLower(stock.Name), query) {
// 				results = append(results, stock)
// 			}
// 	}

// 	return results
// }

// FindStockByCode returns the exact stock that *matches* the code
// func (s *StockStore) FindStockByCode(code string) (*StockIdentity, bool) {
// 	for _, stock := range s.Cache {
// 		if strings.EqualFold(stock.Code, code) {
// 			return &stock, true
// 		}
// 	}
// 	return nil, false
// }

// func Load(path string) (*StockStore, error) {
// 	file, err := os.Open(path)
// 	if err != nil {
// 		return nil, fmt.Errorf("could not open CSV file: %w", err)
// 	}
// 	defer file.Close()

// 	reader := csv.NewReader(file)
// 	_, err = reader.Read()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to read header: %w", err)
// 	}

// 	var stocks []StockIdentity
// 	for {
// 		record, err := reader.Read()
// 		if err == io.EOF {
// 			break
// 		}

// 		if err != nil {
// 			return nil, fmt.Errorf("failed to reac CSV record: %w", err)
// 		}

// 		if len(record) < 4 {
// 			continue
// 		}

// 		stocks = append(stocks, StockIdentity{
// 			Code: record[0],
// 			ISIN: record[1],
// 			Name: record[2],
// 			SecurityType: record[3],
// 		})
// 	}

// 	return &StockStore{Cache: stocks}, nil
// }