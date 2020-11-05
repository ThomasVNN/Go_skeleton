package model

import (
	"encoding/json"
	"github.com/jinzhu/copier"
	"math/rand"
	"reflect"
	"strconv"
	"time"
)

// SearchHits specifies the list of search hits.
type SearchHits struct {
	TotalHits int64        `json:"total"`     // total number of hits found
	MaxScore  *float64     `json:"max_score"` // maximum score of all hits
	Hits      []*SearchHit `json:"hits"`      // the actual hits returned
}

// SearchHit is a single hit.
type SearchHit struct {
	Score          *float64               `json:"_score"`          // computed score
	Index          string                 `json:"_index"`          // index name
	Type           string                 `json:"_type"`           // type meta field
	Id             string                 `json:"_id"`             // external or internal
	Uid            string                 `json:"_uid"`            // uid meta field (see MapperService.java for all meta fields)
	Timestamp      int64                  `json:"_timestamp"`      // timestamp meta field
	TTL            int64                  `json:"_ttl"`            // ttl meta field
	Routing        string                 `json:"_routing"`        // routing meta field
	Parent         string                 `json:"_parent"`         // parent meta field
	Version        *int64                 `json:"_version"`        // version number, when Version is set to true in SearchService
	Sort           []interface{}          `json:"sort"`            // sort information
	Source         *json.RawMessage       `json:"_source"`         // stored document source
	Fields         map[string]interface{} `json:"fields"`          // returned fields
	MatchedQueries []string               `json:"matched_queries"` // matched queries
	// Shard
	// HighlightFields
	// SortValues
	// MatchedFilters
}

type ESType int

type Field struct {
	MgoName string
	Type    ESType
	Handle  func(product *Product)
}

type FieldForMerchant struct {
	EsName string
	Type   ESType
	Handle func(product *Merchant)
}

type ProductES struct {
	CategoryNameSuggest       string                  `json:"category_name_suggest"` //category_name
	ProductId                 uint32                  `json:"product_id"`
	ExternalId                uint32                  `json:"external_id"`
	AdminId                   uint32                  `json:"admin_id"`
	NumberFacets              []NumberFacet           `json:"number_facets"`
	NumberSort                NumberSort              `json:"number_sort"`
	Promotions                []Promotion             `json:"promotions"`
	Name                      string                  `json:"name"`
	DefaultListingScores      DefaultListingScoreEs   `json:"default_listing_scores"`
	ListingScore              ListingScore            `json:"listing_score"`
	OrderCount                OrderCountES            `json:"order_count"`
	CategoryPath              string                  `json:"category_path"` //category_id
	CategoryIds               []int64                 `json:"category_ids"`
	StatusId                  uint32                  `json:"status_id"`      //status_new
	ShopStatusId              int32                   `json:"shop_status_id"` //status of shop
	FilterType                uint32                  `json:"filter_type"`
	Assignee                  string                  `json:"assignee"`
	CreatedAt                 time.Time               `json:"created_at"`
	UpdatedAt                 time.Time               `json:"updated_at"`
	ReviewAt                  time.Time               `json:"review_at"` // Ngày hậu kiểm //review_date
	Quantity                  uint32                  `json:"quantity"`
	BrandId                   uint32                  `json:"brand_id"`
	ShopType                  uint32                  `json:"shop_type"`
	AppDiscountPercentage     float32                 `json:"app_discount_percentage"`
	SellerAdminId             uint32                  `json:"seller_admin_id"`
	SkuDefinedByShop          string                  `json:"sku_defined_by_shop"` //sku_user
	IsReviewType              uint32                  `json:"is_review_type"`
	IsUpdated                 bool                    `json:"is_updated"`
	HasCertificate            bool                    `json:"has_certificate"`
	HasVariant                bool                    `json:"has_variant"` //is_config_variant
	IsReview                  bool                    `json:"is_review"`   // Đã kiểm duyệt
	IsInstallment             bool                    `json:"is_installment"`
	IsInstantShipping         bool                    `json:"is_instant_shipping"`
	IsLoyalty                 bool                    `json:"is_loyalty"`
	IsCertificatedShop        bool                    `json:"is_certificate_shop"`          //is_certified
	IsOff                     bool                    `json:"is_off"`                       //Quá hạn cập nhật.
	IsShopSupportShippingFee  bool                    `json:"is_shop_support_shipping_fee"` //shop_free_shipping
	IsStock                   bool                    `json:"is_stock"`                     //stock_status
	OrderCountRank            float64                 `json:"order_count_rank"`
	RankUpdatedAt             time.Time               `json:"rank_updated_at"`
	RankSearch                float64                 `json:"rank_search"`
	AdsService                AdsService              `json:"ads_service" bson:"ads_service" redis:"ads_service"`
	ExtendedShippingPackage   ExtendedShippingPackage `json:"extended_shipping_package" bson:"extended_shipping_package" redis:"extended_shipping_package"`
	Vas                       Vas                     `json:"vas"`
	MinShopSupportShippingFee int64                   `json:"min_shop_support_shipping_fee"`
	LikeCounter               int64                   `json:"like_counter"`
	WarehouseLocationIds      []int64                 `json:"warehouse_location_ids"`
	CollectionIds             []string                `json:"collection_ids" bson:"collection_ids" redis:"collection_ids"` //DC2 su dung
	CabinetList               []uint32                `json:"cabinet_list" bson:"cabinet_list" redis:"cabinet_list"`       //DC2 su dung
	IsEvent                   bool                    `json:"is_event" bson:"is_event" redis:"is_event"`                   //DC2 su dung
}

