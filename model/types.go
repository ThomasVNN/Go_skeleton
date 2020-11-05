package model

import "time"

type ProductManagerEventTypeEnum int

type ProductSample struct {
	ProductId   string    `json:"product_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ProductManagerEvent struct {
	EventType   ProductManagerEventTypeEnum
	Name        string
	Description string
	Timestamp   time.Time
}

type SearchRequest struct {
	Q string
}

type SearchResult struct {
	Products []ProductSample
}
