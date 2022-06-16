package biz

import (
	"context"
	"fmt"
	_ "github.com/mbobakov/grpc-consul-resolver" // 让 grpc 可以解析consul协议
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"mic-training-lesson-part3/internal"
	"mic-training-lesson-part3/proto/pb"
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
