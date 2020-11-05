package model

type Shop struct {
	SellerAdminId int32 `json:"seller_admin_id" redis:"seller_admin_id" bson:"seller_admin_id"`
}

type Merchant struct {
	Id                int32        `json:"id" bson:"id"`
	MerchantId        int32        `json:"merchant_id" bson:"merchant_id"`
	Status            int32        `json:"status" bson:"status"`
	ShopStatus        int32        `json:"shop_status"`
	ShopType          int32        `json:"shop_type"`
	WareHouseRegionId int32        `json:"warehouse_region_id" bson:"warehouse_region_id"`
	Loyalty           *shopLoyalty `json:"loyalty"`
	IsInstallment     int32        `bson:"is_installment"`
}

type ShopCertified struct {
	AllAttribute map[string]struct {
		AttributeId   int         `json:"attribute_id" bson:"attribute_id"`
		AttributeCode string      `json:"attribute_code" bson:"attribute_code"`
		Value         interface{} `json:"value" bson:"value"`
	} `json:"all_attribute" bson:"all_attribute"`
}

type ShippingNew struct {
	StoreID int32   `bson:"storeid"`
	Levels  []Level `bson:"levels"`
}

type Level struct {
	OrderAmount      int64 `bson:"orderamount"`
	SellerSupportFee int64 `bson:"sellersupportfee"`
	Position         int64 `bson:"position"`
	IsActive         bool  `bson:"isactive"`
}

type ShopSupportPromotionApp struct {
	MerchantID int32      `bson:"merchantid"`
	Level      []AppLevel `bson:"applevel"`
}

type ShopSenPoint struct {
	MerchantID int32   `bson:"merchant_id"`
	Loyalty    Loyalty `bson:"loyalty"`
}

type AppLevel struct {
	DiscountPercent float32 `bson:"discountpercent"`
}

type ShopInstant struct {
	SellerAdminId           int32                    `json:"seller_admin_id" redis:"seller_admin_id" bson:"seller_admin_id"`
	ExtendedShippingPackage *extendedShippingPackage `json:"extended_shipping_package" bson:"extended_shipping_package"`
}

type extendedShippingPackage struct {
	IsUsingInstant int32 `json:"Is_using_instant"`
	IsUsingInDay   int32 `json:"Is_using_in_day"`
	IsSelfShipping int32 `json:"Is_self_shipping"`
}

type shopLoyalty struct {
	IsActive bool  `json:"is_active"`
	Percent  int64 `json:"percent"`
	//UpdatedTime int64 `json:"updated_time"`
}

type Loyalty struct {
	IsActive   bool    `bson:"is_active"`
	Percent    float64 `bson:"percent"`
	UpdateTime int64   `bson:"update_time"`
}

type ShopRatingInfo struct {
	ProductID        uint32  `json:"product_id"`
	RatingPercentage float32 `json:"rating_percentage"`
	TotalRated       int32   `json:"total_rated"`
}

type InstallmentData struct {
	ShopID              int32 `json:"shop_id"`
	IsActive            bool  `json:"is_active"`
	MinInstallmentPrice int64 `json:"min_installment_price"`
	MaxInstallmentTerm  int32 `json:"max_installment_term"`
}

type ShopAds struct {
	SellerAdminId int32           `json:"seller_admin_id" bson:"seller_admin_id"`
	AdsService    *ShopAdsService `json:"ads_service" bson:"ads_service" redis:"ads_service"`
}

type ShopAdsService struct {
	ShopAds int32 `json:"shop_ads" redis:"shop_ads" bson:"shop_ads"`
	AdPlus  int32 `json:"ad_plus" redis:"ad_plus" bson:"ad_plus"`
}
