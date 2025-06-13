package data

type StockIdentity struct {
	Code         string // 단축코드
	ISIN         string // 표준코드
	Name         string // 종목명
	SecurityType string // 증권그룹구분코드
}

type StockStore struct {
	Cache []StockIdentity
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