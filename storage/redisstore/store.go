package redisstore

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/model"
	"strings"
)

const (
	FW2_REDIS_KEY_PRODUCT_DETAIL = "fw2.product.detail.info."
	HASH_PRODUCT_INFO            = "h:product.info:"
	SORTEDSET_MERCHANT_PRODUCT_LIST = "ss:merchant.product.list:"
)

type RedisStore interface {
	GetConnection() *redis.Client
	SetProductsToRedis(products []*model.Product) []*model.Product
}

type redisStore struct {
	client *redis.Client
}

func New(r *redis.Client) RedisStore {
	return &redisStore{client: r}
}

func (r redisStore) GetConnection() *redis.Client {
	return r.client
}

func (r *redisStore) ParseProductToCache(product *model.Product) ([]byte, error) {
	sliceProduct := []map[string]interface{}{}
	mapProduct := make(map[string]interface{})
	productByte, err := json.Marshal(product)
	if err != nil {
		return []byte{}, err
	}
	err = json.Unmarshal(productByte, &mapProduct)
	if err != nil {
		return []byte{}, err
	}
	if _, ok := mapProduct["Attribute"]; ok {
		delete(mapProduct, "Attribute")
	}
	if product.Attribute != nil && product.RequiredOptions != "" {
		attributes := strings.Split(product.RequiredOptions, ",")
		for _, attr := range attributes {
			if val, ok := product.Attribute[attr]; ok {
				mapProduct[attr] = val
			}
		}
	}
	sliceProduct = append(sliceProduct, mapProduct)
	return json.Marshal(sliceProduct)
}

func (r *redisStore) RedisAddProductInfo(product *model.Product) error {
	rdis := r.GetConnection()
	key := SORTEDSET_MERCHANT_PRODUCT_LIST + fmt.Sprint(product.SellerAdminId)
	keyProductInfo := HASH_PRODUCT_INFO + fmt.Sprint(product.ProductId)
	if product.StatusNew == 5 {
		_, err := rdis.Do("ZREM", key, product.ProductId).Result()
		if err != nil {

		}
		//1.2. Remove key Product detail.
		_, err = rdis.Do("DEL", keyProductInfo).Result()
		return err
	}
	err := r.HMSET(key, product)
	if err != nil {
		_, err := rdis.Do("DEL", keyProductInfo).Result()
		return err
	}
	return err
}

func (r *redisStore) SetProductsToRedis(products []*model.Product) []*model.Product {
	var failed []*model.Product
	for _, p := range products {
		err := r.RedisAddProductInfo(p)
		if err != nil {
			fmt.Println(err)
			failed = append(failed, p)
		}
	}
	return failed
}