type Vas struct {
	UppedAt       time.Time `json:"upped_at"`
	SearchUppedAt time.Time `json:"search_upped_at"`
}

type NumberFacet struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

type NumberSort struct {
	Price int64 `json:"final_gross_price"`
}

type Promotion struct {
	Id 		   int64     `json:"id, omitempty" redis:"id"`
	Type 	   int64     `json:"type, omitempty" redis:"type"`
	From       time.Time `json:"from, omitempty" redis:"from"`
	To         time.Time `json:"to, omitempty" redis:"to"`
	FixedPrice int64     `json:"fixed_price, omitempty" redis:"fixed_price"`
	UpdatedAt int64 	 `json:"updated_at" redis:"updated_at"`
}

type DefaultListingScoreEs struct {
	OrderCompleted15BoostCate2 float64   `json:"order_completed_15_boost_cate_2"`
	OrderCompleted15BoostCate3 float64   `json:"order_completed_15_boost_cate_3"`
	OrderCompletionRateBoost   float64   `json:"order_completion_rate_boost"`
	ShippingFeeBoost           float64   `json:"shipping_fee_boost"`
	RatingBoost                float64   `json:"rating_boost"`
	DescriptionScoreBoost      float64   `json:"description_score_boost"`
	DiscountBoost              float64   `json:"discount_boost"`
	DiscountAppBoost           float64   `json:"discount_app_boost"`
	BrandBoost                 float64   `json:"brand_boost"`
	ShopHoaSenBoost            float64   `json:"shop_hoa_sen_boost"`
	ShopTichCucBoost           float64   `json:"shop_tich_cuc_boost"`
	ShopMallBoost              float64   `json:"shop_mall_boost"`
	ScoreCate2                 int64     `json:"score_cate_2"`
	ScoreCate3                 int64     `json:"score_cate_3"`
	IsUpdated                  bool      `json:"is_updated"`
	UpdatedAt                  time.Time `json:"updated_at"`
	SubsidizedShippingFeeBoost float64   `json:"subsidized_shipping_fee_boost"`
}

type ListingScore struct {
	ScoreV2     float64   `json:"score_v2"`
	ScoreV3     float64   `json:"score_v3"`
	UpdatedAtV2 time.Time `json:"updated_at_v2"`
	UpdatedAtV3 time.Time `json:"updated_at_v3"`
}

type OrderCountES struct {
	Dd7       int64   `json:"dd_7"`
	Dd7Cod    int64   `json:"dd_7_cod"`
	Dd10      int64   `json:"dd_10"`
	Dd10Cod   int64   `json:"dd_10_cod"`
	Dd15      int64   `json:"dd_15"`
	Dd15Cod   int64   `json:"dd_15_cod"`
	Dd30      int64   `json:"dd_30"`
	Dd30Cod   int64   `json:"dd_30_cod"`
	Dd1000    int64   `json:"dd_1000"`
	Dd1000Cod int64   `json:"dd_1000_cod"`
	Rank      float64 `json:"rank"`
}

