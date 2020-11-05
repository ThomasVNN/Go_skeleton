package service

import (
	"context"
	"encoding/json"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/usecase/query"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"gitlab.thovnn.vn/protobuf/internal-apis-go/product/es_service"
)

func makeSearchEndpoint(svc query.ProductQuery) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*es_service.ListV2Request)
		result, err := svc.ListV2(ctx, req)
		return result, err
	}
}

//HTTP used only
func NewHTTPHandler(endpoints Set) http.Handler {
	r := mux.NewRouter()
	r.Methods("POST").Path("/v2/product/list").Handler(httptransport.NewServer(
		endpoints.ListV2Endpoint,
		decodeHTTPSumRequest,
		encodeHTTPGenericResponse,
	))

	r.Methods("POST").Path("/v2/product/list_by_score").Handler(httptransport.NewServer(
		endpoints.ListWithShufflingScores,
		decodeHTTPScoreRequest,
		encodeHTTPGenericResponse,
	))

	r.Methods("POST").Path("/v2/product/buyer/search").Handler(httptransport.NewServer(
		endpoints.BuyerSearch,
		decodeHTTPBuyerSearchRequest,
		encodeHTTPGenericResponse,
	))

	return r
}
func decodeHTTPSearchFiltersRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req *es_service.ListSearchFiltersRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

func decodeHTTPSumRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req es_service.ListV2Request
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

func decodeHTTPScoreRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req *es_service.ListingScoreRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

func decodeHTTPBuyerSearchRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req *es_service.BuyerSearchRequest
	err := json.NewDecoder(r.Body).Decode(&req)

	return req, err
}

func encodeHTTPGenericResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	//fmt.Println("=== encodeHTTPGenericResponse, ", response)
	//if f, ok := response.(endpoint.Failer); ok && f.Failed() != nil {
	//	errorEncoder(ctx, f.Failed(), w)
	//	return nil
	//}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func NewGrpcServer(endpoints Set) es_service.ESServiceServer {
	return &grpcServer{
		listV2: grpctransport.NewServer(
			endpoints.ListV2Endpoint,
			decodeGRPCListV2Request,
			encodeGRPCListV2Response,
		),
		listingScore: grpctransport.NewServer(
			endpoints.ListWithShufflingScores,
			decodeGRPCListingScoreRequest,
			encodeGRPCListingScoreResponse,
		),
		buyerSearch: grpctransport.NewServer(
			endpoints.BuyerSearch,
			decodeGRPCBuyerSearchRequest,
			encodeGRPCBuyerSearchResponse,
		),
		listCategoriesFilters: grpctransport.NewServer(
			endpoints.GetCategoryFilters,
			decodeGRPCListCategoriesFiltersRequest,
			decodeGRPCListCategoriesFiltersResponse,
		),

		listSearchFilters: grpctransport.NewServer(
			endpoints.SearchFilters,
			decodeGRPCListSearchFiltersRequest,
			encodeGRPCListSearchFiltersResponse,
		),
		priceRangeQuery: grpctransport.NewServer(
			endpoints.PriceRangeQuery,
			decodeGRPCPriceRangeQueryRequest,
			encodeGRPCPriceRangeQueryResponse,
		),
	}
}

type grpcServer struct {
	listV2                grpctransport.Handler
	listingScore          grpctransport.Handler
	buyerSearch           grpctransport.Handler
	listCategoriesFilters grpctransport.Handler
	listSearchFilters     grpctransport.Handler
	getProductByID    grpctransport.Handler
	priceRangeQuery  grpctransport.Handler

}

func (s *grpcServer) ListSearchFilters(context.Context, *es_service.ListSearchFiltersRequest) (*es_service.ListSearchFiltersResponse, error) {
	return &es_service.ListSearchFiltersResponse{}, nil
}

func (s *grpcServer) GetProductByID(ctx context.Context, req *es_service.GetProductByIDRequest) (res *es_service.GetProductByIDResponse, err error) {
	return &es_service.GetProductByIDResponse{}, nil
}

func (s *grpcServer) GetProductPriceRange(ctx context.Context, req *es_service.ProductPriceRangeRequest) (*es_service.ProductPriceRangeResponse, error) {
	_, rep, err := s.priceRangeQuery.ServeGRPC(ctx, req)

	if err != nil {
		return nil, err
	}

	return rep.(*es_service.ProductPriceRangeResponse), nil
}

