package esstore

import (
	"context"
	"errors"
	"encoding/json"
	"fmt"
	"github.com/go-kit/kit/log"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/util"
	"strings"
	"time"

	"github.com/go-kit/kit/log/level"
	"github.com/olivere/elastic/v7"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/model"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/pkg/helper"
	"gitlab.thovnn.vn/r-d-1/rd/rd"
)

var mapSort = map[string]string{
	//Seller
	"price_asc":             "final_price/asc",
	"price_desc":            "final_price/desc",
	"product_asc":           "product_id/asc",
	"product_desc":          "product_id/desc",
	"product_name_asc":      "name.raw/asc",
	"product_name_desc":     "name.raw/desc",
	"product_sku_user_asc":  "sku_defined_by_shop.raw/asc",
	"product_sku_user_desc": "sku_defined_by_shop.raw/desc",
	"product_status_asc":    "status_id/asc",
	"product_status_desc":   "status_id/desc",
	"product_price_asc":     "final_price/asc",
	"product_price_desc":    "final_price/desc",
	"product_stock_asc":     "is_stock/asc",
	"product_stock_desc":    "is_stock/desc",
	"updated_date_asc":      "updated_at/asc",
	"updated_date_desc":     "updated_at/desc",
	"created_date_asc":      "created_at/asc",
	"created_date_desc":     "created_at/desc",
	"review_date_desc":      "review_at/desc",
	"review_date_asc":       "review_at/asc",
	"seller_admin_id_desc":  "seller_admin_id/desc",
	"seller_admin_id_asc":   "seller_admin_id/asc",
	//Buyer
	"like_desc":           "counter_like/desc",
	"view_desc":           "counter_view/desc",
	"view_asc":            "counter_view/asc",
	"norder_30_desc":      "order_count_rank/desc",
	"norder_30_asc":       "order_count_rank/asc",
	"rank_desc":           "rank_search/desc",
	"rank_asc":            "rank_search/asc",
	"promotion_asc":       "promotion_percent/asc",
	"promotion_desc":      "promotion_percent/desc",
	"vasup_desc":          "vasup/desc",
	"rating_percent_asc":  "rating_percent/asc",
	"rating_percent_desc": "rating_percent/desc",
	"cate2_score":         "default_listing_score.Score_Cate2/desc",
	"cate3_score":         "default_listing_score.Score_Cate3/desc",
	"hotsell_15_desc":     "rank_updated_time/desc",
	"listing_v2_desc":     "listing_score_v2.Score_v2/desc",
	"listing_v3_desc":     "listing_score.Score_v3/desc",
}

const (
	LIMIT = 5
)

type ResponseItemES struct {
	Type       string `json:"type" bson:"type"`
	ProductId  string `json:"product_id" bson:"product_id"`
	Status     int    `json:"status" bson:"status"`
	Error      string `json:"error" bson:"error"`
	RetryCount int    `json:"retry_count"`
}

type ElasticOptions struct {
	Offset   int64
	Size     int64
	Sorters  []elastic.Sorter
	Includes []string
	Excludes []string
}

type GetAttributeFacetsResult struct {
	Attribute string
	Options   []int64
}

type SearchQuery struct {
	source interface{}
}

func NewSearchQuery(src interface{}) *SearchQuery {
	return &SearchQuery{source: src}
}

func (sq *SearchQuery) Source() (interface{}, error) {
	source := make(map[string]interface{})
	source["query"] = sq.source
	return source, nil
}

