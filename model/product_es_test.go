package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestProductES_SetAttributeFacets(t *testing.T) {
	var tests = []struct {
		productMongo         *Product
		productConfiguration []string
		expected             []NumberFacet
	}{
		{
			&Product{
				Attribute: map[string]string{"chat_vai": "{\"attribute_id\":1495,\"name\":\"Chất vải\",\"product_option\":\"17726831_1495\",\"show_required\":1,\"type\":\"String\",\"value\":[{\"option_id\":25108,\"value\":\"Cotton 4 chiều\",\"product_option_id\":\"17726831_25108\",\"image\":\"\",\"is_custom\":false}],\"is_custom\":false}",
					"hoa_tiet": "{\"attribute_id\":1490,\"name\":\"Họa tiết\",\"product_option\":\"17726831_1490\",\"show_required\":1,\"type\":\"String\",\"value\":[{\"option_id\":25010,\"value\":\"Trơn\",\"product_option_id\":\"17726831_25010\",\"image\":\"\",\"is_custom\":false}],\"is_custom\":false}",
					"mau_sac":  "{\"attribute_id\":284,\"name\":\"Màu sắc\",\"product_option\":\"17726831_284\",\"show_required\":1,\"type\":\"Option\",\"value\":[{\"option_id\":46335,\"value\":\"Xám Tiêu\",\"product_option_id\":\"17726831_46335\",\"image\":\"https://media3.scdn.vn/img3/2019/4_26/pUjmSh.jpg\",\"is_custom\":true},{\"option_id\":46336,\"value\":\"Ve Chai\",\"product_option_id\":\"17726831_46336\",\"image\":\"https://media3.scdn.vn/img3/2019/5_13/VLVtDU.jpg\",\"is_custom\":true}],\"is_custom\":false}"},
			},
			[]string{"Kich_thuoc_1", "Kich_thuoc_2", "Kich_thuoc_3", "Kich_thuoc_4", "Kich_thuoc_5", "Kich_thuoc_6", "Kich_thuoc_7", "Chat_lieu_1", "Chat_lieu_2", "Chat_lieu_3", "Chat_lieu_4", "Chat_lieu_5", "Chat_lieu_6", "Chat_lieu_7", "Chat_lieu_8", "Chat_lieu_9", "Chat_lieu_10", "Kieu_dang_1", "Kieu_dang_2", "Kieu_dang_3", "Kieu_dang_4", "Kieu_dang_5", "Kieu_dang_6", "Kieu_dang_7", "Nhan_hieu_1", "Nhan_hieu_2", "Do_tuoi_be_1", "Do_tuoi_be_2", "Do_tuoi_be_3", "Can_nang_be_1", "Dung_tich_be_1", "Kieu_tay_ao_1", "Kich_thuoc_be_1n", "Kich_thuoc_be_1", "Nhan_hieu_be_3", "Nhan_hieu_be_2", "Nhan_hieu_be_4", "Nhan_hieu_be_5", "Nhan_hieu_be_6", "Nhan_hieu_be_7", "Nhan_hieu_be_8", "Phan_loai_be_1", "Phan_loai_be_3", "Phan_loai_be_6", "Phan_loai_be_7", "Phan_loai_be_8", "Mau_sac", "Xuat_xu", "Dung_tich", "Nhan_hieu", "Kich_thuoc_do_lot", "Kieu_giay", "Chat_lieu_giay", "Mua", "Mau_sac_elt", "Xuat_xu_elt", "Nguon_goc_elt", "Dung_luong_the_nho", "Loai_tai_nghe", "Kieu_tai_nghe", "Kieu_loa", "Hang_sx_case", "Hang_sx_modem", "Hang_sx_the_nho", "Hang_sx_pin_laptop", "Hang_sx_usb", "Hang_sx_loadanang", "Hang_sx_loa_vitinh", "Hang_sx_ban_phim", "Hang_sx_sac_dt", "Hang_sx_chuot", "Hang_sx_loa", "Hang_sx_balo", "Loai_bd_mtb", "Tuong_thich_bd_mtb", "Loai_dienthoai", "Dung_luong_nho_elt", "Loai_balo_tui", "Chuan_usb", "Cong_kn_1_mt", "Kieu_tai_nghe", "Cong_kn_tainghe", "Tainghe_tinhnang", "Gthuc_kn_modem", "Modem_nguon", "Clieu_case_bd", "Kieu_dat_dt", "Tuong_thich_bd_dtdd", "Loai_dienthoai", "Sacdt_dung_may_dt", "Loadn_tinhnang", "Loadn_dungcho", "Hang_sx_pin_dtdd", "Kichthuoc_kgs", "Loai_man_hinh", "Chat_Lieu_Do_The_Thao", "Chieu_Dai_Vot_Tennis", "Chat_Lieu_Vot_Tennis", "Chat_Lieu_Dau_Co", "Chat_Lieu_Can_Co", "Chat_Lieu_Kinh_Boi", "Chat_Lieu_Boi_Lan", "Chat_Lieu_Can_Cau", "Chieu_Dai_Can_Cau", "Chieu_Dai_Thu_Gon", "Chat_Lieu_Day_Cau", "Chieu_Dai_Day_Cau", "Loai_Luoi_Cau", "Hang_san_xuat_dd1", "Phan_loai_dd1", "Dung_tich_gd1", "Chat_lieu_gd1", "Phan_loai_gd_1", "Chat_lieu_gd_2", "Phan_loai_gd_3", "Kich_thuoc_gd_2", "Hang_san_xuat_gd3", "Kich_thuoc_gd_1", "Tong_mau_son_", "Weight_book", "Kich_thuoc_1_2_3_4_5_6_7_8_9_10_11_12_13_14_15_16_17", "Doi_tuong_1_2", "Kich_thuoc_1_2_3_4_5_6_7", "Xuat_xu_1_2_3_4_5_6", "Chat_lieu_1_2_3_4_5_6_7_8_9_10_11_12_13", "Chat_vai", "Extended_shipping_package", "Chat_lieu_3572", "Kick_thuoc_be_2", "Kich_co_non", "Kieu_dang_non", "Doi_tuong", "Kich_thuoc_5", "Kich_thuoc_4", "Hinh_thuc_3618", "Loai_ve_vui_choi_3574", "Dia_diem_3630", "Loai_tour_3619", "Loai_xe_3628", "San_bay_3617", "Loai_ve_3629", "Dia_diem_3631", "Noi_nhan_3620", "Quoc_gia_3632", "Loai_visa_3621", "Loai_hinh_di_chuyen_3622", "Loai_hinh_luu_tru_3623", "Tieu_chuan_3624", "Thoi_gian_3626", "Loai_tour_3625", "Tour_bao_gom_3627", "Chat_lieu_kgs", "Thoi_trang_nu___trang_suc___nhan"},
			[]NumberFacet{
				NumberFacet{
					Name:  "Chat_vai_facet",
					Value: 25108,
				},
				NumberFacet{
					Name:  "Mau_sac_facet",
					Value: 46335,
				},
				NumberFacet{
					Name:  "Mau_sac_facet",
					Value: 46336,
				},
			},
		},
	}

	for _, test := range tests {
		actual := setAttributeFacets(test.productMongo, test.productConfiguration)
		assert.Equal(t, test.expected, actual)
	}
}

