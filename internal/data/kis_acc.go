// kis_acc.go
package data

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
)

/*
   GetAccountPortfolio: 주식 잔고조회 (output1 + output2)

   Routes we hit this from
     • GET /accounts/{accNo}/portfolio        → mock = false
     • GET /accounts_mock/{accNo}/portfolio   → mock = true

   Parameters
     accNo : “12345678-01”   (8-2 account string with dash)
     mock  : true  → 모의 URL & TR-ID
              false → 실전 URL & TR-ID

   Returns
     []PortfolioPosition     (output1)
     *AccountSummary         (first element of output2, or nil)
     error
*/
func (c *KISClient) GetAccountPortfolio(accNo string, mock bool) (SlicePortfolioPosition, *AccountSummary, error) {
	//-----------------------------------------------------------------
	// 0.  Choose host & TR-ID and (optionally) swap app-key/secret
	//-----------------------------------------------------------------
	baseURL, trID := KISBaseURL, "TTTC8434R"
	origKey, origSecret := c.AppKey, c.AppSecret

	if mock {
		baseURL = KISBaseURLMock
		trID = "VTTC8434R"
		// Temporarily use the mock credentials that NewKISClient loaded.
		c.AppKey, c.AppSecret = c.MockAppKey, c.MockAppSecret
		defer func() { c.AppKey, c.AppSecret = origKey, origSecret }()
	}

	endpoint := fmt.Sprintf("%s/uapi/domestic-stock/v1/trading/inquire-balance", baseURL)

	//-----------------------------------------------------------------
	// 1.  Validate & split “12345678-01”  →  CANO / ACNT_PRDT_CD
	//-----------------------------------------------------------------
	parts := strings.SplitN(accNo, "-", 2)
	if len(parts) != 2 || len(parts[0]) != 8 || len(parts[1]) != 2 {
		return nil, nil, fmt.Errorf("invalid account format (want 8-2 with dash): %q", accNo)
	}
	cano, acntPrdt := parts[0], parts[1]

	//-----------------------------------------------------------------
	// 2.  Build query params (defaults we agreed on)
	//-----------------------------------------------------------------
	params := url.Values{
		"CANO":                  []string{cano},
		"ACNT_PRDT_CD":          []string{acntPrdt},
		"AFHR_FLPR_YN":          []string{"N"},
		"OFL_YN":                []string{""},
		"INQR_DVSN":             []string{"02"},
		"UNPR_DVSN":             []string{"01"},
		"FUND_STTL_ICLD_YN":     []string{"N"},
		"FNCG_AMT_AUTO_RDPT_YN": []string{"N"},
		"PRCS_DVSN":             []string{"00"},
		"CTX_AREA_FK100":        []string{""},
		"CTX_AREA_NK100":        []string{""},
	}

	//-----------------------------------------------------------------
	// 3.  Ensure we have an access-token (same pattern as kis_stock.go)
	//-----------------------------------------------------------------
	c.AccessToken = os.Getenv(KIS_ACCESS_TOKEN)

	//-----------------------------------------------------------------
	// 4.  Pagination loop until CTX_AREA_NK100 == ""
	//-----------------------------------------------------------------
	var (
		allPositions SlicePortfolioPosition
		summary      *AccountSummary
		pageCount    int
	)

	var didRetry bool

	// Cap at 3 pages to avoid exceeding API limits, even if tr_cont indicates more data.
	for pageCount = 0; pageCount < 3; pageCount++ {
		respBody, err := c.get(endpoint, trID, params)
		if err != nil {
			// If token expired, try to refresh and retry ONCE
			if !didRetry && strings.Contains(err.Error(), "기간이 만료된 token 입니다.") {
				newToken, refreshErr := RefreshKISToken(c.AppKey, c.AppSecret)
				if refreshErr == nil {
					c.AccessToken = newToken
					didRetry = true
					pageCount-- // retry this page
					continue
				}
			}
			return nil, nil, err
		}
		defer respBody.Close()

		var raw struct {
			RtCd      string                 `json:"rt_cd"`
			MsgCd     string                 `json:"msg_cd"`
			Msg       string                 `json:"msg1"`
			CtxFK100  string                 `json:"ctx_area_fk100"`
			CtxNK100  string                 `json:"ctx_area_nk100"`
			Output1   SlicePortfolioPosition `json:"output1"`
			Output2   []AccountSummary       `json:"output2"`
		}

		if err := json.NewDecoder(respBody).Decode(&raw); err != nil {
			return nil, nil, fmt.Errorf("decode error: %w", err)
		}
		if raw.RtCd != "0" {
			return nil, nil, fmt.Errorf("KIS API %s: %s", raw.MsgCd, raw.Msg)
		}

		allPositions = append(allPositions, raw.Output1...)
		if summary == nil && len(raw.Output2) > 0 {
			tmp := raw.Output2[0]
			summary = &tmp
		}

		if raw.CtxNK100 == "" { // last page
			break
		}

		params.Set("CTX_AREA_FK100", raw.CtxFK100)
		params.Set("CTX_AREA_NK100", raw.CtxNK100)
		// continue to next page
	}

	return allPositions, summary, nil
}