type ESStore interface {
	Search(request model.SearchRequest) (model.SearchResult, error)
	Update(data interface{}, Id string) (*elastic.UpdateResponse, error)
	Add(data interface{}, Id string) error
	UpdateBulkByInputData(index string, mapProductAdminIds map[int32]int32, updateData map[string]interface{}) error
	UpdateBulk(changedProducts []map[string]interface{}) (*elastic.BulkResponse, error)
	UpsertBulk(data map[uint32]interface{}) (*elastic.BulkResponse, error)
	SearchWithTermQuery(context context.Context, query elastic.Query, rescore *elastic.Rescore, options ElasticOptions) (*elastic.SearchHits, error)
	AggQuery(ctx context.Context, query elastic.Query, agg elastic.Aggregation) (elastic.Aggregations, error)
	SellerSearch(context context.Context, mapFilters map[string]interface{}) (*elastic.SearchHits, error)
	GetProductByMerchantId(merchantId int32, lastProductId int32) ([]int32, error)
	Updates(data []map[string]interface{}, productIds []string) (*elastic.BulkResponse, error)
	GetAttributeFacets(categoryID int64, attributes []string) ([]GetAttributeFacetsResult, error)
	GetSearchAttributeFilter(key string) ([]GetAttributeFacetsResult, error)
	GetById(productId string) ([]model.Promotion, error)
	AddPromotions(productId string, promotions map[string]interface{}) (*elastic.UpdateResponse, error)
	UpdatePromotions(productId string, promotions map[string]interface{}) (*elastic.UpdateResponse, error)
	DeletePromotions(productId string, promotions map[string]interface{}) (*elastic.UpdateResponse, error)
	UpdatePromotionsForSeller(productId string, promotions map[string]interface{}) (*elastic.UpdateResponse, error)
}

type eSProductStore struct {
	clientFunc func() (*elastic.Client, error)
	index      string
	logger     log.Logger
}

func New(clientFunc func() (*elastic.Client, error), index string, logger log.Logger) (store ESStore) {
	store = &eSProductStore{
		clientFunc: clientFunc,
		index:      index,
		logger:     logger,
	}
	return store
}

func (s *eSProductStore) GetById(productId string) ([]model.Promotion, error) {
	var results []model.Promotion
	client, err := s.clientFunc()
	if err != nil {
		return nil, err
	}
	defer client.Stop()
	get, err := client.Get().
		Index(s.index).
		Id(productId).
		Do(context.Background())
	if err != nil {
		_ = level.Error(s.logger).Log("msg", fmt.Sprint("Getting es results", err.Error()))
		return nil, err
	}
	if get.Found {
		document, _ := json.Marshal(get.Source)
		var productES model.ProductES
		json.Unmarshal(document, &productES)
		results = productES.Promotions
	}
	return results, nil
}

func (s *eSProductStore) AddPromotions(productId string, promotions map[string]interface{}) (*elastic.UpdateResponse, error) {
	client, _ := s.clientFunc()
	defer client.Stop()
	return client.
		Update().
		Index(s.index).
		Id(productId).
		Script(elastic.NewScript("if(ctx._source.promotions == null){ctx._source.promotions=params.promotions;} else {for (def promotion : params.promotions){ctx._source.promotions.add(promotion);}}").
			Params(promotions)).
		Do(context.Background())
}

func (s *eSProductStore) UpdatePromotions(productId string, promotions map[string]interface{}) (*elastic.UpdateResponse, error) {
	client, _ := s.clientFunc()
	defer client.Stop()
	return client.
		Update().
		Index(s.index).
		Id(productId).
		Script(elastic.NewScript("for(def promotion:params.promotions){for(int i=0;i<ctx._source.promotions.length;i++){if(promotion.id==ctx._source.promotions[i].id&&promotion.updated_at>=0&&(ctx._source.promotions[i].updated_at==null||ctx._source.promotions[i].updated_at==0||ctx._source.promotions[i].updated_at<promotion.updated_at)){ctx._source.promotions[i].fixed_price=promotion.fixed_price;ctx._source.promotions[i].from=promotion.from;ctx._source.promotions[i].to=promotion.to;ctx._source.promotions[i].type=promotion.type;ctx._source.promotions[i].updated_at=promotion.updated_at;}}}").
			Params(promotions)).
		Do(context.Background())
}

func (s *eSProductStore) DeletePromotions(productId string, promotions map[string]interface{}) (*elastic.UpdateResponse, error) {
	client, _ := s.clientFunc()
	defer client.Stop()
	return client.
		Update().
		Index(s.index).
		Id(productId).
		Script(elastic.NewScript("ctx._source.promotions.removeIf(promotion -> params.promotion_ids.contains(promotion.id))").
			Params(promotions)).
		Do(context.Background())
}

