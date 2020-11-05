package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/segmentio/kafka-go"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/model"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/storage/esstore"
)

const (
	tick       = 1
	maxPackage = 750
	maxTime    = 10 * time.Second
)

type totalProductScoreCommand struct {
	esStorage esstore.ESStore
	log       log.Logger
	jobQueue  chan map[string]interface{}

	esData   []map[string]interface{}
	lastTime time.Time
	sig      chan int
}

func NewProductTotalScoreCommand(esStorage esstore.ESStore, log log.Logger) model.ProductTotalScoreEvents {
	p := &totalProductScoreCommand{
		esStorage: esStorage,
		log:       log,
		jobQueue:  make(chan map[string]interface{}, 2000),
	}

	go p.UpdateWorker()

	return p
}

func (p *totalProductScoreCommand) logDebugData(prefix string, obj interface{}) {
	_ = level.Info(p.log).Log("msg", func() string {
		str, _ := json.Marshal(obj)
		return fmt.Sprintf("%s: %s", prefix, string(str))
	}())
}

func (p *totalProductScoreCommand) OnProductUpdate(msg kafka.Message) error {
	var products []model.ProductES
	var productsFields []map[string]interface{}


	err := json.Unmarshal(msg.Value, &products)
	if err != nil || products == nil {
		return nil
	}
	err = json.Unmarshal(msg.Value, &productsFields)
	if err != nil || productsFields == nil {
		return nil
	}

	updateData := make([]map[string]interface{}, 0, len(products))
	productIDs := make([]string, 0, len(products))

	for i := 0; i < len(products); i++ {
		data := GetValidUpdateData(products[i], productsFields[i])
		updateData = append(updateData, data)
		productIDs = append(productIDs, fmt.Sprintf("%d", products[i].ProductId))
	}

	res, err := p.esStorage.Updates(updateData, productIDs)

	if res != nil {
		p.logDebugData("OnProductUpdate-kafka", res)

		if res.Errors {
			strErr, _ := json.Marshal(res.Items)
			newErr := errors.New(string(strErr))
			p.logDebugData("OnProductUpdate-kafka-error", newErr)
			return newErr
		}
	}
	if err != nil {
		p.logDebugData("OnProductUpdate-kafka-error", err)
	}

	return err
}

func getValidField(t interface{}, v interface{}) interface{} {
	var result = make(map[string]interface{})
	u, ok := v.(map[string]interface{})
	if !ok {
		return t
	}
	rt, rv := reflect.TypeOf(t), reflect.ValueOf(t)
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		jsonKey := field.Tag.Get("json")
		if _, ok := u[jsonKey]; ok {
			result[jsonKey] = getValidField(rv.Field(i).Interface(), u[jsonKey])
		}
	}

	return result
}

// GetValidUpdateData extract needed fields
func GetValidUpdateData(t interface{}, updateFields map[string]interface{}) map[string]interface{} {

	result, ok := getValidField(t, updateFields).(map[string]interface{})

	if !ok {
		return nil
	}

	return result
}

func (p *totalProductScoreCommand) tryUpdate() {
	dataLen := len(p.esData)

	if dataLen > 0 {
		if dataLen >= maxPackage || p.checkTime() {
			p.bulkUpdate()
		}
	} else {
		p.resetTime()
	}
}

func (p *totalProductScoreCommand) bulkUpdate() {
	res, err := p.esStorage.UpdateBulk(p.esData)
	if res != nil {
		str, _ := json.Marshal(res)
		_ = level.Info(p.log).Log("msg", fmt.Sprintf("Update result: %v", string(str)))
	}
	if err != nil {
		_ = level.Error(p.log).Log("msg", fmt.Sprint("Update error", err))
		return
	}

	p.esData = p.esData[:0]

	p.resetTime()
}

func (p *totalProductScoreCommand) checkTime() bool {
	return time.Since(p.lastTime) > maxTime
}

func (p *totalProductScoreCommand) resetTime() {
	p.lastTime = time.Now()
}

func (p *totalProductScoreCommand) UpdateWorker() {
	_ = level.Info(p.log).Log("msg", "Start update worker")

	p.esData = make([]map[string]interface{}, 0, 2000)
	p.sig = make(chan int)
	p.lastTime = time.Now()
	go func() {
		for _ = range p.sig {
			p.tryUpdate()
		}
	}()

	go func() {
		for {
			time.Sleep(time.Second * tick)
			p.sig <- 1
		}
	}()

	for prod := range p.jobQueue {
		p.esData = append(p.esData, prod)
		p.sig <- 1
	}
}
