package main

import (
	"context"
	"fmt"
	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2"
	log2 "github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/hashicorp/consul/api"
	"log"
	"os"
	"strconv"
	"userOrder/api/userOrder"
	"userOrder/model"
	"userOrder/utils"
)

type UserOrder struct {
	userOrder.UnimplementedUserOrderServer
}

func (*UserOrder) CreateOrder(ctx context.Context, req *userOrder.Request) (*userOrder.Response, error) {
	//获取到相关数据,插入到数据库
	orderId, err := model.InsertOrder(req.HouseId, req.StartDate, req.EndDate, req.UserName)
	if err != nil {
		return &userOrder.Response{
			Errno:  utils.RECODE_DBERR,
			Errmsg: utils.RecodeText(utils.RECODE_DBERR),
		}, err
	}
	resp := &userOrder.Response{}
	resp.Errno = utils.RECODE_OK
	resp.Errmsg = utils.RecodeText(utils.RECODE_OK)
	var orderData userOrder.OrderData
	orderData.OrderId = strconv.Itoa(orderId)
	resp.Data = &orderData

	return resp, nil
}
func (*UserOrder) GetOrderInfo(ctx context.Context, req *userOrder.GetReq) (*userOrder.GetResp, error) {
	//要根据传入数据获取订单信息   mysql
	respData, err := model.GetOrderInfo(req.UserName, req.Role)
	if err != nil {
		return &userOrder.GetResp{
			Errno:  utils.RECODE_DATAERR,
			Errmsg: utils.RecodeText(utils.RECODE_DATAERR),
		}, err
	}

	resp := &userOrder.GetResp{}
	resp.Errno = utils.RECODE_OK
	resp.Errmsg = utils.RecodeText(utils.RECODE_OK)
	var getData userOrder.GetData
	getData.Orders = respData
	resp.Data = &getData

	return resp, err
}
func (*UserOrder) UpdateStatus(ctx context.Context, req *userOrder.UpdateReq) (*userOrder.UpdateResp, error) {
	////根据传入数据,更新订单状态
	err := model.UpdateStatus(req.Action, req.Id, req.Reason)
	if err != nil {
		return &userOrder.UpdateResp{
			Errno:  utils.RECODE_DATAERR,
			Errmsg: utils.RecodeText(utils.RECODE_DATAERR),
		}, err
	}
	return &userOrder.UpdateResp{
		Errno:  utils.RECODE_OK,
		Errmsg: utils.RecodeText(utils.RECODE_OK),
	}, nil
}

func main() {
	model.InitRedis()
	_ = model.InitDb()

	logger := log2.NewStdLogger(os.Stdout)

	consulClient, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Fatal(err)
	}

	grpcSrv := grpc.NewServer(
		grpc.Address(":12346"),
		grpc.Middleware(
			recovery.Recovery(),
			logging.Server(logger),
		),
	)

	s := &UserOrder{}
	userOrder.RegisterUserOrderServer(grpcSrv, s)

	r := consul.New(consulClient)
	app := kratos.New(
		kratos.Name("userOrder"),
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
