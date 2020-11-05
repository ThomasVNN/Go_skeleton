Project for Elastic Search version 7

## Descriptions
ES_Service connect with Elastic Search version 7.2.
ES Service support GRPC

## Requirement 
- MongoDB 3.6+
- Go 1.11+
- Elastic Search version 7
 
## Deploy
Config .env file example: 
```.env
PORT=8080
GRPC_PORT=8081
ES_MGO_URI="mongodb://"
PUB_SUB_URL=""

KAFKA_BROKER_HOSTS="test.thovnn.vn:3021"
KAFKA_PARTITION=0
KAFKA_TOPIC="es.product.added"
KAFKA_MAX_RETRY=2

ES_URL="test.thovnn.vn:3019"
ES_INDEX="product_v1"

HOTSELL_SUBSCRIBE_EVENT="gorest.hotsell.cron"
HOTSELL_SUBSCRIBE_TOKEN="subscribing token"

PUBSUB_EVENT_ES7_SHOP_STATUS = "es7.shop.status"
PUBSUB_TOKEN_ES7_SHOP_STATUS = "subscribing token"

PUBSUB_EVENT_ES7_SHOP_CERTIFICATE = "es7.shop.certificate"
PUBSUB_TOKEN_ES7_SHOP_CERTIFICATE = "subscribing token"

PUBSUB_EVENT_ES7_SHOP_SHIPPINGFEE = "es7.shop.shippingfee"
PUBSUB_TOKEN_ES7_SHOP_SHIPPINGFEE = "subscribing token"

PUBSUB_EVENT_ES7_SHOP_PROMOTION = "es7.shop.promotion"
PUBSUB_TOKEN_ES7_SHOP_PROMOTION = "subscribing token"

PUBSUB_EVENT_ES7_SHOP_INSTALLMENT = "installment.configuration.update"
PUBSUB_TOKEN_ES7_SHOP_INSTALLMENT = "subscribing token"

LISTING_SCORE_SUBSCRIBE_EVENT="api3.cron.product.score.update"
LISTING_SCORE_SUBSCRIBE_TOKEN="4sdf4e5df"
CATEGORY_GRPC_ENDPOINT="13.251.178.90:8080"
```
## Run
run command Es service

- **service:**         Run service  listener GRPC_PORT
- **worker_add:**      Subscriber topic es.product.added from broker
- **worker_update:**   Subscriber topic es.product.updated from broker
- **worker_hotsell_cron:**   Subscriber topic gorest.hotsell.cron from broker
- **worker_shop:**      Subscriber many topics [es7.shop.status, es7.shop.certificate, es7.shop.shippingfee, installment.configuration.update, es7.shop.promotion] from broker rabbitmq
- **worker_listing_score_cron:** Subscribe to api3.cron.product.score.update topic

- help, h:         Shows a list of commands or help for one command
