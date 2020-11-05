package builder

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/fatih/structs"
	kit_log "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	kit_logrus "github.com/go-kit/kit/log/logrus"
	"github.com/olivere/elastic/v7"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/model"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/pkg/helper"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/usecase/helpers"
	"gitlab.thovnn.vn/protobuf/internal-apis-go/base"
	productBase "gitlab.thovnn.vn/protobuf/internal-apis-go/product/base"
	esApi "gitlab.thovnn.vn/protobuf/internal-apis-go/product/es_service"
	"gitlab.thovnn.vn/r-d-1/rd/rd"
)

var (
	MinQuery         = "min"
	MaxQuery         = "max"
	GteQuery         = "gte"
	LteQuery         = "lte"
	RangeQuery       = "range"
	TermsQuery       = "terms"
	TermQuery        = "term"
	MatchPhraseQuery = "match_phrase"
	NestedQuery      = "nested"
)

var specialFields = map[string]string{
	"price": "fixed_price",
}

func (pu *productBuilder) BuildESQuery(req *esApi.ListV2Request) map[string]interface{} {
	mapFilters := buildFilters(req.Filters)
	buildPagination(req.Pagination, mapFilters)
	if len(req.Sorts) > 0 {
		mapFilters["sorts"] = req.GetSorts()
	}
	return mapFilters
}

func buildPagination(pagination *base.Pagination, mapFilters map[string]interface{}) {
	p := helper.GetPagination(int64(pagination.GetPage()), int64(pagination.GetLimit()))
	if p.Limit > 500 {
		p.Limit = 500
	}
	mapFilters["size"] = p.Limit
	mapFilters["offset"] = p.Offset
}

func buildFilters(filters *productBase.Filters) map[string]interface{} {
	if filters == nil {
		return nil
	}
	data := map[string]interface{}{}
	structFields := structs.Fields(filters)
	for _, f := range structFields {
		jsonTag := f.Tag("json")
		tag := strings.Split(jsonTag, ",")[0]
		if !f.IsZero() {
			switch tag {
			case "is_review", "is_promotion", "is_off", "is_certificate_shop", "has_variant", "has_certificate", "is_stock", "is_shop_support_shipping_fee", "is_event":
				buildBoolESFields(data, tag, f.Value())
			case "seller_admin_id":
				data["admin_id"] = f.Value()
			default:
				data[tag] = f.Value()
			}
		}
	}
	return data
}

func buildBoolESFields(data map[string]interface{}, field string, value interface{}) {
	esData := false
	if value == productBase.Filters_IS_TRUE {
		esData = true
	}
	data[field] = esData
}

func (pu *productBuilder) BuildListV2Response(data *elastic.SearchHits) *esApi.ListV2Response {
	products := []uint32{}
	for _, product := range data.Hits {
		source := make(map[string]interface{})
		err := json.Unmarshal(product.Source, &source)
		if err != nil {
			/*devlog.Error(err.Error(), map[string]interface{}{
				"err":    "json.Unmarshal error",
				"source": product.Source,
			}, "getListProduct")*/
			continue
		}
		if _, ok := source["product_id"]; !ok {
			continue
		}
		productId := uint32(source["product_id"].(float64))
		products = append(products, productId)
	}
	response := &esApi.ListV2Response{
		Data:     products,
		MetaData: &productBase.MetaData{Total: data.TotalHits.Value},
	}
	return response
}

func (pb *productBuilder) BuildListingScoreIndexQuery(req *esApi.ListingScoreRequest) *elastic.BoolQuery {
	var esQuery = elastic.NewBoolQuery()
	var filters []filter
	logger := kit_log.WithPrefix(kit_logrus.NewLogrusLogger(logrus.New()), "caller", kit_log.DefaultCaller)
	logger = kit_log.WithPrefix(logger, "prefix", "BuildListingScoreIndexQuery")
	logger = kit_log.WithPrefix(logger, "ts", time.Now())

	level.NewFilter(logger, level.AllowDebug())
	for _, v := range req.Filters {
		if exists := helpers.JsonTagExists(v.Key, "", reflect.TypeOf(model.ProductES{})); !exists {
			//logrus.WithTime(time.Now()).Debug("Field " + v.Key + " cannot be used for filtering")
			_ = level.Debug(logger).Log("msg", "Field "+v.Key+" cannot be used for filtering")
			continue
		}
		filter := filter{
			Key:   v.Key,
			Value: v.Value,
			Type:  v.QueryType,
		}
		filters = append(filters, filter)
	}
	if len(filters) == 0 {
		return nil
	}
	esFilters := buildEsQueryFilters(filters)
	esQuery.Filter(esFilters...)

	return esQuery
}

