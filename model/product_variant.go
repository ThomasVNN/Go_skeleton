package model

type Variant struct {
	Attributes         []*VariantAttribute `json:"attributes" bson:"attributes" redis:"attributes"`
	AttributeHash      string              `json:"attribute_hash" bson:"attribute_hash" redis:"attribute_hash"`
	PromotionStartDate int64               `json:"promotion_start_date" bson:"promotion_start_date" redis:"promotion_start_date"`
	PromotionEndDate   int64               `json:"promotion_end_date" bson:"promotion_end_date" redis:"promotion_end_date"`
	PromotionPercent   float32             `json:"promotion_percent" bson:"promotion_percent" redis:"promotion_percent"`
	IsPromotion        uint32              `json:"is_promotion" bson:"is_promotion" redis:"is_promotion"`
	SkuUser            string              `json:"sku_user" bson:"sku_user" redis:"sku_user"`
	Price              float64             `json:"price" bson:"price" redis:"price"`
	SpecialPrice       float64             `json:"special_price" bson:"special_price" redis:"special_price"`
	FinalPrice         float64             `json:"final_price" bson:"final_price" redis:"final_price"`
	Quantity           uint32              `json:"quantity" bson:"quantity" redis:"quantity"`
	Order              uint32              `json:"order" bson:"order" redis:"order"`
}

type VariantAttribute struct {
	Id       int32  `json:"id" bson:"id" redis:"id"`
	OptionId int32  `json:"option_id" bson:"option_id" redis:"option_id"`
	Code     string `json:"code" bson:"code" redis:"code"`
}

type ProductVariant struct {
	ProductId uint32     `json:"product_id" bson:"product_id" redis:"product_id"`
	Variants  []*Variant `json:"variants" bson:"variants" redis:"variants"`
}