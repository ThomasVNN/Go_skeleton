package command

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/golang/protobuf/jsonpb"
	"github.com/joho/godotenv"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/segmentio/kafka-go"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/pkg/helper"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/repo"
	"sync"

	"gitlab.thovnn.vn/dc1/product-discovery/es_service/model"

	"gitlab.thovnn.vn/dc1/product-discovery/es_service/storage/esstore"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/storage/mgostore"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/usecase/builder"
	productBase "gitlab.thovnn.vn/protobuf/internal-apis-go/product/base"
)

type changeProductCommand struct {
	esStore           esstore.ESStore
	mgoStore          mgostore.MgoStore
	variantStore      mgostore.VariantStore
	shopStore         mgostore.ShopStore
	logStore          mgostore.LogStore
	categoryRepo      repo.CategoryServiceClient
	shopRepo          repo.ShopServiceClient
	installmentRepo          repo.InstallmentServiceClient
	logger            log.Logger
	crossCheckChannel chan []*model.ProductCrossCheck
	batchChannel chan *model.Product
	m            sync.Mutex
}

var esFieldToUpdatePromotion = []string{
	"price",
	"variants",
	"is_promotion",
}

var defaultFieldsToUpdate = []string{"is_loyalty", "shop_status_id", "min_shop_support_shipping_fee", "is_shop_support_shipping_fee","number_facets" }

// OnProductAdded subscribe event es7.product.add
/**
@params: data []byte
1. Build product es data from productId
2. Add to ES
*/
func (m *changeProductCommand) OnProductAdded(msg kafka.Message) error {
	logger := log.With(m.logger, "prefix", "OnProductAdded")
	data := msg.Value
	var product *model.Product
	var shippingSupport *model.ShippingNew
	hashData := fmt.Sprintf("%x", sha1.Sum([]byte(data)))
	ll := map[string]interface{}{"hash": hashData, "data": string(data), "offset": msg.Offset, "partition": msg.Partition}
	logData, _ := json.Marshal(ll)
	level.Info(logger).Log(
		"msg", fmt.Sprint("received data from "+msg.Topic),
		"meta_data", string(logData))

	changeProduct := productBase.ChangedProduct{}
	r := strings.NewReader(string(data))
	err := jsonpb.Unmarshal(r, &changeProduct)
	if err != nil {
		level.Error(logger).Log("msg", fmt.Sprint("add fail, unmarshal json fail, ", err),
			"meta_data", string(logData))
		return err
	}
	pBuilder := builder.New()

	product, err = m.mgoStore.GetProductById(changeProduct.ProductId)

	defer func() {
		logProductEs := pBuilder.BuildProductCrossCheckLog(changeProduct.ProductId, err)
		_ = m.logStore.Add(logProductEs)
	}()
	if err != nil {
		level.Error(logger).Log("msg", fmt.Sprint("add fail, GetProductById error, ", err),
			"meta_data", string(logData))
		return err
	}

	variants, installment, shippingSupport, err := m.getProductExternalDetails(*product, nil)

	if err != nil {
		return err
	}
	product.Variants = variants
	//Get shop info
	merchant, err := m.shopRepo.GetShopByIds([]int32{int32(product.SellerAdminId)})
	if err != nil {
		level.Error(logger).Log("msg", fmt.Sprint("OnProductAdded, GetMerchantFromShopService error, ", err),
			"seller_admin_id", string(product.SellerAdminId))
		return err
	}

	//1. Build product es data from productId
	productES, _ := pBuilder.BuildFullEsData(product, merchant[product.SellerAdminId], installment, shippingSupport)
	productES.CategoryNameSuggest = m.setCategoryNameSuggest(productES.CategoryPath)
	//2. Add to ES
	productId := fmt.Sprint(productES.ProductId)
	err = m.esStore.Add(productES, productId)
	if err != nil {
		level.Error(logger).Log("msg", fmt.Sprint("add fail, esStore.Add error, ", err),
			"meta_data", string(logData))
	} else {
		level.Info(logger).Log("msg", "add success",
			"meta_data", string(logData))
	}

	return err
}

