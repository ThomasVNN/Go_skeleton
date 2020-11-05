package worker

import (
	"context"
	"fmt"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/model"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"

	"github.com/go-kit/kit/log/level"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"gitlab.thovnn.vn/dc1/product-discovery/es_service/pkg/client"

	"gitlab.thovnn.vn/core/sen-kit/pubsub/kafka-go/delaycalculator"

	kafka_go "gitlab.thovnn.vn/core/sen-kit/pubsub/kafka-go"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/repo"

	"gitlab.thovnn.vn/core/sen-kit/storage/elastic"

	"gitlab.thovnn.vn/dc1/product-discovery/es_service/usecase/command"

	"gitlab.thovnn.vn/core/sen-kit/storage/mongo"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/storage/esstore"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/storage/mgostore"
)

var (
	defaultMetricPort = "2112"
)

func RunAdd(logger log.Logger) error {
	//connect mongo
	mgoConfig := mongo.New("es")
	mgoSess, err := mgoConfig.DB()
	if err != nil {
		level.Error(logger).Log("msg", "not connect to mongoDB "+mgoConfig.String())
		return err
	}
	level.Info(logger).Log("msg", "connect to mongoDB success "+mgoConfig.String())
	defer mgoSess.Close()
	mgoStore := mgostore.New(mgoSess)
	variantStore := mgostore.NewVariantStore(mgoSess)
	shopStore := mgostore.NewShopStore(mgoSess)
	//connect es
	//esStore
	//connect mongo Log
	logStore := mgostore.NewLogStore(mgoSess)
	//End
	//esStore
	esConfig := elastic.New("")
	esStore := esstore.New(esConfig.Client(), esConfig.ESIndex, logger)

	//get category service endpoint
	categorySrvUrl := os.Getenv("CATEGORY_GRPC_ENDPOINT")
	shopSrvUrl := os.Getenv("SELLER_SHOP_SERVICE_DOMAIN")
	installmentUrl := os.Getenv("GRPC_ENDPOINT_INSTALLMENT_SERVICE")

	// Init GRPC Client
	grpcClient := client.NewGRPCClient(logger)
	defer grpcClient.Close()

	cateRepo := repo.NewCategoryServiceClient(categorySrvUrl, grpcClient)
	installmentRepo := repo.NewInstallmentServiceClient(installmentUrl, grpcClient)
	shopRepo := repo.NewShopServiceClient(shopSrvUrl)

	productCommand := command.ChangeProductNew(esStore, mgoStore, variantStore, logStore, shopStore, cateRepo, shopRepo, installmentRepo, logger, nil)
	if err != nil {
		level.Error(logger).Log("msg", err)
		return err
	}

	topic := os.Getenv("KAFKA_TOPIC_PRODUCT_ADDED")
	brokerHosts := os.Getenv("KAFKA_BROKER_HOSTS")

	retryWorkerConf := kafka_go.RetryWorkerConfig{
		Topic:           topic,
		Brokers:         strings.Split(brokerHosts, ","),
		Logger:          logger,
		GroupId:         fmt.Sprintf("%s.%s", topic, "group"),
		MaxRetry:        5,
		Metrics:         kafka_go.NewPrometheusMetrics(),
		ProcessFunc:     productCommand.OnProductAdded,
		DelayCalculator: delaycalculator.NewExponentialDelayCalculator(time.Second*10, 2),
	}

	worker, err := kafka_go.NewRetryWorker(context.Background(), retryWorkerConf)
	if err != nil {
		level.Error(logger).Log("msg", "connect kafka fail: "+err.Error())
		return err
	}

	go func() {
		err := httpMetricServer("WORKER_ADD", logger)
		level.Error(logger).Log("msg", fmt.Sprint("http metric server err: ", err))
	}()

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
		level.Error(logger).Log("msg", fmt.Sprint("received kill signal: ", <-ch))
		worker.Close()
	}()

	err = worker.Start()
	level.Error(logger).Log("msg", fmt.Sprint("worker is stopped: ", err))
	return err
}

