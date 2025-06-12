package data

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
	Name       string `json:"hts_kor_isnm"`
	Price      string `json:"stck_prpr"`
	Change     string `json:"prdy_vrss"`
	ChangeSign string `json:"prdy_vrss_sign"`
	ChangeRate string `json:"prdy_ctrt"`
	Volume     string `json:"acml_vol"`
	Rank       string `json:"data_rank"`
}

type PriceStruct struct {
	Date     string
	Open     float64
	High     float64
	Low      float64
	Close    float64
	Volume   int
	Duration string
}

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