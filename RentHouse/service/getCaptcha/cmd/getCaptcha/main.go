package main

import (
	"context"
	"encoding/json"
	"fmt"
	getCaptcha "getCaptcha/api"
	"getCaptcha/model"
	"getCaptcha/utils"
	"github.com/afocus/captcha"
	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/hashicorp/consul/api"
	"image/color"
	"os"
)

type Captcha struct {
	getCaptcha.UnimplementedGreeterServer
}

func (Captcha) Call(ctx context.Context, req *getCaptcha.HelloRequest) (*getCaptcha.HelloReply, error) {

	cap := captcha.New()
	//设置字体
	cap.SetFont("./configs/comic.ttf")
	//设置验证码大小
	cap.SetSize(128, 64)
	//设置干扰强度
	cap.SetDisturbance(captcha.NORMAL)
	//设置前景色
	cap.SetFrontColor(color.RGBA{255, 255, 255, 255})
	//设置背景色
	cap.SetBkgColor(color.RGBA{255, 0, 0, 255}, color.RGBA{0, 0, 255, 255})
	img, rnd := cap.Create(4, captcha.ALL)

	//存储验证码   redis
	err := model.SaveImgRnd(req.Uuid, rnd)
	fmt.Println("图片验证码：", rnd)
	if err != nil {
		return &getCaptcha.HelloReply{
			Errno:  utils.RECODE_DATAERR,
			Errmsg: utils.RecodeText(utils.RECODE_DATAERR),
			Data:   nil,
		}, err
	}
	
	imgBuf, _ := json.Marshal(img)

	return &getCaptcha.HelloReply{
		Errno:  utils.RECODE_OK,
		Errmsg: utils.RecodeText(utils.RECODE_OK),
		Data:   imgBuf,
	}, nil
}

func main() {
	logger := log.NewStdLogger(os.Stdout)

	consulClient, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Fatal(err)
	}

	grpcSrv := grpc.NewServer(
		grpc.Address(":12341"),
		grpc.Middleware(
			recovery.Recovery(),
			logging.Server(logger),
		),
	)

	s := &Captcha{}
	getCaptcha.RegisterGreeterServer(grpcSrv, s)

	r := consul.New(consulClient)
	app := kratos.New(
		kratos.Name("getCaptcha"),
		kratos.Server(
			grpcSrv,
			//httpSrv,
		),
		kratos.Registrar(r),
	)

	if err := app.Run(); err != nil {
		fmt.Println("...err:", err)
		log.Fatal(err)
	}
}
