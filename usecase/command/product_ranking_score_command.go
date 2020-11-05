package command

import (
	"encoding/json"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"gitlab.thovnn.vn/core/golang-sdk/spubsub"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/model"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/storage/esstore"
)

type productRankingScoreCommand struct {
	esStorage esstore.ESStore
}

func NewProductRankingScoreCommand(esStorage esstore.ESStore) model.ProductRankingScoreEvents {
	return &productRankingScoreCommand{
		esStorage: esStorage,
	}
}

func (p *productRankingScoreCommand) OnProductHotSellUpdate(data *spubsub.Message, logger log.Logger) error {
	logging := log.With(logger, "prefix", "OnProductHotSellUpdate")
	var products []model.ProductES
	var pubsubData []map[string]interface{}
	var acceptedField = map[string]interface{}{
		"product_id":             nil,
		"order_count_rank": nil,
		"rank_updated_at":          nil,
		"updated_at":            nil,
		"rank_search":           nil,
	}

	err := json.Unmarshal(data.Data, &products)

	if err != nil {
		_ = level.Error(logging).Log("msg", "Failed to unmarshal product data: "+err.Error())
		return err
	}

	err = json.Unmarshal(data.Data, &pubsubData)
	if err != nil {
		return err
	}
	for i, d := range pubsubData {
		for k := range d {
			if _, ok := acceptedField[k]; !ok {
				delete(d, k)
			}
		}
		pubsubData[i] = d
	}

	res, err := p.esStorage.UpdateBulk(pubsubData)

	if res != nil {
		//TODO: Retry failed product?
	}

	if err != nil {
		_ = level.Error(logging).Log("msg", "Bulk update failed: "+err.Error())
		return err
	}
	return nil
}

func (p *productRankingScoreCommand) OnProductScoreUpdate(data *spubsub.Message, logger log.Logger) error {
	logging := log.With(logger, "prefix", "OnProductScoreUpdate")
	var products []model.ProductES
	var pubsubData []map[string]interface{}
	var acceptedField = map[string]interface{}{
		"product_id":             nil,
		"default_listing_scores": nil,
		"listing_score":          nil,
		"order_count":            nil,
		"like_counter":           nil,
	}

	err := json.Unmarshal(data.Data, &products)
	if err != nil {
		_ = level.Error(logging).Log("msg", "Failed to unmarshal product data: "+err.Error())
		return err
	}
	err = json.Unmarshal(data.Data, &pubsubData)
	if err != nil {
		return err
	}
	for i, d := range pubsubData {
		for k := range d {
			if _, ok := acceptedField[k]; !ok {
				delete(d, k)
			}
		}
		pubsubData[i] = d
	}

	res, err := p.esStorage.UpdateBulk(pubsubData)

	if res != nil {
		//TODO: Retry failed product?
	}

	if err != nil {
		_ = level.Error(logging).Log("msg", "Bulk update failed: "+err.Error())
		return err
	}
	return nil
}
