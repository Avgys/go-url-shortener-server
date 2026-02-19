package model

type ShortenReq struct {
	URL string `json:"url"`
}

type ShortenResp struct {
	URL string `json:"result"`
}
