package mgostore

import (
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/model"
)

type VariantStore interface {
	Get(productId uint32) ([]*model.Variant, error)
	Gets(productIds []uint32) ([]*model.Variant, error)
}

type variantStoreImpl struct {
	sess *mgo.Session
}

func NewVariantStore(sess *mgo.Session) VariantStore {
	return &variantStoreImpl{
		sess: sess,
	}
}

func (v *variantStoreImpl) getCollection() (coll *mgo.Collection) {
	return v.sess.DB("services").C("product_variant")
}

func (v *variantStoreImpl) Get(productId uint32) (variants []*model.Variant, err error) {
	c := v.getCollection()
	var pVariant model.ProductVariant
	err = c.Find(bson.M{"product_id": productId}).One(&pVariant)
	if err != nil {
		return variants, err
	}

	variants = append(variants, pVariant.Variants...)
	return variants, err
}

func (v *variantStoreImpl) Gets(productIds []uint32) ([]*model.Variant, error) {
	c := v.getCollection()
	var variants []*model.Variant
	err := c.Find(bson.M{
		"product_id": bson.M{"$in": productIds},
	}).All(&variants)

	if err != nil && err != mgo.ErrNotFound {
		/*devlog.Error("Get multi variant from database failed", map[string]interface{}{
		"err": err.Error(),
		"productIds": productIds,
		}, "Gets")*/
	}
	return variants, nil
}
