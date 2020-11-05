package default_listing

import "reflect"

type Product struct {
	ProductId           uint32              `json:"product_id" bson:"product_id" redis:"product_id"`
	ExternalId          uint32              `json:"external_id" bson:"external_id" redis:"external_id"`
	AdminId             uint32              `json:"admin_id" bson:"admin_id" redis:"admin_id"`
	CategoryId          string              `json:"category_id" bson:"category_id" redis:"category_id"`
	StatusNew           uint32              `json:"status_new" bson:"status_new" redis:"status_new"`
	StockStatus         uint32              `json:"stock_status" bson:"stock_status" redis:"stock_status"`
	UpdatedAt           string              `json:"updated_at" bson:"updated_at" redis:"updated_at"`
	BrandId             uint32              `json:"brand_id" bson:"brand_id" redis:"brand_id"`
	ShopType            uint32              `json:"shop_type" bson:"shop_type" redis:"shop_type"`
	ShopLevel           uint32              `json:"shop_level" bson:"shop_level" redis:"shop_level"`
	DefaultListingScore DefaultListingScore `json:"default_listing_score" bson:"default_listing_score" redis:"default_listing_score"`
	Price               float64             `json:"price" redis:"price" bson:"price"`
	DescriptionScore    float32             `json:"description_score" bson:"description_score" redis:"description_score"`
}

type EsProduct struct {

}

type Features struct {
	Click                      float64 `json:"click"`
	OrderWithQuantity          float64 `json:"order_with_quantity"`
	Revenue                    float64 `json:"revenue"`
	Position1Count             float64 `json:"position_1_count"`
	Position24Count            float64 `json:"position_2_4_count"`
	Potition560Count           float64 `json:"potition_5_60_count"`
	Ctr                        float64 `json:"ctr"`
	Cr                         float64 `json:"cr"`
	ScoreCate2                 float64 `json:"score_cate2"`
	ScoreCate3                 float64 `json:"score_cate3"`
	OrderCompletionRateBoost   float64 `json:"order_completion_rate_boost"`
	DiscountAppBoost           float64 `json:"discount_app_boost"`
	ShippingFeeBoost           float64 `json:"shipping_fee_boost"`
	RatingBoost                float64 `json:"rating_boost"`
	DiscountBoost              float64 `json:"discount_boost"`
	OrderCompleted15BoostCate2 float64 `json:"order_completed_15_boost_cate_2"`
	OrderCompleted15BoostCate3 float64 `json:"order_completed_15_boost_cate_3"`
	DescriptionScoreBoost      float64 `json:"description_score_boost"`
}

type DefaultListingScore struct {
	CutOffCate2            float64 `json:"cut_off_cate_2" redis:"cut_off_cate_2" bson:"cut_off_cate_2"`
	CutOffCate3            float64 `json:"cut_off_cate_3" redis:"cut_off_cate_3" bson:"cut_off_cate_3"`
	OrderComplete15        int     `json:"order_complete_15" redis:"order_complete_15" bson:"order_complete_15"`
	OrderComplete90        int     `json:"order_complete_90" redis:"order_complete_90" bson:"order_complete_90"`
	OrderComplete30        int     `json:"order_complete_30" redis:"order_complete_30" bson:"order_complete_30"`
	Order90                int     `json:"order_90" redis:"order_90" bson:"order_90"`
	TotalRating            int     `json:"total_rating" redis:"total_rating" bson:"total_rating"`
	PromotionShippingFee   int     `json:"promotion_shipping_fee" redis:"promotion_shipping_fee" bson:"promotion_shipping_fee"`
	RatingPercent          float64 `json:"rating_percent" redis:"rating_percent" bson:"rating_percent"`
	DescriptionScore       float64 `json:"description_score" redis:"description_score" bson:"description_score"`
	AvgPriceCompleted30    float64 `json:"avg_price_completed_30" redis:"avg_price_completed_30" bson:"avg_price_completed_30"`
	TotalPriceCompleted30  float64 `json:"total_price_completed_30" redis:"total_price_completed_30" bson:"total_price_completed_30"`
	TotalOrderCompleted30  float64 `json:"total_order_completed_30" redis:"total_order_completed_30" bson:"total_order_completed_30"`
	OrderCancelledByShop90 int     `json:"order_cancelled_by_shop_90" redis:"order_cancelled_by_shop_90" bson:"order_cancelled_by_shop_90"`
	Price                  float64 `json:"price" redis:"price" bson:"price"`
	DiscountApp            float64 `json:"discount_app" redis:"discount_app" bson:"discount_app"`
	IsUpdated              int     `json:"is_updated" redis:"is_updated" bson:"is_updated"`
	MinPriceForShippingFee int     `json:"min_price_for_shipping_fee" redis:"min_price_for_shipping_fee" bson:"min_price_for_shipping_fee"`
	BrandBoost             float64 `json:"brand_boost" redis:"brand_boost" bson:"brand_boost"` //Field ko còn sử dụng, sử dụng BrandId nằm trong struct Product (struct cha của struct này)
	ShopHoaSenBoost        float64 `json:"shop_hoa_sen_boost" redis:"shop_hoa_sen_boost" bson:"shop_hoa_sen_boost"`
	ShopTichCucBoost       float64 `json:"shop_tich_cuc_boost" redis:"shop_tich_cuc_boost" bson:"shop_tich_cuc_boost"`
	ShopMallBoost          float64 `json:"shop_mall_boost" redis:"shop_mall_boost" bson:"shop_mall_boost"`
	OpsScore               float64 `json:"ops_score" redis:"ops_score" bson:"ops_score"`
}


