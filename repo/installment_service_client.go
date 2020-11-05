package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/model"
	"gitlab.thovnn.vn/protobuf/internal-apis-go/installment"
	"strings"

	"gitlab.thovnn.vn/dc1/product-discovery/es_service/pkg/client")


type InstallmentServiceClient interface {
	GetInstallmentByShopID(shopId uint32) (*model.InstallmentData, error)
	GetInstallmentByShopIDs(shopId []uint32) ([]*model.InstallmentData, error)
}


type installmentService struct {
	url string
	grpcClient         client.Client
}

func (i *installmentService) getClient() (installment.InstallmentServiceClient, error) {
	conn, err := i.grpcClient.GetConnection(i.url)
	if err != nil {
		return nil, err
	}

	return installment.NewInstallmentServiceClient(conn), nil
}

func (i *installmentService) GetInstallmentByShopID(shopID uint32) (*model.InstallmentData, error) {
	c, err := i.getClient()
	if err != nil {
		return nil, err
	}
	req := &installment.GetShopSummaryRequest{
		ShopId:shopID,
	}
	resp, err := c.GetShopSummary(context.Background(), req)
	if err != nil {
		return nil, err
	}
	var data model.InstallmentData

	b, err := json.Marshal(resp)

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &data)

	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (i *installmentService) GetInstallmentByShopIDs(shopIDs []uint32) ([]*model.InstallmentData, error) {
	c, err := i.getClient()
	if err != nil {
		return nil, err
	}
	var data []*model.InstallmentData
	if len (shopIDs) == 0 {
		return data, nil
	}
	var ids []string
	for _, i := range shopIDs {
		ids = append (ids, fmt.Sprint(i))
	}
	req := &installment.GetListShopSummaryRequest{
		ListShopId: strings.Join(ids, "_"),
	}
	resp, err := c.GetListShopSummary(context.Background(), req)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(resp.GetShopSummaryConfig())

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &data)

	if err != nil {
		return nil, err
	}

	return data, nil
}

func NewInstallmentServiceClient(url string, grpcClient client.Client) InstallmentServiceClient {
	return &installmentService{
		url: url,
		grpcClient:         grpcClient,
	}
}
