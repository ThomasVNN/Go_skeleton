package query

import (
	"context"
	"errors"
	"fmt"
	kit_log "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/olivere/elastic/v7"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/model"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/storage/esstore"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/storage/mgostore"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/usecase/builder"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/usecase/helpers"
	esApi "gitlab.thovnn.vn/protobuf/internal-apis-go/product/es_service"
	"reflect"
)

type ProductQuery interface {
	//Search(ctx context.Context, request SearchRequest) (SearchResult, error)
	//SellerSearch(ctx context.Context, request *es_service.ListV2Request) (*es_service.ListV2Response, error)
	ListV2(ctx context.Context, req *esApi.ListV2Request) (res *esApi.ListV2Response, err error)
	ListWithShufflingScores(context.Context, *esApi.ListingScoreRequest) (res *esApi.ListingScoreResponse, err error)
	ListBuyer(context.Context, *esApi.BuyerSearchRequest) (res *esApi.BuyerSearchResponse, err error)
	PriceRangeQuery(context.Context, *esApi.ProductPriceRangeRequest) (res *esApi.ProductPriceRangeResponse, err error)
	ListCategoriesFilters(context.Context, *esApi.ListCategoriesFiltersRequest) (*esApi.ListCategoriesFiltersResponse, error)
	ListSearchFilters(context.Context, *esApi.ListSearchFiltersRequest) (*esApi.ListSearchFiltersResponse, error)
}

type productQuery struct {
	esStore  esstore.ESStore
	mgoStore mgostore.MgoStore
	logger kit_log.Logger
}

func New(esStore esstore.ESStore, mgoStore mgostore.MgoStore,logger kit_log.Logger) (pQuery ProductQuery, err error) {
	pQuery = &productQuery{
		esStore:  esStore,
		mgoStore: mgoStore,
		logger: logger,
	}

	return pQuery, nil
}

func (m *productQuery) Search(ctx context.Context, request model.SearchRequest) (result model.SearchResult, err error) {
	return result, err
}

func (m *productQuery) ListV2(ctx context.Context, req *esApi.ListV2Request) (res *esApi.ListV2Response, err error) {
	result := &esApi.ListV2Response{
		Data:     nil,
		MetaData: nil,
	}
	pBuilder := builder.New()
	mapFilters := pBuilder.BuildESQuery(req)
	data, err := m.esStore.SellerSearch(ctx, mapFilters)
	if err != nil {
		return result, err
	}
	if data == nil || (data.TotalHits == nil) || (data.TotalHits.Value <= 0 || len(data.Hits) <= 0) {
		return result, nil
	}
	result = pBuilder.BuildListV2Response(data)
	return result, err
}

func (m *productQuery) ListWithShufflingScores(ctx context.Context, req *esApi.ListingScoreRequest) (*esApi.ListingScoreResponse, error) {
	result := &esApi.ListingScoreResponse{
		Data:     nil,
		MetaData: nil,
	}
	pBuilder := builder.New()

	esFilter := pBuilder.BuildEsFilterQuery(req.GetFilters())
	operation := req.GetOperations()
	page := int64(operation.GetPage())
	limit := operation.GetQuerySize()
	if limit < operation.GetResponseSize() {
		limit = operation.GetResponseSize()
	}
	if limit <= 0 {
		limit = 10
	}
	if page < 1 {
		page = 1
	}

	sortBy := operation.GetEsSort()

	if sortBy == "" {
		sortBy = "_id"
	} else if !helpers.JsonTagExists(sortBy, "", reflect.TypeOf(model.ProductES{})) {
		return nil, errors.New(fmt.Sprintf("Sort type %s is not supported", sortBy))
	}

	sorter := elastic.NewFieldSort(sortBy).Asc()

	if operation.GetOrderBy() == "desc" {
		sorter.Desc()
	}

	options := esstore.ElasticOptions{
		Offset:   (page - 1) * limit,
		Size:     limit,
		Includes: []string{"default_listing_scores", "product_id", "listing_score", "order_count.rank", "rank_search"},
	}
	options.Sorters = append(options.Sorters, sorter)

	data, err := m.esStore.SearchWithTermQuery(ctx, esFilter, nil, options)
	if err != nil {
		return result, err
	}
	if data == nil || (data.TotalHits == nil) || (data.TotalHits.Value <= 0 || len(data.Hits) <= 0) {
		return result, nil
	}
	result = pBuilder.BuildListingScoreResponse(data)

	if operation.GetBusinessSort() != "" {
		result.Data = helpers.SortProducts(result.Data, operation.GetBusinessSort())
	}

	if operation.GetResponseSize() > 0 && operation.GetResponseSize() < operation.GetQuerySize() {
		result.Data = result.Data[:operation.GetResponseSize()]
	}

	if operation.GetShuffleSize() > 0 {
		result.Data = helpers.ShuffleProducts(operation.GetShuffleSize(), result.Data)
	}

	return result, err
}