func (pb *productBuilder) BuildEsFilterQuery(esListingFilters []*esApi.EsListingFilters) *elastic.BoolQuery {
	var esQuery = elastic.NewBoolQuery()
	var filters []filter

	for _, v := range esListingFilters {
		var f filter
		if v.QueryType == NestedQuery {
			var nestedFilters []filter

			for _, v := range v.GetNestedQuery().GetFilters() {
				nestedFilters = append(nestedFilters, filter{Key:v.GetKey(), Value: v.GetValue()})
			}
			f = filter{
				Type:            v.QueryType,
				NestedFilter:nestedFilter{
					Filters:nestedFilters,
					Path: v.GetNestedQuery().GetNestedPath(),
				},
			}
		} else {
			f = filter{
				Key:             v.Key,
				Value:           v.Value,
				Type:            v.QueryType,
			}
		}

		if k, ok := specialFields[v.Key]; ok {
			f.RealKey = k
		}

		filters = append(filters, f)
	}
	if len(filters) == 0 {
		return nil
	}

	filters = mergePriceFilter(filters)

	esFilters := buildEsQueryFilters(filters)
	esQuery.Filter(esFilters...)
	return esQuery
}

func (pb *productBuilder) BuildEsSearchQuery(req *esApi.SearchQuery) (elastic.Query, *elastic.Rescore) {
	if req.GetQueryString() == "" || req.GetKeywordFormat() == "" {
		return nil, nil
	}
	query, rescore := rd.BuildInnerQueryForProductSearch(rd.ALGO_V11, req.GetIsAccentMark(), req.GetIsInCategory(), req.GetBrandParams(), req.GetShopId(), req.GetKeywordFormat(), req.GetIsSorted(), req.GetFirstTerm(), false)

	return query, rescore
}

func (pb *productBuilder) BuildEsSorters(esSorters []*esApi.Sorter) []elastic.Sorter {
	var sorters []elastic.Sorter
	for _, v := range esSorters {
		sorter, isSpecial := pb.buildSpecialEsSorter(v)
		if !isSpecial {
			sorter.Field = v.GetField()
			sorter.Ascending = v.GetOrder() != "desc"
		}
		sorters = append(sorters, sorter)
	}
	return sorters
}

func (pb *productBuilder) BuildESPriceRangeAggregation() elastic.Aggregation {
	nAggs := elastic.NewNestedAggregation().Path("promotions")
	agg := elastic.NewPercentilesAggregation().
		Field("promotions.fixed_price").
		Percentiles(10, 25, 50, 75, 90)
	nAggs.SubAggregation("prices", agg)
	return nAggs
}

func  (pb *productBuilder) buildSpecialEsSorter(esSorter *esApi.Sorter) (elastic.SortInfo, bool) {
	var sorter elastic.SortInfo
	var isSpecial bool
	switch esSorter.GetField() {
	case "price":
		var promotions = "promotions"
		var sort = elastic.NewNestedSort(promotions)
		boolQuery := elastic.NewBoolQuery()
		isSpecial = true
		from := elastic.NewRangeQuery(promotions + ".from").To(time.Now())
		to := elastic.NewRangeQuery(promotions + ".to").From(time.Now())
		priceFrom := elastic.NewRangeQuery(promotions + ".fixed_price").From(1)
		boolQuery.Must(from, to, priceFrom)
		sort.Filter(boolQuery)
		sorter.Nested = sort
		sorter.SortMode = "min"
		sorter.Ascending = esSorter.GetOrder() != "desc"
		sorter.Field = promotions + ".fixed_price"
	default:
	}
	return sorter, isSpecial
}

