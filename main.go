package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"gorm.io/gorm"
	"mic-training-lesson-part3/biz"
	"mic-training-lesson-part3/internal"
	"mic-training-lesson-part3/internal/register"
	"mic-training-lesson-part3/model"
	"mic-training-lesson-part3/proto/pb"
	"mic-training-lesson-part3/util"
	"net"
	"os"
	"os/signal"
	"syscall"
)

var (
	consulRegistry register.ConsulRegistry
	randomId       string
)

func init() {
	internal.InitDB()
	randomPort := util.GenRandomPort()
	if !internal.AppConf.Debug {
		internal.AppConf.StockSrvConfig.Port = randomPort
	}
	randomId = uuid.NewV4().String()
	consulRegistry = register.NewConsulRegistry(internal.AppConf.ConsulConfig.Host,
		int(internal.AppConf.ConsulConfig.Port))
}

func main() {
	/*
			1、生成proto对应的文件
			2、简历biz目录，生成对应接口。
		    3、拷贝之前main文件的函数、查缺补漏
	*/

	//port := util.GenRandomPort()
	port := internal.AppConf.StockSrvConfig.Port
	addr := fmt.Sprintf("%s:%d", internal.AppConf.StockSrvConfig.Host, port)
	// 将定义的对象注册grpc服务
	server := grpc.NewServer()
	pb.RegisterStockServiceServer(server, &biz.StockService{})
	// 启动服务监听
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		zap.S().Error("stock_srv 启动异常" + err.Error())
		panic(err)
	}
	// grpc 服务的健康检查  注册服务健康检查  启动的grpc  健康检查方法
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())
	// 注册服务
	err = consulRegistry.Register(internal.AppConf.StockSrvConfig.SrvName, randomId,
		internal.AppConf.StockSrvConfig.Port, internal.AppConf.StockSrvConfig.SrvType, internal.AppConf.StockSrvConfig.Tags)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(fmt.Sprintf("%s启动在%d", randomId, port))
	mqAddr := "127.0.0.1:9876"
	pushConsumer, _ := rocketmq.NewPushConsumer(
		consumer.WithNameServer([]string{mqAddr}),
		consumer.WithGroupName("HappyStockGroup"),
	)
	pushConsumer.Subscribe("Happy_BackStockTopic", consumer.MessageSelector{},
		func(ctx context.Context,
			messageExt ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
			for i := range messageExt {
				var order model.Order
				err = json.Unmarshal(messageExt[i].Body, &order)
				if err != nil {
					zap.S().Error("Unmarshal Error:" + err.Error())
					return consumer.ConsumeRetryLater, nil
				}
				tx := internal.DB.Begin()
				var detail model.StockItemDetail
				r := tx.Where(&model.StockItemDetail{
					OrderNo: order.OrderNum,
					Status:  model.HasSell,
				}).First(&detail)
				if r.RowsAffected < 1 {
					return consumer.ConsumeSuccess, nil
				}
				for _, item := range detail.DetailList {
					ret := tx.Model(&model.Stock{ProductId: item.ProductId}).Update(
						"num", gorm.Expr("num+?", item.Num),
					)
					if ret.RowsAffected < 1 {
						return consumer.ConsumeRetryLater, nil
					}
				}
				result := tx.Model(&model.StockItemDetail{}).
					Where(&model.StockItemDetail{OrderNo: order.OrderNum}).
					Update("status", model.HasBack)
				if result.RowsAffected < 1 {
					tx.Rollback()
					return consumer.ConsumeRetryLater, nil
				}
				tx.Commit()
				return consumer.ConsumeSuccess, nil
			}
			return consumer.ConsumeSuccess, nil
		})
	//// 在consul 注册grpc 服务。
	//// consul的相关配置
	//defaultConfig := api.DefaultConfig()
	//// 配置地址
	//defaultConfig.Address = fmt.Sprintf("%s:%d",
	//	internal.AppConf.ConsulConfig.Host,
	//	internal.AppConf.ConsulConfig.Port)
	//// 创建consul的客户端
	//client, err := api.NewClient(defaultConfig)
	//if err != nil {
	//	panic(err)
	//}
	//// 生成健康检查对象
	//checkAddr := fmt.Sprintf("%s:%d",
	//	internal.AppConf.StockSrvConfig.Host,
	//	port)
	//check := api.AgentServiceCheck{
	//	GRPC:                           checkAddr,
	//	Timeout:                        "3s",
	//	Interval:                       "1s",
	//	DeregisterCriticalServiceAfter: "5s",
	//}
	//randUUID := uuid.NewV4().String()
	//reg := api.AgentServiceRegistration{
	//	Name:    internal.AppConf.StockSrvConfig.SrvName,
	//	Address: internal.AppConf.StockSrvConfig.Host,
	//	ID:      randUUID,
	//	Port:    port,
	//	Tags:    internal.AppConf.StockSrvConfig.Tags,
	//	Check:   &check,
	//}
	//// 注册grpc服务
	//err = client.Agent().ServiceRegister(&reg)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(fmt.Sprintf("%s启动在%d", randUUID, port))
	//err = server.Serve(listen)
	//if err != nil {
	//	zap.S().Error("stock_srv 启动异常" + err.Error())
	//	panic(err)
	//}
	//zap.S().Info("stock_srv 启动成功")
	go func() {
		err = server.Serve(listen)
		if err != nil {
			zap.S().Panic(addr + "启动失败" + err.Error())
			fmt.Println(addr + "启动失败" + err.Error())
		} else {
			zap.S().Info(addr + "启动成功")
		}
	}()
	q := make(chan os.Signal)
	signal.Notify(q, syscall.SIGINT, syscall.SIGTERM)
	<-q
	err = consulRegistry.DeRegister(randomId)
	if err != nil {
		zap.S().Panic("注销失败" + randomId + ":" + err.Error())
	} else {
		zap.S().Info("注销成功" + randomId)
	}
}
