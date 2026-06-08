package responses

type ShortenBatchResp []IndexedShortURL

type IndexedShortURL struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
	IsCreated     bool   `json:"is_created"`
}
