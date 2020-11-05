package mgostore

import (
	"errors"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/model"
)

type MgoStore interface {
	ListProduct(ids []int64)
	GetProductById(productId uint32) (*model.Product, error)
	GetMultiProducts(productIds []uint32) ([]*model.Product, error)
}

type mgoStoreImpl struct {
	sess *mgo.Session
}

func New(sess *mgo.Session) MgoStore {
	return &mgoStoreImpl{
		sess: sess,
	}
}

func (s *mgoStoreImpl) ListProduct(ids []int64) {

}

func (s *mgoStoreImpl) getCollection() (coll *mgo.Collection) {
	return s.sess.DB("").C("product_flat")
}

func (s *mgoStoreImpl) GetProductById(productId uint32) (product *model.Product, err error) {
	if productId == 0 {
		return nil, errors.New("not found")
	}
	c := s.getCollection()
	err = c.Find(bson.M{"product_id": productId}).One(&product)
	if err != nil && err != mgo.ErrNotFound {
		/*devlog.Error("Get product from product_flat Failed", map[string]interface{}{
			"err": err.Error(),
			"productId": productId,
		}, "GetProductById")*/
	}
	return product, err
}

func (s *mgoStoreImpl) GetMultiProducts(productIds []uint32) ([]*model.Product, error) {
	c := s.getCollection()
	var products []*model.Product
	err := c.Find(bson.M{
		"product_id": bson.M{"$in": productIds},
	}).All(&products)
	if err != nil && err != mgo.ErrNotFound {
		/*devlog.Error("Get multi product from database failed", map[string]interface{}{
			"err": err.Error(),
			"productIds": productIds,
		}, "GetMultiProducts")*/
	}
	return products, err
}