func TestProductES_SetNumberFacets(t *testing.T) {
	var tests = []struct {
		productMongo *Product
		expected     int
	}{
		{
			&Product{
				FinalPrice:    1500000,
				FinalPriceMax: 2500000,
				Attribute: map[string]string{"chat_vai": "{\"attribute_id\":1495,\"name\":\"Chất vải\",\"product_option\":\"17726831_1495\",\"show_required\":1,\"type\":\"String\",\"value\":[{\"option_id\":25108,\"value\":\"Cotton 4 chiều\",\"product_option_id\":\"17726831_25108\",\"image\":\"\",\"is_custom\":false}],\"is_custom\":false}",
					"hoa_tiet": "{\"attribute_id\":1490,\"name\":\"Họa tiết\",\"product_option\":\"17726831_1490\",\"show_required\":1,\"type\":\"String\",\"value\":[{\"option_id\":25010,\"value\":\"Trơn\",\"product_option_id\":\"17726831_25010\",\"image\":\"\",\"is_custom\":false}],\"is_custom\":false}",
					"mau_sac":  "{\"attribute_id\":284,\"name\":\"Màu sắc\",\"product_option\":\"17726831_284\",\"show_required\":1,\"type\":\"Option\",\"value\":[{\"option_id\":46335,\"value\":\"Xám Tiêu\",\"product_option_id\":\"17726831_46335\",\"image\":\"https://media3.scdn.vn/img3/2019/4_26/pUjmSh.jpg\",\"is_custom\":true},{\"option_id\":46336,\"value\":\"Ve Chai\",\"product_option_id\":\"17726831_46336\",\"image\":\"https://media3.scdn.vn/img3/2019/5_13/VLVtDU.jpg\",\"is_custom\":true}],\"is_custom\":false}"},
			},
			5,
			/*[]*NumberFacet{
				&NumberFacet{
					Name:  "Chat_vai_facet",
					Value: 25108,
				},
				&NumberFacet{
					Name:  "Mau_sac_facet",
					Value: 46335,
				},
				&NumberFacet{
					Name:  "Mau_sac_facet",
					Value: 46336,
				},
				&NumberFacet{
					Name:  "final_price",
					Value: 1500000,
				},
				&NumberFacet{
					Name:  "final_price",
					Value: 2500000,
				},
			},*/
		},
	}

	for _, test := range tests {
		productES := &ProductES{
			ConfigurationCategory: []string{"Kich_thuoc_1", "Kich_thuoc_2", "Kich_thuoc_3", "Kich_thuoc_4", "Kich_thuoc_5", "Kich_thuoc_6", "Kich_thuoc_7", "Chat_lieu_1", "Chat_lieu_2", "Chat_lieu_3", "Chat_lieu_4", "Chat_lieu_5", "Chat_lieu_6", "Chat_lieu_7", "Chat_lieu_8", "Chat_lieu_9", "Chat_lieu_10", "Kieu_dang_1", "Kieu_dang_2", "Kieu_dang_3", "Kieu_dang_4", "Kieu_dang_5", "Kieu_dang_6", "Kieu_dang_7", "Nhan_hieu_1", "Nhan_hieu_2", "Do_tuoi_be_1", "Do_tuoi_be_2", "Do_tuoi_be_3", "Can_nang_be_1", "Dung_tich_be_1", "Kieu_tay_ao_1", "Kich_thuoc_be_1n", "Kich_thuoc_be_1", "Nhan_hieu_be_3", "Nhan_hieu_be_2", "Nhan_hieu_be_4", "Nhan_hieu_be_5", "Nhan_hieu_be_6", "Nhan_hieu_be_7", "Nhan_hieu_be_8", "Phan_loai_be_1", "Phan_loai_be_3", "Phan_loai_be_6", "Phan_loai_be_7", "Phan_loai_be_8", "Mau_sac", "Xuat_xu", "Dung_tich", "Nhan_hieu", "Kich_thuoc_do_lot", "Kieu_giay", "Chat_lieu_giay", "Mua", "Mau_sac_elt", "Xuat_xu_elt", "Nguon_goc_elt", "Dung_luong_the_nho", "Loai_tai_nghe", "Kieu_tai_nghe", "Kieu_loa", "Hang_sx_case", "Hang_sx_modem", "Hang_sx_the_nho", "Hang_sx_pin_laptop", "Hang_sx_usb", "Hang_sx_loadanang", "Hang_sx_loa_vitinh", "Hang_sx_ban_phim", "Hang_sx_sac_dt", "Hang_sx_chuot", "Hang_sx_loa", "Hang_sx_balo", "Loai_bd_mtb", "Tuong_thich_bd_mtb", "Loai_dienthoai", "Dung_luong_nho_elt", "Loai_balo_tui", "Chuan_usb", "Cong_kn_1_mt", "Kieu_tai_nghe", "Cong_kn_tainghe", "Tainghe_tinhnang", "Gthuc_kn_modem", "Modem_nguon", "Clieu_case_bd", "Kieu_dat_dt", "Tuong_thich_bd_dtdd", "Loai_dienthoai", "Sacdt_dung_may_dt", "Loadn_tinhnang", "Loadn_dungcho", "Hang_sx_pin_dtdd", "Kichthuoc_kgs", "Loai_man_hinh", "Chat_Lieu_Do_The_Thao", "Chieu_Dai_Vot_Tennis", "Chat_Lieu_Vot_Tennis", "Chat_Lieu_Dau_Co", "Chat_Lieu_Can_Co", "Chat_Lieu_Kinh_Boi", "Chat_Lieu_Boi_Lan", "Chat_Lieu_Can_Cau", "Chieu_Dai_Can_Cau", "Chieu_Dai_Thu_Gon", "Chat_Lieu_Day_Cau", "Chieu_Dai_Day_Cau", "Loai_Luoi_Cau", "Hang_san_xuat_dd1", "Phan_loai_dd1", "Dung_tich_gd1", "Chat_lieu_gd1", "Phan_loai_gd_1", "Chat_lieu_gd_2", "Phan_loai_gd_3", "Kich_thuoc_gd_2", "Hang_san_xuat_gd3", "Kich_thuoc_gd_1", "Tong_mau_son_", "Weight_book", "Kich_thuoc_1_2_3_4_5_6_7_8_9_10_11_12_13_14_15_16_17", "Doi_tuong_1_2", "Kich_thuoc_1_2_3_4_5_6_7", "Xuat_xu_1_2_3_4_5_6", "Chat_lieu_1_2_3_4_5_6_7_8_9_10_11_12_13", "Chat_vai", "Extended_shipping_package", "Chat_lieu_3572", "Kick_thuoc_be_2", "Kich_co_non", "Kieu_dang_non", "Doi_tuong", "Kich_thuoc_5", "Kich_thuoc_4", "Hinh_thuc_3618", "Loai_ve_vui_choi_3574", "Dia_diem_3630", "Loai_tour_3619", "Loai_xe_3628", "San_bay_3617", "Loai_ve_3629", "Dia_diem_3631", "Noi_nhan_3620", "Quoc_gia_3632", "Loai_visa_3621", "Loai_hinh_di_chuyen_3622", "Loai_hinh_luu_tru_3623", "Tieu_chuan_3624", "Thoi_gian_3626", "Loai_tour_3625", "Tour_bao_gom_3627", "Chat_lieu_kgs", "Thoi_trang_nu___trang_suc___nhan"},
		}
		productES.SetNumberFacets(test.productMongo)
		//jsonByte, _ := json.Marshal(productES.NumberFacets)
		//fmt.Println(string(jsonByte), len(productES.NumberFacets))
		assert.Equal(t, test.expected, len(productES.NumberFacets))
	}
}