func (s *eSProductStore) UpdatePromotionsForSeller(productId string, promotions map[string]interface{}) (*elastic.UpdateResponse, error) {
	client, _ := s.clientFunc()
	defer client.Stop()
	return client.
		Update().
		Index(s.index).
		Id(productId).
		Script(elastic.NewScript("if(ctx._source.promotions == null){ctx._source.promotions=params.promotions;}else{ctx._source.promotions.removeIf(promotion -> {promotion.type==null||promotion.type<=200});for (def promotion : params.promotions){ctx._source.promotions.add(promotion);}}").
			Params(promotions)).
		Do(context.Background())
}


func (s *eSProductStore) Update(data interface{}, Id string) (*elastic.UpdateResponse, error) {
	client, err := s.clientFunc()
	if err != nil {
		return nil, err
	}
	defer client.Stop()
	res, err := client.
		Update().
		Index(s.index).
		Id(Id).
		Doc(data).
		Do(context.Background())
	if err != nil {
		return nil, err
	}

	if res.Status >= 300 {
		if res.GetResult != nil && res.GetResult.Error != nil {
			return nil, errors.New(res.GetResult.Error.Reason)
		}
		return nil, errors.New(res.Result)
	}

	return res, nil
}

func (s *eSProductStore) Updates(data []map[string]interface{}, productIds []string) (*elastic.BulkResponse, error) {
	client, err := s.clientFunc()
	if err != nil {
		return nil, err
	}
	defer client.Stop()
	bulkUpdateRequest := client.Bulk().Index(s.index)
	for i, p := range data {
		updateReq := elastic.NewBulkUpdateRequest().
			Index(s.index).
			Id(productIds[i]).
			Doc(p)
		bulkUpdateRequest = bulkUpdateRequest.Add(updateReq)
	}
	return bulkUpdateRequest.Do(context.Background())
}

func (s *eSProductStore) Add(data interface{}, Id string) error {
	client, err := s.clientFunc()
	if err != nil {
		return err
	}
	defer client.Stop()
	_, err = client.
		Index().
		Index(s.index).
		Id(Id).
		BodyJson(data).
		Do(context.Background())
	if err != nil {
		level.Error(s.logger).Log("msg", "Add err:"+err.Error())
	}
	return err
}

func (s *eSProductStore) UpdateBulk(changedProducts []map[string]interface{}) (*elastic.BulkResponse, error) {
	client, err := s.clientFunc()
	if err != nil {
		return nil, err
	}

	defer client.Stop()

	var response *elastic.BulkResponse
	bulkRequest := client.Bulk()

	for _, mapProduct := range changedProducts {
		if productId, ok := mapProduct["product_id"]; ok {
			id, ok := util.InterfaceToUint(productId)
			if !ok {
				continue
			}
			sid := fmt.Sprint(id)
			updateReq := elastic.NewBulkUpdateRequest().
				Index(s.index).
				Type("_doc").
				Id(sid).
				Doc(mapProduct)
			bulkRequest = bulkRequest.Add(updateReq)
		}
	}

	if bulkRequest.NumberOfActions() > 0 {
		response, err = bulkRequest.Do(context.Background())
		if err != nil {
			return response, err
		}
	}

	return response, nil
}

func (s *eSProductStore) UpsertBulk(data map[uint32]interface{}) (*elastic.BulkResponse, error) {
	client, err := s.clientFunc()
	if err != nil {
		return nil, err
	}

	defer client.Stop()

	var response *elastic.BulkResponse
	bulkRequest := client.Bulk()

	for id, mapProduct := range data {
		id := fmt.Sprint(id)
		updateReq := elastic.NewBulkUpdateRequest().
			Index(s.index).
			Type("_doc").
			Id(id).
			Doc(mapProduct).
			DocAsUpsert(true)
		bulkRequest = bulkRequest.Add(updateReq)
	}

	if bulkRequest.NumberOfActions() > 0 {
		response, err = bulkRequest.Do(context.Background())
		if err != nil {
			return response, err
		}
	}

	return response, nil
}

