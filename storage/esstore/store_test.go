package esstore

import (
	"context"
	"encoding/json"
	"fmt"
	"gitlab.thovnn.vn/core/sen-kit/storage/elastic"
	"os"
	"testing"

	v7 "github.com/olivere/elastic/v7"
)

func TestQueryStringQuery(t *testing.T) {
	os.Setenv("ES_URL", "http://test.thovnn.vn:3019")
	os.Setenv("ES_INDEX", "product_v1")

	productIds := []interface{}{16672237, 16672238}
	boolQuery := v7.NewBoolQuery()
	boolQuery = boolQuery.Must(
		v7.NewTermsQuery("product_id", productIds...),
		v7.NewTermQuery("status_id", 2),
		v7.NewTermQuery("is_stock", true),
	)

	//esStore
	/*esConfig := elastic.New("")
	esStore := New(esConfig.Client(), esConfig.ESIndex)

	results, err := esStore.SearchWithTermQuery(context.Background(), boolQuery, ElasticOptions{})
	if err != nil {
		fmt.Println(err)
	}
	*/
	src, _ := boolQuery.Source()
	query := NewSearchQuery(src)
	ss, _ := query.Source()
	fmt.Println(JsonToString(ss))
}

func JsonToString(value interface{}) string {
	resJson, _ := json.Marshal(value)
	return string(resJson)
}

func TestQueryStringQuery1(t *testing.T) {
	os.Setenv("ES_URL", "http://test.thovnn.vn:3019")
	os.Setenv("ES_INDEX", "product_v1")
	options := ElasticOptions{
		Size:     5,
		Includes: []string{"product_id"},
	}
	boolQuery := v7.NewBoolQuery()
	esConfig := elastic.New("")
	esStore := New(esConfig.Client(), esConfig.ESIndex)

	mapFilters := map[string]interface{}{
		"is_stock":        true,
		"seller_admin_id": 11539,
	}
	for key, value := range mapFilters {
		boolQuery = boolQuery.Must(
			v7.NewTermQuery(key, value),
		)
	}

	termQuery := v7.NewTermQuery("message", "configuration")

	timeRangeFilter := elastic.NewRangeFilter("@timestamp").Gte(1442100154219).Lte(1442704954219)

	// Not sure this one is really useful, I just put it here to mimic what you said is the expected result
	boolFilter := elastic.NewBoolFilter().Must(timeRangeFilter)

	query := elastic.NewFilteredQuery(termQuery).Filter(boolFilter)

	results, err := esStore.SearchWithTermQuery(context.Background(), boolQuery, nil, options)
	if err != nil {
		fmt.Println(err)
	}

	src, _ := boolQuery.Source()
	query := NewSearchQuery(src)
	ss, _ := query.Source()
	fmt.Println(JsonToString(ss))

	jsonByte, _ := json.Marshal(results)
	fmt.Println(string(jsonByte))
}
