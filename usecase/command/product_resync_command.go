package command

import (
	"errors"
	"fmt"
	"github.com/globalsign/mgo/bson"
	"github.com/go-redis/redis"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/usecase/helpers"

	"github.com/go-kit/kit/log/level"

	"github.com/go-kit/kit/log"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/model"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/repo"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/storage/esstore"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/storage/mgostore"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/usecase/builder"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/util"
	"sync"
	"time"
)

const maxRetryCount= 10
const limit = 100
var m sync.Map
var wg sync.WaitGroup
var logTime map[string]int64

func (r *changeProductCommand) MultiSync() error {
	//logs, _ := r.logStore.Query(nil, 1000, "")
	//for _, l := range logs {
	//	l.IsSyncEs = false
	//}
	//_ = r.logStore.UpdateEsStatusBulk(logs, true)
	r.init()
	return r.run()
}

func (r *changeProductCommand) init() {
	r.logIncrement("redis", 0)
	r.logIncrement("redis-failures", 0)
	r.logIncrement("es", 0)
	r.logIncrement("es-failures", 0)
	logTime = make(map[string]int64)
}

func (r *changeProductCommand) run() error {
	logger := log.With(r.logger, "prefix", "runResyncCron")
	now := r.getCurrentTime()
	_ = level.Info(logger).Log("msg", fmt.Sprintf("==Start scan resync script at %v", time.Now().Format("2006-01-02 15:04:05")))

	var n = 3

	wg.Add(1)
	go func() {
		defer wg.Done()
		var s sync.WaitGroup
		s.Add(n)
		for i:= 0; i< n; i++ {
			go func() {
				defer s.Done()
				r.updateMultiProducts(true)
			}()
		}
		s.Wait()
		_ = level.Info(logger).Log("msg", fmt.Sprintf("End of ES resync at %v", time.Now()))
		close(r.crossCheckChannel)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var s sync.WaitGroup
		s.Add(n)
		for i:= 0; i< n; i++ {
			go func() {
				defer s.Done()
				r.updateFailedLogs(true)
			}()
		}
		s.Wait()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		now := r.getCurrentTime()

		r.syncESData()
		end := r.getCurrentTime()
		_ = level.Info(logger).Log("msg", fmt.Sprintf("Elapsed es time %v", end - now))
	}()

	wg.Wait()

	totalEs, _ := m.Load("es")
	totalFailedEs, _ := m.Load("es-failures")
	totalRedis, _ := m.Load("redis")
	totalFailedRedis, _ := m.Load("redis-failures")
	_ = level.Info(logger).Log("msg", fmt.Sprintf("==End scan resync script at %s", time.Now().Format("2006-01-02 15:04:05")))
	_ = level.Info(logger).Log("msg", fmt.Sprintf("Total es processed data %v", totalEs))
	_ = level.Info(logger).Log("msg", fmt.Sprintf("Total es failed data %v", totalFailedEs))
	_ = level.Info(logger).Log("msg", fmt.Sprintf("Total redis success data %v", totalRedis))
	_ = level.Info(logger).Log("msg", fmt.Sprintf("Total redis failure data %v", totalFailedRedis))
	_ = level.Info(logger).Log("msg", fmt.Sprintf("Total time %v", logTime))
	end := r.getCurrentTime()
	_ = level.Info(logger).Log("msg", fmt.Sprintf("Elapsed time %d", end - now))
	return nil
}

func (r *changeProductCommand) syncESData() {
	logger := log.With(r.logger, "prefix", "syncESData")
	var s sync.WaitGroup
	var lastId int
	for {
		var logProductIds []uint32
		var productIds []uint32
		t := r.getCurrentTime()
		logs := r.fetchFailedEsData(lastId)
		r.setLogTime("fetch-es-logs", r.getCurrentTime() - t, true)
		if len(logs) == 0 {
			_ = level.Info(logger).Log("msg", fmt.Sprintf("End of getting ES logs at %v", time.Now()))
			break
		}
		for _, l := range logs {
			logProductIds = append(logProductIds, l.ProductId)
		}
		lastId = int(logProductIds[len(logProductIds) - 1])
		products, err := r.mgoStore.GetMultiProducts(logProductIds)
		for _, p := range products {
			productIds = append(productIds, p.ProductId)
		}

		failed := util.UInt32SliceDifference(logProductIds, productIds)
		if len(failed) > 0 {
			if err == nil {
				err = errors.New("Not found")
			}
			r.buildCrossCheckBulk(failed, err)
		}
		s.Add(1)
		go func(products []*model.Product) {
			defer s.Done()
			for _,p := range products {
				r.batchChannel <- p
			}
		}(products)
	}
	s.Wait()

	close(r.batchChannel)
}

func (r *changeProductCommand) fetchFailedRedisData(lastId int) []*model.ProductCrossCheck {
	query := bson.M{
		"is_sync_redis": false,
		"retry_count_redis": bson.M{
			"$lte": maxRetryCount,
		},
	}
	return r.fetchFailedData(query, lastId)
}

func (r *changeProductCommand) fetchFailedEsData(lastId int) []*model.ProductCrossCheck {
	query := bson.M{
		"is_sync_es": false,
		"retry_count_es": bson.M{
			"$lte": maxRetryCount,
		},
	}
	return r.fetchFailedData(query, lastId)
}