func (s *eSProductStore) Search(request model.SearchRequest) (result model.SearchResult, err error) {
	return result, err
}

func (s *eSProductStore) UpdateBulkByInputData(index string, mapProductAdminIds map[int32]int32, updateData map[string]interface{}) error {
	client, err := s.clientFunc()
	if err != nil {
		return err
	}
	defer client.Stop()
	var limit = 0
	var bulkResponse *elastic.BulkResponse
	var errDoRequest error
	bulkRequest := client.Bulk().Index(index).Type(index)
	for productId, adminId := range mapProductAdminIds {
		level.Debug(s.logger).Log("msg", fmt.Sprint("P:", productId, "Merchant:", adminId))
		updateReq := elastic.NewBulkUpdateRequest().
			Index(index).
			Type("_doc").
			Id(fmt.Sprint(productId)).
			Doc(updateData)
		bulkRequest = bulkRequest.Add(updateReq)

		limit++
		if limit%LIMIT == 0 {
			bulkResponse, errDoRequest = bulkRequest.Do(context.Background())
			if errDoRequest != nil {
				return errDoRequest
			}
		}
	}
	if bulkRequest.NumberOfActions() > 0 {
		bulkResponse, errDoRequest = bulkRequest.Do(context.Background())
		if errDoRequest != nil {
			return errDoRequest
		}
	}

	var responseItems []*ResponseItemES
	for _, f := range bulkResponse.Failed() {
		responseItem := ResponseItemES{
			Type:      f.Type,
			ProductId: f.Id,
			Status:    f.Status,
			Error:     f.Error.Reason,
		}
		responseItems = append(responseItems, &responseItem)
	}
	for _, s := range bulkResponse.Succeeded() {
		responseItem := ResponseItemES{
			Type:      s.Type,
			ProductId: s.Id,
			Status:    200,
			Error:     "",
		}
		responseItems = append(responseItems, &responseItem)
	}

	if len(responseItems) > 0 {
		for _, value := range responseItems {
			fmt.Printf("%+v\n", value.Error)
		}
	}
	return nil
}

func (s *eSProductStore) SearchWithTermQuery(context context.Context, query elastic.Query, rescore *elastic.Rescore, options ElasticOptions) (*elastic.SearchHits, error) {
	client, err := s.clientFunc()
	if err != nil {
		return nil, err
	}
	defer client.Stop()

	var results *elastic.SearchResult

	ss := elastic.NewSearchSource().Query(query).TrackTotalHits(true).From(int(options.Offset)).Size(int(options.Size))
	//Rescorer doesn't work with sort
	if len(options.Sorters) == 0 && rescore != nil {
		ss = ss.Rescorer(rescore.WindowSize(180))
	} else {
		ss = ss.SortBy(options.Sorters...)
	}
	source, _ := ss.Source()
	data, _ := json.Marshal(source)
	results, err = client.Search().
		Index(s.index).
		Source(source).
		//SortBy(options.Sorters...).
		//From(int(options.Offset)). // Starting from this result
		//Size(int(options.Size)).   // Limit of responds
		Pretty(true).
		FetchSourceContext(
			elastic.NewFetchSourceContext(true).Include(options.Includes...).Exclude(options.Excludes...),
		).
		Do(context)
	if err != nil {
		_ = level.Error(s.logger).Log("msg", fmt.Sprint("Getting es results", err.Error(), ", query string: ", string(data)))
	}

	if results == nil {
		return nil, fmt.Errorf("es failed to send correct response")
	}

	return results.Hits, nil
}

func (s *eSProductStore) SellerSearch(context context.Context, mapFilters map[string]interface{}) (*elastic.SearchHits, error) {
	//Get options (offset, limit, sort)
	options := getElasticOptions(mapFilters)
	//Get filters
	filters := getFilters(mapFilters)
	return s.SearchWithTermQuery(context, filters, nil, options)
}

