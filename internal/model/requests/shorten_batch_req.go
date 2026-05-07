package requests

type IndexedFullURL struct {
	CorrelationID string `json:"correlation_id"`
	FullURL       string `json:"original_url"`
}

type ShortenBatchReq []IndexedFullURL
