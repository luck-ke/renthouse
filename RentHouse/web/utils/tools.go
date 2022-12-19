package utils

import (
	"fmt"
	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/hashicorp/consul/api"
)

func GetConsul() *consul.Registry {
	consulCli, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		fmt.Println("NewClient err", err)
		return nil
	}
	return consul.New(consulCli)
}