// OnProductUpdated subscribe event es7.product.update
/**
@param: data []byte
1. Update one product
2. Update multy product
*/
func (m *changeProductCommand) OnProductUpdated(msg kafka.Message) error {
	logger := log.With(m.logger, "prefix", "OnProductUpdated")
	data := msg.Value

	hashData := fmt.Sprintf("%x", sha1.Sum([]byte(data)))
	ll := map[string]interface{}{"hash": hashData, "data": string(data), "offset": msg.Offset, "partition": msg.Partition}
	logData, _ := json.Marshal(ll)

	level.Info(logger).Log(
		"msg", fmt.Sprint("received data from "+msg.Topic),
		"meta_data", string(logData))

	changedProductCollection := &productBase.ChangedProductCollection{}
	r := strings.NewReader(string(data))
	if err := jsonpb.Unmarshal(r, changedProductCollection); err != nil {
		level.Error(logger).Log("msg", fmt.Sprint("updated fail, unmarshal json fail, ", err),
			"meta_data", string(logData))
		return err
	}
	products := changedProductCollection.GetData()
	total := len(products)
	if total <= 0 {
		level.Error(logger).Log("msg", fmt.Sprint("updated fail, total is less 0"),
			"meta_data", string(logData))
		return errors.New("total is less 0")
	}
	//1. Update one product
	if total == 1 {
		err := m.UpdateOne(products[0])
		if err != nil {
			level.Error(logger).Log("msg", fmt.Sprint("update fail, updateOne, ", err),
				"meta_data", string(logData))
			return err
		}
	} else {
		//2. Update multy product
		err := m.UpdateMulty(products)
		if err != nil {
			level.Error(logger).Log("msg", fmt.Sprint("update multi fail, ", err),
				"meta_data", string(logData))
			return err
		}
	}

	level.Info(logger).Log("msg", fmt.Sprint("update success"),
		"meta_data", string(logData))
	return nil
}

func (m *changeProductCommand) buildCrossCheckBulk(productIds []uint32, err error) {
	pBuilder := builder.New()
	var crossCheckBulk []*model.ProductCrossCheck
	for _, id := range productIds {
		crossCheckBulk = append(crossCheckBulk, pBuilder.BuildProductCrossCheckLog(id, err))
	}
	m.crossCheckChannel <- crossCheckBulk
}

func getDataChangedProduct(changedProduct []*productBase.ChangedProduct) ([]uint32, map[uint32]*productBase.ChangedProduct) {
	var productIds = []uint32{}
	mapChangeProduct := map[uint32]*productBase.ChangedProduct{}
	for _, changeProduct := range changedProduct {
		productId := changeProduct.GetProductId()
		productIds = append(productIds, productId)
		mapChangeProduct[productId] = changeProduct
	}
	return productIds, mapChangeProduct
}

func (m *changeProductCommand) buildEsProductData(product *model.Product, installment *model.InstallmentData, shippingSupport *model.ShippingNew, changedFields []string) (map[string]interface{}, error) {
	var err error
	logger := log.With(m.logger, "prefix", "buildEsProductData")
	pBuilder := builder.New()

	merchant, err := m.shopRepo.GetShopByIds([]int32{int32(product.SellerAdminId)})
	if err != nil {
		level.Error(logger).Log("msg", fmt.Sprint("buildEsProductData, GetMerchantFromShopService error, ", err),
			"seller_admin_id", string(product.SellerAdminId))
		return nil, err
	}
	//1. Build final data product es from productId
	productES, _ := pBuilder.BuildFullEsData(product, merchant[int32(product.SellerAdminId)], installment, shippingSupport)
	if helper.IndexOf(changedFields, "category_id") != -1 {
		productES.CategoryNameSuggest = m.setCategoryNameSuggest(productES.CategoryPath)
	}

	//2. Build final product es from product es
	esFields := pBuilder.BuildFieldES(changedFields)
	if installment != nil {
		esFields = append(esFields, "is_installment")
	}
	finalData := productES.SelectFields(append(esFields, defaultFieldsToUpdate...)...)
	finalData["product_id"] = product.ProductId
	return finalData, err
}

/**
@params: changedData *productBase.ChangedProduct
1. Build final data product es from productId
2. Build final product es from product es
3. Update to ES
*/
func (m *changeProductCommand) UpdateOne(changedData *productBase.ChangedProduct) error {
	pBuilder := builder.New()
	var product *model.Product
	var err error
	var variants []*model.Variant
	var installment *model.InstallmentData
	var shippingSupport *model.ShippingNew
	var finalData map[string]interface{}
	logger := log.With(m.logger, "prefix", "UpdateOne")
	product, err = m.mgoStore.GetProductById(changedData.ProductId)
	defer func() {
		logProductEs := pBuilder.BuildProductCrossCheckLog(changedData.ProductId, err)
		m.crossCheckChannel <- []*model.ProductCrossCheck{logProductEs}
	}()
	if err != nil {
		_ = level.Error(logger).Log("msg", fmt.Sprint("Failed to get product with id ", changedData.ProductId))
		return err
	}

	variants, installment, shippingSupport, err = m.getProductExternalDetails(*product, nil)
	if err != nil {
		_ = level.Error(logger).Log("msg", fmt.Sprint("Failed to build product es with id ", changedData.ProductId))
		return nil
	}
	product.Variants = variants

	finalData, err = m.buildEsProductData(product, installment, shippingSupport, changedData.GetFields().GetPaths())

	if err != nil {
		return err
	}
	//3. Update to ES
	productId := fmt.Sprint(product.ProductId)
	_, err = m.esStore.Update(finalData, productId)

	//4. stored promotions to ESv7.2

	// build data promotions for Seller
	if isUpdatePromotion(changedData.GetFields().GetPaths()) {
		promotionsData := model.SetPromotions(product)
		fmt.Println("=======promotionsData=======",promotionsData)
		dataEs, err := m.esStore.UpdatePromotionsForSeller(productId,promotionsData)
		fmt.Println(dataEs,err)
	}
	return err
}

