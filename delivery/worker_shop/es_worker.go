package worker_shop

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/go-kit/kit/log/level"
	kafka_go "gitlab.thovnn.vn/core/sen-kit/pubsub/kafka-go"
	"gitlab.thovnn.vn/core/sen-kit/pubsub/kafka-go/delaycalculator"

	"github.com/go-kit/kit/log"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gitlab.thovnn.vn/core/golang-sdk/spubsub"
	"gitlab.thovnn.vn/core/sen-kit/pubsub"
	"gitlab.thovnn.vn/core/sen-kit/storage/elastic"
	"gitlab.thovnn.vn/core/sen-kit/storage/mongo"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/pkg/pubsub/rabbitmq"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/storage/esstore"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/storage/mgostore"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/usecase/command"
)

var (
	logger            log.Logger
	err               error
	sub               pubsub.Subscriber
	defaultMetricPort = "2112"
)

func RunAll(appLogger log.Logger) error {
	logger = log.With(appLogger)
	g, _ := errgroup.WithContext(context.Background())

	//g.Go(RunShopStatus)
	g.Go(RunShopCertificate)
	g.Go(RunShopShippingFee)
	g.Go(RunShopInstallment)
	g.Go(RunShopSupportPromotionApp)
	g.Go(RunShopRatingInfo)

	if err := g.Wait(); err != nil {
		level.Error(logger).Log("msg", fmt.Sprint("worker is stopped!!!, err: ", err))
		return err
	}
	return nil
}

