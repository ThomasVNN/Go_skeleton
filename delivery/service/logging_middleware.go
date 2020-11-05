package service

import (
	"context"
	"github.com/go-kit/kit/log/level"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/util"
	"time"
	"github.com/go-kit/kit/log"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/usecase/query"
	"gitlab.thovnn.vn/protobuf/internal-apis-go/product/es_service"
)

// implement function to return ServiceMiddleware
func newLoggingMiddleware(logger log.Logger) productManagerMiddleware {
	return func(next query.ProductQuery) query.ProductQuery {
		return loggingMiddleware{next, logger}
	}
}

type loggingMiddleware struct {
	next   query.ProductQuery
	logger log.Logger
}

func (m loggingMiddleware) ListV2(ctx context.Context, req *es_service.ListV2Request) (res *es_service.ListV2Response, err error) {
	defer func(begin time.Time) {
		_ = level.Info(m.logger).Log(
			"method", "ListV2",
			"service_name", "es_service",
			"req", util.JsonToString(req),
			"resp", util.JsonToString(res),
			"response", time.Since(begin).Seconds(),
		)
	}(time.Now())
	res, err = m.next.ListV2(ctx, req)
	return res, err
}

func (m loggingMiddleware) ListWithShufflingScores(ctx context.Context, req *es_service.ListingScoreRequest) (res *es_service.ListingScoreResponse, err error) {
	defer func(begin time.Time) {
		_ = level.Info(m.logger).Log(
			"method", "ListWithShufflingScores",
			"service_name", "es_service",
			"req", util.JsonToString(req),
			"resp", util.JsonToString(res),
			"response", time.Since(begin).Seconds(),
		)
	}(time.Now())
	res, err = m.next.ListWithShufflingScores(ctx, req)
	return res, err
}

func (m loggingMiddleware) ListBuyer(ctx context.Context, req *es_service.BuyerSearchRequest) (res *es_service.BuyerSearchResponse, err error) {
	defer func(begin time.Time) {
		var resLog interface{}
		if res != nil {
			resLog = map[string]interface{}{
				"total" : res.MetaData,
				"respTotal": len(res.Data),
			}
		}
		_ = level.Info(m.logger).Log(
			"method", "ListBuyer",
			"service_name", "es_service",
			"req", util.JsonToString(req),
			"resp", util.JsonToString(resLog),
			"response", time.Since(begin).Seconds(),
		)
	}(time.Now())
	res, err = m.next.ListBuyer(ctx, req)
	return res, err
}

func (m loggingMiddleware) PriceRangeQuery(ctx context.Context, req *es_service.ProductPriceRangeRequest) (res *es_service.ProductPriceRangeResponse, err error) {
	defer func(begin time.Time) {
		_ = m.logger.Log(
			"method", "PriceRangeQuery",
			"service_name", "es_service",
			"req", util.JsonToString(req),
			"resp", util.JsonToString(res),
			"response", time.Since(begin).Seconds(),
		)
	}(time.Now())
	res, err = m.next.PriceRangeQuery(ctx, req)
	return res, err
}

func (m loggingMiddleware) ListCategoriesFilters(ctx context.Context, req *es_service.ListCategoriesFiltersRequest) (res *es_service.ListCategoriesFiltersResponse, err error) {
	defer func(begin time.Time) {
		_ = level.Info(m.logger).Log(
			"method", "ListCategoriesFilters",
			"service_name", "es_service",
			"req", util.JsonToString(req),
			"resp", util.JsonToString(res),
			"response", time.Since(begin).Seconds(),
		)
	}(time.Now())
	res, err = m.next.ListCategoriesFilters(ctx, req)
	return res, err
}

func (m loggingMiddleware) ListSearchFilters(ctx context.Context, req *es_service.ListSearchFiltersRequest) (res *es_service.ListSearchFiltersResponse, err error) {
	defer func(begin time.Time) {
		_ = level.Info(m.logger).Log(
			"method", "ListSearchFilters",
			"service_name", "es_service",
			"req", util.JsonToString(req),
			"resp", util.JsonToString(res),
			"response", time.Since(begin).Seconds(),
		)
	}(time.Now())
	res, err = m.next.ListSearchFilters(ctx, req)
	return res, err
}
