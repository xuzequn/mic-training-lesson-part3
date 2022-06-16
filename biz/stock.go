package biz

import (
	"context"
	"github.com/go-errors/errors"
	"google.golang.org/protobuf/types/known/emptypb"
	"mic-training-lesson-part3/custom_error"
	"mic-training-lesson-part3/internal"
	"mic-training-lesson-part3/model"
	"mic-training-lesson-part3/proto/pb"
)

type StockService struct {
}

func (s StockService) SetStock(ctx context.Context, req *pb.ProductStockItem) (*emptypb.Empty, error) {
	// 参数校验 1、web层，<1 != ""
	var stock model.Stock

	// req.productId  -> productSrv
	//addr := fmt.Sprintf("%s:%d", internal.AppConf.ConsulConfig.Host, internal.AppConf.ConsulConfig.Port)
	//dialAddr := fmt.Sprintf("consul://%s/%s?wait=14s", addr, "product_srv")
	//conn, err := grpc.Dial(dialAddr,
	//	grpc.WithTransportCredentials(insecure.NewCredentials()),
	//	grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robbin"}`),
	//)
	//if err!= nil{
	//	zap.S().Fatal(err)
	//	panic(err)
	//}
	//pb.
	// product_id in stock
	internal.DB.Where("product_id = ?", req.ProductID).First(&stock)
	if stock.ID < 1 {
		stock.ProductId = req.ProductID
	}
	stock.Num = req.Num
	internal.DB.Save(&stock)
	return &emptypb.Empty{}, nil
}

func (s StockService) StockDetail(ctx context.Context, req *pb.ProductStockItem) (*pb.ProductStockItem, error) {
	var stock model.Stock
	r := internal.DB.Where("product_id = ?", req.ProductID).First(&stock)
	if r.RowsAffected < 1 {
		return nil, errors.New(custom_error.ParamError)
	}
	stockPb := ConvertStockModel2pb(stock)
	return &stockPb, nil
}

func (s StockService) Sell(ctx context.Context, req *pb.SellItem) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (s StockService) BackStock(ctx context.Context, req *pb.SellItem) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}
func ConvertStockModel2pb(item model.Stock) pb.ProductStockItem {
	return pb.ProductStockItem{
		ProductID: item.ProductId,
		Num:       item.Num,
	}
}