//overloading
func NewDefaultProductES(product *Product) *ProductES {
	productES := ProductES{}
	copier.Copy(&productES, &product)
	//productES.ConfigurationCategory = []string{"Kich_thuoc_1", "Kich_thuoc_2", "Kich_thuoc_3", "Kich_thuoc_4", "Kich_thuoc_5", "Kich_thuoc_6", "Kich_thuoc_7", "Chat_lieu_1", "Chat_lieu_2", "Chat_lieu_3", "Chat_lieu_4", "Chat_lieu_5", "Chat_lieu_6", "Chat_lieu_7", "Chat_lieu_8", "Chat_lieu_9", "Chat_lieu_10", "Kieu_dang_1", "Kieu_dang_2", "Kieu_dang_3", "Kieu_dang_4", "Kieu_dang_5", "Kieu_dang_6", "Kieu_dang_7", "Nhan_hieu_1", "Nhan_hieu_2", "Do_tuoi_be_1", "Do_tuoi_be_2", "Do_tuoi_be_3", "Can_nang_be_1", "Dung_tich_be_1", "Kieu_tay_ao_1", "Kich_thuoc_be_1n", "Kich_thuoc_be_1", "Nhan_hieu_be_3", "Nhan_hieu_be_2", "Nhan_hieu_be_4", "Nhan_hieu_be_5", "Nhan_hieu_be_6", "Nhan_hieu_be_7", "Nhan_hieu_be_8", "Phan_loai_be_1", "Phan_loai_be_3", "Phan_loai_be_6", "Phan_loai_be_7", "Phan_loai_be_8",
	//	"Mau_sac", "Xuat_xu", "Dung_tich", "Nhan_hieu", "Kich_thuoc_do_lot", "Kieu_giay", "Chat_lieu_giay", "Mua", "Mau_sac_elt", "Xuat_xu_elt", "Nguon_goc_elt", "Dung_luong_the_nho", "Loai_tai_nghe", "Kieu_tai_nghe", "Kieu_loa", "Hang_sx_case", "Hang_sx_modem", "Hang_sx_the_nho", "Hang_sx_pin_laptop", "Hang_sx_usb", "Hang_sx_loadanang", "Hang_sx_loa_vitinh", "Hang_sx_ban_phim", "Hang_sx_sac_dt", "Hang_sx_chuot", "Hang_sx_loa", "Hang_sx_balo", "Loai_bd_mtb", "Tuong_thich_bd_mtb", "Loai_dienthoai", "Dung_luong_nho_elt", "Loai_balo_tui", "Chuan_usb", "Cong_kn_1_mt", "Kieu_tai_nghe", "Cong_kn_tainghe", "Tainghe_tinhnang", "Gthuc_kn_modem", "Modem_nguon", "Clieu_case_bd", "Kieu_dat_dt", "Tuong_thich_bd_dtdd", "Loai_dienthoai", "Sacdt_dung_may_dt", "Loadn_tinhnang", "Loadn_dungcho", "Hang_sx_pin_dtdd", "Kichthuoc_kgs", "Loai_man_hinh", "Chat_Lieu_Do_The_Thao", "Chieu_Dai_Vot_Tennis", "Chat_Lieu_Vot_Tennis", "Chat_Lieu_Dau_Co", "Chat_Lieu_Can_Co", "Chat_Lieu_Kinh_Boi", "Chat_Lieu_Boi_Lan", "Chat_Lieu_Can_Cau", "Chieu_Dai_Can_Cau", "Chieu_Dai_Thu_Gon", "Chat_Lieu_Day_Cau", "Chieu_Dai_Day_Cau", "Loai_Luoi_Cau", "Hang_san_xuat_dd1", "Phan_loai_dd1", "Dung_tich_gd1", "Chat_lieu_gd1", "Phan_loai_gd_1", "Chat_lieu_gd_2", "Phan_loai_gd_3", "Kich_thuoc_gd_2", "Hang_san_xuat_gd3", "Kich_thuoc_gd_1", "Tong_mau_son_", "Weight_book", "Kich_thuoc_1_2_3_4_5_6_7_8_9_10_11_12_13_14_15_16_17", "Doi_tuong_1_2", "Kich_thuoc_1_2_3_4_5_6_7", "Xuat_xu_1_2_3_4_5_6", "Chat_lieu_1_2_3_4_5_6_7_8_9_10_11_12_13", "Chat_vai", "Extended_shipping_package", "Chat_lieu_3572", "Kick_thuoc_be_2", "Kich_co_non", "Kieu_dang_non", "Doi_tuong", "Kich_thuoc_5", "Kich_thuoc_4", "Hinh_thuc_3618", "Loai_ve_vui_choi_3574", "Dia_diem_3630", "Loai_tour_3619", "Loai_xe_3628", "San_bay_3617", "Loai_ve_3629", "Dia_diem_3631", "Noi_nhan_3620", "Quoc_gia_3632", "Loai_visa_3621", "Loai_hinh_di_chuyen_3622", "Loai_hinh_luu_tru_3623", "Tieu_chuan_3624", "Thoi_gian_3626", "Loai_tour_3625", "Tour_bao_gom_3627", "Chat_lieu_kgs", "Thoi_trang_nu___trang_suc___nhan"}
	setters := map[string]Field{
		"SetHasCertificate":        Field{Handle: productES.SetHasCertificate},
		//"SetIsShopShippingFeeSupported":      Field{Handle: productES.SetIsShopShippingFeeSupported},
		"SetIsUpdated":          Field{Handle: productES.SetIsUpdated},
		"SetIsStock":            Field{Handle: productES.SetIsStock},
		"SetIsCertificatedShop": Field{Handle: productES.SetIsCertificatedShop},
		"SetSkuDefinedByShop":   Field{Handle: productES.SetSkuDefinedByShop},
		"SetHasVariant":         Field{Handle: productES.SetHasVariant},
		"SetCategoryPath":       Field{Handle: productES.SetCategoryPath},
		"SetStatusId":           Field{Handle: productES.SetStatusId},
		//"SetShopStatusId":          Field{Handle: productES.SetShopStatusId},
		"SetIsInstantShipping":     Field{Handle: productES.SetIsInstantShipping},
		"SetAppDiscountPercentage": Field{Handle: productES.SetAppDiscountPercentage},
		"SetQuantity":              Field{Handle: productES.SetQuantity},
		"SetCreatedAt":             Field{Handle: productES.SetCreatedAt},
		"SetUpdatedAt":             Field{Handle: productES.SetUpdatedAt},
		"SetReviewAt":              Field{Handle: productES.SetReviewAt},
		"SetVasUp":                 Field{Handle: productES.SetVasUp},
		"SetIsEvent":               Field{Handle: productES.SetIsEvent},
		"SetCollectionIds":         Field{Handle: productES.SetCollectionIds},
		"SetCabinetList":           Field{Handle: productES.SetCabinetList},
	}
	for _, field := range setters {
		field.Handle(product)
	}
	return &productES
}

