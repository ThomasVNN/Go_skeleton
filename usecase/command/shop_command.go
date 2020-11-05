package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/util"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-kit/kit/log/level"

	"github.com/go-kit/kit/log"

	"github.com/segmentio/kafka-go"
	"gitlab.thovnn.vn/core/golang-sdk/spubsub"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/model"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/storage/esstore"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/storage/mgostore"
)

type shopCommand struct {
	esStore   esstore.ESStore
	shopStore mgostore.ShopStore
	logger    log.Logger
}

type Param struct {
	MerchantId int32
	Value      string
}

type ShopAdsParam struct {
	MerchantId  int32
	ShopAdsPlus bool
	ShopAds     bool
}

func NewShop(esStore esstore.ESStore, mgoStore mgostore.ShopStore, logger log.Logger) model.ShopEvents {
	return &shopCommand{
		esStore:   esStore,
		shopStore: mgoStore,
		logger:    logger,
	}
}

/*
	On/Off shop subscribe event es7.product.update
	@params: data *spubsub.Message
	1. Check validate data input
	2. Get merchant by seller admin id from Mongo DB
	3. Get product by merchant id from ES V.7
	4. Update to ES V.7
*/
func (m *shopCommand) OnOffShop(data *spubsub.Message) error {
	shop := model.Shop{}
	err := json.Unmarshal(data.Data, &shop)
	if err != nil {
		return err
	}

	//1. Check validate data input
	if shop.SellerAdminId < 0 {
		level.Error(m.logger).Log("msg", "seller admin id is null")
		return err
	}

	//2. GetMerchantIdBySellerAdminId from Mongo DB
	merchantMgo, err := m.shopStore.GetMerchantByExternalId(shop.SellerAdminId)
	if err != nil {
		fmt.Println("GetMerchantByExternalId error", err.Error())
		return err
	}
	var lastProductId int32 = 0
	for true {
		//3. GetProductByMerchantId
		productEs, err := m.esStore.GetProductByMerchantId(merchantMgo.MerchantId, lastProductId)
		if err != nil {
			fmt.Println("GetProductByMerchantId error", err.Error())
		}
		if len(productEs) == 0 {
			break
		}
		//4.Update status_new in ES
		updateDataEs := make(map[string]interface{})
		mapListProcessProductsIds := make(map[int32]int32)
		for _, product := range productEs {
			mapListProcessProductsIds[product] = merchantMgo.MerchantId
		}
		updateDataEs["shop_status_id"] = merchantMgo.Status
		index := os.Getenv("ES_INDEX")
		err = m.esStore.UpdateBulkByInputData(index, mapListProcessProductsIds, updateDataEs)
		if err != nil {
			fmt.Println("UpdateBulkByInputData", err.Error())
		}
		lastProductId = productEs[len(productEs)-1]
	}
	return nil
}

func (m *shopCommand) OnOffShopCertificate(data *spubsub.Message) error {
	level.Debug(m.logger).Log("msg", "OnOffShopCertificate")
	shop := model.Shop{}
	err := json.Unmarshal(data.Data, &shop)
	if err != nil {
		level.Error(m.logger).Log("msg", fmt.Sprint("cannot unmarshal data from rabbit-mq ", err))
		return err
	}
	if shop.SellerAdminId < 0 {
		level.Error(m.logger).Log("msg", "seller admin id is null")
		return err
	}
	shopCetificate, err := m.shopStore.GetCertifiedByMerchantID(shop.SellerAdminId)
	if err != nil {
		level.Error(m.logger).Log("msg",
			fmt.Sprint("cannot get is_certified from table merchant_info "+strconv.Itoa(int(shop.SellerAdminId))), err.Error())
		return err
	}

	isUpdateCertified := false
	if shopCetificate.AllAttribute["is_certified"].Value.(float64) == 1 {
		isUpdateCertified = true
	}

	var lastProductId int32 = 0
	for true {
		productES, err := m.esStore.GetProductByMerchantId(shop.SellerAdminId, lastProductId)
		if err != nil {
			fmt.Println("GetProductByMerchantId error", err.Error())
			level.Error(m.logger).Log("msg", fmt.Sprint("GetProductByMerchantId error", err.Error()))
			return err
		}
		if len(productES) == 0 {
			break
		}
		updateDataEs := make(map[string]interface{})
		mapListProcessProductsIds := make(map[int32]int32)
		for _, product := range productES {
			mapListProcessProductsIds[product] = shop.SellerAdminId
		}

		// Update field "is_certificate_shop" into ES
		updateDataEs["is_certificate_shop"] = isUpdateCertified
		index := os.Getenv("ES_INDEX")
		err = m.esStore.UpdateBulkByInputData(index, mapListProcessProductsIds, updateDataEs)
		if err != nil {
			level.Error(m.logger).Log("msg", fmt.Sprint("cannot update into es ", err))
			return err
		}
		lastProductId = productES[len(productES)-1]
	}
	return nil
}

