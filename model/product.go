package model

type Product struct {
	ProductId           uint32  `json:"product_id" bson:"product_id" redis:"product_id"`
	ExternalId          uint32  `json:"external_id" bson:"external_id" redis:"external_id"`
	Name                string  `json:"name" bson:"name" redis:"name"`
	AdminId             uint32  `json:"admin_id" bson:"admin_id" redis:"admin_id"`
	SellerAdminId       int32   `json:"seller_admin_id" bson:"seller_admin_id" redis:"seller_admin_id"`
	IsCertified         uint32  `json:"is_certified" bson:"is_certified" redis:"is_certified"`
	SkuUser             string  `json:"sku_user" bson:"sku_user" redis:"sku_user"`
	UnitType            string  `json:"unit_type" bson:"unit_type" redis:"unit_type"`
	CategoryId          string  `json:"category_id" bson:"category_id" redis:"category_id"`
	CategoryNameSuggest string  `json:"category_name_suggest"` //category_name
	StatusNew           uint32  `json:"status_new" bson:"status_new" redis:"status_new"`
	StockStatus         uint32  `json:"stock_status" bson:"stock_status" redis:"stock_status"`
	UpdatedAt           string  `json:"updated_at" bson:"updated_at" redis:"updated_at"`
	TypeProduct         uint32  `json:"type_product" bson:"type_product" redis:"type_product"`
	Quantity            uint32  `json:"quantity" bson:"quantity" redis:"quantity"`
	Price               float64 `json:"price" bson:"price" redis:"price"`
	SpecialPrice        float64 `json:"special_price" bson:"special_price" redis:"special_price"`
	PromotionApp        float32 `json:"promotion_app" bson:"promotion_app" redis:"promotion_app"`
	PromotionStartDate  int64   `json:"promotion_start_date" bson:"promotion_start_date" redis:"promotion_start_date"`
	PromotionToDate     int64   `json:"promotion_to_date" bson:"promotion_to_date" redis:"promotion_to_date"`
	// Cac field de index
	PromotionPercent float32           `json:"promotion_percent" bson:"promotion_percent" redis:"promotion_percent"`
	FinalPrice       float64           `json:"final_price" bson:"final_price" redis:"final_price"`
	BrandId          uint32            `json:"brand_id" bson:"brand_id" redis:"brand_id"`
	Attribute        map[string]string `,inline`
	ShopFreeShipping uint32            `json:"shop_free_shipping" bson:"shop_free_shipping" redis:"shop_free_shipping"`
	RankSearch       float64           `json:"rank_search" bson:"rank_search" redis:"rank_search"`
	HasCertificate   uint32            `json:"has_certificate" bson:"has_certificate" redis:"has_certificate"`
	ShopType         uint32            `json:"shop_type" bson:"shop_type" redis:"shop_type"`
	ShopStatus       int32            	`json:"-" bson:"-"`
	FilterType       uint32            `json:"filter_type" bson:"filter_type" redis:"filter_type"` // Loại sản phẩm (Inside dùng )
	CreatedAt        string            `json:"created_at" bson:"created_at" redis:"created_at"`
	ReviewDate       int64             `json:"review_date" bson:"review_date" redis:"review_date"`          // Ngày hậu kiểm
	IsReview         bool              `json:"is_review" bson:"is_review" redis:"is_review"`                // Đã kiểm duyệt
	IsReviewType     uint32            `json:"is_review_type" bson:"is_review_type" redis:"is_review_type"` // Loại hậu kiểm
	Assignee         string            `json:"assignee" bson:"assignee" redis:"assignee"`                   //Người hậu kiểm
	IsOff            bool              `json:"is_off" bson:"is_off" redis:"is_off"`                         //Sử dụng cho tab quá hạn cập nhật
	IsUpdated        uint32            `json:"is_updated" bson:"is_updated" redis:"is_updated"`             //Sử dụng cho filter loại product bên seller
	CollectionIds    []string      	   `json:"collection_ids" bson:"collection_ids" redis:"collection_ids"` //DC2 su dung
	CabinetList      []uint32      	   `json:"cabinet_list" bson:"cabinet_list" redis:"cabinet_list"` 		//DC2 su dung
	IsEvent          uint32            `json:"is_event" bson:"is_event" redis:"is_event"`					//DC2 su dung
	// Field cho Default Listing
	// Ghi nhan o mongo
	PriceSort               float64                 `json:"price_sort" bson:"price_sort" redis:"price_sort"`
	ExtendedShippingPackage ExtendedShippingPackage `json:"extended_shipping_package" bson:"extended_shipping_package" redis:"extended_shipping_package"`
	PriceMax                float64                 `json:"price_max" bson:"price_max" redis:"price_max"`
	FinalPriceMax           float64                 `json:"final_price_max" bson:"final_price_max" redis:"final_price_max"`
	IsConfigVariant         bool                    `json:"is_config_variant" bson:"is_config_variant" redis:"is_config_variant"`
	IsPromotion             uint32                  `json:"is_promotion" bson:"is_promotion" redis:"is_promotion"`
	Variants                []*Variant              `json:"variants" bson:"-"`
	IsInvalidVariant        bool                    `json:"is_invalid_variant" bson:"-" redis:"-"`
	AdsService              AdsService              `json:"ads_service" bson:"ads_service" redis:"ads_service"`
	Voucher                 Voucher                 `json:"voucher" bson:"voucher" redis:"voucher"`
	DefaultListingScore     DefaultListingScore     `json:"default_listing_score" bson:"default_listing_score" redis:"default_listing_score"`
	Vasup                   int64                   `json:"vasup" bson:"vasup" redis:"vasup"` // time that shop clicked upting
	VasupSearch             int64                   `json:"vasup_search" bson:"vasup_search" redis:"vasup_search"`
	RequiredOptions     string           `json:"required_options" bson:"required_options" redis:"required_options"`
	//
	OrderCount				OrderCount				`json:"order_count" bson:"order_count" redis:"order_count"`
}

