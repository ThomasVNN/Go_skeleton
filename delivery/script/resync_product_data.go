package script

import (
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"gitlab.thovnn.vn/core/sen-kit/storage/elastic"
	"gitlab.thovnn.vn/core/sen-kit/storage/mongo"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/pkg/client"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/repo"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/storage/esstore"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/storage/mgostore"
	"gitlab.thovnn.vn/dc1/product-discovery/es_service/usecase/command"
	"os"
)

func Resync(logger log.Logger) error {
	//_ = godotenv.Load()
	//connect mongo
	mgoConfig := mongo.New("es")
	mgoSess, err := mgoConfig.DB()
	if err != nil {
		_ = level.Error(logger).Log("msg", fmt.Sprintf("not connect to mongoDB %v %v", mgoConfig.String(), err.Error()))
		return err
	}

	defer mgoSess.Close()
	mgoStore := mgostore.New(mgoSess)
	variantStore := mgostore.NewVariantStore(mgoSess)
	//connect mongo Log
	logMgoSess, err := mgoConfig.DB()
	if err != nil {
		_ = level.Error(logger).Log("msg", fmt.Sprintf("not connect to log mongoDB %v", err.Error()))
		return err
	}
	defer logMgoSess.Close()
	logStore := mgostore.NewLogStore(mgoSess)
	//End
	//connect esStore
	esConfig := elastic.New("")
	esStore := esstore.New(esConfig.Client(), esConfig.ESIndex, logger)
	//get category service endpoint
	categorySrvUrl := os.Getenv("CATEGORY_GRPC_ENDPOINT")
	if categorySrvUrl == "" {
		_ = level.Error(logger).Log("msg", "CATEGORY_GRPC_ENDPOINT is empty")
		return err
	}
	// Init GRPC Client
	grpcClient := client.NewGRPCClient(logger)
	defer grpcClient.Close()
	cateRepo := repo.NewCategoryServiceClient(categorySrvUrl, grpcClient)

	commands := command.NewProductResyncCommand(esStore, mgoStore, variantStore, nil, logStore, cateRepo, logger)
	err = commands.MultiSync()
	if err != nil {
		_ = level.Error(logger).Log("msg", fmt.Sprintf("Resync cron failed %v", err.Error()))
		return err
	}
	return err
}
