package worker_shop_ads

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	kafka_go "gitlab.thovnn.vn/core/sen-kit/pubsub/kafka-go"
	"gitlab.thovnn.vn/core/sen-kit/pubsub/kafka-go/delaycalculator"
	"gitlab.thovnn.vn/core/sen-kit/storage/mongo"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/storage/mgostore"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"gitlab.thovnn.vn/core/sen-kit/storage/elastic"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/storage/esstore"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/usecase/command"
)

var (
	defaultMetricPort = "2112"
)

func RunShopAds(logger log.Logger) error {
	level.Info(logger).Log("msg", "Run Shop Ads starting...")
	//connect mongo
	mgoConfig := mongo.New("es")
	mgoSess, err := mgoConfig.DB()
	if err != nil {
		level.Error(logger).Log("msg", fmt.Sprint("cannot connect to mongo ", mgoConfig.String()))
		return err
	}
	level.Info(logger).Log("msg", fmt.Sprint("connect to mongoDB success, ", mgoConfig.String()))
	defer mgoSess.Close()
	shopStore := mgostore.NewShopStore(mgoSess)

	//connect esStore
	esConfig := elastic.New("")
	esStore := esstore.New(esConfig.Client(), esConfig.ESIndex, logger)
	shopCommand := command.NewShop(esStore, shopStore, logger)

	// Init GRPC Client
	topic := os.Getenv("KAFKA_TOPIC_PRODUCT_ES_SHOP_ADS")
	brokerHosts := os.Getenv("KAFKA_BROKER_HOSTS")
	retryWorkerConf := kafka_go.RetryWorkerConfig{
		Topic:       topic,
		Brokers:     strings.Split(brokerHosts, ","),
		Logger:      logger,
		GroupId:     fmt.Sprintf("%s.%s", topic, "group_abc"),
		MaxRetry:    3,
		ProcessFunc: shopCommand.ShopAds,
		Metrics:     kafka_go.NewPrometheusMetrics(),
		IsDebug:     true,
		//DelayCalculator: delaycalculator.NewExponentialDelayCalculator(5*time.Minute, 3),
		DelayCalculator: delaycalculator.NewLinearDelayCalculator(5 * time.Second),
	}

	level.Info(logger).Log("msg", fmt.Sprint("Subscribe on: ", topic))

	worker, err := kafka_go.NewRetryWorker(context.Background(), retryWorkerConf)
	if err != nil {
		level.Error(logger).Log("msg", fmt.Sprint("connect retry kafka fail: ", err))
	}

	go func() {
		err := httpMetricServer("WORKER_SHOP_ADS", logger)
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

func httpMetricServer(prefix string, logger log.Logger) error {
	metricPort := os.Getenv(fmt.Sprintf("%s_HTTP_METRIC_PORT", prefix))
	if metricPort == "" {
		metricPort = defaultMetricPort
	}
	level.Info(logger).Log("msg", fmt.Sprint("listener http metrics, ", "address", ":"+metricPort))
	http.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(":"+metricPort, nil)
}