func (m *shopCommand) ShopSupportShippingFee(data *spubsub.Message) error {
	level.Debug(m.logger).Log("msg", "ShopSupportShippingFee")
	shop := model.Shop{}
	err := json.Unmarshal(data.Data, &shop)
	if err != nil {
		level.Error(m.logger).Log("msg", "cannot unmarshal, "+err.Error())
		return err
	}
	if shop.SellerAdminId < 0 {
		level.Error(m.logger).Log("msg", "seller admin id is null")
		return err
	}
	listStoreID, err := m.shopStore.GetListShippingNewByStoreID(shop.SellerAdminId)
	if err != nil {
		level.Error(m.logger).Log("msg", "cannot get store id from mongo, "+err.Error())
		return err
	}

	merchantID, err := m.shopStore.GetMerchantByExternalId(shop.SellerAdminId)
	if err != nil {
		level.Error(m.logger).Log("msg", "cannot get store id from mongo, "+err.Error())
		return err
	}
	var lastProductId int32 = 0
	for true {
		productES, err := m.esStore.GetProductByMerchantId(merchantID.MerchantId, lastProductId)
		if err != nil {
			level.Error(m.logger).Log("msg", "cannot get product from ES, "+err.Error())
			return err
		}
		if len(productES) == 0 {
			break
		}
		updateDataES := make(map[string]interface{})
		mapListProcessProductsIds := make(map[int32]int32)
		for _, product := range productES {
			mapListProcessProductsIds[product] = merchantID.MerchantId
		}
		updateDataES["is_shop_support_shipping_fee"] = false
		for _, level := range listStoreID.Levels {
			if level.IsActive {
				updateDataES["is_shop_support_shipping_fee"] = level.IsActive
			}
		}
		index := os.Getenv("ES_INDEX")
		err = m.esStore.UpdateBulkByInputData(index, mapListProcessProductsIds, updateDataES)
		if err != nil {
			level.Error(m.logger).Log("msg", "cannot update bulk data into ES , "+err.Error())
			return err
		}
		lastProductId = productES[len(productES)-1]
	}
	return nil
}

func (m *shopCommand) ShopInstallment(data *spubsub.Message) error {
	level.Debug(m.logger).Log("msg", "ShopInstallment")
	shopInstallment := model.Shop{}
	err := json.Unmarshal(data.Data, &shopInstallment)
	if err != nil {
		level.Error(m.logger).Log("msg", "cannot unmarshal, "+err.Error())
		return err
	}
	if shopInstallment.SellerAdminId < 0 {
		level.Error(m.logger).Log("msg", "seller admin id is null")
		return err
	}
	merchantID, err := m.shopStore.GetMerchantByExternalId(shopInstallment.SellerAdminId)
	isInstallment := false
	if err != nil {
		level.Error(m.logger).Log("msg", "cannot get store id from mongo, "+err.Error())
		return err
	}
	if merchantID.IsInstallment == 1 {
		isInstallment = true
	}

	var lastProductId int32 = 0
	for true {
		productEs, err := m.esStore.GetProductByMerchantId(merchantID.MerchantId, lastProductId)
		if err != nil {
			level.Error(m.logger).Log("msg", "cannot get product from ES, "+err.Error())
			return err
		}
		if len(productEs) == 0 {
			break
		}
		updateDataEs := make(map[string]interface{})
		mapListProcessProductsIds := make(map[int32]int32)
		for _, product := range productEs {
			mapListProcessProductsIds[product] = merchantID.MerchantId
		}
		updateDataEs["is_installment"] = isInstallment
		index := os.Getenv("ES_INDEX")
		err = m.esStore.UpdateBulkByInputData(index, mapListProcessProductsIds, updateDataEs)
		if err != nil {
			level.Error(m.logger).Log("msg", "cannot update bulk data into ES, "+err.Error())
			return err
		}
		lastProductId = productEs[len(productEs)-1]
	}
	return nil
}

