package default_listing

import (
	"encoding/json"
	"fmt"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/model"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/storage/esstore"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/storage/redisstore"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/util"
	"math"
	"strconv"
	"sync"
	"time"
)

const DL_PRODUCT_FEATURES_PATTERN = "product-features/%d?collection=%s"

var wg sync.WaitGroup

type DefaultListingService interface {
	ProcessCronProductsData(products []model.Product)
	ProcessProductsData(products []model.Product)
	CalculateProductScores(product model.Product) (model.DefaultListingScoreEs, model.ListingScore)
}

type defaultListingService struct {
	featureV2Url string
	featureV3Url string
	redis        redisstore.RedisStore
	radMgo       RadStore
	db           *mgo.Session
	esStore      esstore.ESStore
}

func NewDefaultListingService(redis redisstore.RedisStore, esstore esstore.ESStore) DefaultListingService {
	radStore := NewRadStore()
	return &defaultListingService{redis: redis, radMgo: radStore, esStore: esstore}
}

func (dl *defaultListingService) ProcessCronProductsData(products []model.Product) {
	var esProducts []map[string]interface{}
	var esProductIds []string

	minId, maxId := dl.getProductRange(products)
	productFeaturesV2 := dl.GetProductFeaturesByRange(minId, maxId, "v2")
	productFeaturesV3 := dl.GetProductFeaturesByRange(minId, maxId, "v3")
	defaultConfig := dl.getGlobalConfig()
	histogram := dl.getHistogramMap()

	for _, p := range products {
		var product = make(map[string]interface{})
		cateId := dl.getCategoryIDByPath(p.CategoryId)
		featuresV2 := productFeaturesV2[p.ProductId]
		featuresV3 := productFeaturesV3[p.ProductId]
		listingMlScore := dl.calculateListingScore(p, defaultConfig, featuresV2, featuresV3)
		defaultListingScore := dl.calculateDefaultScore(p, defaultConfig, histogram[cateId])
		product["listing_score"] = listingMlScore
		product["default_listing_score"] = defaultListingScore
		esProducts = append(esProducts, product)
		esProductIds = append(esProductIds, fmt.Sprint(p.ProductId))
		if len(esProductIds)%1000 == 0 {
			_, _ = dl.esStore.Updates(esProducts, esProductIds)
			esProducts = []map[string]interface{}{}
			esProductIds = []string{}
		}
	}
	_, _ = dl.esStore.Updates(esProducts, esProductIds)
}

func (dl *defaultListingService) ProcessProductsData(products []model.Product) {
	var productIds []uint32
	var esProducts []map[string]interface{}
	var esProductIds []string

	for _, p:= range products {
		productIds = append(productIds, p.ProductId)
	}

	productFeaturesV2 := dl.GetProductFeatures("v2", productIds...)
	productFeaturesV3 := dl.GetProductFeatures("v3", productIds...)

	defaultConfig := dl.getGlobalConfig()
	histogram := dl.getHistogramMap()


	for _, p := range products {
		var product = make(map[string]interface{})
		cateId := dl.getCategoryIDByPath(p.CategoryId)
		featuresV2 := productFeaturesV2[p.ProductId]
		featuresV3 := productFeaturesV3[p.ProductId]
		listingMlScore := dl.calculateListingScore(p, defaultConfig, featuresV2, featuresV3)
		defaultListingScore := dl.calculateDefaultScore(p, defaultConfig, histogram[cateId])
		product["listing_score"] = listingMlScore
		product["default_listing_score"] = defaultListingScore
		esProducts = append(esProducts, product)
		esProductIds = append(esProductIds, fmt.Sprint(p.ProductId))
		if len(esProductIds)%1000 == 0 {
			_, _ = dl.esStore.Updates(esProducts, esProductIds)
			esProducts = []map[string]interface{}{}
			esProductIds = []string{}
		}
	}
	_, _ = dl.esStore.Updates(esProducts, esProductIds)
}

