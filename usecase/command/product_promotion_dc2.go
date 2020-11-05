package command

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gogo/protobuf/types"
	"github.com/jinzhu/copier"
	"github.com/segmentio/kafka-go"
	"gitlab.thovnn.vn/core/golang-sdk/spubsub"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/model"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/pkg/pubsub/rabbitmq"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/storage/esstore"
	productModel "gitlab.thovnn.vn/protobuf/internal-apis-go/product"
	"gitlab.thovnn.vn/protobuf/internal-apis-go/product/es_service"
	"os"
	"strconv"
)

type productPromotionDc2 struct {
	esStorage esstore.ESStore
	logger    log.Logger
}

const (
	addPromotion               = "1"
	updatePromotion            = "2"
	deletePromotion            = "3"
	RetryTimes                 = 3
	SuffixTopicPromotionChange = "_retry_"
)

func NewProductPromotionDC2Command(esStorage esstore.ESStore, logger log.Logger) model.ProductPromotionDC2Events {
	return &productPromotionDc2{
		esStorage: esStorage,
		logger:    logger,
	}
}

type PromotionResponse struct {
	Id         int64
	Type       int64
	From       *types.Timestamp
	To         *types.Timestamp
	FixedPrice int64
	UpdatedAt  int64
	Result     bool
	Message    string
}

type PromotionResponseDataPubSubDC2 struct {
	ProductId  int64
	Action     int
	Promotions []PromotionResponse
	Result     bool
	Message    string
}

func (m *productPromotionDc2) OnPromotionChange(msg kafka.Message) error {
	header := msg.Headers[0]
	typeChangePromotion := string(header.Value)
	var err error = nil
	responseData := productModel.PromotionResponseDataPubSubDC2{}
	switch typeChangePromotion {
	case addPromotion:
		responseData, err = m.OnPromotionAdd(msg)
	case updatePromotion:
		responseData, err = m.OnPromotionUpdate(msg)
	case deletePromotion:
		responseData, err = m.OnPromotionDelete(msg)
	}

	topic := os.Getenv("KAFKA_TOPIC_ES_PROMOTION_CHANGED") + SuffixTopicPromotionChange + fmt.Sprintf("%d", RetryTimes)

	if topic == msg.Topic || responseData.Result == true {
		fmt.Println("================")
		fmt.Println(topic, err, responseData, msg.Topic)
		fmt.Println("================")
		//call pub/sub to DC2
		buffer, _ := json.Marshal(responseData)
		pubSubUrl := os.Getenv("PUB_SUB_URL")
		rabbitMqConfig := rabbitmq.NewPubSubClient(pubSubUrl)
		pubResult, err := rabbitMqConfig.Publish(&spubsub.Publishing{
			Event: os.Getenv("PROMOTION_PUBLISH_EVENT"),
			Token: os.Getenv("PROMOTION_PUBLISH_TOKEN"),
			Data:  buffer,
		})
		fmt.Println(pubResult, err)
	}
	return err
}

func (m *productPromotionDc2) OnPromotionAdd(msg kafka.Message) (productModel.PromotionResponseDataPubSubDC2, error) {
	logger := log.With(m.logger, "prefix", "OnPromotionAdd")
	data := msg.Value

	hashData := fmt.Sprintf("%x", sha1.Sum([]byte(data)))
	ll := map[string]interface{}{"hash": hashData, "data": string(data), "offset": msg.Offset, "partition": msg.Partition}
	logData, _ := json.Marshal(ll)
	level.Info(logger).Log(
		"msg", fmt.Sprint("received data from "+msg.Topic),
		"meta_data", string(logData))

	promotionAdd := productModel.AddProductPromotionRequest{}
	err := json.Unmarshal(data, &promotionAdd)

	responseData := productModel.PromotionResponseDataPubSubDC2{
		ProductId:  promotionAdd.ProductId,
		Action:     1,
		Promotions: []*productModel.PromotionResponseDc2{},
		Result:     true,
	}

	if err != nil {
		level.Error(logger).Log("msg", fmt.Sprint("add fail, unmarshal json fail, ", err),
			"meta_data", string(logData))
		responseData.Result = false
		responseData.Message = err.Error()
		return responseData, err
	}

	productId := strconv.FormatInt(promotionAdd.ProductId, 10)
	promotionsRequest := promotionAdd.Promotions

	var promotions []model.Promotion
	promotions = parsePromotionDataFromProto(promotionsRequest)

	promotionResponse, error := m.addPromotions(productId, promotions)
	responseData.Promotions = promotionResponse

	if error != nil {
		level.Error(logger).Log("msg", fmt.Sprint("add fail, esStore.addPromotions error, ", error),
			"meta_data", string(logData))
		responseData.Result = false
		responseData.Message = "add failed"
	} else {
		level.Info(logger).Log("msg", "update promotion success",
			"meta_data", string(logData))
	}
	return responseData, error
}