func getElasticOptions(mapFilters map[string]interface{}) ElasticOptions {
	offset := getNumberFieldValue("offset", mapFilters)
	size := getNumberFieldValue("size", mapFilters)
	options := ElasticOptions{
		Offset:   offset,
		Size:     size,
		Includes: []string{"product_id"},
	}
	if sorts, ok := mapFilters["sorts"]; ok {
		delete(mapFilters, "sorts")
		for _, sort := range sorts.([]string) {
			fieldSort, isSortAsc := getSortFieldValue(sort)
			if fieldSort == "" {
				continue
			}
			if fieldSort == "final_price" {
				options.Sorters = append(options.Sorters, getSorterByFinalPrice(isSortAsc))
			} else {
				options.Sorters = append(options.Sorters, elastic.SortInfo{Field: fieldSort, Ascending: isSortAsc})
			}
		}
	}
	return options
}

func getSorterByFinalPrice(isSortAsc bool) elastic.SortInfo {
	boolQuery := elastic.NewBoolQuery()
	boolQuery.Must(elastic.NewRangeQuery("promotions.from").Lte("now"))
	boolQuery.Must(elastic.NewRangeQuery("promotions.to").Gte("now"))
	nestedSort := elastic.NewNestedSort("promotions").Filter(boolQuery)
	return elastic.SortInfo{
		Field:     "promotions.fixed_price",
		Ascending: isSortAsc,
		Nested:    nestedSort,
		SortMode:  "min",
	}
}

//getFilters: Only fields are filtered. It's has data filter.
/**
@function: getFilters
@params: mapFilters
*/
func getFilters(mapFilters map[string]interface{}) *elastic.BoolQuery {
	filters := elastic.NewBoolQuery()
	finalPriceQuery := getFinalPriceQuery(mapFilters)
	if finalPriceQuery != nil {
		filters = filters.Filter(getFinalPriceQuery(mapFilters))
	}
	//Case filter final_price
	for key, value := range mapFilters {
		filters = getFilter(filters, key, value)
	}
	return filters
}

func getFilter(filters *elastic.BoolQuery, key string, value interface{}) *elastic.BoolQuery {
	switch key {
	case "status_id", "product_id", "shop_type", "category_path":
		filters = filters.Filter(getTermsQueryByField(key, value.(string)))
	case "category_ids":
		filters = filters.Filter(getTermsQueryByField(key, fmt.Sprint(value)))
	case "updated_at_from", "review_at_from":
		filters = filters.Filter(getRangeQuery(key, true, time.Unix(value.(int64), 0)))
	case "updated_at_to", "review_at_to":
		filters = filters.Filter(getRangeQuery(key, false, time.Unix(value.(int64), 0)))
	case "final_price_from", "final_price_to":
		return filters
	case "order_complete_30_from":
		filters = filters.Filter(getRangeQuery("order_count.dd_30_from", true, value))
	case "order_complete_30_to":
		filters = filters.Filter(getRangeQuery("order_count.dd_30_to", false, value))
	case "rating_from":
		filters = filters.Filter(getRangeQuery("rating_percentage_from", true, value))
	case "rating_to":
		filters = filters.Filter(getRangeQuery("rating_percentage_to", false, value))
	case "is_promotion":
		var promotions = "promotions"
		boolQuery := elastic.NewBoolQuery()
		from := elastic.NewRangeQuery(promotions + ".from").To(time.Now()).Gt(time.Unix(946685800,0))
		//Before 3000-01-01
		to := elastic.NewRangeQuery(promotions + ".to").From(time.Now()).Lt(time.Unix(32503670000,0))
		priceFrom := elastic.NewRangeQuery(promotions + ".fixed_price").Gt(0)
		exists := elastic.NewExistsQuery(promotions + ".type")
		//Buyer type ranges from 0 to 199, hence < 200
		buyerType := elastic.NewRangeQuery(promotions + ".type").Lt(200)
		subBoolQuery := elastic.NewBoolQuery()
		subBoolQuery.MustNot(exists)
		subShouldQuery := elastic.NewBoolQuery()
		subShouldQuery.Should(subBoolQuery, buyerType)
		boolQuery.Filter(from, to, priceFrom, subShouldQuery)
		nestedQuery := elastic.NewNestedQuery(promotions, boolQuery)
		filters.Filter(nestedQuery)
	case "promotion_start_date":
		filters = filters.Filter(getNestedQuery("promotions", "promotions.from", false, time.Unix(value.(int64), 0)))
	case "promotion_to_date":
		filters = filters.Filter(getNestedQuery("promotions", "promotions.to", true, time.Unix(value.(int64), 0)))
	case "keyword":
		filters = filters.Must(getKeywordQuery(value.(string)))
	case "name", "sku_defined_by_shop":
		filters = filters.Must(getStringQuery(key, value.(string)))
	case "inside_product_type":
		filters = filters.Filter(getInsideProductType(value.(string)))
	case "extended_shipping_package":
		query := getExtendedShippingPackageQuery(value.(int64))
		filters = filters.Filter(query...)
	default:
		filters = filters.Filter(
			elastic.NewTermQuery(key, value),
		)
	}
	return filters
}

