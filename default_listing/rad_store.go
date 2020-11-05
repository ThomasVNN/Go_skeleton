package default_listing

import (
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/sirupsen/logrus"
	"gitlab.thovnn.vn/core/sen-kit/storage/mongo"
	"sync"
)
const (
	CollectionDefaultFeaturesByCategoryName          = "features_default_by_category"
	CollectionListingV3DefaultFeaturesByCategoryName = "listing_v3_features_default_by_category"
	CollectionDefaultFeatures                        = "features_default"
	CollectionListingV3DefaultFeatures               = "listing_v3_features_default"
	CollectionProductFeatures                        = "product_features"
	CollectionListingV3ProductFeatures               = "listing_v3_product_features"
)

var logger = logrus.New().WithField("prefix", "es_worker")
type RadStore interface {
 	//DB () *mgo.Database
 	//C(name string) *mgo.Collection
	GetProductFeaturesByRange (min, max uint32, collectionType string) ([]ProductFeatures, error)
	GetProductFeatures (collectionType string, ids... uint32) ([]ProductFeatures, error)
}

type radStore struct {
	m sync.Mutex
	collection string
	db *mgo.Database
}

func NewRadStore() RadStore {
	var rad radStore
	rad.Connect("ListingML")
	return &rad
}

func (r *radStore) Connect (name string) {
	mgoConfig := mongo.New("Rad")
	mgoSess, err := mgoConfig.DB()
	if err != nil {
		logger.Println("not connect to mongoDB ", mgoConfig.String())
		return
	}
	r.db = mgoSess.DB(name)
}

func (r *radStore) C (name string) *mgo.Collection {
	return r.db.C(name)
}

func (r *radStore) DB () *mgo.Database {
	return r.db
}

func (r *radStore) GetProductFeaturesByRange (min, max uint32, collectionType string) ([]ProductFeatures, error){
	var collection = CollectionProductFeatures
	var features []ProductFeatures
	if collectionType == "v3" {
		collection = CollectionListingV3ProductFeatures
	}

	var query bson.M
	query = bson.M{
		"product_id" : bson.M{
			"$gte": min,
			"$lte": max,
		},
	}

	err := r.C(collection).Find(query).All(&features)
	if err != nil {
		return nil, err
	}
	return features, nil
}

func (r *radStore) GetProductFeatures (collectionType string, ids... uint32) ([]ProductFeatures, error){
	var collection = CollectionProductFeatures
	var features []ProductFeatures
	if collectionType == "v3" {
		collection = CollectionListingV3ProductFeatures
	}

	var query bson.M
	query = bson.M{
		"product_id" : bson.M{
			"$in": ids,
		},
	}

	err := r.C(collection).Find(query).All(&features)
	if err != nil {
		return nil, err
	}
	return features, nil
}