/**
@params: changedData []*productBase.ChangedProduct
1. Build final data product es from productId
2. Build final product es from product es
3. Update batch 50 items
*/
func (m *changeProductCommand) UpdateMulty(changedData []*productBase.ChangedProduct) error {
	logger := log.With(m.logger, "prefix", "UpdateMulty")
	pBuilder := builder.New()
	var shippingSupports = make(map[int32]*model.ShippingNew)
	productIds, mapChangedProduct := getDataChangedProduct(changedData)
	products, err := m.mgoStore.GetMultiProducts(productIds)
	if err != nil {
		m.buildCrossCheckBulk(productIds, err)
		return err
	}

	var sellerAdminIds []int32
	for _, product := range products {
		sellerAdminIds = append(sellerAdminIds, int32(product.SellerAdminId))
	}

	merchants, err := m.shopRepo.GetShopByIds(sellerAdminIds)
	if err != nil {
		level.Error(logger).Log("msg", fmt.Sprint("UpdateMulty, GetMerchantFromShopService error, ", err),
			"seller_admin_ids", string(sellerAdminIds))
		return err
	}

	var sliceData []map[string]interface{}
	var promotionsData = make(map[uint32]map[string]interface{})
	count := 0
	for _, product := range products {
		productId := product.ProductId
		if changedData, ok := mapChangedProduct[productId]; ok {
			variants, installment, shippingSupport, err := m.getProductExternalDetails(*product,shippingSupports[product.SellerAdminId])

			if err != nil {
				_ = level.Error(logger).Log("msg", fmt.Sprint("Failed to get product with id ", changedData.ProductId))
				continue
			}
			shippingSupports[product.SellerAdminId] = shippingSupport
			product.Variants = variants
			//1. Build final data product es from productId
			productES, err := pBuilder.BuildFullEsData(product, merchants[product.SellerAdminId], installment, shippingSupport)
			//1.1 build data promotions for Seller
			if isUpdatePromotion(changedData.GetFields().GetPaths()) {
				promotionsData[productId] = model.SetPromotions(product)
			}
			if err != nil {
				continue
			}
			if helper.IndexOf(changedData.GetFields().GetPaths(), "category_id") != -1 {
				productES.CategoryNameSuggest = m.setCategoryNameSuggest(productES.CategoryPath)
			}
			count++
			//2. Build final product es from product es
			esFields := pBuilder.BuildFieldES(changedData.GetFields().GetPaths())
			if installment != nil {
				esFields = append(esFields, "is_installment")
			}
			finalData := productES.SelectFields(append(esFields, defaultFieldsToUpdate...)...)
			finalData["product_id"] = productES.ProductId
			level.Debug(logger).Log("msg", fmt.Sprintf("%+v, %d", finalData, productES.ProductId))

			sliceData = append(sliceData, finalData)
			//3. Update batch 50 items
			if count%50 == 0 {
				bulkResponse, _ := m.esStore.UpdateBulk(sliceData)
				logsProductEs := pBuilder.BuildESResponseCrossCheckLog(bulkResponse)
				m.crossCheckChannel <- logsProductEs
				sliceData = []map[string]interface{}{}
			}
		}
	}
	if len(sliceData) > 0 {
		bulkResponse, _ := m.esStore.UpdateBulk(sliceData)
		logsProductEs := pBuilder.BuildESResponseCrossCheckLog(bulkResponse)
		m.crossCheckChannel <- logsProductEs
	}
	//4. stored promotions to ESv7.2
	for _,productId := range productIds {
		if promotion, ok := promotionsData[productId]; ok {
			_, err = m.esStore.UpdatePromotionsForSeller(fmt.Sprint(productId),promotion)
		}
	}

	return nil
}