func getTermQueryByField(key string, value interface{}) (query *elastic.TermQuery) {
	return elastic.NewTermQuery(key, value)
}

func getInsideProductType(productType string) (query *elastic.TermQuery) {
	if productType == "2" {
		return getTermQueryByField("has_certificate", false)
	}
	return getTermQueryByField("is_updated", productType)
}

func getKeywordQuery(keyword string) (query elastic.Query) {
	firstTerm := ""
	keywordFormat := strings.ToLower(strings.TrimSpace(keyword))
	i := strings.Index(keywordFormat, " ")
	if i > -1 {
		firstTerm = keywordFormat[:i]
	}
	query, _ = rd.BuildInnerQueryForProductSearch("algo4", true, false, []string{}, []string{}, keywordFormat, false, firstTerm, false)
	return query
}

func getStringQuery(key, stringValue string) (query *elastic.MatchPhraseQuery) {
	return elastic.NewMatchPhraseQuery(key, stringValue)
}

func getNumberFieldValue(field string, mapFilters map[string]interface{}) int64 {
	if number, ok := mapFilters[field]; ok {
		delete(mapFilters, field)
		return number.(int64)
	}
	return 0
}

func getSortFieldValue(field string) (string, bool) {
	isSortAsc := false
	sorts := strings.Split(field, ",")
	for key, value := range mapSort {
		sliceSort := strings.Split(value, "/")
		for _, sortBy := range sorts {
			if key == sortBy {
				if sliceSort[1] == "asc" {
					isSortAsc = true
				}
				return sliceSort[0], isSortAsc
			}
		}
	}
	return "", isSortAsc
}

func getsliceStringValue(strValue string) []interface{} {
	sliceValue := strings.Split(strValue, ",")
	strValues := make([]interface{}, len(sliceValue))
	for index, value := range sliceValue {
		strValues[index] = value
	}
	return strValues
}

func getTermsQueryByField(fieldSet, value string) *elastic.TermsQuery {
	sliceString := getsliceStringValue(value)
	return elastic.NewTermsQuery(fieldSet, sliceString...)
}

//getRangeQuery: get filter by range
/**
@function: getRangeQuery
@params: field string, isGte bool, value interface{}
*/
func getRangeQuery(field string, isGte bool, value interface{}) *elastic.RangeQuery {
	sliceFieldName := helper.RemoveLastString(field, "_")
	rangeQuery := elastic.NewRangeQuery(sliceFieldName)
	if isGte {
		rangeQuery.Gte(value)
	} else {
		rangeQuery.Lte(value)
	}
	return rangeQuery
}

func getNestedQuery(field, nestedField string, isGte bool, value interface{}) (nestedQuery *elastic.NestedQuery) {
	if isGte {
		nestedQuery = elastic.NewNestedQuery(field, elastic.NewRangeQuery(nestedField).Gte(value))
	} else {
		nestedQuery = elastic.NewNestedQuery(field, elastic.NewRangeQuery(nestedField).Lte(value))
	}
	return nestedQuery
}