func (dl *defaultListingService) CalculateProductScores(p model.Product) (model.DefaultListingScoreEs, model.ListingScore) {
	productFeaturesV2 := dl.GetProductFeatures("v2", p.ProductId)
	productFeaturesV3 := dl.GetProductFeatures("v3", p.ProductId)
	defaultConfig := dl.getGlobalConfig()
	histogram := dl.getHistogramMap()
	cateId := dl.getCategoryIDByPath(p.CategoryId)
	featuresV2 := productFeaturesV2[p.ProductId]
	featuresV3 := productFeaturesV3[p.ProductId]
	listingMlScore := dl.calculateListingScore(p, defaultConfig, featuresV2, featuresV3)
	defaultListingScore := dl.calculateDefaultScore(p, defaultConfig, histogram[cateId])

	return defaultListingScore, listingMlScore
}

func (dl *defaultListingService) GetProductFeaturesByRange(minId, maxId uint32, ver string) map[uint32]Features {
	var featureMap = make(map[uint32]Features)
	features, _ := dl.radMgo.GetProductFeaturesByRange(minId, maxId, ver)
	for _, f := range features {
		featureMap[f.ProductId] = f.Data
	}
	return featureMap
}

func (dl *defaultListingService) GetProductFeatures(ver string, productIds... uint32) map[uint32]Features {
	var featureMap = make(map[uint32]Features)

	features, _ := dl.radMgo.GetProductFeatures(ver, productIds...)
	for _, f := range features {
		featureMap[f.ProductId] = f.Data
	}
	return featureMap
}

func calculateRawListingScore(config ListingML, features Features) float64 {
	return (config.ClickWeight * features.Click) +
		(config.OrderWithQuantityWeight * features.OrderWithQuantity) +
		(config.RevenueWeight * features.Revenue) +
		(config.Position1CountWeight * features.Position1Count) +
		(config.Position24CountWeight * features.Position24Count) +
		(config.Position560CountWeight * features.Potition560Count) +
		(config.CtrWeight * features.Ctr) +
		(config.CRWeight * features.Ctr) +
		(config.ScoreCate2Weight * features.ScoreCate2) +
		(config.ScoreCate3Weight * features.ScoreCate3) +
		(config.OrderCompleted15BoostCate2weight * features.OrderCompleted15BoostCate2) +
		(config.OrderCompleted15BoostCate3weight * features.OrderCompleted15BoostCate3) +
		(config.ShippingFeeBoostWeight * features.ShippingFeeBoost) +
		(config.RatingBoostWeight * features.RatingBoost) +
		(config.DescriptionScoreBoostWeight * features.DescriptionScoreBoost) +
		(config.DiscountBoostWeight * features.DiscountBoost) +
		(config.DiscountAppBoostWeight * features.DiscountAppBoost) +
		(config.OrderCompletionRateBoostWeight * features.OrderCompletionRateBoost)
}

func (dl *defaultListingService) getProductFeatures(productId uint32, categoryId, typeFunc string, listingType string) ProductFeatures {
	//Listing ML init values
	productFeatures := ProductFeatures{}
	//realtime
	//-get product features
	productFeaturesURL := fmt.Sprintf(DL_PRODUCT_FEATURES_PATTERN, int64(productId), listingType)
	productResponse, _ := util.SendHttpRequest("get", productFeaturesURL, nil)

	_ = json.Unmarshal([]byte(productResponse), &productFeatures)

	return productFeatures
}