func (m *productQuery) ListBuyer(ctx context.Context, req *esApi.BuyerSearchRequest) (*esApi.BuyerSearchResponse, error) {
	var result = &esApi.BuyerSearchResponse{}
	var page, limit int64 = 1, 20
	var query elastic.Query
	pBuilder := builder.New()
	pagination := req.GetPagination()
	if pagination.GetPage() > 1 {
		page = int64(pagination.GetPage())
	}
	if req.Pagination != nil && pagination.GetLimit() >= 0 {
		limit = int64(pagination.GetLimit())
	}
	if pagination.GetLimit() > 500 {
		limit = 500
	}

	esFilter := pBuilder.BuildEsFilterQuery(req.GetFilters())
	esSearch, esRescore := pBuilder.BuildEsSearchQuery(req.GetQuery())

	if esSearch != nil {
		if esFilter == nil {
			esFilter = elastic.NewBoolQuery()
		}
		esFilter.Must(esSearch)
	}

	if esFilter == nil {
		query = elastic.NewMatchAllQuery()
	} else {
		query = esFilter
	}

	sorters := pBuilder.BuildEsSorters(req.GetSorters())
	if len(sorters) > 0 {
		esRescore = nil
	}
	includes := []string{"product_id"}
	if len(req.GetIncludes()) > 0 {
		includes = append(includes, req.GetIncludes()...)
	}
	options := esstore.ElasticOptions{
		Offset:(page -1)* limit,
		Size: limit,
		Includes: includes,
		Sorters: sorters,
		Excludes: req.GetExcludes(),
	}
	data, err := m.esStore.SearchWithTermQuery(ctx, query, esRescore, options)
	if err != nil {
		_ = level.Error(m.logger).Log("msg", fmt.Sprint("error on ListBuyer query ", err.Error()))
		return result, err
	}

	response := pBuilder.BuildListingScoreResponse(data)
	result.Data = response.GetData()
	result.MetaData = response.GetMetaData()
	return result, nil
}

func (m *productQuery) PriceRangeQuery(ctx context.Context, req *esApi.ProductPriceRangeRequest) (*esApi.ProductPriceRangeResponse, error) {
	pBuilder := builder.New()
	esFilter := pBuilder.BuildEsFilterQuery(req.GetFilters())
	esSearch, _ := pBuilder.BuildEsSearchQuery(req.GetQuery())
	agg := pBuilder.BuildESPriceRangeAggregation()

	if esSearch != nil {
		if esFilter == nil {
			esFilter = elastic.NewBoolQuery()
		}
		esFilter.Must(esSearch)
	}

	data, err := m.esStore.AggQuery(ctx, esFilter, agg)

	if err != nil {
		return &esApi.ProductPriceRangeResponse{}, err
	}

	res := pBuilder.BuildPriceRangeAggregationResponse(data)
	return res, nil
}

func (m *productQuery) ListCategoriesFilters(ctx context.Context, req *esApi.ListCategoriesFiltersRequest) (res *esApi.ListCategoriesFiltersResponse, err error) {
	ca := make([]*esApi.CategoryAttribute, 0)
	awo, err := m.esStore.GetAttributeFacets(req.CategoryId, req.Attributes)
	if err != nil {
		return nil, err
	}

	for _, attribute := range awo {
		ca = append(ca, &esApi.CategoryAttribute{Attribute: attribute.Attribute, Options: attribute.Options})
	}
	return &esApi.ListCategoriesFiltersResponse{CategoryId: req.CategoryId, Attributes: ca}, nil
}

func (m *productQuery) ListSearchFilters(ctx context.Context, req *esApi.ListSearchFiltersRequest) (*esApi.ListSearchFiltersResponse, error) {
	var list []*esApi.CategoryAttribute
	result, err := m.esStore.GetSearchAttributeFilter(req.Keyword)
	if err != nil {
		return nil, err
	}

	for _, item := range result {
		list = append(list, &esApi.CategoryAttribute{Attribute: item.Attribute, Options: item.Options})
	}
	return &esApi.ListSearchFiltersResponse{Attributes: list}, nil
}