func getFinalPriceQuery(mapFilters map[string]interface{}) *elastic.NestedQuery {
	boolQuery := elastic.NewBoolQuery()
	isFilterFinalPrice := false
	var finalPriceFrom float64 = 0
	var finalPriceTo float64 = 0
	if value, ok := mapFilters["final_price_from"]; ok {
		isFilterFinalPrice = true
		finalPriceFrom = value.(float64)
	}
	if value, ok := mapFilters["final_price_to"]; ok {
		isFilterFinalPrice = true
		finalPriceTo = value.(float64)
	}
	if !isFilterFinalPrice {
		return nil
	}
	boolQuery = boolQuery.Must(elastic.NewRangeQuery("promotions.from").Lte("now"))
	boolQuery = boolQuery.Must(elastic.NewRangeQuery("promotions.to").Gte("now"))
	if finalPriceFrom > 0 {
		boolQuery = boolQuery.Must(elastic.NewRangeQuery("promotions.fixed_price").Gte(finalPriceFrom))
	}
	if finalPriceTo > 0 {
		boolQuery = boolQuery.Must(elastic.NewRangeQuery("promotions.fixed_price").Lte(finalPriceTo))
	}
	return elastic.NewNestedQuery("promotions", boolQuery)
}

/*
	1: Hoả Tốc
	2: Trong ngày
	3: Hoả Tốc + Trong Ngày
*/
func getExtendedShippingPackageQuery(extendedShippingPackage int64) (query []elastic.Query) {
	if extendedShippingPackage == 1 {
		query = append(query, elastic.NewTermQuery("extended_shipping_package.is_using_instant", true))
	} else if extendedShippingPackage == 2 {
		query = append(query, elastic.NewTermQuery("extended_shipping_package.is_using_in_day", true))
	} else if extendedShippingPackage == 3 {
		query = append(query, elastic.NewTermQuery("extended_shipping_package.is_using_instant", true))
		query = append(query, elastic.NewTermQuery("extended_shipping_package.is_using_in_day", true))
	}
	return query
}

func JsonToString(value interface{}) string {
	resJson, _ := json.Marshal(value)
	return string(resJson)
}

func (s *eSProductStore) GetProductByMerchantId(merchantId int32, lastProductId int32) ([]int32, error) {
	sorter := elastic.NewFieldSort("product_id").Asc()
	var productIdArr []int32
	options := ElasticOptions{
		Size:     100,
		Includes: []string{"product_id"},
	}
	options.Sorters = append(options.Sorters, sorter)

	boolQuery := elastic.NewBoolQuery()
	boolQuery = boolQuery.Must(
		elastic.NewTermQuery("seller_admin_id", merchantId),
	)
	timeRange := elastic.NewRangeQuery("product_id").Gt(lastProductId)
	boolQuery = boolQuery.Filter(timeRange)
	results, err := s.SearchWithTermQuery(context.Background(), boolQuery, nil, options)

	if results == nil {
		return productIdArr, err
	}

	for _, product := range results.Hits {
		source := make(map[string]interface{})
		err := json.Unmarshal(product.Source, &source)
		if err != nil {
			return productIdArr, err
		}
		productId := int32(source["product_id"].(float64))
		productIdArr = append(productIdArr, productId)
	}

	return productIdArr, nil
}

func (s *eSProductStore) AggQuery(ctx context.Context, query elastic.Query, agg elastic.Aggregation) (elastic.Aggregations, error) {
	client, err := s.clientFunc()
	if err != nil {
		return nil, err
	}
	defer client.Stop()
	ss := elastic.NewSearchSource().Query(query).TrackTotalHits(false).Size(0)
	ss.Aggregation("price_ranges", agg)

	source, _ := ss.Source()
	data, _ := json.Marshal(source)
	fmt.Println(string(data))
	results, err := client.Search().
		Index(s.index).
		Source(source).
		Pretty(true).
		Do(ctx)
	if err != nil {
		_ = level.Info(s.logger).Log("msg", "SearchWithTermQuery, query: "+string(data))
		_ = level.Error(s.logger).Log("msg", fmt.Sprint("Getting es results", err.Error()))
		return nil, err
	}

	return results.Aggregations, nil
}

