package biz

import (
	"context"
	"fmt"
	_ "github.com/mbobakov/grpc-consul-resolver" // 让 grpc 可以解析consul协议
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"mic-training-lesson-part3/internal"
	"mic-training-lesson-part3/model"
	"mic-training-lesson-part3/proto/pb"
	"sync"
	"testing"
)

var (
	client pb.StockServiceClient
)

func init() {
	addr := fmt.Sprintf("%s:%d", internal.AppConf.ConsulConfig.Host, internal.AppConf.ConsulConfig.Port)
	dialAddr := fmt.Sprintf("consul://%s/%s?wait=14s", addr, internal.AppConf.StockSrvConfig.SrvName)
	conn, err := grpc.Dial(dialAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robbin"}`),
	)
	if err != nil {
		zap.S().Fatal(err)
		panic(err)
	}
	client = pb.NewStockServiceClient(conn)
}

func TestStockService_SetStock(t *testing.T) {
	_, err := client.SetStock(context.Background(), &pb.ProductStockItem{
		ProductID: 1,
		Num:       2,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestStockService_StockDetail(t *testing.T) {
	r, err := client.StockDetail(context.Background(), &pb.ProductStockItem{
		ProductID: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(r)
}

var wg sync.WaitGroup

func TestStockService_Sell(t *testing.T) {
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			var itemList []*pb.ProductStockItem
			item := &pb.ProductStockItem{
				ProductID: 1,
				Num:       1,
			}
			itemList = append(itemList, item)
			sellItem := &pb.SellItem{
				StockItemList: itemList,
			}
			res, err := client.Sell(context.Background(), sellItem)
			if err != nil {
				t.Fatal(err)
			}
			fmt.Println(res)
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestStockService_BackStock(t *testing.T) {
	var itemList []*pb.ProductStockItem
	item := &pb.ProductStockItem{
		ProductID: 1,
		Num:       18,
	}
	itemList = append(itemList, item)
	res, err := client.BackStock(context.Background(), &pb.SellItem{StockItemList: itemList})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(res)
}
func TestCreateStockItemDetail(t *testing.T) {
	item := model.StockItemDetail{
		OrderNo: "12345",
		Status:  model.HasSell,
		DetailList: []model.ProductDetail{
			{ProductId: 1, Num: 6},
			{ProductId: 2, Num: 8},
		},
	}
	internal.DB.Save(&item)
}
func FindCreateStockItemDetail(t *testing.T) {
	var itemDetail model.StockItemDetail
	internal.DB.Model(model.StockItemDetail{}).Where(model.StockItemDetail{
		OrderNo: "12345",
	}).First(&itemDetail)
	fmt.Println(itemDetail)
}
