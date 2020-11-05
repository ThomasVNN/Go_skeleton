package worker_total_cron

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

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"gitlab.thovnn.vn/core/sen-kit/storage/elastic"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/storage/esstore"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/usecase/command"
)

var (
	defaultMetricPort = "2112"
)

func RunTotalScoreSubscriber(logger log.Logger) error {
	level.Info(logger).Log("msg", "RunTotalScoreSubscriber starting...")
	//connect esStore
	esConfig := elastic.New("")
	esStore := esstore.New(esConfig.Client(), esConfig.ESIndex, logger)
	productCommand := command.NewProductTotalScoreCommand(esStore, logger)
	// Init GRPC Client
	topic := os.Getenv("KAFKA_TOPIC_PRODUCT_SCORE_TOTAL")
	brokerHosts := os.Getenv("KAFKA_BROKER_HOSTS")
	retryWorkerConf := kafka_go.RetryWorkerConfig{
		Topic:           topic,
		Brokers:         strings.Split(brokerHosts, ","),
		Logger:          logger,
		GroupId:         fmt.Sprintf("%s.%s", topic, "group"),
		MaxRetry:        5,
		ProcessFunc:     productCommand.OnProductUpdate,
		Metrics:         kafka_go.NewPrometheusMetrics(),
		//DelayCalculator: delaycalculator.NewExponentialDelayCalculator(5*time.Minute, 3),
		DelayCalculator: delaycalculator.NewLinearDelayCalculator(5*time.Second),
	}

	level.Info(logger).Log("msg", fmt.Sprint("Subscribe on: ", topic))

	worker, err := kafka_go.NewRetryWorker(context.Background(), retryWorkerConf)
	if err != nil {
		level.Error(logger).Log("msg", fmt.Sprint("connect retry kafka fail: ", err))
	}

	go func() {
		err := httpMetricServer("WORKER_SCORE_UPDATE", logger)
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