type CategoryFeatures struct {
	CatePath2   int64   `bson:"cate_path_2" json:"cate_path_2"`
	Ctr         float64 `bson:"ctr" json:"ctr"`
	CR          float64 `bson:"cr" json:"cr"`
	LastUpdated int64   `bson:"last_updated" json:"last_updated"`
}

type DefaultListingConfig struct {
	MinOrderCompleted15BoostCate2  float64 `json:"min_order_completed_15_boost_cate_2"`
	MaxOrderCompleted15BoostCate2  float64 `json:"max_order_completed_15_boost_cate_2"`
	MinOrderCompleted15BoostCate3  float64 `json:"min_order_completed_15_boost_cate_3"`
	MaxOrderCompleted15BoostCate3  float64 `json:"max_order_completed_15_boost_cate_3"`
	OrderCompletionRateBoostLevel1 float64 `json:"order_completion_rate_boost_level_1"`
	OrderCompletionRateBoostLevel2 float64 `json:"order_completion_rate_boost_level_2"`
	OrderCompletionRateBoostLevel3 float64 `json:"order_completion_rate_boost_level_3"`
	MaxBrandOrShopHoaSenBoost      float64 `json:"max_brand_or_shop_hoa_sen_boost"`
	ShippingFeeBoostLevel1         float64 `json:"shipping_fee_boost_level_1"`
	ShippingFeeBoostLevel2         float64 `json:"shipping_fee_boost_level_2"`
	ShippingFeeBoostLevel3         float64 `json:"shipping_fee_boost_level_3"`
	ShippingFeeBoostLevel4         float64 `json:"shipping_fee_boost_level_4"`
	RatingBoostLevel1              float64 `json:"rating_boost_level_1"`
	RatingBoostLevel2              float64 `json:"rating_boost_level_2"`
	RatingBoostLevel4              float64 `json:"rating_boost_level_4"`
	RatingBoostLevel5              float64 `json:"rating_boost_level_5"`
	RatingBoostMinNumRatings       float64 `json:"rating_boost_min_num_ratings"`
	DescriptionScoreBoostLevel2    float64 `json:"description_score_boost_level_2"`
	DescriptionScoreBoostLevel3    float64 `json:"description_score_boost_level_3"`
	DescriptionScoreBoostLevel4    float64 `json:"description_score_boost_level_4"`
	DescriptionScoreBoostLevel5    float64 `json:"description_score_boost_level_5"`
	DiscountBoostLevel1            float64 `json:"discount_boost_level_1"`
	DiscountBoostLevel2            float64 `json:"discount_boost_level_2"`
	DiscountBoostLevel3            float64 `json:"discount_boost_level_3"`
	DiscountBoostLevel4            float64 `json:"discount_boost_level_4"`
	DiscountAppBoostLevel1         float64 `json:"discount_app_boost_level_1"`
	DiscountAppBoostLevel2         float64 `json:"discount_app_boost_level_2"`
	DiscountAppBoostLevel3         float64 `json:"discount_app_boost_level_3"`
	DescriptionScoreBoostDefault   float64 `json:"description_score_boost_default"`
	MaxShippingFeeBoostRatio       float64 `json:"max_shipping_fee_boost_ratio"`
	MaxBrandBoost                  float64 `json:"max_brand_boost"`
	MaxShopHoaSenBoost             float64 `json:"max_shop_hoa_sen_boost"`
	MaxShopTichCucBoost            float64 `json:"max_shop_tich_cuc_boost"`
	MinShopMallBoost               float64 `json:"min_shop_mall_boost"`
	MaxShopMallBoost               float64 `json:"max_shop_mall_boost"`
	MaxShopMallOps                 float64 `json:"max_shop_mall_ops"`
	//Listing ML
	ListingMLV2 ListingML `json:"listing_ml_v2"`
	ListingMLV3 ListingML `json:"listing_ml_v3"`
}

