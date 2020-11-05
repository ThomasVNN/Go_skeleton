package worker_dc2

import (
	"context"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	kafka_go "gitlab.thovnn.vn/core/sen-kit/pubsub/kafka-go"
	"gitlab.thovnn.vn/core/sen-kit/storage/elastic"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/storage/esstore"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/usecase/command"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var (
	defaultMetricPort = "2112"
)

type delayCalculator struct {
	Interval time.Duration
}

func (d delayCalculator) CalculateDelay(numAttempts int) time.Duration {
	l := []int{1,1,1}
	if numAttempts > len(l) {
		return 0
	}
	return time.Duration(l[numAttempts-1])* time.Minute
}

func RunPromotionChange(logger log.Logger) error {
	//connect esStore
	esConfig := elastic.New("")
	esStore := esstore.New(esConfig.Client(), esConfig.ESIndex, logger)
	productPromotionDc2Command := command.NewProductPromotionDC2Command(esStore,logger)

	topic := os.Getenv("KAFKA_TOPIC_ES_PROMOTION_CHANGED")
	brokerHosts := os.Getenv("KAFKA_BROKER_HOSTS")

	retryWorkerConf := kafka_go.RetryWorkerConfig{
		Topic:           topic,
		Brokers:         strings.Split(brokerHosts, ","),
		Logger:          logger,
		GroupId:         fmt.Sprintf("%s.%s", topic, "group"),
		MaxRetry:        command.RetryTimes,
		Metrics:         kafka_go.NewPrometheusMetrics(),
		ProcessFunc:     productPromotionDc2Command.OnPromotionChange,
		DelayCalculator: delayCalculator{},
	}

	worker, err := kafka_go.NewRetryWorker(context.Background(), retryWorkerConf)
	if err != nil {
		level.Error(logger).Log("msg", "connect kafka fail: "+err.Error())
		return err
	}

	go func() {
		err := httpMetricServer("WORKER_PROMOTION_CHANGE", logger)
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

//func RunPromotionChangeTest(logger log.Logger) error {
//	pubsubURL := os.Getenv("SUB_URL")
//	rabbitMqConfig := rabbitmq.NewPubSubClient(pubsubURL)
//	msgs, err := rabbitMqConfig.Subscribe(context.Background(), &spubsub.SubscribeOption{
//		Event:         os.Getenv("PROMOTION_PUBLISH_EVENT"),
//		Token:         os.Getenv("PROMOTION_SUB_TOKEN"),
//		MaxConcurrent: 1,
//	})
//
//	if err != nil {
//		level.Error(logger).Log("msg", fmt.Sprint("connect rabbitmq fail: ", err))
//		return err
//	}
//
//	level.Info(logger).Log("msg", fmt.Sprintf("Init success, %#v", rabbitMqConfig))
//
//	for m := range msgs {
//		fmt.Println(string(m.Data))
//		go func(m *spubsub.Message) {
//			fmt.Println(string(m.Data))
//			m.Ack("")
//		}(m)
//	}
//
//	return err
//}