func BuildShopRelatedData(product *ProductES, shop *Merchant) *ProductES {
	if shop == nil {
		return product
	}
	setters := map[string]FieldForMerchant{
		"IsLoyalty":  FieldForMerchant{Handle: product.SetIsLoyalty},
		"ShopStatus": FieldForMerchant{Handle: product.SetShopStatus},
		"ShopType":   FieldForMerchant{Handle: product.SetShopType},
		//"ShopInstallment" : FieldForMerchant{Handle: product.SetShopInstallment},
	}
	for _, field := range setters {
		field.Handle(shop)
	}
	return product
}

func BuildShopShippingSupport(product *ProductES, shippingSupport *ShippingNew) *ProductES {
	if shippingSupport == nil {
		return product
	}

	var min int64 = 0
	var support = false

	if len(shippingSupport.Levels) > 0 {
		min = shippingSupport.Levels[0].OrderAmount
	}

	for _, level := range shippingSupport.Levels {
		if level.OrderAmount <= 0 {
			continue
		}
		if level.IsActive == true {
			support = true
		}
		if level.OrderAmount < min {
			min = level.OrderAmount
		}
	}
	product.MinShopSupportShippingFee = min
	product.IsShopSupportShippingFee = support
	return product
}

func SetPromotions(productMongo *Product) map[string]interface{} {
	var promotions []Promotion
	currentTime := time.Now().Unix()
	var finalPrice, finalPriceMax float64 = productMongo.Price, productMongo.Price
	if productMongo.Variants != nil && productMongo.IsConfigVariant {
		finalPrice = productMongo.Variants[0].Price
		finalPriceMax = productMongo.Variants[0].Price
		var quantity uint32 = 0
		for _, variant := range productMongo.Variants {
			quantity += variant.Quantity
			if variant.Quantity == 0 {
				//Sản phẩm hết hàng thì chỉ lấy giá gốc.
				if finalPrice > variant.Price {
					finalPrice = variant.Price
				}
				if finalPriceMax < variant.Price {
					finalPriceMax = variant.Price
				}
				continue
			}
			if variant.PromotionEndDate > currentTime {
				promotion := setPromotion(variant.PromotionStartDate, variant.PromotionEndDate, variant.SpecialPrice)
				promotions = append(promotions, promotion)
			}
			if finalPrice > variant.Price {
				finalPrice = variant.Price
			}
			if finalPriceMax < variant.Price {
				finalPriceMax = variant.Price
			}
			price := setPromotion(946684800, 32503680000, variant.Price)
			promotions = append(promotions, price)
		}
		if quantity == 0 {
			//Min, Max get time 1/1/2000, 1/1/3000 for sort price.
			promotionMax := setPromotion(946684800, 32503680000, finalPrice)
			promotions = append(promotions, promotionMax)
		}
	} else {
		if productMongo.PromotionToDate > currentTime {
			promotion := setPromotion(productMongo.PromotionStartDate, productMongo.PromotionToDate, productMongo.SpecialPrice)
			promotions = append(promotions, promotion)
		}
		//Min, Max get time 1/1/2000, 1/1/3000 for sort price.
		promotionMin := setPromotion(946684800, 32503680000, finalPrice)
		promotions = append(promotions, promotionMin)
		if finalPriceMax > finalPrice {
			promotionMax := setPromotion(946684800, 32503680000, finalPriceMax)
			promotions = append(promotions, promotionMax)
		}
	}
	promotions = uniquePromotions(promotions)
	responseData := make(map[string]interface{})
	responseData["promotions"] = promotions
	return responseData
}

