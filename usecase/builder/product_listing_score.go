package builder

import (
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/default_listing"
)

func (b *productBuilder) BuildProductEsScore(product default_listing.Product) (map[string]interface{}, error) {
	var err error
	return nil, err
}

func (b *productBuilder) BuildProductsEsScore(products []default_listing.Product) ([]map[string]interface{}, error) {
	var err error
	var productMap map[string]interface{}
	var productMapList []map[string]interface{}
	for _, product := range products {
		productMap, err = b.BuildProductEsScore(product)
		if err != nil {
			continue
		}
		productMapList = append(productMapList, productMap)
	}

	return productMapList, err
}