func RunShopInstant(appLogger log.Logger) error {
	logger = appLogger
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

	// connect to es
	esConfig := elastic.New("")
	esStore := esstore.New(esConfig.Client(), esConfig.ESIndex, logger)
	shopCommand := command.NewShop(esStore, shopStore, logger)

	topic := os.Getenv("KAFKA_TOPIC_SHOP_INSTANT")
	brokerHosts := os.Getenv("KAFKA_BROKER_HOSTS")
	retryWorkerConf := kafka_go.RetryWorkerConfig{
		Topic:           topic,
		Brokers:         strings.Split(brokerHosts, ","),
		Logger:          logger,
		GroupId:         fmt.Sprintf("%s.%s", topic, "group"),
		MaxRetry:        3,
		ProcessFunc:     shopCommand.ShopInstant,
		Metrics:         kafka_go.NewPrometheusMetrics(),
		DelayCalculator: delaycalculator.NewLinearDelayCalculator(time.Second * 30),
	}

	worker, err := kafka_go.NewRetryWorker(context.Background(), retryWorkerConf)
	if err != nil {
		level.Error(logger).Log("msg", fmt.Sprint("connect kafka fail: ", err))
		return err
	}

	go func() {
		err := httpMetricServer("WORKER_SHOP_INSTANT", logger)
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

//func RunShopStatus() error {
//	//connect mongo
//	mgoConfig := mongo.New("es")
//	mgoSess, err := mgoConfig.DB()
//	if err != nil {
//		level.Error(logger).Log("msg", fmt.Sprint("not connect to mongoDB ", mgoConfig.String()))
//		return err
//	}
//	level.Info(logger).Log("msg", fmt.Sprint("connect to mongoDB success ", mgoConfig.String()))
//	defer mgoSess.Close()
//	shopStore := mgostore.NewShopStore(mgoSess)
//
//	//connect es
//	//esStore
//	esConfig := elastic.New("")
//	esStore := esstore.New(esConfig.Client(), esConfig.ESIndex, logger)
//
//	shopCommand := command.NewShop(esStore, shopStore, logger)
//	if err != nil {
//		level.Error(logger).Log("msg", err)
//		return err
//	}
//
//	rabbitMqConfig := rabbitmq.NewPubSubClient(os.Getenv("PUB_SUB_URL"))
//	listMessage, err := rabbitMqConfig.Subscribe(context.Background(), &spubsub.SubscribeOption{
//		Event:         os.Getenv("PUBSUB_EVENT_ES7_SHOP_STATUS"),
//		Token:         os.Getenv("PUBSUB_TOKEN_ES7_SHOP_STATUS"),
//		MaxConcurrent: 1,
//	})
//
//	if err != nil {
//		level.Error(logger).Log("msg", fmt.Sprint("connect rabbitmq fail: ", err))
//		return err
//	}
//	level.Info(logger).Log("msg", fmt.Sprint("Init success, ", rabbitMqConfig))
//
//	return processPubSubMessages(listMessage, shopCommand.OnOffShop)
//}

func RunShopCertificate() error {
	//connect mongo
	mgoConfig := mongo.New("es")
	mgoSess, err := mgoConfig.DB()
	if err != nil {
		level.Error(logger).Log("msg", fmt.Sprint("not connect to mongoDB ", mgoConfig.String()))
		return err
	}
	level.Info(logger).Log("msg", fmt.Sprint("connect to mongoDB success ", mgoConfig.String()))
	defer mgoSess.Close()
	shopStore := mgostore.NewShopStore(mgoSess)
	//connect es
	//esStore
	esConfig := elastic.New("")
	esStore := esstore.New(esConfig.Client(), esConfig.ESIndex, logger)
	shopCommand := command.NewShop(esStore, shopStore, logger)
	if err != nil {
		level.Error(logger).Log("msg", err)
		return err
	}

	rabbitMqConfig := rabbitmq.NewPubSubClient(os.Getenv("PUB_SUB_URL"))
	listMessage, err := rabbitMqConfig.Subscribe(context.Background(), &spubsub.SubscribeOption{
		Event:         os.Getenv("PUBSUB_EVENT_ES7_SHOP_CERTIFICATE"),
		Token:         os.Getenv("PUBSUB_TOKEN_ES7_SHOP_CERTIFICATE"),
		MaxConcurrent: 10,
	})

	if err != nil {
		level.Error(logger).Log("msg", fmt.Sprint("connect rabbit-mq fail: ", err))
		return err
	}
	level.Info(logger).Log("msg", fmt.Sprint("Init success, ", rabbitMqConfig))

	return processPubSubMessages(listMessage, shopCommand.OnOffShopCertificate)
}

func RunShopShippingFee() error {
	//connect mongo
	mgoConfig := mongo.New("es")
	mgoSess, err := mgoConfig.DB()
	if err != nil {
		level.Error(logger).Log("msg", fmt.Sprint("cannot connect to mongo ", mgoConfig.String()))
		return err
	}
	level.Info(logger).Log("msg", "connect to mongoDB success")
	defer mgoSess.Close()
	shopStore := mgostore.NewShopStore(mgoSess)

	// connect to es
	esConfig := elastic.New("")
	esStore := esstore.New(esConfig.Client(), esConfig.ESIndex, logger)
	shopCommand := command.NewShop(esStore, shopStore, logger)

	// connect to rabbit-mq
	newRabbitMQ := rabbitmq.NewPubSubClient(os.Getenv("PUB_SUB_URL"))
	listMessage, err := newRabbitMQ.Subscribe(context.Background(), &spubsub.SubscribeOption{
		Event:         os.Getenv("PUBSUB_EVENT_ES7_SHOP_SHIPPINGFEE"),
		Token:         os.Getenv("PUBSUB_TOKEN_ES7_SHOP_SHIPPINGFEE"),
		MaxConcurrent: 10,
	})
	if err != nil {
		level.Error(logger).Log("msg", fmt.Sprint("connect rabbit-mq fail: ", err))
		return err
	}
	level.Info(logger).Log("msg", fmt.Sprint("init success, ", newRabbitMQ))

	return processPubSubMessages(listMessage, shopCommand.ShopSupportShippingFee)
}

func RunShopInstallment() error {
	//connect mongo
	mgoConfig := mongo.New("es")
	mgoSess, err := mgoConfig.DB()
	if err != nil {
		level.Error(logger).Log("msg", fmt.Sprint("cannot connect to mongo ", mgoConfig.String()))
		return err
	}
	level.Info(logger).Log("msg", fmt.Sprint("connect to mongoDB success ", mgoConfig.String()))
	defer mgoSess.Close()
	shopStore := mgostore.NewShopStore(mgoSess)

	// connect to es
	esConfig := elastic.New("")
	esStore := esstore.New(esConfig.Client(), esConfig.ESIndex, logger)
	shopCommand := command.NewShop(esStore, shopStore, logger)

	// connect to rabbit-mq
	newRabbitMQ := rabbitmq.NewPubSubClient(os.Getenv("PUB_SUB_URL"))
	listMessage, err := newRabbitMQ.Subscribe(context.Background(), &spubsub.SubscribeOption{
		Event:         os.Getenv("PUBSUB_EVENT_ES7_SHOP_INSTALLMENT"),
		Token:         os.Getenv("PUBSUB_TOKEN_ES7_SHOP_INSTALLMENT"),
		MaxConcurrent: 10,
	})
	if err != nil {
		level.Error(logger).Log("msg", err)
		return err
	}
	level.Info(logger).Log("msg", fmt.Sprint("init success, ", newRabbitMQ))

	return processPubSubMessages(listMessage, shopCommand.ShopInstallment)
}

func RunShopSupportPromotionApp() error {
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

	// connect to es
	esConfig := elastic.New("")
	esStore := esstore.New(esConfig.Client(), esConfig.ESIndex, logger)
	shopCommand := command.NewShop(esStore, shopStore, logger)

	// connect to rabbit-mq
	newRabbitMQ := rabbitmq.NewPubSubClient(os.Getenv("PUB_SUB_URL"))
	listMessage, err := newRabbitMQ.Subscribe(context.Background(), &spubsub.SubscribeOption{
		Event:         os.Getenv("PUBSUB_EVENT_ES7_SHOP_PROMOTION"),
		Token:         os.Getenv("PUBSUB_TOKEN_ES7_SHOP_PROMOTION"),
		MaxConcurrent: 10,
	})
	if err != nil {
		level.Error(logger).Log("msg", fmt.Sprint("connect rabbit-mq fail: ", err))
		return err
	}
	level.Info(logger).Log("msg", fmt.Sprint("init success, ", newRabbitMQ))

	return processPubSubMessages(listMessage, shopCommand.ShopPromotionApp)
}

func RunShopRatingInfo() error {
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

	// connect to es
	esConfig := elastic.New("")
	esStore := esstore.New(esConfig.Client(), esConfig.ESIndex, logger)
	shopCommand := command.NewShop(esStore, shopStore, logger)
	// connect to rabbit-mq
	newRabbitMQ := rabbitmq.NewPubSubClient(os.Getenv("PUB_SUB_URL"))
	listMessage, err := newRabbitMQ.Subscribe(context.Background(), &spubsub.SubscribeOption{
		Event:         os.Getenv("PUBSUB_EVENT_ES7_SHOP_RATING_INFO"),
		Token:         os.Getenv("PUBSUB_TOKEN_ES7_SHOP_RATING_INFO"),
		MaxConcurrent: 10,
	})
	if err != nil {
		level.Error(logger).Log("msg", fmt.Sprint("connect rabbit-mq fail: ", err))
		return err
	}

	return processPubSubMessages(listMessage, shopCommand.ShopRatingInfo)
}

func processPubSubMessages(listMessage <-chan *spubsub.Message, process func(*spubsub.Message) error) error {
	var err error
	for message := range listMessage {
		err = process(message)
		if err != nil {
			_ = level.Error(logger).Log("msg", fmt.Sprintf("failed to update data to es, retry count at %d with err %v", message.DeliveredCount, err.Error()))
			if message.DeliveredCount >= 5 {
				message.Ack("")
				continue
			}
			message.Redeliver("", time.Second*10*time.Duration(message.DeliveredCount))
			continue
		}
		message.Ack("")
	}
	return err
}

func httpMetricServer(prefix string, logger log.Logger) error {
	metricPort := os.Getenv(fmt.Sprintf("%s_HTTP_METRIC_PORT", prefix))
	if metricPort == "" {
		metricPort = defaultMetricPort
	}
	level.Info(logger).Log("msg", "listener http metrics at :"+metricPort)
	http.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(":"+metricPort, nil)
}
