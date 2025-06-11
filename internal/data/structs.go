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