func (m *shopCommand) ShopPromotionApp(data *spubsub.Message) error {
	level.Debug(m.logger).Log("msg", "shop support promotion app")
	shop := model.Shop{}
	err := json.Unmarshal(data.Data, &shop)
	if err != nil {
		level.Error(m.logger).Log("msg", "cannot unmarshal, "+err.Error())
		return err
	}
	if shop.SellerAdminId < 0 {
		level.Error(m.logger).Log("msg", "seller admin id is null")
		return err
	}
	appLevel, err := m.shopStore.GetAppLevelByStoreID(shop.SellerAdminId)
	if err != nil {
		level.Error(m.logger).Log("msg", "cannot get storeID from mongo in shipping_payment_config table, "+err.Error())
		return err
	}
	var lastProductId int32 = 0
	for true {
		productES, err := m.esStore.GetProductByMerchantId(appLevel.MerchantID, lastProductId)
		if err != nil {
			level.Error(m.logger).Log("msg", "cannot get product from ES, "+err.Error())
			return err
		}
		if len(productES) == 0 {
			break
		}
		updateDataES := make(map[string]interface{})
		mapListProcessProductsIds := make(map[int32]int32)
		for _, product := range productES {
			mapListProcessProductsIds[product] = appLevel.MerchantID
		}

		for _, level := range appLevel.Level {
			updateDataES["app_discount_percentage"] = level.DiscountPercent
		}
		index := os.Getenv("ES_INDEX")
		err = m.esStore.UpdateBulkByInputData(index, mapListProcessProductsIds, updateDataES)
		if err != nil {
			level.Error(m.logger).Log("msg", "cannot update bulk data into ES, "+err.Error())
			return err
		}
		lastProductId = productES[len(productES)-1]
	}
	return nil
}

func (m *shopCommand) ShopInstant(msg kafka.Message) error {
	changeProduct := model.ShopInstant{}
	data := msg.Value
	err := json.Unmarshal(data, &changeProduct)
	if err != nil {
		return err
	}

	merchant, err := m.shopStore.GetMerchantByExternalId(changeProduct.SellerAdminId)
	if err != nil {
		level.Error(m.logger).Log("msg", "cannot get store id from mongo, "+err.Error())
		return err
	}
	var lastProductId int32 = 0
	for true {
		time.Sleep(time.Millisecond * 500)
		productEs, err := m.esStore.GetProductByMerchantId(merchant.MerchantId, lastProductId)
		if err != nil {
			level.Error(m.logger).Log("msg", "cannot get product from ES, "+err.Error())
			return err
		}
		if len(productEs) == 0 {
			break
		}
		updateDataEs := make(map[string]interface{})
		mapListProcessProductsIds := make(map[int32]int32)
		for _, product := range productEs {
			mapListProcessProductsIds[product] = merchant.MerchantId
		}

		extendedShippingPackage := make(map[string]interface{})
		if changeProduct.ExtendedShippingPackage.IsUsingInstant == 0 {
			extendedShippingPackage["is_using_instant"] = false
		} else if changeProduct.ExtendedShippingPackage.IsUsingInstant == 1 {
			extendedShippingPackage["is_using_instant"] = true
		}

		if changeProduct.ExtendedShippingPackage.IsUsingInDay == 0 {
			extendedShippingPackage["is_using_in_day"] = false
		} else if changeProduct.ExtendedShippingPackage.IsUsingInDay == 1 {
			extendedShippingPackage["is_using_in_day"] = true
		}

		if changeProduct.ExtendedShippingPackage.IsSelfShipping == 0 {
			extendedShippingPackage["is_self_shipping"] = false
		} else if changeProduct.ExtendedShippingPackage.IsSelfShipping == 1 {
			extendedShippingPackage["is_self_shipping"] = true
		}
		updateDataEs["extended_shipping_package"] = extendedShippingPackage
		index := os.Getenv("ES_INDEX")

		err = m.esStore.UpdateBulkByInputData(index, mapListProcessProductsIds, updateDataEs)
		if err != nil {
			level.Error(m.logger).Log("msg", "cannot update bulk data into ES, "+err.Error())
			return err
		}
		lastProductId = productEs[len(productEs)-1]
	}
	return nil
}