func (dl *defaultListingService) getListingScore(productInfo model.Product, defaultConfig DefaultListingConfig, mlvConfig ListingML, features Features) (float64, time.Time) {
	defaultListingScore := productInfo.DefaultListingScore
	//Brand_boost
	var brandBoost float64 = 1
	if productInfo.BrandId > 0 {
		brandBoost = mlvConfig.MaxBrandBoostV2
	}

	//Shop_hoa_sen_boost
	var shopHoaSenBoost float64 = 1
	if defaultListingScore.ShopHoaSenBoost == 1 {
		shopHoaSenBoost = mlvConfig.MaxShopHoaSenBoostV2
	}
	//Shop_tich_cuc_boost
	var shopTichCucBoost float64 = 1
	if defaultListingScore.ShopTichCucBoost == 1 {
		shopTichCucBoost = mlvConfig.MaxShopTichCucBoostV2
	}
	//Shop_mall_boost
	shopMallBoost := dl.getListingMlShopMallBoosts(productInfo.DefaultListingScore, defaultConfig, mlvConfig)

	rawScore := calculateRawListingScore(mlvConfig, features)

	//Listing ML
	scoreV2 := (1.0 / (1.0 + math.Exp(-rawScore))) *
		brandBoost * shopHoaSenBoost * shopTichCucBoost * shopMallBoost *
		math.Pow10(9)
	scoreV2 = dl.toFixed(scoreV2, 0)
	return scoreV2, time.Now()
}

func (dl *defaultListingService) calculateListingScore(productInfo model.Product, defaultConfig DefaultListingConfig, featuresV2, featuresV3 Features) model.ListingScore {
	var listingScore model.ListingScore
	var scoreV2, scoreV3 float64
	var timeV2, timeV3 time.Time
	wg.Add(2)
	go func() {
		scoreV2, timeV2 = dl.getListingScore(productInfo, defaultConfig, defaultConfig.ListingMLV2, featuresV2)
		wg.Done()
	}()
	go func() {
		scoreV3, timeV3 = dl.getListingScore(productInfo, defaultConfig, defaultConfig.ListingMLV3, featuresV3)
		wg.Done()
	}()
	wg.Wait()
	listingScore.ScoreV2 = scoreV2
	listingScore.ScoreV3 = scoreV3
	listingScore.UpdatedAtV2 = timeV2
	listingScore.UpdatedAtV3 = timeV3
	return listingScore
}

func (dl *defaultListingService) calculateDefaultScore(productInfo model.Product, defaultConfig DefaultListingConfig, histogram Histogram) model.DefaultListingScoreEs {
	defaultListingScore := productInfo.DefaultListingScore
	//Brand_boost
	var brandBoost float64 = 1
	if productInfo.BrandId > 0 {
		brandBoost = defaultConfig.MaxBrandBoost
	}

	//Shop_hoa_sen_boost
	var shopHoaSenBoost float64 = 1
	if defaultListingScore.ShopHoaSenBoost == 1 {
		shopHoaSenBoost = defaultConfig.MaxShopHoaSenBoost
	}
	//Shop_tich_cuc_boost
	var shopTichCucBoost float64 = 1
	if defaultListingScore.ShopTichCucBoost == 1 {
		shopTichCucBoost = defaultConfig.MaxShopTichCucBoost
	}
	//Shop_mall_boost
	shopMallBoost := dl.getShopMallBoosts(productInfo.DefaultListingScore, defaultConfig)
	orderCompleted15BoostCate2, orderCompleted15BoostCate3 := dl.getOrder15BoostCate(productInfo, defaultConfig, histogram)
	orderCompletionRateBoost := dl.getOrderCompletionRateBoost(productInfo, defaultConfig)
	shippingFeeBoost := dl.getShippingFeeBoost(productInfo, defaultConfig)
	ratingBoost := dl.getRatingBoost(productInfo, defaultConfig)
	descriptionScoreBoost := dl.getDescriptionScoreBoost(productInfo, defaultConfig)
	discountBoost := dl.getDiscountBoost(productInfo, defaultConfig)
	discountAppBoost := dl.getDiscountAppBoost(productInfo, defaultConfig)

	var resultScore model.DefaultListingScoreEs
	resultScore.OrderCompleted15BoostCate2 = orderCompleted15BoostCate2
	resultScore.OrderCompleted15BoostCate3 = orderCompleted15BoostCate3
	resultScore.OrderCompletionRateBoost = orderCompletionRateBoost
	resultScore.BrandBoost = brandBoost
	resultScore.ShopHoaSenBoost = shopHoaSenBoost
	resultScore.ShopTichCucBoost = shopTichCucBoost
	resultScore.ShopMallBoost = shopMallBoost
	resultScore.ShippingFeeBoost = shippingFeeBoost
	resultScore.RatingBoost = ratingBoost
	resultScore.DescriptionScoreBoost = descriptionScoreBoost
	resultScore.DiscountBoost = discountBoost
	resultScore.DiscountAppBoost = discountAppBoost
	resultScore.CalculateCate2Score()
	resultScore.CalculateCate3Score()
	resultScore.UpdatedAt = time.Now()

	return resultScore
}

