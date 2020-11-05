package builder

import (
	"github.com/olivere/elastic/v7"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/default_listing"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/model"
	esApi "gitlab.thovnn.vn/protobuf/internal-apis-go/product/es_service"
)

type ProductBuilder interface {
	BuildFullEsData(product *model.Product, shop *model.Merchant, installment *model.InstallmentData, shippingSupport *model.ShippingNew)  (*model.ProductES, error)
	BuildFieldES(mongoFields []string) []string
	BuildProductEsScore(product default_listing.Product) (map[string]interface{}, error)
	BuildProductsEsScore(product []default_listing.Product) ([]map[string]interface{}, error)
	BuildESQuery(request *esApi.ListV2Request) map[string]interface{}
	BuildListV2Response(data *elastic.SearchHits) *esApi.ListV2Response
	BuildListingScoreIndexQuery(req *esApi.ListingScoreRequest) *elastic.BoolQuery
	BuildEsFilterQuery(req []*esApi.EsListingFilters) *elastic.BoolQuery
	BuildEsSearchQuery(req *esApi.SearchQuery) (elastic.Query, *elastic.Rescore)
	BuildListingScoreResponse(data *elastic.SearchHits) *esApi.ListingScoreResponse
	BuildEsSorters(sorter []*esApi.Sorter) []elastic.Sorter
	BuildProductCrossCheckLog(productId uint32, err error) *model.ProductCrossCheck
	BuildESPriceRangeAggregation() elastic.Aggregation
	BuildPriceRangeAggregationResponse(aggregations elastic.Aggregations) *esApi.ProductPriceRangeResponse
	BuildProductCrossCheckLogs(productId []uint32, err error) []*model.ProductCrossCheck
	BuildESResponseCrossCheckLog(*elastic.BulkResponse) []*model.ProductCrossCheck
}

type productBuilder struct {
}

func New() ProductBuilder {
	return &productBuilder{}
}

type filter struct {
	Value           []string `json:"value"`
	Key             string   `json:"key"`
	Type            string   `json:"type"`
	RealKey         string   `json:"real_key"`
	NestedFilter      nestedFilter   `json:"nested_query"`
}

type nestedFilter struct {
	Path string `json:"path"`
	Filters []filter
}
