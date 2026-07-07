package responses

type GetURLsStatsResponse struct {
	UniqueLongUrlCount int64 `json:"urls"`
	UniqueUserIDCount  int64 `json:"users"`
}
