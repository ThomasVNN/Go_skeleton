package service

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/kit/log/level"

	"github.com/go-kit/kit/log"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/kelseyhightower/envconfig"
	"gitlab.thovnn.vn/protobuf/internal-apis-go/product/es_service"
	"google.golang.org/grpc"

	"gitlab.thovnn.vn/dc1/product-discovery/es_service/usecase/query"

	"gitlab.thovnn.vn/core/sen-kit/storage/elastic"
	"gitlab.thovnn.vn/core/sen-kit/storage/mongo"

	"gitlab.thovnn.vn/dc1/product-discovery/es_service/storage/esstore"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/storage/mgostore"

	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

type productManagerMiddleware func(manager query.ProductQuery) query.ProductQuery

const (
	SRVName = "es_service"
	Version = "0.0.1"
)

var (
	//logger            = log.NewLogging(SRVName, "info")
	defaultMetricPort = "2112"
)

type Config struct {
	GrpcPort string `split_words:"true"`
}

func loadConfig() *Config {
	var conf Config
	_ = envconfig.Process("", &conf)
	if conf.GrpcPort == "" {
		conf.GrpcPort = "8080"
	}
	return &conf
}

func Run(logger log.Logger) error {
	var appConfig = loadConfig()

	//add metrics for grpc service
	requestCount := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "",
		Subsystem: SRVName,
		Name:      "grpc_request_total",
		Help:      "Number of requests received.",
	}, []string{"method", "success"})
	requestLatency := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "",
		Subsystem: SRVName,
		Name:      "grpc_request_duration_seconds",
		Help:      "Total duration of requests in seconds.",
	}, []string{"method"})
	requestLatencyHistogram := kitprometheus.NewHistogramFrom(stdprometheus.HistogramOpts{
		Namespace: "",
		Subsystem: SRVName,
		Name:      "grpc_request_duration_histogram_seconds",
		Help:      "Total duration of requests histogram in seconds.",
		Buckets:   []float64{0.03, 0.06, 0.1, 0.15, 0.2, 0.5, 1, 2, 3},
	}, []string{"method"})
	serviceRunningGauge := kitprometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
		Namespace: "",
		Subsystem: SRVName,
		Name:      "grpc_server_running",
		Help:      "Number of server that currently exist.",
	}, []string{})
	serviceRunningGauge.Set(1)
	defer serviceRunningGauge.Set(0)

	//mgoStore
	mgoConfig := mongo.New("es")
	mgoSess, err := mgoConfig.DB()
	if err != nil {
		level.Error(logger).Log("msg", fmt.Sprint("not connect to mongoDB ", mgoConfig.String()))
		return err
	}
	level.Info(logger).Log("msg", "connect to mongoDB success "+mgoConfig.String())
	defer mgoSess.Close()

	mgoStore := mgostore.New(mgoSess)

	//esStore
	esConfig := elastic.New("")
	esStore := esstore.New(esConfig.Client(), esConfig.ESIndex, logger)

	svc, err := query.New(esStore, mgoStore, logger)
	if err != nil {
		level.Error(logger).Log("msg", fmt.Sprint("init productQuery fail, ", err))
		return err
	}

	svc = Metrics(requestCount, requestLatency, requestLatencyHistogram)(svc)
	svc = newLoggingMiddleware(logger)(svc)

	grpcListener, err := net.Listen("tcp", ":"+appConfig.GrpcPort)
	if err != nil {
		level.Error(logger).Log("msg", fmt.Sprint("grpc listener error, ", err))
		return err
	}

	endpoints := NewEndpoint(svc)

	grpcServer := NewGrpcServer(endpoints)

	baseServer := grpc.NewServer(grpc.UnaryInterceptor(grpctransport.Interceptor))
	es_service.RegisterESServiceServer(baseServer, grpcServer)
	errChan := make(chan error)

	go func() {

		errChan <- httpMetricServer("SERVICE", logger)
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	go func() {
		level.Info(logger).Log("msg", fmt.Sprint("grpc listener at address, ", ":"+appConfig.GrpcPort))
		errChan <- baseServer.Serve(grpcListener)
	}()

	level.Error(logger).Log("msg", fmt.Sprint("Service is stopped: ", <-errChan))
	return nil
}

func httpMetricServer(prefix string, logger log.Logger) error {
	metricPort := os.Getenv(fmt.Sprintf("%s_HTTP_METRIC_PORT", prefix))
	if metricPort == "" {
		metricPort = defaultMetricPort
	}

	level.Info(logger).Log("msg", "listener http metrics at :"+metricPort, "prefix", "abc")
	http.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(":"+metricPort, nil)
}
