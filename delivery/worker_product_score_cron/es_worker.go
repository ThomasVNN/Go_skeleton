package worker_product_score_cron

import (
	"context"
	"fmt"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"gitlab.thovnn.vn/core/golang-sdk/spubsub"
	"gitlab.thovnn.vn/core/sen-kit/storage/elastic"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/pkg/pubsub/rabbitmq"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/storage/esstore"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/usecase/command"
)

func RunHotSellSubscriber(logger log.Logger) error {
	//connect es
	//esStore
	esConfig := elastic.New("")

	esStore := esstore.New(esConfig.Client(), esConfig.ESIndex, logger)

	productRankingScore := command.NewProductRankingScoreCommand(esStore)

	pubsubURL := os.Getenv("PUB_SUB_URL")

	rabbitMqConfig := rabbitmq.NewPubSubClient(pubsubURL)

	msgs, err := rabbitMqConfig.Subscribe(context.Background(), &spubsub.SubscribeOption{
		Event:         os.Getenv("HOTSELL_SUBSCRIBE_EVENT"),
		Token:         os.Getenv("HOTSELL_SUBSCRIBE_TOKEN"),
		MaxConcurrent: 1,
	})

	if err != nil {
		_ = level.Error(logger).Log("msg", fmt.Sprint("connect rabbitmq fail: ", err))
		return err
	}

	_ = level.Info(logger).Log("msg", fmt.Sprintf("Init success, %#v", rabbitMqConfig))

	for m := range msgs {
		go func(m *spubsub.Message) {
			err := productRankingScore.OnProductHotSellUpdate(m ,logger)
			if err != nil {
				if m.DeliveredCount < 5 {
					return
				}
			}
			m.Ack("")
		}(m)
	}
	return err
}

func runListingScoreSubscriber(event string, token string, logger log.Logger) error {
	esConfig := elastic.New("")

	esStore := esstore.New(esConfig.Client(), esConfig.ESIndex, logger)

	productRankingScore := command.NewProductRankingScoreCommand(esStore)

	pubsubURL := os.Getenv("PUB_SUB_URL")

	_ = level.Info(logger).Log("msg", fmt.Sprint(pubsubURL, event, token))
	rabbitMqConfig := rabbitmq.NewPubSubClient(pubsubURL)
	msgs, err := rabbitMqConfig.Subscribe(context.Background(), &spubsub.SubscribeOption{
		Event:         event,
		Token:         token,
		MaxConcurrent: 1,
	})

	if err != nil {
		_ = level.Error(logger).Log("msg", fmt.Sprint("connect rabbitmq fail: ", err))
		return err
	}

	_ = level.Info(logger).Log("msg", fmt.Sprint("Init success, ", fmt.Sprintf("%#v", rabbitMqConfig)))
	guard := make(chan interface{}, 10)
	for m := range msgs {
		guard <- nil
		go func(m *spubsub.Message) {
			err := productRankingScore.OnProductScoreUpdate(m, logger)
			<- guard
			if err != nil {
				if m.DeliveredCount < 5 {
					return
				}
			}
			m.Ack("")
		}(m)
	}
	return err
}

func RunListingScoreSubscriber(logger log.Logger) error {
	return runListingScoreSubscriber(
		os.Getenv("LISTING_SCORE_SUBSCRIBE_EVENT"),
		os.Getenv("LISTING_SCORE_SUBSCRIBE_TOKEN"),
		logger)
}

func RunRTListingScoreSubscriber(logger log.Logger) error {
	event := os.Getenv("REALTIME_SCORE_SUBSCRIBE_EVENT")
	if event == "" {
		event = "api3.realtime.product.update"
	}
	token := os.Getenv("REALTIME_SCORE_SUBSCRIBE_TOKEN")
	if token == "" {
		token = "4yVQUAO2"
	}
	return runListingScoreSubscriber(event, token, logger)
}
