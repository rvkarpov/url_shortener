package handler

type OriginURLInfo struct {
	URL string `json:"url"`
}

type ShortURLInfo struct {
	Result string `json:"result"`
}

type OriginURLBatchItem struct {
	ID  string `json:"correlation_id"`
	URL string `json:"original_url"`
}

type ShortURLBatchItem struct {
	ID  string `json:"correlation_id"`
	URL string `json:"short_url"`
}
