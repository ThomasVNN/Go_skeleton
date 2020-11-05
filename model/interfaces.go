package model

import (
	"github.com/segmentio/kafka-go"
	"gitlab.thovnn.vn/core/golang-sdk/spubsub"
	"github.com/go-kit/kit/log"
)

type ProductEvents interface {
	OnProductAdded(m kafka.Message) error
	OnProductUpdated(m kafka.Message) error
	OnProductScoreUpdated(data []byte)
}
type ShopEvents interface {
	OnOffShop(m *spubsub.Message) error
	OnOffShopCertificate(m *spubsub.Message) error
	ShopSupportShippingFee(data *spubsub.Message) error
	ShopInstallment(data *spubsub.Message) error
	ShopPromotionApp(data *spubsub.Message) error
	ShopRatingInfo(data *spubsub.Message) error
	ShopInstant(m kafka.Message) error
	ShopWareHouse(m kafka.Message) error
	ShopAds(m kafka.Message) error
}

type ProductRankingScoreEvents interface {
	OnProductHotSellUpdate(m *spubsub.Message, logger log.Logger) error
	OnProductScoreUpdate(m *spubsub.Message, logger log.Logger) error
}

type ProductTotalScoreEvents interface {
	OnProductUpdate(msg kafka.Message) error
}

type ProductResynchronization interface {
	MultiSync() error
}

type ProductPromotionDC2Events interface {
	OnPromotionChange(m kafka.Message) error
}