func uniquePromotions(promotions []Promotion) []Promotion {
	var unique []Promotion
	for _, v := range promotions {
		skip := false
		for _, u := range unique {
			if v.From == u.From && v.To == u.To && v.FixedPrice == u.FixedPrice {
				skip = true
				break
			}
		}
		if !skip {
			unique = append(unique, v)
		}
	}
	return unique
}

func setPromotion(from, to int64, specialPrice float64) Promotion {
	return Promotion{
		Id:                   int64(rand.Intn(100)),
		Type:                 1,
		From:                 time.Unix(from,0),
		To:                   time.Unix(to,0),
		UpdatedAt:            time.Now().Unix(),
		FixedPrice:           int64(specialPrice),
	}
}

func (p *ProductES) SetReviewDate(product *Product) {
	p.ReviewAt = time.Unix(product.ReviewDate, 0)
}

func (p *ProductES) SetCreatedAt(product *Product) {
	createdAt, err := strconv.ParseInt(product.CreatedAt, 10, 64)
	if err != nil {
		return
	}
	p.CreatedAt = time.Unix(createdAt, 0)
}

func (p *ProductES) SetUpdatedAt(product *Product) {
	updatedAt, err := strconv.ParseInt(product.UpdatedAt, 10, 64)
	if err != nil {
		return
	}
	p.UpdatedAt = time.Unix(updatedAt, 0)
}

func (p *ProductES) SetReviewAt(product *Product) {
	p.ReviewAt = time.Unix(product.ReviewDate, 0)
}

func (p *ProductES) SetHasCertificate(product *Product) {
	p.HasCertificate = false
	if product.HasCertificate == 1 {
		p.HasCertificate = true
	}
}

func (p *ProductES) SetIsShopShippingFeeSupported(product *Product) {
	p.IsShopSupportShippingFee = false
	if product.ShopFreeShipping == 1 {
		p.IsShopSupportShippingFee = true
	}
}

func (p *ProductES) SetIsStock(product *Product) {
	p.IsStock = false
	if product.StockStatus == 1 {
		p.IsStock = true
	}
}

func (p *ProductES) SetIsUpdated(product *Product) {
	p.IsUpdated = false
	if product.IsUpdated == 1 {
		p.IsUpdated = true
	}
}

func (p *ProductES) SetIsCertificatedShop(product *Product) {
	p.IsCertificatedShop = false
	if product.IsCertified == 1 {
		p.IsCertificatedShop = true
	}
}

func (p *ProductES) SetIsInstantShipping(product *Product) {
	p.IsInstantShipping = product.ExtendedShippingPackage.IsUsingInstant
}

func (p *ProductES) SetSkuDefinedByShop(product *Product) {
	p.SkuDefinedByShop = product.SkuUser
}