func (m *productPromotionDc2) OnPromotionUpdate(msg kafka.Message) (productModel.PromotionResponseDataPubSubDC2, error) {
	logger := log.With(m.logger, "prefix", "OnPromotionUpdate")
	data := msg.Value

	hashData := fmt.Sprintf("%x", sha1.Sum([]byte(data)))
	ll := map[string]interface{}{"hash": hashData, "data": string(data), "offset": msg.Offset, "partition": msg.Partition}
	logData, _ := json.Marshal(ll)
	level.Info(logger).Log(
		"msg", fmt.Sprint("received data from "+msg.Topic),
		"meta_data", string(logData))

	promotionUpdate := productModel.UpdateProductPromotionRequest{}
	err := json.Unmarshal(data, &promotionUpdate)

	responseData := productModel.PromotionResponseDataPubSubDC2{
		ProductId:  promotionUpdate.ProductId,
		Action:     2,
		Promotions: []*productModel.PromotionResponseDc2{},
		Result:     true,
	}

	if err != nil {
		level.Error(logger).Log("msg", fmt.Sprint("update fail, unmarshal json fail, ", err),
			"meta_data", string(logData))
		responseData.Result = false
		responseData.Message = err.Error()
		return responseData, err
	}

	productId := strconv.FormatInt(promotionUpdate.ProductId, 10)
	promotionsRequest := promotionUpdate.Promotions

	var promotions []model.Promotion
	promotions = parsePromotionUpdateDataFromProto(promotionsRequest)

	promotionResponse, error := m.updatePromotions(productId, promotions)
	responseData.Promotions = promotionResponse

	if error != nil {
		level.Error(logger).Log("msg", fmt.Sprint("add fail, esStore.addPromotions error, ", error),
			"meta_data", string(logData))
		responseData.Result = false
		responseData.Message = "update failed"
	} else {
		level.Info(logger).Log("msg", "update promotion success",
			"meta_data", string(logData))
	}
	return responseData, error
}

func (m *productPromotionDc2) OnPromotionDelete(msg kafka.Message) (productModel.PromotionResponseDataPubSubDC2, error) {
	logger := log.With(m.logger, "prefix", "OnPromotionDelete")
	data := msg.Value

	hashData := fmt.Sprintf("%x", sha1.Sum([]byte(data)))
	ll := map[string]interface{}{"hash": hashData, "data": string(data), "offset": msg.Offset, "partition": msg.Partition}
	logData, _ := json.Marshal(ll)
	level.Info(logger).Log(
		"msg", fmt.Sprint("received data from "+msg.Topic),
		"meta_data", string(logData))

	promotionDelete := productModel.DeleteProductPromotionRequest{}
	err := json.Unmarshal(data, &promotionDelete)

	responseData := productModel.PromotionResponseDataPubSubDC2{
		ProductId:  promotionDelete.ProductId,
		Action:     3,
		Promotions: []*productModel.PromotionResponseDc2{},
		Result:     true,
	}

	if err != nil {
		level.Error(logger).Log("msg", fmt.Sprint("delete fail, unmarshal json fail, ", err),
			"meta_data", string(logData))
		responseData.Result = false
		responseData.Message = err.Error()
		return responseData, err
	}

	productId := strconv.FormatInt(promotionDelete.ProductId, 10)
	promotionsData, err := m.esStorage.GetById(productId)
	var promotionsEs []*es_service.ProductPromotion
	copier.Copy(&promotionsEs, &promotionsData)

	promotionResponse, error := m.deletePromotions(productId, promotionDelete.PromotionIds)
	responseData.Promotions = promotionResponse
	if error != nil {
		level.Error(logger).Log("msg", fmt.Sprint("delete failed. ", error),
			"meta_data", string(logData))
		responseData.Result = false
		responseData.Message = "delete failed"
	} else {
		level.Info(logger).Log("msg", "addPromotions success",
			"meta_data", string(logData))
	}
	return responseData, error
}

