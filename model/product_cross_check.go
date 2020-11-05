package model

import (
	"time"
)

type ProductCrossCheck struct {
	ProductId       uint32    `json:"product_id" bson:"product_id"`
	ErrorRedis      string    `json:"error_redis" bson:"error_redis"`
	ErrorElastic    string    `json:"error_elastic" bson:"error_elastic"`
	IsSyncEs        bool      `json:"is_sync_es" bson:"is_sync_es"`
	IsSyncRedis     bool      `json:"is_sync_redis" bson:"is_sync_redis"`
	RetryCountRedis int       `json:"retry_count_redis" bson:"retry_count_redis"`
	RetryCountEs    int       `json:"retry_count_es" bson:"retry_count_es"`
	CreatedAt       int64     `json:"created_at" bson:"created_at"`
	UpdatedAt       int64     `json:"updated_at" bson:"updated_at"`
	ExpiredAt       time.Time `json:"expired_at" bson:"expired_at"`
}