func (m *shopCommand) ShopSenPoint(data *spubsub.Message) error {
	level.Debug(m.logger).Log("shop support sen point")
	shop := model.Shop{}
	err := json.Unmarshal(data.Data, &shop)
	if err != nil {
		level.Error(m.logger).Log("msg", "cannot unmarshal, "+err.Error())
		return err
	}

	point, err := m.shopStore.GetSenPointExternalID(shop.SellerAdminId)
	if err != nil {
		level.Error(m.logger).Log("msg", "cannot get storeID from mongo in shipping_payment_config table, "+err.Error())
		return err
	}
	var lastProductId int32 = 0
	for true {
		time.Sleep(time.Millisecond * 500)
		productES, err := m.esStore.GetProductByMerchantId(point.MerchantID, lastProductId)
		if err != nil {
			level.Error(m.logger).Log("msg", "cannot get product from ES, "+err.Error())
			return err
		}

		if len(productES) == 0 {
			break
		}

		updateDataES := make(map[string]interface{})
		mapListProcessProductsIds := make(map[int32]int32)
		for _, product := range productES {
			mapListProcessProductsIds[product] = point.MerchantID
		}
		updateDataES["is_loyalty"] = point.Loyalty.IsActive

		index := os.Getenv("ES_INDEX")
		err = m.esStore.UpdateBulkByInputData(index, mapListProcessProductsIds, updateDataES)
		if err != nil {
			level.Error(m.logger).Log("msg", "cannot update bulk data into ES, "+err.Error())
			return err
		}
		lastProductId = productES[len(productES)-1]
	}
	return nil
}

func (m *shopCommand) ShopRatingInfo(data *spubsub.Message) error {
	level.Debug(m.logger).Log("msg", "shop rating info")
	shops := []model.ShopRatingInfo{}
	err := json.Unmarshal(data.Data, &shops)
	if err != nil {
		level.Error(m.logger).Log("msg", "cannot unmarshal, "+err.Error())
		return err
	}

	updateDataES := make([]map[string]interface{}, len(shops))
	for _, s := range shops {
		dataUpdate := make(map[string]interface{})
		dataUpdate["product_id"] = s.ProductID
		dataUpdate["rating_percentage"] = s.RatingPercentage
		dataUpdate["total_rated"] = s.TotalRated
		updateDataES = append(updateDataES, dataUpdate)
	}
	_, err = m.esStore.UpdateBulk(updateDataES)
	if err != nil {
		level.Error(m.logger).Log("msg", "cannot update bulk data into ES, "+err.Error())
		return err
	}
	return nil
}

func (m *shopCommand) ShopWareHouse(msg kafka.Message) error {
	shop := model.Shop{}
	data := msg.Value
	err := json.Unmarshal(data, &shop)
	if err != nil {
		return err
	}

	if shop.SellerAdminId < 0 {
		level.Error(m.logger).Log("msg", "seller admin id is null")
		return err
	}

	merchant, err := m.shopStore.GetMerchantByExternalId(shop.SellerAdminId)
	if err != nil {
		level.Error(m.logger).Log("msg", fmt.Sprint("cannot get merchantID from table merchant_flat", err))
		return err
	}
	m.UpdateESByQueryShopWarehouse(merchant.MerchantId, shop.SellerAdminId)
	/*if shop.Type == 1 { // warehouse

	} else if shop.Type == 2 { // shop status
		m.UpdateESByQueryShopStatus(merchant.MerchantId, merchant.Status)
	}*/
	return nil
}

