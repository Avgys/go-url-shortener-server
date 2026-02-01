package model

type ShortenReq struct {
	Url string `json:"url"`
}

type ShortenResp struct {
	Url string `json:"result"`
}
