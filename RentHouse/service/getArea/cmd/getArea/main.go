package main

import (
	"context"
	"fmt"
	getArea "getArea/api/area"
	"getArea/model"
	"getArea/utils"
	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/hashicorp/consul/api"
	"os"
)

type GetArea struct {
	getArea.UnimplementedGetAreaServer
}

func (*GetArea) GetAreaSer(ctx context.Context, req *getArea.Request) (*getArea.Response, error) {
	//获取数据并返回给调用者
	areas, err := model.GetArea()
	if err != nil {
		return &getArea.Response{
			Errno:  utils.RECODE_DBERR,
			Errmsg: utils.RecodeText(utils.RECODE_DBERR),
		}, err
	}

	response := &getArea.Response{}
	for _, v := range areas {
		var areaInfo getArea.AreaInfo
		areaInfo.Aid = int32(v.Id)
		areaInfo.Aname = v.Name

		response.Data = append(response.Data, &areaInfo)
	}
	response.Errno = utils.RECODE_OK
	response.Errmsg = utils.RecodeText(utils.RECODE_OK)
	//return &getArea.Response{
	//	Errno:  utils.RECODE_OK,
	//	Errmsg: utils.RecodeText(utils.RECODE_OK),
	//	Data: response,
	//}, nil
	return response, nil
}

func main() {

	model.InitRedis()
	model.InitDb()

	logger := log.NewStdLogger(os.Stdout)

	consulClient, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Fatal(err)
	}

	grpcSrv := grpc.NewServer(
		grpc.Address(":12343"),
		grpc.Middleware(
			recovery.Recovery(),
			logging.Server(logger),
		),
	)

	s := &GetArea{}
	getArea.RegisterGetAreaServer(grpcSrv, s)

	r := consul.New(consulClient)
	app := kratos.New(
		kratos.Name("getArea"),
		kratos.Server(
			grpcSrv,
		),
		kratos.Registrar(r),
	)

	if err := app.Run(); err != nil {
		fmt.Println("Run err")
		log.Fatal(err)
	}

}