func (r *changeProductCommand) fetchFailedData(query bson.M, lastId int) []*model.ProductCrossCheck {
	var logs []*model.ProductCrossCheck
	var err error
	query["product_id"] = bson.M{
		"$gt": lastId,
	}
	for i := 0; i < 4; i++ {
		logs, err = r.logStore.Query(query, limit, "product_id")
		if err == nil {
			continue
		}
		time.Sleep(time.Second * 10)
	}

	return logs
}

func (r *changeProductCommand) getCurrentTime()  int64 {

	return time.Now().Unix()
}

func (r *changeProductCommand) logIncrement(t string, count int) {
	v, _ := m.Load(t)
	if v == nil {
		v = 0
	}
	c := v.(int)
	c += count
	m.Store(t, c)
}

func (r *changeProductCommand) setLogTime(key string, value int64, increment bool) {
	r.m.Lock()
	defer r.m.Unlock()
	if !increment {
		logTime[key] = value
		return
	}
	logTime[key] += value
}

func (r *changeProductCommand) getLogTime(key string) int64 {
	r.m.Lock()
	defer r.m.Unlock()
	return logTime[key]
}

func (r *changeProductCommand) setRedisDataResults(productIds []uint32, err error) error {
	var logs []*model.ProductCrossCheck
	var errStr string
	if err != nil {
		errStr = err.Error()
	}
	for _, id := range productIds {
		l := &model.ProductCrossCheck{
			ProductId:id,
			IsSyncRedis: err == nil,
			ErrorRedis:errStr,
		}
		logs = append(logs, l)
	}
	return r.logStore.UpdateRedisStatusBulk(logs, true)
}

func (r *changeProductCommand) updateFailedLogs(retry bool) {
	for c := range r.crossCheckChannel {
		if len(c) > 0 {
			r.logIncrement("es-failures", len(c))
			_ = r.logStore.UpdateEsStatusBulk(c, retry)
		}
	}
}

func (m *changeProductCommand) updateMultiProductLogs(data map[uint32]interface{}, retry bool) error {
	logger := log.With(m.logger, "prefix", "updateMultiProductLogs")
	pBuilder := builder.New()
	bulkResponse, err := m.esStore.UpsertBulk(data)
	if err != nil {
		_ = level.Error(logger).Log("msg", err)
		return err
	}
	logsProductEs := pBuilder.BuildESResponseCrossCheckLog(bulkResponse)
	_ = m.logStore.UpdateEsStatusBulk(logsProductEs, retry)
	return nil
}

func (r *changeProductCommand) updateMultiProducts(retry bool) {
	var sliceData = make(map[uint32]interface{})
	var count int
	var b = builder.New()
	var variants []*model.Variant
	var installment *model.InstallmentData
	var shippingSupport *model.ShippingNew
	var err error
	var failed []*model.ProductCrossCheck
	fields := helpers.BuildTagList(model.Product{}, []string{"description", "default_listing_score"})
	for c:= range r.batchChannel {
		count++
		variants, installment, shippingSupport, err = r.getProductExternalDetails(*c,  nil)
		if err == nil {
			c.Variants = variants
		}
		p, err := r.buildEsProductData(c, installment, shippingSupport, fields)
		if err != nil {
			failed = append(failed, b.BuildProductCrossCheckLog(c.ProductId, err))
		} else {
			sliceData[c.ProductId] = p
		}
		if count%50 == 0 {
			err := r.updateMultiProductLogs(sliceData, retry)
			if err != nil {
				var ids []uint32
				for _, d := range sliceData {
					p:=d.(*model.ProductES)
					ids = append(ids, p.ProductId)
				}
				r.buildCrossCheckBulk(ids, err)
			} else {
				r.logIncrement("es", len(sliceData))
			}
			r.crossCheckChannel <- failed
			failed = []*model.ProductCrossCheck{}
			sliceData = make(map[uint32]interface{})
		}
	}
	if len(failed) > 0 {
		r.crossCheckChannel<-failed
	}
	if len(sliceData) > 0 {
		err := r.updateMultiProductLogs(sliceData, retry)
		if err != nil {
			var ids []uint32
			for _, d := range sliceData {
				p := d.(*model.ProductES)
				ids = append(ids, p.ProductId)
			}
			r.buildCrossCheckBulk(ids, err)
		} else {
			r.logIncrement("es", len(sliceData))
		}
	}
}

func NewProductResyncCommand (esStore esstore.ESStore, mgoStore mgostore.MgoStore,
	variantStore mgostore.VariantStore, redisConn *redis.Client, logStore mgostore.LogStore,
	categoryRepo repo.CategoryServiceClient, logger log.Logger) model.ProductResynchronization {
	return &changeProductCommand{
		esStore:      esStore,
		variantStore: variantStore,
		logStore:     logStore,
		mgoStore:     mgoStore,
		categoryRepo: categoryRepo,
		logger:       logger,
		batchChannel: make(chan *model.Product, 1000),
		crossCheckChannel: make(chan []*model.ProductCrossCheck, 1000),
	}
}