func buildSpecialEsFilters(f filter) (elastic.Query, bool, bool) {
	var esQuery elastic.Query
	var isSpecial bool
	var skip = false
	var err error
	logger := kit_log.WithPrefix(kit_logrus.NewLogrusLogger(logrus.New()), "caller", kit_log.DefaultCaller)
	logger = kit_log.WithPrefix(logger, "prefix", "buildSpecialEsFilters")
	logger = kit_log.WithPrefix(logger, "ts", time.Now())
	level.NewFilter(logger, level.AllowDebug())
	switch f.Key {
	case "price":
		esQuery, err = buildPriceFilter(f)
	case "is_promotion":
		esQuery, err = buildPromotionFilter(f)
	case "is_installment":
		esQuery, err = buildIsInstallmentFilter(f)
	}

	if err != nil {
		skip = true
		_ = level.Debug(logger).Log("msg", "Error on "+f.Key, err.Error())
	}
	if esQuery != nil {
		isSpecial = true
	}

	return esQuery, isSpecial, skip
}

func buildEsQueryFilters(query []filter) []elastic.Query {
	var esQuery []elastic.Query
	logger := kit_log.WithPrefix(kit_logrus.NewLogrusLogger(logrus.New()), "caller", kit_log.DefaultCaller)
	logger = kit_log.WithPrefix(logger, "prefix", "buildEsQueryFilters")
	logger = kit_log.WithPrefix(logger, "caller", kit_log.DefaultCaller)
	logger = kit_log.WithPrefix(logger, "ts", time.Now())
	level.NewFilter(logger, level.AllowDebug())

	for _, f := range query {
		q, ok, skip := buildSpecialEsFilters(f)
		if ok {
			esQuery = append(esQuery, q)
			continue
		}
		if skip {
			continue
		}
		q, err := buildEsQueryFilter(f)
		if err != nil {
			_ = level.Debug(logger).Log("msg", "Error on "+f.Key, err.Error())
			continue
		}
		esQuery = append(esQuery, q)
	}

	return esQuery
}

func buildEsQueryFilter(f filter) (elastic.Query, error) {
	var q elastic.Query

	if len(f.Value) == 0 && f.Type != NestedQuery {
		return q, errors.New("Value cannot be empty.")
	}
	switch f.Type {
	case RangeQuery:
		if len(f.Value) < 2 {
			return q, errors.New("Range query requires 2 values to build")
		}
		q = elastic.NewRangeQuery(f.Key).From(f.Value[0]).To(f.Value[1])
	case MinQuery:
		q = elastic.NewRangeQuery(f.Key).Gt(f.Value[0])
	case MaxQuery:
		q = elastic.NewRangeQuery(f.Key).Lt(f.Value[0])
	case GteQuery:
		q = elastic.NewRangeQuery(f.Key).Gte(f.Value[0])
	case LteQuery:
		q = elastic.NewRangeQuery(f.Key).Lte(f.Value[0])
	case MatchPhraseQuery:
		q = elastic.NewMatchPhraseQuery(f.Key, f.Value[0])
	case TermQuery:
		q = elastic.NewTermQuery(f.Key, f.Value[0])
	case TermsQuery:
		var val []interface{}
		for _, v := range f.Value {
			val = append(val, v)
		}
		q = elastic.NewTermsQuery(f.Key, val...)
	case NestedQuery:
		nestedFilter := f.NestedFilter

		if nestedFilter.Path == "" {
			return q, errors.New("Query type " + f.Type + " requires nested path")
		}
		var nestedQuery = elastic.NewBoolQuery()
		for _, f := range nestedFilter.Filters {
			var innerFilter = filter{
				Key:   nestedFilter.Path + "." + f.Key,
				Value: f.Value,
				Type:TermsQuery,
			}
			innerQuery, err := buildEsQueryFilter(innerFilter)
			if err != nil {
				continue
			}
			nestedQuery.Filter(innerQuery)
		}

		q = elastic.NewNestedQuery(nestedFilter.Path, nestedQuery)
	default:
		return q, errors.New("Query type " + f.Type + " is not supported")
	}

	return q, nil
}
