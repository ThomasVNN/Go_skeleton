package mgostore

import (
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/sirupsen/logrus"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/model"
)

type ShopStore interface {
	GetMerchantByExternalId(sellerAdminId int32) (*model.Merchant, error)
	GetCertifiedByMerchantID(merchantID int32) (*model.ShopCertified, error)
	GetListShippingNewByStoreID(storeID int32) (newShip *model.ShippingNew, err error)
	GetAppLevelByStoreID(storeID int32) (appLevel *model.ShopSupportPromotionApp, err error)
	GetSenPointExternalID(int32) (*model.ShopSenPoint, error)
}

type shopStoreImpl struct {
	sess *mgo.Session
}

func NewShopStore(sess *mgo.Session) ShopStore {
	return &shopStoreImpl{
		sess: sess,
	}
}

func (s *shopStoreImpl) getCollection() (coll *mgo.Collection) {
	return s.sess.DB("services").C("merchant_flat")
}

func (s *shopStoreImpl) getCollectionShippingFee() (coll *mgo.Collection) {
	return s.sess.DB("services").C("shipping_new")
}

func (s *shopStoreImpl) getCollectionMerchantInfo() (coll *mgo.Collection) {
	return s.sess.DB("services").C("merchant_info")
}

func (s *shopStoreImpl) getCollectionShippingPaymentConfig() (coll *mgo.Collection) {
	return s.sess.DB("services").C("shipping_payment_config")
}

func (s *shopStoreImpl) GetCertifiedByMerchantID(sellerAdminId int32) (merchant *model.ShopCertified, err error) {
	c := s.getCollection()
	err = c.Find(bson.M{"external_id": sellerAdminId}).One(&merchant)
	if err != nil && err != mgo.ErrNotFound {
		logrus.Error("cannot find is_certified in merchant_info table ", err)
	}
	return merchant, err
}

func (s *shopStoreImpl) GetMerchantByExternalId(sellerAdminId int32) (merchant *model.Merchant, err error) {
	c := s.getCollection()
	err = c.Find(bson.M{"external_id": sellerAdminId}).One(&merchant)
	if err != nil && err != mgo.ErrNotFound {
		logrus.Error("cannot find external_id in merchant_flat table ", err)
	}
	return merchant, err
}

func (s *shopStoreImpl) GetListShippingNewByStoreID(storeID int32) (newShip *model.ShippingNew, err error) {
	newCollection := s.getCollectionShippingFee()
	err = newCollection.Find(bson.M{"storeid": storeID}).One(&newShip)
	if err != nil && err != mgo.ErrNotFound {
		logrus.Info("record not found")
	}
	return newShip, err
}

func (s *shopStoreImpl) GetAppLevelByStoreID(storeID int32) (appLevel *model.ShopSupportPromotionApp, err error) {
	newCollection := s.getCollectionShippingPaymentConfig()
	err = newCollection.Find(bson.M{"storeid": storeID}).One(&appLevel)
	if err != nil && err != mgo.ErrNotFound {
		logrus.Info("record not found")
	}
	return appLevel, err
}

func (s *shopStoreImpl) GetSenPointExternalID(sellerAdminID int32) (p *model.ShopSenPoint, err error) {
	newCollection := s.getCollection()
	err = newCollection.Find(bson.M{"external_id": sellerAdminID}).One(&p)
	if err != nil && err != mgo.ErrNotFound {
		logrus.Info("record not found")
	}
	return p, err
}
