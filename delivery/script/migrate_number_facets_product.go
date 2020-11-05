package script

import (
	"fmt"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"gitlab.thovnn.vn/core/sen-kit/storage/elastic"
	"gitlab.thovnn.vn/core/sen-kit/storage/mongo"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/model"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/storage/esstore"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/usecase/builder"
	"sync"
	"sync/atomic"
)

var lmGetMerchant = 20
var lmGetProduct = 100
var maxWorker = 10
var merchantFlat = "merchant_flat"
var productFlat = "product_flat"
var DB = "services"

type MigrateNumberFacet struct {
	pBuilder builder.ProductBuilder
	mgoSes   *mgo.Session
	esStore  esstore.ESStore
	logger   log.Logger
}

func NewMigrateNumberFacet(logger log.Logger) (*MigrateNumberFacet, error) {
	//mgo
	mgoConfig := mongo.New("es")
	mgoSes, err := mgoConfig.DB()
	if err != nil {
		return nil, err
	}
	//es
	esConfig := elastic.New("")
	esStore := esstore.New(esConfig.Client(), esConfig.ESIndex, logger)

	return &MigrateNumberFacet{
		pBuilder: builder.New(),
		mgoSes:   mgoSes,
		esStore:  esStore,
		logger:   logger,
	}, nil
}

type Shop struct {
	MerchantId int64 `json:"merchant_id" bson:"merchant_id"`
	ExternalId int64 `json:"external_id" bson:"external_id"`
	Attributes struct {
		WarehouseCity struct {
			AttributeId   int64  `json:"attribute_id" bson:"attribute_id"`
			AttributeCode string `json:"attribute_code" bson:"attribute_code"`
			Value         int64  `json:"value" bson:"value"`
		} `json:"warehouse_city" bson:"warehouse_city"`
	} `json:"all_attribute" bson:"all_attribute"`
}

func RunMigrateNumberFacet(logger log.Logger) error {
	//st := time.Now()
	//m, err := NewMigrateNumberFacet(logger)
	//if err != nil {
	//	return err
	//}
	//defer m.mgoSes.Close()
	//
	//var wg sync.WaitGroup
	//c := make(chan *model.Merchant, 200)
	//var totalShop, totalProduct uint64
	//
	////Update product using 5 worker
	//for i := 1; i <= maxWorker; i++ {
	//	wg.Add(1)
	//	go m.UpdateProducts(&wg, c, &totalShop, &totalProduct, i)
	//}
	//wg.Add(1)
	//go m.GetShops(&wg, c)
	//
	//wg.Wait()
	//et := time.Now()
	//processT := et.Sub(st)
	//fmt.Println("msg", fmt.Sprintf("Total Shop : %v , Total Product : %v ,  Total Run Time : %v", totalShop, totalProduct, processT.Minutes()))
	return nil
}

func (m *MigrateNumberFacet) GetShops(wg *sync.WaitGroup, c chan *model.Merchant) {
	collection := m.mgoSes.DB(DB).C(merchantFlat)
	var lastId int64
	var total int64
	for {
		var shops []Shop
		err := collection.Find(bson.M{"status": bson.M{"$in": []int{2, 3}}, "merchant_id": bson.M{"$gt": lastId}}).Limit(lmGetMerchant).Sort("merchant_id").
			Select(bson.M{"merchant_id": 1, "external_id": 1, "all_attribute.warehouse_city": 1}).All(&shops)
		if err != nil {
			level.Error(m.logger).Log("msg", fmt.Sprintf("err get shop : %v , lastID : %v", err, lastId))
		}

		if len(shops) == 0 {
			break
		}

		lastId = shops[len(shops)-1].MerchantId
		total += int64(len(shops))
		for _, s := range shops {
			merchant := &model.Merchant{
				Id:                int32(s.ExternalId),
				MerchantId:        int32(s.MerchantId),
				WareHouseRegionId: int32(s.Attributes.WarehouseCity.Value),
			}
			c <- merchant
		}

		if len(shops) < lmGetMerchant {
			break
		}
	}
	close(c)
	wg.Done()
}

func (m *MigrateNumberFacet) UpdateProducts(wg *sync.WaitGroup, c chan *model.Merchant, ttShop, ttProduct *uint64, w int) {
	defer wg.Done()
	collection := m.mgoSes.DB(DB).C(productFlat)

	for item := range c {
		var lastId uint32 = 0
		for {
			products, err := m.GetProducts(collection, uint32(item.MerchantId), lastId)
			if err != nil {
				level.Error(m.logger).Log("msg", fmt.Sprintf("err get product : %v , lastID : %v", err, lastId))
			}
			if len(products) <= 0 {
				break
			}

			lastId = products[len(products)-1].ProductId
			atomic.AddUint64(ttProduct, uint64(len(products)))

			var list []map[string]interface{}
			var ids []uint32
			for _, p := range products {
				ids = append(ids, p.ProductId)
				item, err := m.BuildData(&p, item)
				if err != nil {
					level.Error(m.logger).Log("msg", fmt.Sprintf("build data product es : %v , lastID : %v", err, p.ProductId))
				} else {
					list = append(list, item)
				}
			}

			fmt.Println("migrate_number_facets_product", fmt.Sprintf("Worker : %v  , Shop : %v   , Product  : %v", w, item.Id, ids))
			_, err = m.esStore.UpdateBulk(list)
			if err != nil {
				level.Error(m.logger).Log("msg", fmt.Sprintf("Error update product  : %v ", err))
			}
			if len(products) < lmGetProduct {
				break
			}
		}
		atomic.AddUint64(ttShop, 1)
	}
}

func (m *MigrateNumberFacet) GetProducts(collection *mgo.Collection, merchantId uint32, lastId uint32) ([]model.Product, error) {
	var products []model.Product
	err := collection.Find(bson.M{"admin_id": merchantId, "status_new": bson.M{"$in": []uint32{1, 2, 3}}, "stock_status": bson.M{"$in": []uint32{0, 1}}, "product_id": bson.M{"$gt": lastId}}).
		Limit(lmGetProduct).Sort("product_id").All(&products)

	if err != nil {
		return nil, err
	}
	return products, nil
}

func (m *MigrateNumberFacet) BuildData(p *model.Product, s *model.Merchant) (map[string]interface{}, error) {
	item, err := m.pBuilder.BuildFullEsData(p, s, nil, nil)
	if err != nil {
		return nil, err
	}
	fields := []string{"number_facets"}
	rs := item.SelectFields(fields...)
	rs["product_id"] = item.ProductId
	return rs, nil
}