func RunUpdate(logger log.Logger) error {
	//connect mongo
	mgoConfig := mongo.New("es")
	mgoSess, err := mgoConfig.DB()
	if err != nil {
		level.Error(logger).Log("msg", fmt.Sprint("not connect to mongoDB ", mgoConfig.String()))
		return err
	}
	level.Info(logger).Log("msg", fmt.Sprint("connect to mongoDB success", mgoConfig.String()))
	defer mgoSess.Close()
	mgoStore := mgostore.New(mgoSess)
	variantStore := mgostore.NewVariantStore(mgoSess)
	shopStore := mgostore.NewShopStore(mgoSess)
	//connect es
	//esStore
	//connect mongo Log
	logStore := mgostore.NewLogStore(mgoSess)
	//End
	//connect esStore
	esConfig := elastic.New("")
	esStore := esstore.New(esConfig.Client(), esConfig.ESIndex, logger)
	//get category service endpoint
	categorySrvUrl := os.Getenv("CATEGORY_GRPC_ENDPOINT")
	shopSrvUrl := os.Getenv("SELLER_SHOP_SERVICE_DOMAIN")
	installmentUrl := os.Getenv("GRPC_ENDPOINT_INSTALLMENT_SERVICE")
	// Init GRPC Client
	grpcClient := client.NewGRPCClient(logger)
	defer grpcClient.Close()
	cateRepo := repo.NewCategoryServiceClient(categorySrvUrl, grpcClient)
	installmentRepo := repo.NewInstallmentServiceClient(installmentUrl, grpcClient)
	shopRepo := repo.NewShopServiceClient(shopSrvUrl)

	crossCheckChannel := make(chan []*model.ProductCrossCheck, 1000)
	productCommand := command.ChangeProductNew(esStore, mgoStore, variantStore, logStore, shopStore, cateRepo, shopRepo, installmentRepo, logger, crossCheckChannel)

	if err != nil {
		level.Error(logger).Log("msg", err)
		return err
	}
	topic := os.Getenv("KAFKA_TOPIC_PRODUCT_UPDATED")
	brokerHosts := os.Getenv("KAFKA_BROKER_HOSTS")

	retryWorkerConf := kafka_go.RetryWorkerConfig{
		Topic:           topic,
		Brokers:         strings.Split(brokerHosts, ","),
		Logger:          logger,
		GroupId:         fmt.Sprintf("%s.%s", topic, "group"),
		MaxRetry:        5,
		ProcessFunc:     productCommand.OnProductUpdated,
		Metrics:         kafka_go.NewPrometheusMetrics(),
		DelayCalculator: delaycalculator.NewExponentialDelayCalculator(5*time.Minute, 3),
	}

	worker, err := kafka_go.NewRetryWorker(context.Background(), retryWorkerConf)
	if err != nil {
		level.Error(logger).Log("msg", fmt.Sprint("connect retry kafka fail: ", err))
	}

	go func() {
		err := httpMetricServer("WORKER_UPDATE", logger)
		level.Error(logger).Log("msg", fmt.Sprint("http metric server err: ", err))
	}()

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
		level.Error(logger).Log("msg", fmt.Sprint("received kill signal: ", <-ch))
		worker.Close()
	}()

	go func() {
		for m := range crossCheckChannel {
			err := logStore.UpdateBulk(m, false)
			if err != nil {
				level.Error(logger).Log("msg", fmt.Sprint("Failed to write crosscheck log: ", err))
			}
		}
	}()

	err = worker.Start()
	level.Error(logger).Log("msg", fmt.Sprint("worker is stopped: ", err))
	return err
}

func httpMetricServer(prefix string, logger log.Logger) error {
	metricPort := os.Getenv(fmt.Sprintf("%s_HTTP_METRIC_PORT", prefix))
	if metricPort == "" {
		metricPort = defaultMetricPort
	}
	level.Info(logger).Log("msg", fmt.Sprint("listener http metrics, ", "address", ":"+metricPort))
	http.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(":"+metricPort, nil)
}