type ListingML struct {
	ClickWeight                      float64 `bson:"click_weight" json:"click_weight"`
	OrderWithQuantityWeight          float64 `bson:"order_with_quantity_weight" json:"order_with_quantity_weight"`
	RevenueWeight                    float64 `bson:"revenue_weight" json:"revenue_weight"`
	Position1CountWeight             float64 `bson:"position_1_count_weight" json:"position_1_count_weight"`
	Position24CountWeight            float64 `bson:"position_2_4_count_weight" json:"position_2_4_count_weight"`
	Position560CountWeight           float64 `bson:"position_5_60_count_weight" json:"position_5_60_count_weight"`
	CtrWeight                        float64 `bson:"ctr_weight" json:"ctr_weight"`
	CRWeight                         float64 `bson:"cr_weight" json:"cr_weight"`
	ScoreCate2Weight                 float64 `bson:"score_cate2_weight" json:"score_cate2_weight"`
	ScoreCate3Weight                 float64 `bson:"score_cate3_weight" json:"score_cate3_weight"`
	OrderCompleted15BoostCate2weight float64 `bson:"order_completed_15_boost_cate2_weight" json:"order_completed_15_boost_cate2_weight"`
	OrderCompleted15BoostCate3weight float64 `bson:"order_completed_15_boost_cate3_weight" json:"order_completed_15_boost_cate3_weight"`
	ShippingFeeBoostWeight           float64 `bson:"shipping_fee_boost_weight" json:"shipping_fee_boost_weight"`
	RatingBoostWeight                float64 `bson:"rating_boost_weight" json:"rating_boost_weight"`
	DescriptionScoreBoostWeight      float64 `bson:"description_score_boost_weight" json:"description_score_boost_weight"`
	DiscountBoostWeight              float64 `bson:"discount_boost_weight" json:"discount_boost_weight"`
	DiscountAppBoostWeight           float64 `bson:"discount_app_boost_weight" json:"discount_app_boost_weight"`
	OrderCompletionRateBoostWeight   float64 `bson:"order_completion_rate_boost_weight" json:"order_completion_rate_boost_weight"`
	MaxBrandBoostV2                  float64 `bson:"max_brand_boost_v2" json:"max_brand_boost_v2"`
	MaxShopHoaSenBoostV2             float64 `bson:"max_shop_hoa_sen_boost_v2" json:"max_shop_hoa_sen_boost_v2"`
	MaxShopTichCucBoostV2            float64 `bson:"max_shop_tich_cuc_boost_v2" json:"max_shop_tich_cuc_boost_v2"`
	MinShopMallBoostV2               float64 `bson:"min_shop_mall_boost_v2" json:"min_shop_mall_boost_v2"`
	MaxShopMallBoostV2               float64 `bson:"max_shop_mall_boost_v2" json:"max_shop_mall_boost_v2"`
}

type GlobalConfig struct {
	Key   string
	Value string
}

type ProductFeatures struct {
	Data Features `json:"data"`
	ProductId uint32 `json:"product_id"`
}

type Histogram struct {
	CatePath2   string `json:"cate_path_2"`
	CatePath3    string `json:"cate_path_3"`
	CutOff2      int `json:"cut_off_2"`
	CutOff3      int `json:"cut_off_3"`
	StatusUpdate bool `json:"status_update"`
}

func copyIdenticalFields(a, b interface{}) {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b).Elem()

	at := av.Type()
	for i := 0; i < at.NumField(); i++ {
		name := at.Field(i).Name

		bf := bv.FieldByName(name)
		if bf.IsValid() {
			bf.Set(av.Field(i))
		}
	}
}