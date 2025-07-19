package data

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
    // Real URL should be "https://openapi.koreainvestment.com:9443"
    KISBaseURL = "https://openapi.koreainvestment.com:9443"
    KISBaseURLMock = "https://openapivts.koreainvestment.com:29443"
    KIS_ACCESS_TOKEN = "KIS_ACCESS_TOKEN"
)
var indexCodeToName = map[string]string{
    "0001": "Kospi",
    "1001": "Kosdaq",
    "2001": "Kospi200",
    "4001": "KRX100",
}

func NewKISClient() *KISClient {
    return &KISClient{
        AppKey: os.Getenv("KIS_APP_KEY"),
        AppSecret: os.Getenv("KIS_APP_SECRET"),
        MockAppKey: os.Getenv("KIS_MOCK_APP_KEY"),
        MockAppSecret: os.Getenv("KIS_MOCK_APP_SECRET"),
    }
}

// GetRecentDailyPrice: 주식현재가 일자별
func (c *KISClient) GetRecentDailyPrice(symbol string) (SlicePriceStruct, error) {
    endpoint := fmt.Sprintf("%s/uapi/domestic-stock/v1/quotations/inquire-daily-price", KISBaseURL)

    c.AccessToken = os.Getenv(KIS_ACCESS_TOKEN)

    params := url.Values{}
    params.Add("FID_COND_MRKT_DIV_CODE", "J")
    params.Add("FID_INPUT_ISCD", symbol)
    params.Add("FID_PERIOD_DIV_CODE", "D")
    params.Add("FID_ORG_ADJ_PRC", "0")

    resp_body, err := c.get(endpoint, "FHKST01010400", params)
    if err != nil {
        return nil, err
    }
    defer resp_body.Close()

    var raw struct {
        Output SlicePriceStruct `json:"output"`
    }

    if err := json.NewDecoder(resp_body).Decode(&raw); err != nil {
        return nil, err
    }

    for i := range raw.Output {
        raw.Output[i].Duration = "D"
    }

    return raw.Output, nil
}

// GetDailyPrice: 국내주식기간별시세(일/주/월/년)
// Retrieves historical stock prices for a given symbol between two dates.
// It calls the "국내주식기간별시세(일/주/월/년)" API and returns a slice of PriceStruct.
//
// Parameters:
//   - symbol: the stock symbol (e.g., "005930" for Samsung Electronics)
//   - from: the start date in "YYYYMMDD" format (e.g., "20240101")
//   - to: the end date in "YYYYMMDD" format (e.g., "20240601")
//   - duration: one of "D" (daily), "W" (weekly), "M" (monthly), or "Y" (yearly)
//
// Returns:
//   - A slice of PriceStruct containing date, open, high, low, close, volume, and duration fields
//   - An error if the API call fails or the response cannot be parsed
func (c *KISClient) GetDailyPrice(symbol, from, to, duration string) (SlicePriceStruct, error) {
    endpoint := fmt.Sprintf("%s/uapi/domestic-stock/v1/quotations/inquire-daily-itemchartprice", KISBaseURL)

    c.AccessToken = os.Getenv(KIS_ACCESS_TOKEN)

    params := url.Values{}
	params.Add("FID_COND_MRKT_DIV_CODE", "J")
	params.Add("FID_INPUT_ISCD", symbol)
	params.Add("FID_INPUT_DATE_1", from)
	params.Add("FID_INPUT_DATE_2", to)
	params.Add("FID_PERIOD_DIV_CODE", duration)
	params.Add("FID_ORG_ADJ_PRC", "0")

    resp_body, err := c.get(endpoint, "FHKST03010100", params)
    if err != nil {
        return nil, err
    }
    defer resp_body.Close()

	var raw struct {
        Output SlicePriceStruct `json:"output2"`
    }

    if err := json.NewDecoder(resp_body).Decode(&raw); err != nil {
        return nil, err
    }

    for i := range raw.Output {
        raw.Output[i].Duration = "D"
    }

    return raw.Output, nil
}

