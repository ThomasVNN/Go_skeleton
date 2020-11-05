package mgostore

import (
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/sirupsen/logrus"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/model"
	"time"
)

type LogStore interface {
	Add(log *model.ProductCrossCheck) error
	UpdateBulk(products []*model.ProductCrossCheck, isRetry bool) error
	UpdateRedisStatusBulk(logs []*model.ProductCrossCheck, isRetry bool) error
	UpdateEsStatusBulk(logs []*model.ProductCrossCheck, isRetry bool) error
	Query(query bson.M, limit int, order string) ([]*model.ProductCrossCheck, error)
}

type logStoreImpl struct {
	sess *mgo.Session
}

func NewLogStore(sess *mgo.Session) LogStore {
	return &logStoreImpl{
		sess: sess,
	}
}

func (this *logStoreImpl) getCollection() (coll *mgo.Collection) {
	return this.sess.DB("").C("product_data_cross_check")
}

func (this *logStoreImpl) Add(log *model.ProductCrossCheck) error {
	c := this.getCollection()
	var updateQuery = make(bson.M)
	setData := bson.M{
		"updated_at":    log.UpdatedAt,
		"expired_at":    log.ExpiredAt,
		"is_sync_es":    log.IsSyncEs,
		"error_elastic": log.ErrorElastic,
		//Function này không phải là cron nên retry_count_es = 0
		"retry_count_es": 0,
	}
	updateQuery["$set"] = setData
	updateQuery["$setOnInsert"] = bson.M{"created_at": log.CreatedAt}
	_, err := c.Upsert(bson.M{"product_id": log.ProductId}, updateQuery)
	if err != nil {
		logrus.Error("Upsert log Failed ", err)
	}
	return err
}

func (this *logStoreImpl) UpdateBulk(logs []*model.ProductCrossCheck, isRetry bool) error {
	c := this.getCollection()
	bulk := c.Bulk()
	bulk.Unordered()
	for _, log := range logs {
		var updateQuery = make(bson.M)
		setData := bson.M{
			"updated_at":    log.UpdatedAt,
			"expired_at":    log.ExpiredAt,
			"is_sync_es":    log.IsSyncEs,
			"error_elastic": log.ErrorElastic,
		}
		if !log.IsSyncEs && isRetry {
			updateQuery["$inc"] = bson.M{"retry_count_es": 1}
		} else {
			//Function này không phải là cron nên retry_count_es = 0
			setData["retry_count_es"] = 0
		}
		updateQuery["$set"] = setData
		updateQuery["$setOnInsert"] = bson.M{"created_at": log.CreatedAt}
		bulk.Upsert(bson.M{"product_id": log.ProductId}, updateQuery)
	}
	_, err := bulk.Run()
	if err != nil {
		logrus.Error("UpdateBulk log Failed ", err)
	}
	return err
}


func (this *logStoreImpl) Query(q bson.M, limit int, order string) ([]*model.ProductCrossCheck, error) {
	c := this.getCollection()
	if order == "" {
		order = "_id"
	}
	var logs []*model.ProductCrossCheck
	err := c.Find(q).Limit(limit).Sort(order).All(&logs)
	if err != nil {
		logrus.Error("Fail to Query logs data", err)
		return nil, err
	}
	return logs, nil
}

func (this *logStoreImpl) UpdateRedisStatusBulk(logs []*model.ProductCrossCheck, isRetry bool) error {
	c := this.getCollection()

	bulk := c.Bulk()
	bulk.Unordered()
	for _, log := range logs {
		var updateQuery = make(bson.M)
		setData := bson.M{
			"updated_at":    time.Now().Unix(),
			"is_sync_redis": log.IsSyncRedis,
			"error_redis":   log.ErrorRedis,
		}
		if !log.IsSyncRedis && isRetry {
			updateQuery["$inc"] = bson.M{"retry_count_redis": 1}
		} else {
			//Function này không phải là cron nên retry_count = 0
			setData["retry_count_redis"] = 1
		}
		updateQuery["$set"] = setData
		updateQuery["$setOnInsert"] = bson.M{"created_at": log.CreatedAt}
		bulk.Update(bson.M{"product_id": log.ProductId}, updateQuery)
	}
	_, err := bulk.Run()
	if err != nil {
		logrus.Error("UpdateRedisStatusBulk log Failed ", err)
	}
	return err
}

func (this *logStoreImpl) UpdateEsStatusBulk(logs []*model.ProductCrossCheck, isRetry bool) error {
	c := this.getCollection()

	bulk := c.Bulk()
	bulk.Unordered()
	for _, log := range logs {
		var updateQuery = make(bson.M)
		setData := bson.M{
			"updated_at":    time.Now().Unix(),
			"is_sync_es":    log.IsSyncEs,
			"error_elastic": log.ErrorElastic,
		}
		if !log.IsSyncEs && isRetry {
			updateQuery["$inc"] = bson.M{"retry_count_es": 1}
		} else {
			//Function này không phải là cron nên retry_count = 0
			setData["retry_count_es"] = 1
		}
		updateQuery["$set"] = setData
		updateQuery["$setOnInsert"] = bson.M{"created_at": log.CreatedAt}
		bulk.UpdateAll(bson.M{"product_id": log.ProductId}, updateQuery)
	}
	_, err := bulk.Run()
	if err != nil {
		logrus.Error("UpdateEsStatusBulk log Failed ", err)
	}
	return err
}
