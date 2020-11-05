package default_listing

import (
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/model"
	"math"
	"strings"
)

func (dl *defaultListingService) getCateScore(params ...float64) int64 {
	var score float64 = 1
	for _, param := range params {
		score *= param
	}
	return int64(score)
}

func (dl *defaultListingService) getProductRange(products []model.Product) (uint32, uint32) {
	var max, min uint32
	for _, v := range products {
		if min == 0 || min > v.ProductId {
			min = v.ProductId
		}
		if max < v.ProductId {
			max = v.ProductId
		}
	}
	return min, max
}

func (dl *defaultListingService) toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(dl.round(num*output)) / output
}

func (dl *defaultListingService) round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func (dl *defaultListingService) getCategoryIDByPath(catePath string) string {
	if strings.TrimSpace(catePath) == "" {
		return "0"
	}
	cateIds := strings.Split(catePath, "/")
	if len(cateIds) != 5 {
		return "0"
	}
	return cateIds[3]
}
