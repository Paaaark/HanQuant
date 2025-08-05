package data

type StockMeta struct {
	Code         string // 단축코드
	ISIN         string // 표준코드
	Name         string // 종목명
	SecurityType string // 증권그룹
	CapSize      string // 시총규모 (0,1,2,3)
	IndLarge     string // 업종 대분류
	IndMedium    string // 업종 중분류
	IndSmall     string // 업종 소분류
}

type StockStore struct {
	Cache []StockMeta
}

type KISClient struct {
	AppKey        string
	AppSecret     string
	MockAppKey    string
	MockAppSecret string
	AccessToken   string
	TrID          string
}

type RankingStock struct {
	Code       string `json:"stck_shrn_iscd"`
	Code2      string `json:"mksc_shrn_iscd"`
	Name       string `json:"hts_kor_isnm"`
	Price      string `json:"stck_prpr"`
	Change     string `json:"prdy_vrss"`
	ChangeSign string `json:"prdy_vrss_sign"`
	ChangeRate string `json:"prdy_ctrt"`
	Volume     string `json:"acml_vol"`
	MarketCap  string `json:"stck_avls"`
	Rank       string `json:"data_rank"`
}

type SliceRankingStock []RankingStock

type PriceStruct struct {
	Date     string `json:"stck_bsop_date"`
	Open     string `json:"stck_oprc"`
	High     string `json:"stck_hgpr"`
	Low      string `json:"stck_lwpr"`
	Close    string `json:"stck_clpr"`
	Volume   string `json:"acml_vol"`
	Duration string
}

type SlicePriceStruct []PriceStruct

type IndexStruct struct {
	IndexCode string
	IndexName string

	Date           string
	CurrentPrice   string `json:"bstp_nmix_prpr"`
	ChangeFromPrev string `json:"bstp_nmix_prdy_vrss"`
	ChangeSign     string `json:"prdy_vrss_sign"`
	ChangeRate     string `json:"bstp_nmix_prdy_ctrt"`

	Open string `json:"bstp_nmix_oprc"`
	High string `json:"bstp_nmix_hgpr"`
	Low  string `json:"bstp_nmix_lwpr"`

	Volume string `json:"acml_vol"`

	RisingCnt     string `json:"ascn_issu_cnt"`
	UpperLimitCnt string `json:"uplm_issu_cnt"`
	FlatCnt       string `json:"stnr_issu_cnt"`
	FallingCnt    string `json:"down_issu_cnt"`
	LowerLimitCnt string `json:"lslm_issu_cnt"`
}

type StockSnapshot struct {
	Code       string `json:"inter_shrn_iscd"`
	Name       string `json:"inter_kor_isnm"`
	Price      string `json:"inter2_prpr"`
	Change     string `json:"inter2_prdy_vrss"`
	ChangeSign string `json:"prdy_vrss_sign"`
	ChangeRate string `json:"prdy_ctrt"`

	Open   string `json:"inter2_oprc"`
	High   string `json:"inter2_hgpr"`
	Low    string `json:"inter2_lwpr"`
	Volume string `json:"acml_vol"`

	SecurityType string `json:"mrkt_trtm_cls_name"`

	AskPrice         string `json:"inter2_askp"`
	BidPrice         string `json:"inter2_bidp"`
	AskVolume        string `json:"seln_rsqn"`
	BidVolume        string `json:"shnu_rsqn"`
	TotalAskVolume   string `json:"total_askp_rsqn"`
	TotalBidVolume   string `json:"total_bidp_rsqn"`
	TotalTradedValue string `json:"acml_tr_pbmn"`
}

type SliceStockSnapshot []StockSnapshot