func (dl *defaultListingService) getGlobalConfig() DefaultListingConfig {
	var dataScore DefaultListingConfig
	var scoreConfig GlobalConfig
	err := json.Unmarshal([]byte(scoreConfig.Value), &dataScore)
	if err != nil {
		return DefaultListingConfig{}
	}
	return DefaultListingConfig{}
}

func (dl *defaultListingService) getListingMlShopMallBoosts(defaultListingScore model.DefaultListingScore, scoreConfig DefaultListingConfig, mlConfig ListingML) float64 {
	var shopMallBoost float64 = 1
	if defaultListingScore.ShopMallBoost == 1 {
		if defaultListingScore.OpsScore == scoreConfig.MaxShopMallOps {
			shopMallBoost = mlConfig.MaxShopMallBoostV2
		} else if defaultListingScore.OpsScore > 0 &&
			defaultListingScore.OpsScore < scoreConfig.MaxShopMallOps {
			shopMallBoost = mlConfig.MinShopMallBoostV2 + defaultListingScore.OpsScore/scoreConfig.MaxShopMallOps*(mlConfig.MaxShopMallBoostV2-mlConfig.MinShopMallBoostV2)
			shopMallBoost, _ = strconv.ParseFloat(fmt.Sprintf("%.3f", shopMallBoost), 64)
		} else {
			shopMallBoost = mlConfig.MinShopMallBoostV2
		}
	}
	return shopMallBoost
}

func (dl *defaultListingService) getOrder15BoostCate(productInfo model.Product, dataScore DefaultListingConfig, histogram Histogram) (float64, float64) {
	var orderCompleted15BoostCate2 float64 = 1
	var orderCompleted15BoostCate3 float64 = 1
	var defaultListingScore = productInfo.DefaultListingScore
	var cutOffCate2Float = float64(histogram.CutOff2)
	var cutOffCate3Float = float64(histogram.CutOff3)

	if defaultListingScore.OrderComplete15 > 0 {
		if float64(defaultListingScore.OrderComplete15) < cutOffCate2Float {
			orderCompleted15BoostCate2 = dataScore.MinOrderCompleted15BoostCate2 + (float64(defaultListingScore.OrderComplete15)-1)/(cutOffCate2Float-1)*(dataScore.MaxOrderCompleted15BoostCate2-dataScore.MinOrderCompleted15BoostCate2)
		} else {
			orderCompleted15BoostCate2 = dataScore.MaxOrderCompleted15BoostCate2
		}

		if float64(defaultListingScore.OrderComplete15) < cutOffCate3Float {
			orderCompleted15BoostCate3 = dataScore.MinOrderCompleted15BoostCate3 + (float64(defaultListingScore.OrderComplete15)-1)/(cutOffCate3Float-1)*(dataScore.MaxOrderCompleted15BoostCate3-dataScore.MinOrderCompleted15BoostCate3)
		} else {
			orderCompleted15BoostCate3 = dataScore.MaxOrderCompleted15BoostCate3
		}
	}

	return orderCompleted15BoostCate2, orderCompleted15BoostCate3
}

