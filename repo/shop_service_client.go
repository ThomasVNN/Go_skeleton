package repo

import (
	"encoding/json"
	"errors"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/model"
	"io/ioutil"
	"net/http"
	"strconv"
)

const (
	GetStoresPath = "/stores/get-stores?ids="
)

type ShopInfo struct {
}

type ShopServiceClient interface {
	GetShopByIds(ids []int32) (map[int32]*model.Merchant, error)
}

type ShopServiceImpl struct {
	url string
}

func NewShopServiceClient(url string) ShopServiceClient {
	return &ShopServiceImpl{
		url: url,
	}
}

func (s *ShopServiceImpl) GetShopByIds(ids []int32) (map[int32]*model.Merchant, error) {
	if s.url == "" {
		return nil, errors.New("can not call shop service")
	}

	if len(ids) <= 0 {
		return nil, errors.New("invalid argument")
	}

	str := ""
	for _, id := range ids {
		str += strconv.Itoa(int(id)) + ","
	}

	response, err := http.Get(s.url + GetStoresPath + str)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	type Response struct {
		Code int              `json:"code"`
		Data []model.Merchant `json:"data"`
	}

	var resp = new(Response)
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}

	var mapResp = make(map[int32]*model.Merchant)
	for _, item := range resp.Data {
		if _, ok := mapResp[item.Id]; !ok {
			mapResp[item.Id] = &item
		}
	}

	return mapResp, nil
}
