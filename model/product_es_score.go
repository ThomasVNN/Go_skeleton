package model

import "time"

type ProductESScore struct {
	ProductId            uint32                `json:"product_id"`
	DefaultListingScores DefaultListingScoreEs `json:"default_listing_scores"`
	ListingScore         ListingScore          `json:"listing_score"`
	OrderCountRank       float64               `json:"order_count_rank"`
	RankUpdatedTime      time.Time             `json:"rank_updated_time"`
	RankSearch           float64               `json:"rank_search"`
}