func (dl *defaultListingService) getOrderCompletionRateBoost(productInfo model.Product, dataScore DefaultListingConfig) float64 {
	var defaultListingScore = productInfo.DefaultListingScore
	var orderCompletionRateBoost = dataScore.OrderCompletionRateBoostLevel2
	if defaultListingScore.Order90 >= 5 && defaultListingScore.OrderComplete90 > 0 && defaultListingScore.OrderCancelledByShop90 > 0 {
		var orderCompletionRateRaw float64
		orderCompletionRateRaw = float64(float64(defaultListingScore.OrderComplete90) / float64(defaultListingScore.OrderComplete90+defaultListingScore.OrderCancelledByShop90))
		if orderCompletionRateRaw >= 0 && orderCompletionRateRaw <= 0.2 {
			orderCompletionRateBoost = dataScore.OrderCompletionRateBoostLevel1
		} else if orderCompletionRateRaw < 0.5 {
			orderCompletionRateBoost = dataScore.OrderCompletionRateBoostLevel2
		} else if orderCompletionRateRaw < 0.75 {
			orderCompletionRateBoost = dataScore.OrderCompletionRateBoostLevel3
		} else {
			orderCompletionRateBoost = 1
		}
	}
	return orderCompletionRateBoost
}

func (dl *defaultListingService) getShippingFeeBoost(productInfo model.Product, dataScore DefaultListingConfig) float64 {
	var defaultListingScore = productInfo.DefaultListingScore
	var shippingFeeBoostRatio float64 = 1
	var shippingFeeBoostLevel float64

	if defaultListingScore.Price < float64(defaultListingScore.MinPriceForShippingFee) {
		shippingFeeBoostRatio = float64(defaultListingScore.Price/float64(defaultListingScore.MinPriceForShippingFee)) * dataScore.MaxShippingFeeBoostRatio
	}

	if defaultListingScore.PromotionShippingFee > 0 && defaultListingScore.PromotionShippingFee < 10000 {
		shippingFeeBoostLevel = dataScore.ShippingFeeBoostLevel1-1
	} else if defaultListingScore.PromotionShippingFee < 20000 {
		shippingFeeBoostLevel = dataScore.ShippingFeeBoostLevel2-1
	} else if defaultListingScore.PromotionShippingFee < 30000 {
		shippingFeeBoostLevel = dataScore.ShippingFeeBoostLevel3-1
	} else {
		shippingFeeBoostLevel = dataScore.ShippingFeeBoostLevel4-1
	}

	return 1 + shippingFeeBoostLevel*shippingFeeBoostRatio
}
func (dl *defaultListingService) getRatingBoost(productInfo model.Product, dataScore DefaultListingConfig) float64 {
	var ratingBoost float64 = 1
	var defaultListingScore = productInfo.DefaultListingScore
	if float64(defaultListingScore.TotalRating) >= dataScore.RatingBoostMinNumRatings {
		if defaultListingScore.RatingPercent >= 4.5 && defaultListingScore.RatingPercent <= 5 {
			ratingBoost = dataScore.RatingBoostLevel5
		} else if defaultListingScore.RatingPercent >= 3.5 {
			ratingBoost = dataScore.RatingBoostLevel4
		} else if defaultListingScore.RatingPercent >= 2.5 {
			ratingBoost = 1
		} else if defaultListingScore.RatingPercent >= 1.5 {
			ratingBoost = dataScore.RatingBoostLevel2
		} else if defaultListingScore.RatingPercent >= 1 {
			ratingBoost = dataScore.RatingBoostLevel1
		}
	}
	return ratingBoost
}

func (dl *defaultListingService) getDescriptionScoreBoost(productInfo model.Product, dataScore DefaultListingConfig) float64 {
	//Description_Score_Boost
	defaultListingScore := productInfo.DefaultListingScore
	var descriptionScoreBoost = dataScore.DescriptionScoreBoostDefault

	if defaultListingScore.DescriptionScore >= 1 && defaultListingScore.DescriptionScore < 20 {
		descriptionScoreBoost = 1
	} else if defaultListingScore.DescriptionScore < 40 {
		descriptionScoreBoost = dataScore.DescriptionScoreBoostLevel2
	} else if defaultListingScore.DescriptionScore < 60 {
		descriptionScoreBoost = dataScore.DescriptionScoreBoostLevel3
	} else if defaultListingScore.DescriptionScore < 80 {
		descriptionScoreBoost = dataScore.DescriptionScoreBoostLevel4
	} else if defaultListingScore.DescriptionScore <= 100 {
		descriptionScoreBoost = dataScore.DescriptionScoreBoostLevel5
	}

	return descriptionScoreBoost
}