func (m *changeProductCommand) setCategoryNameSuggest(categoryPath string) string {
	logger := log.With(m.logger, "prefix", "setCategoryNameSuggest")
	if categoryPath == "" {
		level.Error(logger).Log("msg", "categoryId empty: "+categoryPath)
		return ""
	}
	paths := []string{"category_id", "name"}
	categoryId, err := helper.GetLastCategoryId(categoryPath)
	if err != nil {
		level.Error(logger).Log("msg", fmt.Sprint("GetLastCategoryId err: ", err))
		return ""
	}
	categoryInfo, err := m.categoryRepo.GetCategoryById(categoryId, paths)
	if err != nil || categoryInfo == nil {
		level.Error(logger).Log("msg", fmt.Sprint("GetCategoryById err: ", err))
		return ""
	}
	return categoryInfo.Name
}

func (m *changeProductCommand) OnProductScoreUpdated(data []byte) {
	return
}

func ChangeProductNew(esStore esstore.ESStore, mgoStore mgostore.MgoStore,
	variantStore mgostore.VariantStore, logStore mgostore.LogStore, shopStore mgostore.ShopStore,
	categoryRepo repo.CategoryServiceClient, shopRepo repo.ShopServiceClient, installmentRepo repo.InstallmentServiceClient, logger log.Logger,
	crossCheckChannel chan []*model.ProductCrossCheck) model.ProductEvents {
	return &changeProductCommand{
		esStore:           esStore,
		variantStore:      variantStore,
		logStore:          logStore,
		mgoStore:          mgoStore,
		categoryRepo:      categoryRepo,
		shopRepo:          shopRepo,
		logger:            logger,
		shopStore:         shopStore,
		installmentRepo: installmentRepo,
		crossCheckChannel: crossCheckChannel,
	}
}

func (m *changeProductCommand) getProductExternalDetails(product model.Product, shippingSupport *model.ShippingNew) ([]*model.Variant, *model.InstallmentData, *model.ShippingNew, error) {
	logger := log.With(m.logger, "prefix", "getProductExternalDetails")
	var err error
	var variants []*model.Variant
	var shopInstallment *model.InstallmentData = &model.InstallmentData{}
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if product.IsConfigVariant {
			variants, err = m.variantStore.Get(product.ProductId)
			if err != nil && err.Error() != "not found" {
				_ = level.Error(logger).Log("msg", fmt.Sprintf("Failed to get product variants with id %d", product.ProductId))
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		res, er := m.installmentRepo.GetInstallmentByShopID(uint32(product.SellerAdminId))
		if er != nil {
			shopInstallment = nil
			_ = level.Error(logger).Log("msg", fmt.Sprintf("Failed to get installment with shop id %d, err: %s", product.SellerAdminId, er.Error()))
			return
		}
		shopInstallment = res
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if shippingSupport != nil {
			return
		}

		var er error
		shippingSupport, er = m.shopStore.GetListShippingNewByStoreID(product.SellerAdminId)
		if er != nil {
			_ = level.Error(logger).Log("msg", fmt.Sprintf("Failed to get shop shipping support with id %d", product.SellerAdminId))
		}
	}()
	wg.Wait()

	return variants, shopInstallment, shippingSupport, err
}

func GetMerchantFromShopService(ids []int32) (map[int32]*model.Merchant, error) {
	_ = godotenv.Load()
	domain := os.Getenv("SELLER_SHOP_SERVICE_DOMAIN")
	if domain == "" {
		return nil, errors.New("can not call shop seller")
	}

	str := ""
	var checkId = make(map[int32]int32)
	for _, id := range ids {
		if _, ok := checkId[id]; !ok {
			str += strconv.Itoa(int(id)) + ","
			checkId[id] = id
		}

	}
	response, err := http.Get(domain + "/stores/get-stores?ids=" + str)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	type Response struct {
		Code int              `json:"code"`
		Data []model.Merchant `json:"data"`
	}

	var resp = new(Response)
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}

	var mapResp = make(map[int32]*model.Merchant)
	for _, item := range resp.Data {
		if _, ok := mapResp[item.Id]; !ok {
			mapResp[item.Id] = &item
		}
	}

	return mapResp, nil
}

func isUpdatePromotion(requestFields []string) bool {
	for _,field := range requestFields{
		if isExists(esFieldToUpdatePromotion,field){
			return true
		}
	}
	return false
}

func isExists(array []string, item string) bool {
	for _,value := range array{
		if item == value {
			return true
		}
	}

	return false
}
