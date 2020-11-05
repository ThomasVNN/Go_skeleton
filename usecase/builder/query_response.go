package builder

import (
	"encoding/json"
	"fmt"
	"github.com/olivere/elastic/v7"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/model"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/util"
	productBase "gitlab.thovnn.vn/protobuf/internal-apis-go/product/base"
	esApi "gitlab.thovnn.vn/protobuf/internal-apis-go/product/es_service"
	"sort"
	"strconv"
)

func (pu *productBuilder) BuildListingScoreResponse(data *elastic.SearchHits) *esApi.ListingScoreResponse {
	var products []*esApi.ProductData
	response := &esApi.ListingScoreResponse{
		Data:     products,
		MetaData: &productBase.MetaData{Total: data.TotalHits.Value},
	}
	if data == nil {
		return response
	}
	for _, hit := range data.Hits {
		var product esApi.ProductData
		err := json.Unmarshal(hit.Source, &product)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if hit.Score != nil {
			product.SearchScore = *hit.Score
		}
		products = append(products, &product)
	}
	response.Data = products
	return response
}

func (pu *productBuilder) BuildPriceRangeAggregationResponse(aggregations elastic.Aggregations) *esApi.ProductPriceRangeResponse {
	var priceRanges []float64
	defaultRanges := []*esApi.PriceRange{
		{
			To: 100000,
		},
		{
			From: 100000,
			To:   1000000,
		},
		{
			From: 1000000,
			To:   10000000,
		},
		{
			From: 10000000,
			To:   100000000,
		},
		{
			From: 100000000,
		},
	}
	res := &esApi.ProductPriceRangeResponse{
		Ranges: defaultRanges,
	}
	ranges, ok := aggregations["price_ranges"]
	if !ok {
		return res
	}

	var m model.EsPriceAggregation

	err := json.Unmarshal(ranges, &m)

	if err != nil {
		return res
	}
	val := m.Prices.Values

	var keyRanges []string

	for k := range val {
		keyRanges = append(keyRanges, k)
	}

	sort.Slice(keyRanges, func(i, j int) bool {
		numA, _ := strconv.ParseFloat(keyRanges[i], 64)
		numB, _ := strconv.ParseFloat(keyRanges[j], 64)
		return numA < numB
	})

	for _, k := range keyRanges {
		priceRanges = append(priceRanges, util.ProcessPrice(val[k]))
	}

	resRanges := pu.generatePriceRanges(priceRanges)
	return &esApi.ProductPriceRangeResponse{
		Ranges: resRanges,
		Total: m.DocCount,
	}
}

func (pu *productBuilder) generatePriceRanges(priceRanges []float64) []*esApi.PriceRange {
	var ranges []*esApi.PriceRange
	for i := range priceRanges {
		if i == 0 {
			tmp := &esApi.PriceRange{
				To: priceRanges[i],
			}
			ranges = append(ranges, tmp)
			continue
		}
		if priceRanges[i-1] == priceRanges[i] {
			//cho nay khong can hien thi vi range cua no qua thap
			continue
		}
		tmp := &esApi.PriceRange{
			From: priceRanges[i-1],
			To:   priceRanges[i],
		}
		ranges = append(ranges, tmp)
		if i == len(priceRanges)-1 {
			tmp := &esApi.PriceRange{
				From: priceRanges[i],
			}
			ranges = append(ranges, tmp)
		}
	}

	return ranges
}