func TestProductES_SetPromotion(t *testing.T) {
	var tests = []struct {
		from, to     int64
		specialPrice float64
		expected     Promotion
	}{
		{
			from: 1568937600, to: 1576800000,
			specialPrice: 99000,
			expected: Promotion{
				From:       time.Unix(1557939600, 0),
				To:         time.Unix(1558285199, 0),
				FixedPrice: int64(99000),
			},
		},
	}

	for _, test := range tests {
		actual := setPromotion(test.from, test.to, test.specialPrice)
		assert.Equal(t, test.expected, actual)
	}
}

func TestProductES_SetPromotions(t *testing.T) {
	var tests = []struct {
		productMongo *Product
		expected     []Promotion
	}{
		{
			&Product{
				Variants: []*Variant{
					&Variant{
						IsPromotion:        0,
						PromotionStartDate: 1557939600,
						PromotionEndDate:   1558285199,
					},
				},
			},
			[]Promotion{},
		},
		{
			&Product{
				Variants: []*Variant{
					&Variant{
						IsPromotion:        1,
						PromotionStartDate: 1557939600,
						PromotionEndDate:   1558285199,
						SpecialPrice:       50000,
					},
				},
			},
			[]Promotion{Promotion{
				From:       time.Unix(1557939600, 0),
				To:         time.Unix(1558285199, 0),
				FixedPrice: 50000,
			}},
		},
	}

	for _, test := range tests {
		productES := &ProductES{}
		productES.SetPromotions(test.productMongo)
		assert.Equal(t, test.expected, productES.Promotions)
	}
}