func (dl *defaultListingService) getDiscountBoost(productInfo model.Product, dataScore DefaultListingConfig) float64 {
	//Discount_Boost
	defaultListingScore := productInfo.DefaultListingScore
	var discountBoost float64 = 1
	var realDiscount float64
	if defaultListingScore.OrderComplete30 == 0 || defaultListingScore.Price >= defaultListingScore.AvgPriceCompleted30 {
		realDiscount = 0
	} else if defaultListingScore.Price < defaultListingScore.AvgPriceCompleted30 {
		realDiscount = defaultListingScore.AvgPriceCompleted30 - defaultListingScore.Price/defaultListingScore.AvgPriceCompleted30
	}

	if realDiscount > 0.3 {
		discountBoost = dataScore.DiscountBoostLevel4
	} else if realDiscount > 0.2 && realDiscount <= 0.3 {
		discountBoost = dataScore.DiscountBoostLevel3
	} else if realDiscount >= 0.1 && realDiscount <= 0.2 {
		discountBoost = dataScore.DiscountBoostLevel2
	} else if realDiscount > 0 && realDiscount < 0.1 {
		discountBoost = dataScore.DiscountBoostLevel1

	}

	return discountBoost
}

func (dl *defaultListingService) getDiscountAppBoost(productInfo model.Product, dataScore DefaultListingConfig) float64 {
	//Discount_Boost
	defaultListingScore := productInfo.DefaultListingScore
	//Discount_App_Boost
	var discountAppBoost float64
	if defaultListingScore.DiscountApp == 10 {
		discountAppBoost = dataScore.DiscountAppBoostLevel3
	} else if defaultListingScore.DiscountApp == 5 {
		discountAppBoost = dataScore.DiscountAppBoostLevel2
	} else if defaultListingScore.DiscountApp == 2 {
		discountAppBoost = dataScore.DiscountAppBoostLevel1
	} else {
		discountAppBoost = 1
	}

	return discountAppBoost
}

func (dl *defaultListingService) getShopMallBoosts(defaultListingScore model.DefaultListingScore, scoreConfig DefaultListingConfig) float64 {
	var shopMallBoost float64 = 1
	if defaultListingScore.ShopMallBoost == 1 {
		if defaultListingScore.OpsScore == scoreConfig.MaxShopMallOps {
			shopMallBoost = scoreConfig.MaxShopMallBoost
		} else if defaultListingScore.OpsScore > 0 &&
			defaultListingScore.OpsScore < scoreConfig.MaxShopMallOps {
			shopMallBoost = scoreConfig.MinShopMallBoost + defaultListingScore.OpsScore/scoreConfig.MaxShopMallOps*(scoreConfig.MaxShopMallBoost-scoreConfig.MinShopMallBoost)
			shopMallBoost, _ = strconv.ParseFloat(fmt.Sprintf("%.3f", shopMallBoost), 64)
		} else {
			shopMallBoost = scoreConfig.MinShopMallBoost
		}
	}
	return shopMallBoost
}

func (dl *defaultListingService) getHistogramMap() map[string]Histogram {
	histogramMap := make(map[string]Histogram)
	var dataHistogram []Histogram
	var err error
	var page = 1
	var limit = 500
	var skip = 0
	for {
		if page < 2 {
			skip = 0
		} else {
			skip = (page - 1) * limit
		}

		err = dl.db.DB("").C("histogram_flat").Find(bson.M{}).Sort("last_update").Skip(skip).Limit(limit).All(&dataHistogram)
		if err != nil {
			fmt.Println("Error: data histogram: ", err)
			break
		}
		if len(dataHistogram) < 1 {
			break
		}

		for _, val := range dataHistogram {
			histogramMap[val.CatePath3] = val

		}
		page++
	}

	return histogramMap
}
