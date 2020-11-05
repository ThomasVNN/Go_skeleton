package service

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/usecase/query"
	"gitlab.thovnn.vn/protobuf/internal-apis-go/product/es_service"
)

func makeListingScoreIndexEndpoint(svc query.ProductQuery) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*es_service.ListingScoreRequest)
		result, err := svc.ListWithShufflingScores(ctx, req)
		return result, err
	}
}

func makeBuyerSearchEndpoint(svc query.ProductQuery) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*es_service.BuyerSearchRequest)
		result, err := svc.ListBuyer(ctx, req)
		return result, err
	}
}

func newListCategoriesFiltersEndpoint(svc query.ProductQuery) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*es_service.ListCategoriesFiltersRequest)
		resp, err := svc.ListCategoriesFilters(ctx, req)
		return resp, err
	}
}

func newListSearchFiltersEndpoint(svc query.ProductQuery) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*es_service.ListSearchFiltersRequest)
		resp, err := svc.ListSearchFilters(ctx, req)
		return resp, err
	}
}

func makePriceRangeQueryEndpoint(svc query.ProductQuery) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*es_service.ProductPriceRangeRequest)
		result, err := svc.PriceRangeQuery(ctx, req)
		return result, err
	}
}
