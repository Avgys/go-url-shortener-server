package model

type ShortenBatchReq []IndexedFullURL

type IndexedFullURL struct {
	CorrelationID string `json:"correlation_id"`
	FullURL       string `json:"original_url"`
}

type ShortenBatchResp []IndexedShortURL

type IndexedShortURL struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
	IsCreated     bool   `json:"is_created"`
}