func (s *grpcServer) ListV2(ctx context.Context, req *es_service.ListV2Request) (res *es_service.ListV2Response, err error) {
	_, rep, err := s.listV2.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*es_service.ListV2Response), nil
}

func (s *grpcServer) ListWithShufflingScores(ctx context.Context, req *es_service.ListingScoreRequest) (res *es_service.ListingScoreResponse, err error) {
	_, rep, err := s.listingScore.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*es_service.ListingScoreResponse), nil
}

func (s *grpcServer) ListBuyer(ctx context.Context, req *es_service.BuyerSearchRequest) (res *es_service.BuyerSearchResponse, err error) {
	_, rep, err := s.buyerSearch.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*es_service.BuyerSearchResponse), nil
}

func (s *grpcServer) ListCategoriesFilters(ctx context.Context, req *es_service.ListCategoriesFiltersRequest) (*es_service.ListCategoriesFiltersResponse, error) {
	_, rep, err := s.listCategoriesFilters.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*es_service.ListCategoriesFiltersResponse), nil
}

type Set struct {
	ListV2Endpoint          endpoint.Endpoint
	ListWithShufflingScores endpoint.Endpoint
	BuyerSearch             endpoint.Endpoint
	PriceRangeQuery endpoint.Endpoint
	GetCategoryFilters      endpoint.Endpoint
	SearchFilters           endpoint.Endpoint
	GetProductByID          endpoint.Endpoint
}

func NewEndpoint(service query.ProductQuery) Set {
	var listV2Endpoint endpoint.Endpoint
	var listWithShufflingScores endpoint.Endpoint
	var buyerSearch endpoint.Endpoint
	var getCategoryFilters endpoint.Endpoint
	var getSearchFilters endpoint.Endpoint
	var priceRangeQuery endpoint.Endpoint
	listV2Endpoint = makeSearchEndpoint(service)
	listWithShufflingScores = makeListingScoreIndexEndpoint(service)
	buyerSearch = makeBuyerSearchEndpoint(service)
	priceRangeQuery = makePriceRangeQueryEndpoint(service)
	//listV2Endpoint = LoggingMiddleware(logkit.With(logger, "method", "ListV2"))(listV2Endpoint)
	getCategoryFilters = newListCategoriesFiltersEndpoint(service)
	getSearchFilters = newListSearchFiltersEndpoint(service)

	return Set{
		ListV2Endpoint:          listV2Endpoint,
		ListWithShufflingScores: listWithShufflingScores,
		BuyerSearch:             buyerSearch,
		GetCategoryFilters:      getCategoryFilters,
		SearchFilters:           getSearchFilters,
		PriceRangeQuery: priceRangeQuery,
	}
}

func decodeGRPCListV2Request(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*es_service.ListV2Request)
	return req, nil
}

func encodeGRPCListV2Response(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(*es_service.ListV2Response)
	return resp, nil
}

func decodeGRPCListingScoreRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*es_service.ListingScoreRequest)
	return req, nil
}

func encodeGRPCListingScoreResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(*es_service.ListingScoreResponse)
	return resp, nil
}

func decodeGRPCBuyerSearchRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*es_service.BuyerSearchRequest)
	return req, nil
}

func encodeGRPCBuyerSearchResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(*es_service.BuyerSearchResponse)
	return resp, nil
}

func decodeGRPCPriceRangeQueryRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*es_service.ProductPriceRangeRequest)
	return req, nil
}

func encodeGRPCPriceRangeQueryResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(*es_service.ProductPriceRangeResponse)
	return resp, nil
}

func decodeGRPCListCategoriesFiltersRequest(ctx context.Context, req interface{}) (interface{}, error) {
	return req.(*es_service.ListCategoriesFiltersRequest), nil
}

func decodeGRPCListCategoriesFiltersResponse(ctx context.Context, resp interface{}) (interface{}, error) {
	return resp.(*es_service.ListCategoriesFiltersResponse), nil
}

func decodeGRPCListSearchFiltersRequest(ctx context.Context, req interface{}) (interface{}, error) {
	return req.(*es_service.ListSearchFiltersRequest), nil
}

func encodeGRPCListSearchFiltersResponse(ctx context.Context, resp interface{}) (interface{}, error) {
	return resp.(*es_service.ListSearchFiltersResponse), nil
}