func (m *shopCommand) UpdateESByQueryShopWarehouse(merchantId int32, sellerAdminId int32) error {
	warehouse, err := m.shopStore.GetCertifiedByMerchantID(sellerAdminId)
	if err != nil {
		level.Error(m.logger).Log("msg",
			fmt.Sprint("cannot get is_certified from table merchant_info "+strconv.Itoa(int(sellerAdminId))), err.Error())
		return err
	}
	if warehouse == nil {
		return nil
	}
	value, ok := util.InterfaceToUint(warehouse.AllAttribute["warehouse_city"].Value)
	if !ok {
		return errors.New("ware house city error")
	}
	dataParam := &Param{
		MerchantId: merchantId,
		Value:      strconv.Itoa(int(value)),
	}
	query := `{
  				"query": {
    				"term": {
      					"admin_id": {{.MerchantId }}   
    					}
  					},
  					"script": {
    					"source": "ctx._source.number_facets.removeIf(attribute -> attribute instanceof List);ctx._source.number_facets.removeIf(attr -> attr.name == 'shop_warehouse_city_id');ctx._source.number_facets.add(params.warehouse);",
    					"params": {
      						"warehouse":{
	          					"name": "shop_warehouse_city_id",
								"value": {{.Value}}
	        				}
    					}
  					}
				}`
	builder := &strings.Builder{}
	t := template.Must(template.New("updateWarehouse").Parse(query))
	if err := t.Execute(builder, dataParam); err != nil {
		return nil
	}
	s := builder.String()
	//fmt.Println("Query:", s)
	body := strings.NewReader(s)
	url := os.Getenv("ES_URL") + "/" + os.Getenv("ES_INDEX") + "/" + "_update_by_query"
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("Data return: ", err, bodyBytes)
	}
	return nil
}

func (m *shopCommand) UpdateESByQueryShopStatus(merchantId int32, value int32) error {
	dataParam := &Param{
		MerchantId: merchantId,
		Value:      strconv.Itoa(int(value)),
	}
	query := `{
  				"query": {
    				"term": {
      					"admin_id": {{.MerchantId }}   
    					}
  					},
  					"script": {
    					"source": "ctx._source.shop_status_id = params.shop_status_id",
						"params": {
         				   "shop_status_id": {{.Value}}
						}
  					}
				}`
	builder := &strings.Builder{}
	t := template.Must(template.New("updateShopStatus").Parse(query))
	if err := t.Execute(builder, dataParam); err != nil {
		return nil
	}
	s := builder.String()
	//fmt.Println("Query:", s)
	body := strings.NewReader(s)
	url := os.Getenv("ES_URL") + "/" + os.Getenv("ES_INDEX") + "/" + "_update_by_query"
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("Data return: ", err, bodyBytes)
	}
	return nil
}

func (m *shopCommand) ShopAds(msg kafka.Message) error {
	shop := model.ShopAds{}
	shopAdsPlus := false
	shopAds := false
	data := msg.Value
	err := json.Unmarshal(data, &shop)
	fmt.Println("HoanTK log shop ads:", shop)
	if err != nil {
		return err
	}
	if shop.SellerAdminId < 0 {
		level.Error(m.logger).Log("msg", "seller admin id is null")
		return err
	}

	if shop.AdsService == nil {
		level.Error(m.logger).Log("msg", "ads service is null")
		return err
	}

	merchant, err := m.shopStore.GetMerchantByExternalId(shop.SellerAdminId)
	if err != nil {
		level.Error(m.logger).Log("msg", fmt.Sprint("cannot get merchantID from table merchant_flat", err))
		return err
	}

	if shop.AdsService.AdPlus == 1 {
		shopAdsPlus = true
	}

	if shop.AdsService.ShopAds == 1 {
		shopAds = true
	}

	dataParam := &ShopAdsParam{
		MerchantId:  merchant.MerchantId,
		ShopAdsPlus: shopAdsPlus,
		ShopAds:     shopAds,
	}

	query := `{
  				"query": {
    				"term": {
      					"admin_id": {{.MerchantId }}   
    					}
  					},
  					"script": {
    					"source": "ctx._source.shop_ads_plus = params.shop_ads_plus;ctx._source.shop_ads = params.shop_ads;",
    					"params": {
      						"shop_ads_plus": {{.ShopAdsPlus}},
							"shop_ads" : {{.ShopAds}}
    					}
  					}
				}`

	t := template.Must(template.New("updateShopAds").Parse(query))
	builder := &strings.Builder{}
	if err := t.Execute(builder, dataParam); err != nil {
		return nil
	}
	s := builder.String()
	body := strings.NewReader(s)
	url := os.Getenv("ES_URL") + "/" + os.Getenv("ES_INDEX") + "/" + "_update_by_query?wait_for_completion=false&scroll_size=100&conflicts=proceed"
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("Data return: ", err, bodyBytes)
	}

	return nil
}
