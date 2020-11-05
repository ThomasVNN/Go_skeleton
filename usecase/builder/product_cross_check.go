package builder

import (
	"errors"
	"github.com/olivere/elastic/v7"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/model"
	"strconv"
	"time"
)

func (pb *productBuilder) BuildProductCrossCheckLog(productId uint32, err error) *model.ProductCrossCheck {
	currentTime := time.Now()
	isSyncES := true
	var strErr string
	if err != nil {
		strErr = err.Error()
		isSyncES = false
	}
	return &model.ProductCrossCheck{
		ProductId:    productId,
		IsSyncEs:     isSyncES,
		ErrorElastic: strErr,
		CreatedAt:    currentTime.Unix(),
		UpdatedAt:    currentTime.Unix(),
		ExpiredAt:    currentTime,
	}
}

func (pb *productBuilder) BuildProductCrossCheckLogs(productIds []uint32, err error) []*model.ProductCrossCheck {
	var logs []*model.ProductCrossCheck
	for _, productId := range productIds {
		currentTime := time.Now()
		isSyncES := true
		var strErr string
		if err != nil {
			strErr = err.Error()
			isSyncES = false
		}
		logs = append(logs, &model.ProductCrossCheck{
			ProductId:    productId,
			IsSyncEs:     isSyncES,
			ErrorElastic: strErr,
			CreatedAt:    currentTime.Unix(),
			UpdatedAt:    currentTime.Unix(),
			ExpiredAt:    currentTime,
		})
	}
	return logs
}

func (pb *productBuilder) BuildESResponseCrossCheckLog(bulkResponse *elastic.BulkResponse) []*model.ProductCrossCheck {
	if bulkResponse == nil {
		return nil
	}
	var logsProductEs []*model.ProductCrossCheck
	for _, f := range bulkResponse.Failed() {
		pId, _ := strconv.ParseInt(f.Id, 10, 32)
		logProductEs := pb.BuildProductCrossCheckLog(uint32(pId), errors.New(f.Error.Reason))
		logsProductEs = append(logsProductEs, logProductEs)
	}

	for _, f := range bulkResponse.Succeeded() {
		pId, _ := strconv.ParseInt(f.Id, 10, 32)
		logProductEs := pb.BuildProductCrossCheckLog(uint32(pId), nil)
		logsProductEs = append(logsProductEs, logProductEs)
	}
	return logsProductEs
}