func (p *ProductES) SetHasVariant(product *Product) {
	p.HasVariant = product.IsConfigVariant
}

func (p *ProductES) SetCategoryPath(product *Product) {
	p.CategoryPath = product.CategoryId
}

func (p *ProductES) SetStatusId(product *Product) {
	p.StatusId = product.StatusNew
}

func (p *ProductES) SetShopStatusId(product *Product) {
	p.ShopStatusId = product.ShopStatus
}

func (p *ProductES) SetAppDiscountPercentage(product *Product) {
	p.AppDiscountPercentage = product.PromotionApp
}

func (p *ProductES) SetQuantity(product *Product) {
	p.Quantity = product.Quantity
	if product.Quantity < 0 {
		p.Quantity = 0
	}
}

func (p *ProductES) SetIsEvent(product *Product) {
	p.IsEvent = false
	if product.IsEvent == 1 {
		p.IsEvent = true
	}
}

func (p *ProductES) SetCollectionIds(product *Product) {
	p.CollectionIds = product.CollectionIds
}

func (p *ProductES) SetCabinetList(product *Product) {
	p.CabinetList = product.CabinetList
}

func (p *ProductES) SetVasUp(product *Product) {
	if product.Vasup > 0 {
		p.Vas.UppedAt = time.Unix(product.Vasup, 0)
	}
	if product.VasupSearch > 0 {
		p.Vas.SearchUppedAt = time.Unix(product.VasupSearch, 0)
	}
}

func (p *ProductES) SetShopStatus(shop *Merchant) {
	p.ShopStatusId = shop.ShopStatus
}

func (p *ProductES) SetShopType(shop *Merchant) {
	p.ShopType = uint32(shop.ShopType)
}

func (p *ProductES) SetIsLoyalty(shop *Merchant) {
	p.IsLoyalty = shop.Loyalty != nil && shop.Loyalty.IsActive
}

func fieldSet(fields ...string) map[string]bool {
	set := make(map[string]bool, len(fields))
	for _, s := range fields {
		set[s] = true
	}
	return set
}

func (p *ProductES) SelectFields(fields ...string) map[string]interface{} {
	fs := fieldSet(fields...)
	rt, rv := reflect.TypeOf(*p), reflect.ValueOf(*p)
	out := make(map[string]interface{}, rt.NumField())
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		jsonKey := field.Tag.Get("json")
		if fs[jsonKey] {
			out[jsonKey] = rv.Field(i).Interface()
		}
	}
	return out
}

func (d *DefaultListingScoreEs) CalculateCate2Score() {
	orderCompleted15BoostCate2 := d.OrderCompleted15BoostCate2
	orderCompletionRateBoost := d.OrderCompletionRateBoost
	shippingFeeBoost := d.ShippingFeeBoost
	ratingBoost := d.RatingBoost
	descriptionScoreBoost := d.DescriptionScoreBoost
	discountBoost := d.DiscountBoost
	discountAppBoost := d.DiscountAppBoost
	d.ScoreCate2 = int64(1000 * orderCompleted15BoostCate2 * orderCompletionRateBoost * shippingFeeBoost * ratingBoost * descriptionScoreBoost * discountBoost * discountAppBoost)
}

func (d *DefaultListingScoreEs) CalculateCate3Score() {
	orderCompleted15BoostCate3 := d.OrderCompleted15BoostCate3
	orderCompletionRateBoost := d.OrderCompletionRateBoost
	shippingFeeBoost := d.ShippingFeeBoost
	ratingBoost := d.RatingBoost
	descriptionScoreBoost := d.DescriptionScoreBoost
	discountBoost := d.DiscountBoost
	discountAppBoost := d.DiscountAppBoost
	d.ScoreCate2 = int64(1000 * orderCompleted15BoostCate3 * orderCompletionRateBoost * shippingFeeBoost * ratingBoost * descriptionScoreBoost * discountBoost * discountAppBoost)
}

func (p *ProductES) GetValidUpdateData(data []byte) map[string]interface{} {
	var result = make(map[string]interface{})
	var updateFields map[string]interface{}
	_ = json.Unmarshal(data, &updateFields)
	rt, rv := reflect.TypeOf(*p), reflect.ValueOf(*p)
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		jsonKey := field.Tag.Get("json")
		if _, ok := updateFields[jsonKey]; ok {
			result[jsonKey] = rv.Field(i).Interface()
		}
	}
	return result
}
