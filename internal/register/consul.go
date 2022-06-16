package register

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"go.uber.org/zap"
	"mic-training-lesson-part3/internal"
)

type IRegister interface {
	Register(name, id string, port int, tags []string) error
	DeRegister(serviceId string) error
}

type ConsulRegistry struct {
	Host string
	Port int
}

func NewConsulRegistry(host string, port int) ConsulRegistry {
	return ConsulRegistry{
		Host: host,
		Port: port,
	}
}

func (cr ConsulRegistry) Register(name, id string, port int, srvType string, tags []string) error {
	defaultConfig := api.DefaultConfig()
	h := cr.Host
	p := cr.Port
	defaultConfig.Address = fmt.Sprintf("%s:%d", h, p)
	client, err := api.NewClient(defaultConfig)
	if err != nil {
		zap.S().Error(err)
		return err
	}
	agentServiceReg := new(api.AgentServiceRegistration)
	agentServiceReg.Address = internal.AppConf.StockSrvConfig.Host
	agentServiceReg.Port = port
	agentServiceReg.ID = id
	agentServiceReg.Name = name
	agentServiceReg.Tags = tags
	if srvType == "grpc" {
		checkAddr := fmt.Sprintf("%s:%d", internal.AppConf.StockSrvConfig.Host, port)
		check := api.AgentServiceCheck{
			GRPC:                           checkAddr,
			Timeout:                        "3s",
			Interval:                       "1s",
			DeregisterCriticalServiceAfter: "5s",
		}
		agentServiceReg.Check = &check
	} else if srvType == "restful" {
		serverAdder := fmt.Sprintf("http://%s:%d/health", internal.AppConf.StockSrvConfig.Host,
			internal.AppConf.StockSrvConfig.Port)
		check := api.AgentServiceCheck{
			HTTP:                           serverAdder,
			Timeout:                        "3s",
			Interval:                       "1s",
			DeregisterCriticalServiceAfter: "5s",
		}
		agentServiceReg.Check = &check
	} else {
		zap.S().Error("服务属性错误")
		return err
	}

	err = client.Agent().ServiceRegister(agentServiceReg)
	if err != nil {
		zap.S().Error(err)
		return err
	}
	return nil
}

func (cr ConsulRegistry) DeRegister(serviceId string) error {
	defaultConfig := api.DefaultConfig()
	h := internal.AppConf.ConsulConfig.Host
	p := internal.AppConf.ConsulConfig.Port
	defaultConfig.Address = fmt.Sprintf("%s:%d", h, p)
	client, err := api.NewClient(defaultConfig)
	if err != nil {
		zap.S().Error(err)
		return err
	}
	return client.Agent().ServiceDeregister(serviceId)
}