func (m *productPromotionDc2) addPromotions(productId string, promotions []model.Promotion) (dataResponse []*productModel.PromotionResponseDc2, err error) {
	promotionsEs, err := m.esStorage.GetById(productId)
	if err != nil {
		return dataResponse, err
	}
	var promotionsDataFinal []model.Promotion
	promotionsDataFinal, dataResponse, err = getPromotionsDataForAddEs(promotions, promotionsEs)

	data := make(map[string]interface{})
	data["promotions"] = promotionsDataFinal

	_, err = m.esStorage.AddPromotions(productId, data)
	if err != nil {
		for k, v := range dataResponse {
			if v.Result == true {
				dataResponse[k].Result = false
				dataResponse[k].Message = "Update elastic failed"
			}
		}
	}
	return dataResponse, err
}

func (m *productPromotionDc2) updatePromotions(productId string, promotions []model.Promotion) (dataResponse []*productModel.PromotionResponseDc2, err error) {
	promotionsEs, err := m.esStorage.GetById(productId)
	if err != nil {
		return dataResponse, err
	}
	var promotionsDataFinal []model.Promotion
	promotionsDataFinal, dataResponse, err = getPromotionsDataForUpdateEs(promotions, promotionsEs)
	if err != nil {
		return dataResponse, err
	}

	data := make(map[string]interface{})
	data["promotions"] = promotionsDataFinal
	_, err = m.esStorage.UpdatePromotions(productId, data)
	if err != nil {
		for k, v := range dataResponse {
			if v.Result == true {
				dataResponse[k].Result = false
				dataResponse[k].Message = "Update elastic failed"
			}
		}
	}
	return dataResponse, err
}

func (m *productPromotionDc2) deletePromotions(productId string, promotionsIds []int64) (dataResponse []*productModel.PromotionResponseDc2, err error) {
	data := make(map[string]interface{})
	promotionsEs, err := m.esStorage.GetById(productId)
	if err != nil {
		return dataResponse, err
	}
	data["promotion_ids"], dataResponse, err = getPromotionsDataForDeleteEs(promotionsEs, promotionsIds)

	_, err = m.esStorage.DeletePromotions(productId, data)
	return dataResponse, err
}

func getPromotionsDataForAddEs(promotionReq []model.Promotion, promotionEs []model.Promotion) (result []model.Promotion, dataResponse []*productModel.PromotionResponseDc2, err error) {
	var dataNeedToAdd []model.Promotion
	dataResponse = []*productModel.PromotionResponseDc2{}
	for _, vReq := range promotionReq {
		isValid := true
		from, _ := types.TimestampProto(vReq.From)
		to, _ := types.TimestampProto(vReq.To)
		temp := &productModel.PromotionResponseDc2{
			Id:         vReq.Id,
			Type:       vReq.Type,
			From:       from,
			To:         to,
			UpdatedAt:  vReq.UpdatedAt,
			FixedPrice: vReq.FixedPrice,
			Result:     true,
			Message:    "",
		}
		if vReq.Type <= 200 {
			isValid = false
		} else {
			for _, vEs := range promotionEs {
				if vReq.Id == vEs.Id {
					isValid = false
					continue
				}
			}
		}
		if isValid {
			dataNeedToAdd = append(dataNeedToAdd, vReq)
		} else {
			temp.Result = false
			temp.Message = "Promotion is invalid"
		}
		dataResponse = append(dataResponse, temp)
	}
	if len(dataNeedToAdd) == 0 {
		return dataNeedToAdd, dataResponse, errors.New("Nothing to change")
	}
	return dataNeedToAdd, dataResponse, nil
}