type PortfolioPosition struct {
	Symbol            string `json:"pdno"`           // 종목코드
	Name              string `json:"prdt_name"`      // 종목명
	TradeType         string `json:"trad_dvsn_name"` // 매매구분
	HoldingQty        string `json:"hldg_qty"`       // 보유수량
	OrderableQty      string `json:"ord_psbl_qty"`   // 주문가능수량
	AvgPrice          string `json:"pchs_avg_pric"`  // 매입평균가
	PurchaseAmount    string `json:"pchs_amt"`       // 매입금액
	CurrentPrice      string `json:"prpr"`           // 현재가
	EvaluationAmount  string `json:"evlu_amt"`       // 평가금액
	UnrealizedPnl     string `json:"evlu_pfls_amt"`  // 평가손익금액
	UnrealizedPnlRate string `json:"evlu_pfls_rt"`   // 평가손익율
	FluctuationRate   string `json:"fltt_rt"`        // 등락율
}

type SlicePortfolioPosition []PortfolioPosition

type AccountSummary struct {
	TotalDeposit          string `json:"dnca_tot_amt"`       // 예수금
	D2Deposit             string `json:"prvs_rcdl_excc_amt"` // D+2 예수금
	TotalPurchaseAmount   string `json:"pchs_amt_smtl_amt"`  // 매입금액합계
	TotalEvaluationAmount string `json:"evlu_amt_smtl_amt"`  // 평가금액합계
	TotalUnrealizedPnl    string `json:"evlu_pfls_smtl_amt"` // 평가손익합계
	NetAsset              string `json:"nass_amt"`           // 순자산금액
	AssetChangeAmount     string `json:"asst_icdc_amt"`      // 자산증감액
	AssetChangeRate       string `json:"asst_icdc_erng_rt"`  // 자산증감수익율
}

type OrderRequest struct {
	Symbol    string `json:"symbol"`
	Qty       string `json:"qty"`
	Price     string `json:"price"`
	OrderType string `json:"order_type"` // e.g. 00=limit, 01=market
	Side      string `json:"side"`       // "buy" or "sell"
	Mock      bool   `json:"mock"`       // true for paper trading
}

type OrderResponse struct {
	OrderNo   string `json:"order_no"`
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
	Success   bool   `json:"success"`
}

type User struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
	CreatedAt    string `json:"created_at"`
}

type UserAccount struct {
	ID          int64  `json:"id"`
	UserID      int64  `json:"user_id"`
	AccountID   string `json:"account_id"`
	EncCANO     []byte `json:"enc_cano"`
	EncAppKey   []byte `json:"enc_app_key"`
	EncAppSecret []byte `json:"enc_app_secret"`
	IsMock      bool   `json:"is_mock"`
	CreatedAt   string `json:"created_at"`
}

type KISAccessToken struct {
	UserAccountID int64  `json:"user_account_id"`
	Token         string `json:"token"`
	ExpiresAt     string `json:"expires_at"`
	RefreshedAt   string `json:"refreshed_at"`
}

type Order struct {
	ID            int64   `json:"id"`
	UserAccountID int64   `json:"user_account_id"`
	Symbol        string  `json:"symbol"`
	Side          string  `json:"side"`
	Qty           float64 `json:"qty"`
	OrderType     string  `json:"order_type"`
	LimitPrice    *float64 `json:"limit_price,omitempty"`
	Status        string  `json:"status"`
	KISOrderID    string  `json:"kis_order_id"`
	CreatedAt     string  `json:"created_at"`
}

// MinutePriceStruct for minute-by-minute stock data
type MinutePriceStruct struct {
	DateTime string `json:"stck_cntg_hour"` // YYYYMMDDHHMMSS format
	Open     string `json:"stck_oprc"`
	High     string `json:"stck_hgpr"`
	Low      string `json:"stck_lwpr"`
	Close    string `json:"stck_prpr"`
	Volume   string `json:"cntg_vol"`
	Duration string // Always "M" for minute data
}

type SliceMinutePriceStruct []MinutePriceStruct

// HistoricalDataRequest for fetching historical data
type HistoricalDataRequest struct {
	Symbol   string `json:"symbol"`
	FromDate string `json:"from_date"` // YYYYMMDD format
	ToDate   string `json:"to_date"`   // YYYYMMDD format
	Duration string `json:"duration"`  // "D" for daily, "M" for minute
}

// S3Config for AWS S3 configuration
type S3Config struct {
	BucketName string
	Region     string
	AccessKey  string
	SecretKey  string
}