package builder

import (
	"errors"
	"github.com/olivere/elastic/v7"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/util"
	"math"
	"strconv"
	"time"
)

func mergePriceFilter(filters []filter) []filter {
	var priceFilters []filter
	var otherFilters []filter
	var priceFilter filter
	var prices []string
	var min = -1
	var max int
	for _, f := range filters {
		if f.Key == "price" {
			priceFilters = append(priceFilters, f)
			priceFilter.Key = f.Key
			priceFilter.RealKey = f.RealKey
			continue
		}
		otherFilters = append(otherFilters, f)
	}

	for _, f := range priceFilters {
		if len(f.Value) == 0 {
			continue
		}
		price, ok := util.InterfaceToInt(f.Value[0])
		if !ok {
			continue
		}
		if min == -1 || min > price {
			min = price
		}
		if max < price {
			max = price
		}
		prices = append(prices, strconv.Itoa(price))
		if priceFilter.Type == RangeQuery {
			continue
		}
		if ((f.Type == MaxQuery || f.Type == LteQuery) && (priceFilter.Type == MinQuery || priceFilter.Type == GteQuery)) ||
			((f.Type == MinQuery || f.Type == GteQuery) && (priceFilter.Type == MaxQuery || priceFilter.Type == LteQuery)){
			priceFilter.Type = RangeQuery
			continue
		}
		priceFilter.Type = f.Type
	}

	if priceFilter.Type == RangeQuery {
		priceFilter.Value = []string{
			strconv.Itoa(min),strconv.Itoa(max),
		}
	} else {
		priceFilter.Value = prices
	}

	return append(otherFilters, priceFilter)
}

func buildPriceFilter(f filter) (elastic.Query, error) {
	var promotions = "promotions"
	var outerBoolQuery = elastic.NewBoolQuery()
	var boolQuery = elastic.NewBoolQuery()
	f.Key = promotions + "." + f.RealKey
	from := elastic.NewRangeQuery(promotions + ".from").To(time.Now())
	to := elastic.NewRangeQuery(promotions + ".to").From(time.Now())
	q, err := buildEsQueryFilter(f)
	if err != nil {
		return nil, err
	}

	nextNestedQuery := elastic.NewNestedQuery(promotions, q).ScoreMode("min")
	boolQuery.Filter(from, to, q)
	nestedQuery := elastic.NewNestedQuery(promotions, boolQuery)
	outerBoolQuery.Filter(nestedQuery, nextNestedQuery)
	return outerBoolQuery, nil
}

func buildPromotionFilter (f filter) (elastic.Query, error) {
	if len(f.Value) == 0 || f.Value[0] != "true" {
		return nil, errors.New("no filter required")
	}
	var promotions = "promotions"
	boolQuery := elastic.NewBoolQuery()
	f.Key = promotions + ".fixed_price"
	f.Type = MinQuery
	f.Value = []string{"0"}
	//After 2000-01-01
	from := elastic.NewRangeQuery(promotions + ".from").To(time.Now()).Gt(time.Unix(946685800, 0))
	//Before 3000-01-01
	to := elastic.NewRangeQuery(promotions + ".to").From(time.Now()).Lt(time.Unix(32503670000, 0))
	q, err := buildEsQueryFilter(f)
	if err != nil {
		return nil, err
	}
	exists := elastic.NewExistsQuery(promotions + ".type")
	//Buyer type ranges from 0 to 199, hence < 200
	buyerType := elastic.NewRangeQuery(promotions + ".type").Lt(200)
	subBoolQuery := elastic.NewBoolQuery()
	subBoolQuery.MustNot(exists)
	subShouldQuery := elastic.NewBoolQuery()
	subShouldQuery.Should(subBoolQuery, buyerType)
	boolQuery.Filter(from, to, q, subShouldQuery)
	nestedQuery := elastic.NewNestedQuery(promotions, boolQuery)
	return nestedQuery, nil
}

func buildIsInstallmentFilter(f filter) (elastic.Query, error) {
	var promotions = "promotions"
	boolQuery := elastic.NewBoolQuery()
	outerQuery := elastic.NewBoolQuery()
	from := elastic.NewRangeQuery(promotions + ".from").To(time.Now())
	to := elastic.NewRangeQuery(promotions + ".to").From(time.Now())
	fromPrice := elastic.NewRangeQuery(promotions + ".fixed_price").From(3 * math.Pow(10, 6))
	q, err := buildEsQueryFilter(f)
	if err != nil {
		return nil, err
	}
	boolQuery.Filter(from, to, fromPrice)
	nestedQuery := elastic.NewNestedQuery(promotions, boolQuery)
	outerQuery.Filter(nestedQuery, q)
	return outerQuery, nil
}