func getPromotionsDataForUpdateEs(promotionReq []model.Promotion, promotionEs []model.Promotion) (result []model.Promotion, dataResponse []*productModel.PromotionResponseDc2, err error) {
	var dataNeedToUpdate []model.Promotion
	dataResponse = []*productModel.PromotionResponseDc2{}
	for _, vReq := range promotionReq {
		isValid := true
		from, _ := types.TimestampProto(vReq.From)
		to, _ := types.TimestampProto(vReq.To)
		temp := &productModel.PromotionResponseDc2{
			Id:         vReq.Id,
			Type:       vReq.Type,
			From:       from,
			To:         to,
			FixedPrice: vReq.FixedPrice,
			UpdatedAt:  vReq.UpdatedAt,
			Result:     true,
			Message:    "",
		}
		if vReq.Type < 200 {
			isValid = false
		} else {
			for _, vEs := range promotionEs {
				if (vReq.Id == vEs.Id) && (vEs.Type <= 200) && (vReq.UpdatedAt <= vEs.UpdatedAt) {
					isValid = false
					continue
				}
			}
		}
		if isValid {
			dataNeedToUpdate = append(dataNeedToUpdate, vReq)
		} else {
			temp.Result = false
			temp.Message = "Promotion is invalid"
		}
		dataResponse = append(dataResponse, temp)
	}
	if len(dataNeedToUpdate) == 0 {
		return dataNeedToUpdate, dataResponse, errors.New("Nothing to change")
	}
	return dataNeedToUpdate, dataResponse, nil
}

func getPromotionsDataForDeleteEs(promotionEs []model.Promotion, promotionIds []int64) (promotionsIds []int64, dataResponse []*productModel.PromotionResponseDc2, err error) {
	for _, id := range promotionIds {
		isValid := false
		for _, pES := range promotionEs {
			if (id == pES.Id) && (pES.Type > 200) {
				isValid = true
			}
		}
		temp := &productModel.PromotionResponseDc2{
			Id:      id,
			Result:  true,
			Message: "",
		}
		if isValid {
			promotionsIds = append(promotionsIds, id)
		} else {
			temp.Result = false
			temp.Message = "Promotion is invalid"
		}
		dataResponse = append(dataResponse, temp)
	}
	if len(promotionsIds) == 0 {
		return promotionsIds, dataResponse, errors.New("promotion data not found")
	}
	return promotionsIds, dataResponse, nil
}

func parsePromotionDataFromProto(promotionsRequest []*es_service.ProductPromotion) (promotions []model.Promotion) {
	for _, v := range promotionsRequest {
		from, _ := types.TimestampFromProto(v.From)
		to, _ := types.TimestampFromProto(v.To)
		promotion := &model.Promotion{
			Id:         v.Id,
			Type:       v.Type,
			From:       from,
			To:         to,
			FixedPrice: v.FixedPrice,
			UpdatedAt:  v.UpdatedAt,
		}
		promotions = append(promotions, *promotion)
	}
	return promotions
}

func parsePromotionUpdateDataFromProto(promotionsRequest []*es_service.ProductPromotion) (promotions []model.Promotion) {
	for _, v := range promotionsRequest {
		from, _ := types.TimestampFromProto(v.From)
		to, _ := types.TimestampFromProto(v.To)
		promotion := &model.Promotion{
			Id:         v.Id,
			Type:       v.Type,
			From:       from,
			To:         to,
			FixedPrice: v.FixedPrice,
			UpdatedAt:  v.UpdatedAt,
		}
		promotions = append(promotions, *promotion)
	}
	return promotions
}