type ExtendedShippingPackage struct {
	IsUsingInstant bool `json:"is_using_instant" redis:"is_using_instant" bson:"is_using_instant"`
	IsUsingInDay   bool `json:"is_using_in_day" redis:"is_using_in_day" bson:"is_using_in_day"`
	IsSelfShipping bool `json:"is_self_shipping" redis:"is_self_shipping" bson:"is_self_shipping"`
}

type AdsService struct {
	ShopAds int32 `json:"shop_ads" redis:"shop_ads" bson:"shop_ads"`
	AdPlus  int32 `json:"ad_plus" redis:"ad_plus" bson:"ad_plus"`
}

type Voucher struct {
	ProductType uint32 `json:"product_type" redis:"product_type" bson:"product_type"`
	StartDate   int64  `json:"start_date" redis:"start_date" bson:"start_date"`
	EndDate     int64  `json:"end_date" redis:"end_date" bson:"end_date"`
	IsCheckDate bool   `json:"is_check_date" redis:"is_check_date" bson:"is_check_date"`
}

type AttributeMapping struct {
	AttributeId   int                      `json:"attribute_id" bson:"attribute_id"`
	Name          string                   `json:"name" bson:"name"`
	ProductOption string                   `json:"product_option" bson:"product_option"`
	ShowRequired  int                      `json:"show_required" bson:"show_required"`
	Type          string                   `json:"type" bson:"type"`
	Value         []AttributeOptionMapping `json:"value" bson:"value"`
	IsCustom      bool                     `json:"is_custom" bson:"is_custom"`
}

type AttributeOptionMapping struct {
	OptionId        int32  `json:"option_id" bson:"option_id"`
	Value           string `json:"value" bson:"value"`
	ProductOptionId string `json:"product_option_id" bson:"product_option_id"`
	Image           string `json:"image" bson:"image"`
	IsCustom        bool   `json:"is_custom" bson:"is_custom"`
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

type OrderCount struct {
	Dd90	  int64   `json:"dd_90" bson:"dd_90" redis:"dd_90"`
	Dd90Cod	  int64   `json:"dd_90_cod" bson:"dd_90_cod" redis:"dd_90_cod"`
	Dd30      int64   `json:"dd_30" bson:"dd_30" redis:"dd_30"`
}
