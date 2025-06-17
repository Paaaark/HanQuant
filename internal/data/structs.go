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