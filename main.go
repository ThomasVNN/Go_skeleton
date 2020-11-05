package main

import (
	"fmt"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/delivery/worker_shop_ads"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/delivery/worker_total_cron"
	"os"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/joho/godotenv"
	"github.com/urfave/cli"
	"gitlab.thovnn.vn/core/sen-kit/senlog"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/delivery/script"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/delivery/service"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/delivery/worker"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/delivery/worker_dc2"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/delivery/worker_product_score_cron"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/delivery/worker_shop"
)

const (
	ServiceName = "es_service"
)

func main() {
	//load .env file to env
	_ = godotenv.Load()
	//init log
	logLevel := os.Getenv("LOGLEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	logConfig := senlog.Config{
		Level: logLevel,
	}
	logger, closeLogFunc := senlog.New(&logConfig)
	logger = senlog.WithService(logger, ServiceName)
	defer closeLogFunc()

	_ = level.Info(logger).Log("msg", "Loglevel: %s", logLevel)

	app := cli.NewApp()
	app.Version = service.Version
	app.Usage = "elastic search version 7 service"

	app.Commands = []cli.Command{
		{
			Name:    "service",
			Aliases: []string{""},
			Usage:   "elastic service",
			Action: func(c *cli.Context) error {
				service.Run(log.With(logger, "command", "service"))
				return nil
			},
		},
		{
			Name:    "worker_add",
			Aliases: []string{""},
			Usage:   "Subscriber from broker",
			Action: func(c *cli.Context) error {
				return worker.RunAdd(log.With(logger, "command", "worker_add"))
			},
		},
		{
			Name:    "worker_update",
			Aliases: []string{""},
			Usage:   "Subscriber from broker",
			Action: func(c *cli.Context) error {
				return worker.RunUpdate(log.With(logger, "command", "worker_update"))
			},
		},
		{
			Name:    "worker_shop",
			Aliases: []string{""},
			Usage:   `Subscriber many topics from rabbitmq`,
			Action: func(c *cli.Context) error {
				return worker_shop.RunAll(log.With(logger, "command", "worker_shop"))
			},
		},
		{
			Name:    "worker_hotsell_cron",
			Aliases: []string{""},
			Usage:   "Subscriber from broker",
			Action: func(c *cli.Context) error {
				return worker_product_score_cron.RunHotSellSubscriber(log.With(logger, "command", "worker_hotsell_cron"))
			},
		},
		{
			Name:    "worker_listing_score_cron",
			Aliases: []string{""},
			Usage:   "Subscriber from broker",
			Action: func(c *cli.Context) error {
				return worker_product_score_cron.RunListingScoreSubscriber(log.With(logger, "command", "worker_listing_score_cron"))
			},
		},
		{
			Name:    "worker_listing_score_realtime",
			Aliases: []string{""},
			Usage:   "Subscriber from broker",
			Action: func(c *cli.Context) error {
				return worker_product_score_cron.RunRTListingScoreSubscriber(log.With(logger, "command", "worker_listing_score_realtime"))
			},
		},
		{
			Name:    "worker_update_total_score",
			Aliases: []string{""},
			Usage:   "Subscriber from kafka product listing",
			Action: func(c *cli.Context) error {
				return worker_total_cron.RunTotalScoreSubscriber(log.With(logger, "command", "worker_update_total_score"))
			},
		},
		{
			Name:    "worker_shop_instant",
			Aliases: []string{""},
			Usage:   "Subscriber from broker",
			Action: func(c *cli.Context) error {
				return worker_shop.RunShopInstant(log.With(logger, "command", "worker_shop_instant"))
			},
		},
		{
			Name:    "migrate_number_facets_product",
			Aliases: []string{""},
			Usage:   "Migrate data attributes & warehouse",
			Action: func(c *cli.Context) error {
				return script.RunMigrateNumberFacet(log.With(logger, "command", "migrate_number_facets_product"))
			},
		},
		/*{
			Name:    "worker_shop_warehouse",
			Aliases: []string{""},
			Usage:   "Subscriber from broker",
			Action: func(c *cli.Context) error {
				return worker_shop_warehouse.RunShopWareHouse(log.With(logger, "command", "worker_shop_warehouse"))
			},
		},
		{
			Name:    "migrate_number_facets_product",
			Aliases: []string{""},
			Usage:   "Migrate data attributes & warehouse",
			Action: func(c *cli.Context) error {
				return script.RunMigrateNumberFacet(log.With(logger, "command", "migrate_number_facets_product"))
			},
		},*/
		{
			Name:    "worker_promotion_change",
			Aliases: []string{""},
			Usage:   "Subscriber from broker",
			Action: func(c *cli.Context) error {
				return worker_dc2.RunPromotionChange(log.With(logger, "command", "worker_promotion_change"))
			},
		},
		{
			Name:    "worker_shop_ads",
			Aliases: []string{""},
			Usage:   "Subscriber from broker",
			Action: func(c *cli.Context) error {
				return worker_shop_ads.RunShopAds(log.With(logger, "command", "worker_shop_ads"))
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		level.Error(logger).Log("msg", fmt.Sprint("es-service is stopped, ", err))
	}

}
