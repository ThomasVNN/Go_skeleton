package builder

import (
	"encoding/json"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/model"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/pkg/helper"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/util"
	"strings"
)

var mapESFields = map[string]string{
	"attribute":          "number_facets",
	"category_id":        "number_facets",
	"status_new":         "status_id",
	"stock_status":       "is_stock",
	"shop_free_shipping": "is_shop_support_shipping_fee",
	"is_config_variant":  "has_variant",
	"sku_user":           "sku_defined_by_shop",
	"review_date":        "review_at",
	"is_certified":       "is_certificate_shop",
}

// BuildUpdateEsData
/**
BuildUpdateEsData using for subscribe events es7.product.update, es7.product.add
@params: product *model.Product
1. Build product struct es from mongo
*/
func (pu *productBuilder) BuildFullEsData(product *model.Product, merchant *model.Merchant, installment *model.InstallmentData, shippingSupport *model.ShippingNew) (*model.ProductES, error) {
	//1. Build product struct es from mongo
	productES := model.NewDefaultProductES(product)
	productES = model.BuildShopRelatedData(productES, merchant)
	productES = model.BuildShopShippingSupport(productES, shippingSupport)
	if installment != nil {
		productES.IsInstallment = installment.IsActive
	}
	//2. Build number facet
	attributeFacets := buildAttributeFacets(product)
	if attributeFacets != nil {
		productES.NumberFacets = append(productES.NumberFacets, attributeFacets...)
	}

	warehouseFacet := buildWarehouseFacets(merchant)
	if warehouseFacet != nil {
		productES.NumberFacets = append(productES.NumberFacets, warehouseFacet...)
	}
	//add cate ids \
	cateIds, err := util.GetCategory(product.CategoryId)
	if err == nil {
		productES.CategoryIds = cateIds
	}

	return productES, nil
}

// BuildFieldES
/**
BuildFieldES using for parse to es fields
@params: mongoFields []string
@Return: esFields []string
*/
func (pu *productBuilder) BuildFieldES(mongoFields []string) []string {
	esFields := []string{"updated_at"}
	if helper.IndexOf(mongoFields, "category_id") != -1 {
		esFields = append(esFields, "category_name_suggest", "category_ids")
	}
	for _, field := range mongoFields {
		if esField, ok := mapESFields[field]; ok {
			esFields = append(esFields, esField)
		} else {
			esFields = append(esFields, field)
		}
	}
	return esFields
}

func buildAttributeFacets(product *model.Product) []model.NumberFacet {
	if product == nil {
		return nil
	}

	var numberFacets []model.NumberFacet
	attributes := product.Attribute
	if attributes == nil {
		return numberFacets
	}
	for key, value := range attributes {
		attribute := model.AttributeMapping{}
		err := json.Unmarshal([]byte(value), &attribute)
		if err != nil {
			continue
		}
		if attribute.ProductOption == "" || len(attribute.Value) <= 0 {
			continue
		}
		for _, valueOption := range attribute.Value {
			if valueOption.OptionId > 0 {
				numberFacet := model.NumberFacet{
					Name:  strings.ToLower(key),
					Value: int64(valueOption.OptionId),
				}
				numberFacets = append(numberFacets, numberFacet)
			}
		}
	}
	return numberFacets
}

func buildWarehouseFacets(merchant *model.Merchant) []model.NumberFacet {
	if merchant == nil {
		return nil
	}

	var numberFacets []model.NumberFacet

	numberFacet := model.NumberFacet{
		Name:  "shop_warehouse_city_id",
		Value: int64(merchant.WareHouseRegionId),
	}
	numberFacets = append(numberFacets, numberFacet)
	return numberFacets
}
