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

func (s SliceStockSnapshot) EncodeJSON() []byte {
	var buf bytes.Buffer
	buf.WriteByte('[')

	for i, snap := range s {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(fmt.Sprintf(
			`{"Code":"%s","Name":"%s","Price":"%s",`+
				`"Change":"%s","ChangeSign":"%s","ChangeRate":"%s",`+
				`"Open":"%s","High":"%s","Low":"%s",`+
				`"Volume":"%s","SecurityType":"%s","AskPrice":"%s",`+
				`"BidPrice":"%s","AskVolume":"%s","BidVolume":"%s",`+
				`"TotalAskVolume":"%s","TotalBidVolume":"%s","TotalTradedValue":"%s"}`,
			escape(snap.Code), escape(snap.Name), escape(snap.Price),
			escape(snap.Change), escape(snap.ChangeSign), escape(snap.ChangeRate),
			escape(snap.Open), escape(snap.High), escape(snap.Low),
			escape(snap.Volume), escape(snap.SecurityType), escape(snap.AskPrice),
			escape(snap.BidPrice), escape(snap.AskVolume), escape(snap.BidVolume),
			escape(snap.TotalAskVolume), escape(snap.TotalBidVolume), escape(snap.TotalTradedValue),
		))
	}

	buf.WriteByte(']')
	return buf.Bytes()
}

func (s SlicePortfolioPosition) EncodeJSON() []byte {
	var buf bytes.Buffer
	buf.WriteByte('[')

	for i, p := range s {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(fmt.Sprintf(
			`{"Symbol":"%s","Name":"%s","TradeType":"%s","HoldingQty":"%s","OrderableQty":"%s",`+
				`"AvgPrice":"%s","PurchaseAmount":"%s","CurrentPrice":"%s","EvaluationAmount":"%s",`+
				`"UnrealizedPnl":"%s","UnrealizedPnlRate":"%s","FluctuationRate":"%s"}`,
			escape(p.Symbol), escape(p.Name), escape(p.TradeType),
			escape(p.HoldingQty), escape(p.OrderableQty), escape(p.AvgPrice),
			escape(p.PurchaseAmount), escape(p.CurrentPrice), escape(p.EvaluationAmount),
			escape(p.UnrealizedPnl), escape(p.UnrealizedPnlRate), escape(p.FluctuationRate),
		))
	}

	buf.WriteByte(']')
	return buf.Bytes()
}

func (a AccountSummary) EncodeJSON() []byte {
	return []byte(fmt.Sprintf(
		`{"TotalDeposit":"%s","D2Deposit":"%s","TotalPurchaseAmount":"%s","TotalEvaluationAmount":"%s",`+
			`"TotalUnrealizedPnl":"%s","NetAsset":"%s","AssetChangeAmount":"%s","AssetChangeRate":"%s"}`,
		escape(a.TotalDeposit), escape(a.D2Deposit), escape(a.TotalPurchaseAmount),
		escape(a.TotalEvaluationAmount), escape(a.TotalUnrealizedPnl), escape(a.NetAsset),
		escape(a.AssetChangeAmount), escape(a.AssetChangeRate),
	))
}

// escape puts a quotation around a string
func escape(s string) string {
	return strconv.Quote(s)[1 : len(strconv.Quote(s))-1]
}
