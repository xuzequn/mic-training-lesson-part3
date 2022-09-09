package biz

import (
	"context"
	"fmt"
	"github.com/go-errors/errors"
	"github.com/go-redsync/redsync/v4"
	"google.golang.org/protobuf/types/known/emptypb"
	"mic-training-lesson-part3/custom_error"
	"mic-training-lesson-part3/internal"
	"mic-training-lesson-part3/model"
	"mic-training-lesson-part3/proto/pb"
	"sync"
	"time"
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

var m sync.Mutex

func (s StockService) Sell(ctx context.Context, req *pb.SellItem) (*emptypb.Empty, error) {
	// 面试必问：mutex锁- 》 悲观锁- 》 乐观锁- 》分布式锁， 重要 重要 重要
	tx := internal.DB.Begin() //  事务控制
	//m.Lock()                  // 为了保证并发安全，添加并发互斥锁，但是性能差，如果100w人来竞争，考虑分布式锁
	//defer m.Unlock()
	for _, item := range req.StockItemList {
		var stock model.Stock
		// 乐观锁事务

		// 分布式锁
		mutex := internal.RedSync.NewMutex(fmt.Sprintf("product_%d", item.ProductID), redsync.WithExpiry(16*time.Second))
		err := mutex.Lock()
		if err != nil {
			return nil, errors.New(custom_error.RedisLockErr)
		}
		//r := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("product_id = ?", item.ProductID).First(&stock) // 悲观锁 独占锁
		r := internal.DB.Where("product_id = ?", item.ProductID).First(&stock)
		if r.RowsAffected == 0 {
			tx.Rollback()
			return nil, errors.New(custom_error.ProductNotFound)
		}
		if stock.Num < item.Num {
			tx.Rollback()
			return nil, errors.New(custom_error.StockNotEnough)
		}
		stock.Num -= item.Num
		tx.Save(&stock)
		ok, err := mutex.Unlock()
		if !ok || err != nil {
			return nil, errors.New(custom_error.StockNotEnough)
		}
		// update num = 88 and version = 1 where id =1 and version =0
		//  更新选定字段，保证有零值。
		// 乐观锁
		//r = tx.Where(&model.Stock{}).Select("num").Where("product_id=? and version = ?",
		//	item.ProductID, stock.Version).Updates(model.Stock{
		//	Num:     stock.Num,
		//	Version: stock.Version + 1})
		//if r.RowsAffected == 0 {
		//	zap.S().Info("扣减库存失败")
		//} else {
		//	break
		//}
	}
	tx.Commit()
	return &emptypb.Empty{}, nil
}

func (s StockService) BackStock(ctx context.Context, req *pb.SellItem) (*emptypb.Empty, error) {
	/*
		什么时候触发回滚？
		超时怎么办？
		定点创建失败？
		手动归还
	*/
	tx := internal.DB.Begin()
	for _, item := range req.StockItemList {
		var stock model.Stock
		r := internal.DB.Where("product_id=?", item.ProductID).First(&stock)
		if r.RowsAffected < 1 {
			tx.Rollback()
			return nil, errors.New(custom_error.ProductNotFound)
		}
		stock.Num += item.Num
		tx.Save(&stock)
	}
	tx.Commit()
	return &emptypb.Empty{}, nil
}
func ConvertStockModel2pb(item model.Stock) pb.ProductStockItem {
	return pb.ProductStockItem{
		ProductID: item.ProductId,
		Num:       item.Num,
	}
}
