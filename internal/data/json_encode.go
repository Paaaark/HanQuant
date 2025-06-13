package data

import (
	"bytes"
	"fmt"
	"strconv"
)

func (s SlicePriceStruct) EncodeJSON() []byte {
	var buf bytes.Buffer
	buf.WriteByte('[')

	for i, p := range s { 
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(fmt.Sprintf(`{"Date":"%s","Open":"%s","High":"%s","Low":"%s","Close":"%s","Volume":"%s","Duration":"%s"}`,
			escape(p.Date), escape(p.Open), escape(p.High),
			escape(p.Low), escape(p.Close), escape(p.Volume), escape(p.Duration)))
	}

	buf.WriteByte(']')
	return buf.Bytes()
}

func (s SliceRankingStock) EncodeJSON() []byte {
	var buf bytes.Buffer
	buf.WriteByte('[')

	for i, stock := range s {
		if i > 0 {
			buf.WriteByte(',')
		}
		code := stock.Code
		if code == "" {
			code = stock.Code2
		}
		buf.WriteString(fmt.Sprintf(`{"Code":"%s","Name":"%s","Price":"%s","Change":"%s","ChangeSign":"%s","ChangeRate":"%s","Volume":"%s","MarketCap":"%s","Rank":"%s"}`,
			escape(code), escape(stock.Name), escape(stock.Price),
			escape(stock.Change), escape(stock.ChangeSign), escape(stock.ChangeRate),
			escape(stock.Volume), escape(stock.MarketCap), escape(stock.Rank)))
	}

	buf.WriteByte(']')
	return buf.Bytes()
}

func (idx IndexStruct) EncodeJSON() []byte {
	return []byte(fmt.Sprintf(
		`{"IndexCode":"%s","IndexName":"%s","Date":"%s","CurrentPrice":"%s","ChangeFromPrev":"%s","ChangeSign":"%s","ChangeRate":"%s","Open":"%s","High":"%s","Low":"%s","Volume":"%s","RisingCnt":"%s","UpperLimitCnt":"%s","FlatCnt":"%s","FallingCnt":"%s","LowerLimitCnt":"%s"}`,
		escape(idx.IndexCode), escape(idx.IndexName), escape(idx.Date),
		escape(idx.CurrentPrice), escape(idx.ChangeFromPrev), escape(idx.ChangeSign),
		escape(idx.ChangeRate), escape(idx.Open), escape(idx.High), escape(idx.Low),
		escape(idx.Volume), escape(idx.RisingCnt), escape(idx.UpperLimitCnt),
		escape(idx.FlatCnt), escape(idx.FallingCnt), escape(idx.LowerLimitCnt),
	))
}

// escape puts a quotation around a string
func escape(s string) string {
	return strconv.Quote(s)[1 : len(strconv.Quote(s))-1]
}