func (s *eSProductStore) GetAttributeFacets(categoryID int64, attributes []string) ([]GetAttributeFacetsResult, error) {
	//convert []string to []interface{}
	af := make([]interface{}, len(attributes))
	for i, v := range attributes {
		af[i] = v
	}
	client, err := s.clientFunc()
	if err != nil {
		return nil, err
	}
	defer client.Stop()

	q := elastic.NewTermQuery("category_ids", categoryID)
	saf := elastic.NewTermsAggregation().Field("number_facets.value")
	sa := elastic.NewTermsAggregation().Field("number_facets.name").IncludeValues(af...).SubAggregation("facet_value", saf)
	a := elastic.NewNestedAggregation().Path("number_facets").SubAggregation("facet_name", sa)
	r, err := client.Search().Index(s.index).Query(q).Aggregation("number_facets", a).Size(0).Do(context.Background())
	if err != nil {
		return nil, err
	}
	if r.Error != nil {
		return nil, errors.New(r.Error.Reason)
	}
	awo := make([]GetAttributeFacetsResult, 0, len(attributes))
	if nfAgg, ok := r.Aggregations.Filter("number_facets"); ok {
		if fnAgg, ok := nfAgg.Filters("facet_name"); ok {
			for _, bucket := range fnAgg.Buckets {
				if facetValue, ok := bucket.Terms("facet_value"); ok {
					bv := make([]int64, 0, len(facetValue.Buckets))
					for _, bk := range facetValue.Buckets {
						if keyAsf64, err := bk.KeyNumber.Float64(); err == nil {
							bv = append(bv, int64(keyAsf64))
						}
					}
					awo = append(awo, GetAttributeFacetsResult{Attribute: bucket.KeyNumber.String(), Options: bv})
				}
			}
		}
	}
	return awo, nil
}

func (s *eSProductStore) GetSearchAttributeFilter(keyword string) ([]GetAttributeFacetsResult, error) {
	var resp []GetAttributeFacetsResult
	client, err := s.clientFunc()
	if err != nil {
		return resp, err
	}
	defer client.Stop()
	firstTerm := ""
	keywordFormat := strings.ToLower(strings.TrimSpace(keyword))
	i := strings.Index(keywordFormat, " ")
	if i > -1 {
		firstTerm = keywordFormat[:i]
	}
	q, _ := rd.BuildInnerQueryForProductSearch("algo4", util.CheckIsAccentMarks(keyword), false, []string{}, []string{}, keywordFormat, false, firstTerm, false)
	saf := elastic.NewTermsAggregation().Field("number_facets.value")
	sa := elastic.NewTermsAggregation().Field("number_facets.name").SubAggregation("option_ids", saf)
	a := elastic.NewNestedAggregation().Path("number_facets").SubAggregation("attribute_id", sa)
	r, err := client.Search().Index(s.index).Query(q).Aggregation("number_facets", a).Size(0).Do(context.Background())
	if err != nil {
		fmt.Println("aggregation has error:", err.Error())
		return resp, err
	}

	awo := make([]GetAttributeFacetsResult, 0)
	if nfa, ok := r.Aggregations.Filters("number_facets"); ok {
		if ata, ok := nfa.Filters("attribute_id"); ok {
			for _, a := range ata.Buckets {
				if opts, ok := a.Filters("option_ids"); ok {
					o := make([]int64, 0, len(opts.Buckets))
					for _, opt := range opts.Buckets {
						if fo, err := opt.KeyNumber.Float64(); err == nil {
							o = append(o, int64(fo))
						}
					}
					awo = append(awo, GetAttributeFacetsResult{
						Attribute: a.KeyNumber.String(),
						Options:   o,
					})
				}
			}
		}
	}
	return awo, nil
}
