package internal

import (
	"fmt"
	"testing"
)

func TestReg(t *testing.T) {
	err := Reg(AppConf.StockSrvConfig.Host, AppConf.StockSrvConfig.SrvName,
		AppConf.StockSrvConfig.SrvName, AppConf.StockSrvConfig.Port,
		AppConf.StockSrvConfig.Tags)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("注册成功")
	}
}

func TestGetServiceList(t *testing.T) {
	GetServiceList()
}

func TestFilterService(t *testing.T) {
	FilterService("stock_srv")
}
