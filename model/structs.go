package model

type EsPriceAggregation struct {
	DocCount int64 `json:"doc_count"`
	Prices struct{
		Values map[string]float64 `json:"values"`
	} `json:"prices"`
}
