package data

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

type KISClient struct {
    AppKey     string
    AppSecret  string
    AccessToken string
    TrID       string
}

const (
    // Real URL should be "https://openapi.koreainvestment.com:9443"
    KISBaseURL = "https://openapivts.koreainvestment.com:29443"
    KIS_ACCESS_TOKEN = "KIS_ACCESS_TOKEN"
)

func NewKISClient() *KISClient {
    return &KISClient{
        AppKey: os.Getenv("KIS_APP_KEY"),
        AppSecret: os.Getenv("KIS_APP_SECRET"),
    }
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

// GetRecentDailyPrice: 주식현재가 일자별
func (c *KISClient) GetRecentDailyPrice(symbol string) ([]PriceStruct, error) {
    endpoint := fmt.Sprintf("%s/uapi/domestic-stock/v1/quotations/inquire-daily-price", KISBaseURL)

    c.AccessToken = os.Getenv(KIS_ACCESS_TOKEN)

    params := url.Values{}
    params.Add("FID_COND_MRKT_DIV_CODE", "J")
    params.Add("FID_INPUT_ISCD", symbol)
    params.Add("FID_PERIOD_DIV_CODE", "D")
    params.Add("FID_ORG_ADJ_PRC", "0")

    req, err := http.NewRequest("GET", endpoint+"?"+params.Encode(), nil)
    if err != nil {
        return nil, err
    }

    c.prepareRequestHeader(req, "FHKST01010400")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    fmt.Println(resp.Request.URL)

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API request failed: %s", resp.Status)
    }

    return parsePriceBody(resp.Body, "output", "D")
}

// GetDailyPrice retrieves historical stock prices for a given symbol between two dates.
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
func (c *KISClient) GetDailyPrice(symbol, from, to, duration string) ([]PriceStruct, error) {
    endpoint := fmt.Sprintf("%s/uapi/domestic-stock/v1/quotations/inquire-daily-itemchartprice", KISBaseURL)

    c.AccessToken = os.Getenv(KIS_ACCESS_TOKEN)

    params := url.Values{}
	params.Add("FID_COND_MRKT_DIV_CODE", "J")
	params.Add("FID_INPUT_ISCD", symbol)
	params.Add("FID_INPUT_DATE_1", from)
	params.Add("FID_INPUT_DATE_2", to)
	params.Add("FID_PERIOD_DIV_CODE", duration)
	params.Add("FID_ORG_ADJ_PRC", "0")
    req, err := http.NewRequest("GET", endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.prepareRequestHeader(req, "FHKST03010100")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API call error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API call failed: %s\n%s", resp.Status, string(body))
	}

	return parsePriceBody(resp.Body, "output2", duration)
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

// parsePriceBody parses a KIS API HTTP response body into a slice of PriceStruct.
//
// Parameters:
//   - body: the response body (io.Reader) from the KIS API call
//   - targetField: the JSON field name containing the price data array (e.g., "output", "output2")
//     If you're unsure which field to use, pass "output" as the default.
//   - duration: a string indicating the time granularity (e.g., "D" for daily, "W" for weekly)
//
// Returns:
//   - A slice of PriceStruct containing date, open, high, low, close, volume, and duration
//   - An error if decoding or data parsing fails
func parsePriceBody(body io.Reader, targetField, duration string) ([]PriceStruct, error) {
    var raw map[string]json.RawMessage

    if err := json.NewDecoder(body).Decode(&raw); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }

    data, ok := raw[targetField]
    if !ok {
        return nil, fmt.Errorf("field %q not found in response", targetField)
    }

    var rawPrices []map[string]string
    if err := json.Unmarshal(data, &rawPrices); err != nil {
        return nil, fmt.Errorf("failed to parse %s: %w", targetField, err)
    }

    var prices []PriceStruct
	for _, entry := range rawPrices {
		open, _ := strconv.ParseFloat(entry["stck_oprc"], 64)
		high, _ := strconv.ParseFloat(entry["stck_hgpr"], 64)
		low, _ := strconv.ParseFloat(entry["stck_lwpr"], 64)
		closeVal, _ := strconv.ParseFloat(entry["stck_clpr"], 64)
		volume, _ := strconv.Atoi(entry["acml_vol"])

		prices = append(prices, PriceStruct{
			Date:     entry["stck_bsop_date"],
			Open:     open,
			High:     high,
			Low:      low,
			Close:    closeVal,
			Volume:   volume,
			Duration: duration,
		})
	}
    return prices, nil
}