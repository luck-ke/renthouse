package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/hashicorp/consul/api"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111" // 引入sms
	"math/rand"
	"os"
	register "register/api"
	"register/model"
	"register/utils"
	"time"
)

type Register struct {
	register.UnimplementedRegisterServer
}

func (e *Register) SmsCode(ctx context.Context, req *register.Request) (*register.Response, error) {
	//验证图片验证码是否输入正确  从redis中获取到存储的图片验证码
	rnd, err := model.GetImgCode(req.Uuid)
	if err != nil {
		fmt.Println("redis 获取验证码错误", err)
		return &register.Response{
			Errno:  utils.RECODE_NODATA,
			Errmsg: utils.RecodeText(utils.RECODE_NODATA),
		}, err
	}

	//判断输入的图片验证码是否正确
	if req.Text != rnd {
		//返回自定义的error数据
		fmt.Println("图片验证码错误", err)
		return &register.Response{
			Errno:  utils.RECODE_DATAERR,
			Errmsg: utils.RecodeText(utils.RECODE_DATAERR),
		}, err
	}

	//发送短信
	credential := common.NewCredential(
		"AKIDjJm5D9yfBPl1Ad8XmzJvOgngLo6HPOAj",
		"PYfdTLfv2uaAYjlstoaEoQ3c8hAUl7Ui",
	)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.ReqMethod = "POST"
	cpf.HttpProfile.Endpoint = "sms.tencentcloudapi.com"
	cpf.SignMethod = "HmacSHA1"
	client, _ := sms.NewClient(credential, "ap-guangzhou", cpf)
	request := sms.NewSendSmsRequest()
	request.SmsSdkAppId = common.StringPtr("1400741476")
	request.SignName = common.StringPtr("开发者成长足迹公众号")
	request.TemplateId = common.StringPtr("1554203")
	//获取6位数随机码
	myRnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	vcode := fmt.Sprintf("%06d", myRnd.Int31n(1000000))

	fmt.Println("短信验证码：", vcode)
	request.TemplateParamSet = common.StringPtrs([]string{vcode})
	phone := "+86" + req.Mobile
	request.PhoneNumberSet = common.StringPtrs([]string{phone})
	request.SessionContext = common.StringPtr("")
	request.ExtendCode = common.StringPtr("")
	request.SenderId = common.StringPtr("")
	_, err = client.SendSms(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		fmt.Printf("An API error has returned: %s", err)
		return &register.Response{
			Errno:  utils.RECODE_SMSERR,
			Errmsg: utils.RecodeText(utils.RECODE_SMSERR),
		}, err
	}
	if err != nil {
		fmt.Println("client.SendSms err", err)
		return &register.Response{
			Errno:  utils.RECODE_SMSERR,
			Errmsg: utils.RecodeText(utils.RECODE_SMSERR),
		}, err
		//panic(err)
	}
	//b, _ := json.Marshal(response.Response)

	err = model.SaveSmsCode(req.Mobile, vcode)
	if err != nil {
		fmt.Println("数据库存储失败", err)
		return &register.Response{
			Errno:  utils.RECODE_DATAERR,
			Errmsg: utils.RecodeText(utils.RECODE_DATAEXIST),
		}, err
	}

	return &register.Response{
		Errno:  utils.RECODE_OK,
		Errmsg: utils.RecodeText(utils.RECODE_OK),
	}, nil
}

//注册
func (e *Register) Register(ctx context.Context, req *register.RegRequest) (*register.RegResponse, error) {
	//校验短信验证码会否正确
	smsCode, err := model.GetSmsCode(req.Mobile)
	if err != nil {
		fmt.Println("验证码查询错误", err)
		return &register.RegResponse{
			Errno:  utils.RECODE_DATAERR,
			Errmsg: utils.RecodeText(utils.RECODE_DATAERR),
		}, err
	}

	if smsCode != req.SmsCode {
		fmt.Println("验证码错误", err)
		return &register.RegResponse{
			Errno:  utils.RECODE_SMSERR,
			Errmsg: utils.RecodeText(utils.RECODE_SMSERR),
		}, err
	}

	//存储用户数据到Mysql上
	//给密码加密
	pwdByte := sha256.Sum256([]byte(req.Password))
	pwd_hash := string(pwdByte[:])
	//要把sha256得到的数据转换之后存储  转换16进制的
	pwdHash := fmt.Sprintf("%x", pwd_hash)

	err = model.SaveUser(req.Mobile, pwdHash)
	if err != nil {
		fmt.Println("存储密码错误", err)
		return &register.RegResponse{
			Errno:  utils.RECODE_DBERR,
			Errmsg: utils.RecodeText(utils.RECODE_DBERR),
		}, err
	}

	return &register.RegResponse{
		Errno:  utils.RECODE_OK,
		Errmsg: utils.RecodeText(utils.RECODE_OK),
	}, nil
}

//登录
func (e *Register) Login(ctx context.Context, req *register.RegRequest) (*register.RegResponse, error) {

	//查询输入手机号和密码是否正确  mysql
	//给密码加密
	pwdByte := sha256.Sum256([]byte(req.Password))
	pwd_hash := string(pwdByte[:])
	//要把sha256得到的数据转换之后存储  转换16进制的
	pwdHash := fmt.Sprintf("%x", pwd_hash)

	user, err := model.CheckUser(req.Mobile, pwdHash)
	if err != nil {
		return &register.RegResponse{
			Errno:  utils.RECODE_LOGINERR,
			Errmsg: utils.RecodeText(utils.RECODE_LOGINERR),
		}, err
	}

	//查询成功  登录成功  把用户名存储到session中  把用户名传给web端
	return &register.RegResponse{
		Errno:  utils.RECODE_OK,
		Errmsg: utils.RecodeText(utils.RECODE_OK),
		Name:   user.Name,
	}, nil
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
		grpc.Address(":12342"),
		grpc.Middleware(
			recovery.Recovery(),
			logging.Server(logger),
		),
	)

	s := &Register{}
	register.RegisterRegisterServer(grpcSrv, s)

	r := consul.New(consulClient)
	app := kratos.New(
		kratos.Name("register"),
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