func TestProductES_SetReviewDate(t *testing.T) {
	var tests = []struct {
		productMongo *Product
		expected     time.Time
	}{
		{
			&Product{
				ReviewDate: 1557734036,
			},
			time.Unix(1557734036, 0),
		},
	}

	for _, test := range tests {
		productES := &ProductES{}
		productES.SetReviewDate(test.productMongo)
		assert.Equal(t, test.expected, productES.ReviewAt)
	}
}

func TestProductES_FieldSet(t *testing.T) {
	var tests = []struct {
		fields   []string
		expected map[string]bool
	}{
		{
			[]string{"status_new", "stock_status"},
			map[string]bool{
				"status_new":   true,
				"stock_status": true,
			},
		},
		{
			[]string{"attribute", "price"},
			map[string]bool{
				"attribute": true,
				"price":     true,
			},
		},
	}

	for _, test := range tests {
		actual := fieldSet(test.fields...)
		assert.Equal(t, test.expected, actual)
	}
}

func TestProductES_SelectFields(t *testing.T) {
	var tests = []struct {
		fields       []string
		productMongo *Product
		expected     map[string]interface{}
	}{
		{
			[]string{"status_new", "stock_status"},
			&Product{
				StatusNew:   2,
				StockStatus: 1,
			},
			map[string]interface{}{
				"status_new":   int64(2),
				"stock_status": int64(1),
			},
		},
	}

	for _, test := range tests {
		productES := NewDefaultProductES(test.productMongo)
		actual := productES.SelectFields(test.fields...)
		assert.Equal(t, test.expected, actual)
	}
}
