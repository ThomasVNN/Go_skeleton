package service

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kit/kit/metrics"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/usecase/query"
	esApi "gitlab.thovnn.vn/protobuf/internal-apis-go/product/es_service"
)

type ServiceMiddleware func(query.ProductQuery) query.ProductQuery

func Metrics(requestCount metrics.Counter,
	requestLatency metrics.Histogram,
	requestLatencyHistogram metrics.Histogram) ServiceMiddleware {
	return func(next query.ProductQuery) query.ProductQuery {
		return &metricsMiddleware{
			next,
			requestCount,
			requestLatency,
			requestLatencyHistogram,
		}
	}
}

type metricsMiddleware struct {
	query.ProductQuery
	requestCount            metrics.Counter
	requestLatency          metrics.Histogram
	requestLatencyHistogram metrics.Histogram
}

func (mw *metricsMiddleware) ListV2(ctx context.Context, req *esApi.ListV2Request) (res *esApi.ListV2Response, err error) {
	defer func(begin time.Time) {
		mw.requestCount.With(
			"method", "ListV2",
			"success", fmt.Sprintf("%t", err == nil)).Add(1)
		mw.requestLatency.With("method", "ListV2").Observe(time.Since(begin).Seconds())
		mw.requestLatencyHistogram.With("method", "ListV2").Observe(time.Since(begin).Seconds())
	}(time.Now())
	res, err = mw.ProductQuery.ListV2(ctx, req)
	return res, err
}

func (mw *metricsMiddleware) ListWithShufflingScores(ctx context.Context, req *esApi.ListingScoreRequest) (res *esApi.ListingScoreResponse, err error) {
	defer func(begin time.Time) {
		mw.requestCount.With(
			"method", "ListWithShufflingScores",
			"success", fmt.Sprintf("%t", err == nil)).Add(1)
		mw.requestLatency.With("method", "ListWithShufflingScores").Observe(time.Since(begin).Seconds())
		mw.requestLatencyHistogram.With("method", "ListWithShufflingScores").Observe(time.Since(begin).Seconds())
	}(time.Now())
	res, err = mw.ProductQuery.ListWithShufflingScores(ctx, req)
	return res, err
}

func (mw *metricsMiddleware) ListBuyer(ctx context.Context, req *esApi.BuyerSearchRequest) (res *esApi.BuyerSearchResponse, err error) {
	defer func(begin time.Time) {
		mw.requestCount.With("method", "ListBuyer",
			"success", fmt.Sprintf("%t", err == nil)).Add(1)
		mw.requestLatency.With("method", "ListBuyer").Observe(time.Since(begin).Seconds())
		mw.requestLatencyHistogram.With("method", "ListBuyer").Observe(time.Since(begin).Seconds())

	}(time.Now())
	res, err = mw.ProductQuery.ListBuyer(ctx, req)
	return res, err
}