// GetTopFluctuationStocks: 국내주식 등락률 순위
// Fetches the top 30 ranked stocks by price fluctuation (e.g. 상승률 순)
func (c *KISClient) GetTopFluctuationStocks() (SliceRankingStock, error) {
	endpoint := fmt.Sprintf("%s/uapi/domestic-stock/v1/ranking/fluctuation", KISBaseURL)

    c.AccessToken = os.Getenv(KIS_ACCESS_TOKEN)

	params := url.Values{}
    params.Add("fid_rsfl_rate2", "")
    params.Add("fid_cond_mrkt_div_code", "J")
    params.Add("fid_cond_scr_div_code", "20170")
    params.Add("fid_input_iscd", "0000")
    params.Add("fid_rank_sort_cls_code", "0") // 상승율 순
    params.Add("fid_input_cnt_1", "0")
    params.Add("fid_prc_cls_code", "1")
    params.Add("fid_input_price_1", "")
    params.Add("fid_input_price_2", "")
    params.Add("fid_vol_cnt", "")
    params.Add("fid_trgt_cls_code", "0")
    params.Add("fid_trgt_exls_cls_code", "0")
    params.Add("fid_div_cls_code", "0")
    params.Add("fid_rsfl_rate1", "")

    resp_body, err := c.get(endpoint, "FHPST01700000", params)
    if err != nil {
        return nil, err
    }
    defer resp_body.Close()

	var result struct {
		Output SliceRankingStock `json:"output"`
	}

	if err := json.NewDecoder(resp_body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	return result.Output, nil
}

// GetMostTradedStocks: 거래량순위
// Fetches the top 30 ranked stocks by volumes traded 
func (c *KISClient) GetMostTradedStocks() (SliceRankingStock, error) {
	endpoint := fmt.Sprintf("%s/uapi/domestic-stock/v1/quotations/volume-rank", KISBaseURL)

    c.AccessToken = os.Getenv(KIS_ACCESS_TOKEN)

	params := url.Values{}
    params.Add("FID_COND_MRKT_DIV_CODE", "J")
    params.Add("FID_COND_SCR_DIV_CODE", "20171")
    params.Add("FID_INPUT_ISCD", "0000")
    params.Add("FID_DIV_CLS_CODE", "0")
    params.Add("FID_BLNG_CLS_CODE", "0") // 0: 평균거래량, 3: 거래금액순
    params.Add("FID_TRGT_CLS_CODE", "111111111")
    params.Add("FID_TRGT_EXLS_CLS_CODE", "0000000000")
    params.Add("FID_INPUT_PRICE_1", "")
    params.Add("FID_INPUT_PRICE_2", "")
    params.Add("FID_VOL_CNT", "")
    params.Add("FID_INPUT_DATE_1", "")

    resp_body, err := c.get(endpoint, "FHPST01710000", params)
    if err != nil {
        return nil, err
    }
    defer resp_body.Close()

	var result struct {
		Output SliceRankingStock `json:"output"`
	}

	if err := json.NewDecoder(resp_body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

    fmt.Println(result.Output)

	return result.Output, nil
}

// GetTopMarketCapStocks: 국내주식 시가총액 상위위
// Fetches the top 30 ranked stocks by volumes traded 
func (c *KISClient) GetTopMarketCapStocks() (SliceRankingStock, error) {
	endpoint := fmt.Sprintf("%s/uapi/domestic-stock/v1/ranking/market-cap", KISBaseURL)

    c.AccessToken = os.Getenv(KIS_ACCESS_TOKEN)

	params := url.Values{}
    params.Add("fid_input_price_2", "")
    params.Add("fid_cond_mrkt_div_code", "J")
    params.Add("fid_cond_scr_div_code", "20174")
    params.Add("fid_div_cls_code", "0")
    params.Add("fid_input_iscd", "0000") // 0: 평균거래량, 3: 거래금액순
    params.Add("fid_trgt_cls_code", "0")
    params.Add("fid_trgt_exls_cls_code", "0")
    params.Add("fid_input_price_1", "")
    params.Add("fid_vol_cnt", "")

    resp_body, err := c.get(endpoint, "FHPST01740000", params)
    if err != nil {
        return nil, err
    }
    defer resp_body.Close()

	var result struct {
		Output SliceRankingStock `json:"output"`
	}

	if err := json.NewDecoder(resp_body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	return result.Output, nil
}

// GetMultipleStockSnapshot: 관심종목(멀티종목) 시세조회
// Fetches stock snapshot of multiple stocks
func (c *KISClient) GetMultipleStockSnapshot(targetCode []string) (SliceStockSnapshot, error) {
	endpoint := fmt.Sprintf("%s/uapi/domestic-stock/v1/quotations/intstock-multprice", KISBaseURL)

    c.AccessToken = os.Getenv(KIS_ACCESS_TOKEN)

	params := url.Values{}
    marketDivCode := "FID_COND_MRKT_DIV_CODE_"
    inputIscd := "FID_INPUT_ISCD_"
    for i, v := range targetCode {
        params.Add(fmt.Sprintf("%s%d", marketDivCode, i + 1), "J")
        params.Add(fmt.Sprintf("%s%d", inputIscd, i + 1), v)
    }

    resp_body, err := c.get(endpoint, "FHKST11300006", params)
    if err != nil {
        return nil, err
    }
    defer resp_body.Close()

    // body, err := io.ReadAll(resp_body)
    // fmt.Println(string(body))

	var result struct {
		Output SliceStockSnapshot `json:"output"`
	}

	if err := json.NewDecoder(resp_body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	return result.Output, nil
}

// GetIndexPrice: 국내업종 현재지수
// Fetches index price of the targetIndex (0001: Kospi, 1001: Kosdaq, 2001: Kospi200, 4001: KRX100, and more)
func (c *KISClient) GetIndexPrice(targetIndex string) (*IndexStruct, error) {
	endpoint := fmt.Sprintf("%s/uapi/domestic-stock/v1/quotations/inquire-index-price", KISBaseURL)

    c.AccessToken = os.Getenv(KIS_ACCESS_TOKEN)

	params := url.Values{}
    params.Add("FID_COND_MRKT_DIV_CODE", "U")
    params.Add("FID_INPUT_ISCD", targetIndex)

    resp_body, err := c.get(endpoint, "FHPUP02100000", params)
    if err != nil {
        return nil, err
    }
    defer resp_body.Close()

    // body, err := io.ReadAll(resp_body)
    // fmt.Println(string(body))

	var result struct {
		Output IndexStruct `json:"output"`
	}

	if err := json.NewDecoder(resp_body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

    result.Output.Date = time.Now().Format("20060102")
    result.Output.IndexCode = targetIndex
    if v, ok := indexCodeToName[targetIndex]; ok {
        result.Output.IndexName = v
    } else {
        result.Output.IndexName = targetIndex
    }

	return &result.Output, nil
}

// PlaceOrder places a buy or sell order using the KIS API.
func (c *KISClient) PlaceOrder(accNo string, req OrderRequest) (*OrderResponse, error) {
	baseURL := KISBaseURL
	trID := "TTTC0012U" // Buy (real)
	if req.Side == "sell" {
		trID = "TTTC0011U" // Sell (real)
	}
	if req.Mock {
		baseURL = KISBaseURLMock
		if req.Side == "buy" {
			trID = "VTTC0012U"
		} else {
			trID = "VTTC0011U"
		}
	}

	endpoint := baseURL + "/uapi/domestic-stock/v1/trading/order-cash"

	// Split accNo into CANO and ACNT_PRDT_CD
	parts := strings.SplitN(accNo, "-", 2)
	if len(parts) != 2 || len(parts[0]) != 8 || len(parts[1]) != 2 {
		return nil, fmt.Errorf("invalid account format (want 8-2 with dash): %q", accNo)
	}
	cano, acntPrdt := parts[0], parts[1]

	// Build JSON body (all keys UPPERCASE)
	body := map[string]string{
		"CANO":         cano,
		"ACNT_PRDT_CD": acntPrdt,
		"PDNO":         req.Symbol,
		"ORD_DVSN":     req.OrderType,
		"ORD_QTY":      req.Qty,
		"ORD_UNPR":     req.Price,
	}
	if req.Side == "sell" {
		body["SLL_TYPE"] = "01" // normal sell
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal order body: %w", err)
	}

	c.AccessToken = os.Getenv(KIS_ACCESS_TOKEN)

	httpReq, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create order request: %w", err)
	}
	c.prepareRequestHeader(httpReq, trID)
	httpReq.Header.Set("content-type", "application/json; charset=utf-8")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("order request failed: %w", err)
	}
	defer resp.Body.Close()

	var raw struct {
		RtCd   string `json:"rt_cd"`
		MsgCd  string `json:"msg_cd"`
		Msg1   string `json:"msg1"`
		Output struct {
			OrderNo   string `json:"odno"`
			OrdTmd    string `json:"ord_tmd"`
		} `json:"output"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode order response: %w", err)
	}
	if raw.RtCd != "0" {
		return &OrderResponse{
			OrderNo:   "",
			Timestamp: "",
			Message:   raw.Msg1,
			Success:   false,
		}, nil
	}
	return &OrderResponse{
		OrderNo:   raw.Output.OrderNo,
		Timestamp: raw.Output.OrdTmd,
		Message:   raw.Msg1,
		Success:   true,
	}, nil
}

// prepareRequestHeaders sets the standard headers required for KIS API requests.
//
// It adds headers for content type, authorization, app key, app secret, and transaction ID (tr_id).
// This helps avoid duplication across different API calls that require similar headers.
//
// Parameters:
//   - req: the HTTP request to which the headers will be applied
//   - trID: the transaction ID specific to the API endpoint being called (e.g., "FHKST03010100")
func (c *KISClient) prepareRequestHeader(req *http.Request, trID string) {
    req.Header.Set("content-type", "application/json; charset=utf-8")
	req.Header.Set("authorization", "Bearer "+c.AccessToken)
	req.Header.Set("appkey", c.AppKey)
	req.Header.Set("appsecret", c.AppSecret)
	req.Header.Set("tr_id", trID)
}

// get sends a GET request to the given KIS API endpoint with headers and query parameters,
// and returns the raw response body if the status is 200 OK. Otherwise, it returns an error.
//
// ⚠️ Caller MUST close the returned body to avoid resource leaks.
func (c *KISClient) get(endpoint, trID string, params url.Values) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	c.prepareRequestHeader(req, trID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("KIS error: %s\n%s", resp.Status, string(b))
	}

	return resp.Body, nil
}

func (c *KISClient) GetKISAccessToken() (string, error) {
    payload := map[string]string{
        "grant_type": "client_credentials",
        "appkey":     c.AppKey,
        "appsecret":  c.AppSecret,
    }

    body, err := json.Marshal(payload)
    if err != nil {
        return "", fmt.Errorf("failed to encode JSON: %w", err)
    }

    req, err := http.NewRequest("POST", KISBaseURL+"/oauth2/tokenP", bytes.NewBuffer(body))
    if err != nil {
        return "", fmt.Errorf("failed to create request: %w", err)
    }

    req.Header.Set("Content-Type", "application/json; charset=UTF-8")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return "", fmt.Errorf("request error: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        responseBody, _ := io.ReadAll(resp.Body)
        return "", fmt.Errorf("unexpected status %s: %s", resp.Status, responseBody)
    }

    var authResp struct {
        AccessToken string `json:"access_token"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
        return "", fmt.Errorf("failed to parse response: %w", err)
    }

    c.AccessToken = authResp.AccessToken
    return authResp.AccessToken, nil
}
