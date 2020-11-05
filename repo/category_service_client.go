package repo

import (
	"context"
	"fmt"

	"gitlab.thovnn.vn/dc1/product-discovery/es_service/pkg/client"

	"github.com/gogo/protobuf/types"
	"github.com/jinzhu/copier"
	"gitlab.thovnn.vn/protobuf/internal-apis-go/category"
)

type CategoryInfo struct {
	CategoryId int32  `json:"category_id,omitempty"`
	Name       string `json:"name,omitempty"`
}

type CategoryServiceClient interface {
	GetCategoryById(categoryId int32, paths []string) (*CategoryInfo, error)
	GetConfigurationCategory(categoryId uint64) ([]string, error)
}

type CategoryServiceImpl struct {
	categoryServiceUrl string
	grpcClient         client.Client
}

func NewCategoryServiceClient(url string, grpcClient client.Client) CategoryServiceClient {
	return &CategoryServiceImpl{
		categoryServiceUrl: url,
		grpcClient:         grpcClient,
	}
}

func (this *CategoryServiceImpl) getCategoryClient() (category.CategoryServiceClient, error) {
	conn, err := this.grpcClient.GetConnection(this.categoryServiceUrl)
	if err != nil {
		return nil, err
	}

	return category.NewCategoryServiceClient(conn), nil
}

func (this *CategoryServiceImpl) GetCategoryById(categoryId int32, paths []string) (*CategoryInfo, error) {
	client, err := this.getCategoryClient()
	if err != nil {
		return nil, err
	}
	categoryRequest := &category.GetCategoryRequest{
		Id: categoryId,
		Fields: &types.FieldMask{
			Paths: paths,
		},
	}
	grpcCategory, err := client.GetCategory(context.Background(), categoryRequest)
	if err != nil {
		fmt.Println("client.GetCategory error", err.Error(), categoryId)
		return nil, err
	}
	categoryInfo := CategoryInfo{}
	copier.Copy(&categoryInfo, &grpcCategory)
	return &categoryInfo, nil
}

func (this *CategoryServiceImpl) GetConfigurationCategory(categoryId uint64) ([]string, error) {
	client, err := this.getCategoryClient()
	if err != nil {
		return nil, err
	}
	listAttributeFilterRequest := &category.ListAttributeFilterRequest{
		CategoryId: categoryId,
	}
	filter, err := client.ListAttributeFilter(context.Background(), listAttributeFilterRequest)
	if err != nil {
		return nil, err
	}
	var codes []string
	for _, item := range filter.Attributes {
		codes = append(codes, item.Code)
	}
	return codes, nil
}
