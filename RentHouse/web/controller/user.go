package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/afocus/captcha"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/hashicorp/consul/api"
	"image/png"
	"path"
	getArea "test05kratos/web/api/area"
	getCaptcha2 "test05kratos/web/api/getCaptcha"
	"test05kratos/web/api/house"
	"test05kratos/web/api/register"
	"test05kratos/web/api/user"
	"test05kratos/web/api/userOrder"
	"test05kratos/web/model"
	"test05kratos/web/utils"
	"time"

	"net/http"
)

func GetSession(ctx *gin.Context) {
	//构造未登录
	resp := make(map[string]interface{})

	//查询session数据,如果查询到了,返回数据
	//初始化session对象
	session := sessions.Default(ctx)

	//获取session数据
	userName := session.Get("userName")
	if userName == nil {
		resp["errno"] = utils.RECODE_LOGINERR
		resp["errmsg"] = utils.RecodeText(utils.RECODE_LOGINERR)
	} else {
		resp["errno"] = utils.RECODE_OK
		resp["errmsg"] = utils.RecodeText(utils.RECODE_OK)

		//可以是结构体,可以是map
		tempMap := make(map[string]interface{})
		tempMap["name"] = userName.(string)
		resp["data"] = tempMap
	}

	ctx.JSON(http.StatusOK, resp)
}

//获取图片验证码
func GetImageCd(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	consulCli, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		fmt.Println("NewClient err", err)
		return
	}
	r := consul.New(consulCli)
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("discovery:///getCaptcha"),
		grpc.WithDiscovery(r),
	)
	if err != nil {
		fmt.Println("grpc.Dial err", err)
		return
	}
	defer conn.Close()
	gClient := getCaptcha2.NewGreeterClient(conn)
	ctx1, cel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cel()
	resp, err := gClient.Call(ctx1, &getCaptcha2.HelloRequest{
		Uuid: uuid,
	})
	if err != nil {
		fmt.Println("获取远端数据失败", err)
		ctx.JSON(http.StatusOK, resp)
		return
	}
	var img captcha.Image
	json.Unmarshal(resp.Data, &img)
	png.Encode(ctx.Writer, img)
}

//发送短信
func GetSmscd(ctx *gin.Context) {
	//获取数据
	mobile := ctx.Param("mobile")
	//获取输入的图片验证码
	text := ctx.Query("text")
	//获取验证码图片的uuid
	uuid := ctx.Query("id")

	//校验数据
	if mobile == "" || text == "" || uuid == "" {
		fmt.Println("传入数据不完整")
		return
	}

	consulCli, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		fmt.Println("NewClient err", err)
		return
	}
	r := consul.New(consulCli)
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("discovery:///register"),
		grpc.WithDiscovery(r),
	)
	if err != nil {
		fmt.Println("grpc.Dial err", err)
		return
	}
	defer conn.Close()
	gClient := register.NewRegisterClient(conn)
	ctx1, cel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cel()
	resp, err := gClient.SmsCode(ctx1, &register.Request{
		Mobile: mobile,
		Uuid:   uuid,
		Text:   text,
	})

	if err != nil {
		fmt.Println("调用远程服务错误", err)
	}

	ctx.JSON(http.StatusOK, resp)
}

type RegUser struct {
	Mobile   string `json:"mobile"`
	PassWord string `json:"password"`
	SmsCode  string `json:"sms_code"`
}

//注册
func PostRet(ctx *gin.Context) {
	var reg RegUser
	err := ctx.Bind(&reg)

	//校验数据
	if err != nil {
		fmt.Println("获取前段传递数据失败")
		return
	}

	r := utils.GetConsul()
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("discovery:///register"),
		grpc.WithDiscovery(r),
	)
	if err != nil {
		fmt.Println("grpc.Dial err", err)
		return
	}
	defer conn.Close()
	gClient := register.NewRegisterClient(conn)
	ctx1, cel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cel()
	resp, err := gClient.Register(ctx1, &register.RegRequest{
		Mobile:   reg.Mobile,
		Password: reg.PassWord,
		SmsCode:  reg.SmsCode,
	})
	if err != nil {
		fmt.Println("调用远程服务错误", err)
	}

	//返回数据
	ctx.JSON(http.StatusOK, resp)
}

//获取地区信息
func GetArea(ctx *gin.Context) {
	r := utils.GetConsul()
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("discovery:///getArea"),
		grpc.WithDiscovery(r),
	)
	if err != nil {
		fmt.Println("grpc.Dial err", err)
		return
	}
	defer conn.Close()
	gClient := getArea.NewGetAreaClient(conn)
	ctx1, cel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cel()
	resp, err := gClient.GetAreaSer(ctx1, &getArea.Request{})
	if err != nil {
		fmt.Println("调用远程服务错误", err)
	}
	//返回数据
	ctx.JSON(http.StatusOK, resp)
}

type LogUser struct {
	Mobile   string `json:"mobile"`
	PassWord string `json:"password"`
}

//登录
func PostLogin(ctx *gin.Context) {
	var log LogUser
	err := ctx.Bind(&log)
	//校验数据
	if err != nil {
		fmt.Println("获取数据失败")
		return
	}
	//处理数据   把业务放在为服务中
	//初始化客户端
	r := utils.GetConsul()
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("discovery:///register"),
		grpc.WithDiscovery(r),
	)
	if err != nil {
		fmt.Println("grpc.Dial err", err)
		return
	}
	defer conn.Close()
	gClient := register.NewRegisterClient(conn)
	ctx1, cel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cel()

	resp, err := gClient.Login(ctx1, &register.RegRequest{Mobile: log.Mobile, Password: log.PassWord})

	if err != nil {
		fmt.Println("调用远程服务错误", err)
	}
	//返回数据  存储session  并返回数据给web端
	session := sessions.Default(ctx)
	session.Set("userName", resp.Name)
	session.Save()
	ctx.JSON(http.StatusOK, resp)
}

//退出登录，删除session
func DeleteSession(ctx *gin.Context) {
	//删除session
	session := sessions.Default(ctx)

	//删除session
	session.Delete("userName")
	err := session.Save()
	if err != nil {
		fmt.Println("session删除失败", err)
	}

	//fmt.Println("控制器函数执行....")

	resp := make(map[string]interface{})
	defer ctx.JSON(http.StatusOK, resp)
	if err != nil {
		resp["errno"] = utils.RECODE_DATAERR
		resp["errmsg"] = utils.RecodeText(utils.RECODE_DATAERR)
		return
	}

	resp["errno"] = utils.RECODE_OK
	resp["errmsg"] = utils.RecodeText(utils.RECODE_OK)
}

//上传用户头像
func PostAvatar(ctx *gin.Context) {
	//获取数据  获取图片  文件流  文件头  err
	fileHeader, err := ctx.FormFile("avatar")

	//检验数据
	if err != nil {
		fmt.Println("文件上传失败")
		return
	}

	//三种校验 大小,类型,防止重名
	if fileHeader.Size > 50000000 {
		fmt.Println("文件过大,请重新选择")
		return
	}

	fileExt := path.Ext(fileHeader.Filename)
	if fileExt != ".png" && fileExt != ".jpg" {
		fmt.Println("文件类型错误,请重新选择")
		return
	}

	//获取用户名
	session := sessions.Default(ctx)
	userName := session.Get("userName")

	r := utils.GetConsul()
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("discovery:///user"),
		grpc.WithDiscovery(r),
	)
	if err != nil {
		fmt.Println("grpc.Dial err", err)
		return
	}
	defer conn.Close()
	gClient := user.NewUserClient(conn)
	ctx1, cel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cel()
	resp, err := gClient.UploadAvatar(ctx1, &user.UploadReq{
		Avatar:   nil,
		UserName: userName.(string),
		FileName: fileHeader.Filename,
	})
	if err != nil {
		fmt.Println("调用远程服务错误", err)
	}

	ctx.JSON(http.StatusOK, resp)
}

//获取用户信息
func GetUserInfo(ctx *gin.Context) {
	//获取用户名
	session := sessions.Default(ctx)
	userName := session.Get("userName")

	r := utils.GetConsul()
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("discovery:///user"),
		grpc.WithDiscovery(r),
	)
	if err != nil {
		fmt.Println("grpc.Dial err", err)
		return
	}
	defer conn.Close()
	gClient := user.NewUserClient(conn)
	ctx1, cel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cel()
	resp, err := gClient.GetUser(ctx1, &user.Request{
		Name: userName.(string),
	})
	if err != nil {
		fmt.Println("调用远程服务错误", err)
	}
	ctx.JSON(http.StatusOK, resp)

}

type UpdateUser struct {
	Name string `json:"name"`
}

//更新用户名
func PutUserInfo(ctx *gin.Context) {
	//获取数据
	var nameData UpdateUser
	err := ctx.Bind(&nameData)
	//校验数据
	if err != nil {
		fmt.Println("获取数据错误")
		return
	}

	//从session中获取原来的用户名
	session := sessions.Default(ctx)
	userName := session.Get("userName")

	r := utils.GetConsul()
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("discovery:///user"),
		grpc.WithDiscovery(r),
	)
	if err != nil {
		fmt.Println("grpc.Dial err", err)
		return
	}
	defer conn.Close()
	gClient := user.NewUserClient(conn)
	ctx1, cel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cel()
	resp, err := gClient.UpdateUserName(ctx1, &user.UpdateReq{
		NewName: nameData.Name,
		OldName: userName.(string),
	})
	if err != nil {
		fmt.Println("调用远程服务错误", err)
	}

	//更新session数据
	if resp.Errno == utils.RECODE_OK {
		//更新成功,session中的用户名也需要更新一下
		session.Set("userName", nameData.Name)
		session.Save()
	}

	ctx.JSON(http.StatusOK, resp)

}

type AuthUser struct {
	IdCard   string `json:"id_card"`
	RealName string `json:"real_name"`
}

//实名认证
func PutUserAuth(ctx *gin.Context) {
	//获取数据
	var auth AuthUser
	err := ctx.Bind(&auth)
	//校验数据
	if err != nil {
		fmt.Println("获取数据错误", err)
		return
	}

	session := sessions.Default(ctx)
	userName := session.Get("userName")

	r := utils.GetConsul()
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("discovery:///user"),
		grpc.WithDiscovery(r),
	)
	if err != nil {
		fmt.Println("grpc.Dial err", err)
		return
	}
	defer conn.Close()
	gClient := user.NewUserClient(conn)
	ctx1, cel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cel()
	resp, err := gClient.AuthUpdate(ctx1, &user.AuthReq{
		UserName: userName.(string),
		RealName: auth.RealName,
		IdCard:   auth.IdCard,
	})
	if err != nil {
		fmt.Println("调用远程服务错误", err)
	}

	//返回数据
	ctx.JSON(http.StatusOK, resp)
}

//获取已发布房源信息  假数据
func GetUserHouses(ctx *gin.Context) {
	//获取用户名
	userName := sessions.Default(ctx).Get("userName")

	////调用远程服务
	r := utils.GetConsul()
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("discovery:///house"),
		grpc.WithDiscovery(r),
	)
	if err != nil {
		fmt.Println("grpc.Dial err", err)
		return
	}
	defer conn.Close()
	gClient := house.NewHouseClient(conn)
	ctx1, cel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cel()
	resp, err := gClient.GetHouseInfo(ctx1, &house.GetReq{UserName: userName.(string)})

	//返回数据
	ctx.JSON(http.StatusOK, resp)
}

type House struct {
	Acreage   string   `json:"acreage"`
	Address   string   `json:"address"`
	AreaId    string   `json:"area_id"`
	Beds      string   `json:"beds"`
	Capacity  string   `json:"capacity"`
	Deposit   string   `json:"deposit"`
	Facility  []string `json:"facility"`
	MaxDays   string   `json:"max_days"`
	MinDays   string   `json:"min_days"`
	Price     string   `json:"price"`
	RoomCount string   `json:"room_count"`
	Title     string   `json:"title"`
	Unit      string   `json:"unit"`
}

//发布房源
func PostHouses(ctx *gin.Context) {
	//获取数据   bind数据的时候不带自动转换   c.getInt()
	var houses House
	err := ctx.Bind(&houses)
	fmt.Println(houses.Facility)

	//var strF string
	//for i, s := range houses.Facility {
	//	if i == 0 {
	//		strF = s
	//	} else {
	//		strF += "," + s
	//	}
	//}
	//fmt.Println(strF)
	//
	////strS := []string{}
	//arr := strings.Split(strF, ",")
	//fmt.Println(arr)
	//
	//for _, s := range arr {
	//	fmt.Println(s)
	//}

	//校验数据
	if err != nil {
		fmt.Println("获取数据错误", err)
		return
	}

	//获取用户名
	userName := sessions.Default(ctx).Get("userName")

	r := utils.GetConsul()
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("discovery:///house"),
		grpc.WithDiscovery(r),
	)
	if err != nil {
		fmt.Println("grpc.Dial err", err)
		return
	}
	defer conn.Close()
	gClient := house.NewHouseClient(conn)
	ctx1, cel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cel()
	resp, err := gClient.PubHouse(ctx1, &house.Request{
		Acreage:   houses.Acreage,
		Address:   houses.Address,
		AreaId:    houses.AreaId,
		Beds:      houses.Beds,
		Capacity:  houses.Capacity,
		Deposit:   houses.Deposit,
		Facility:  houses.Facility,
		MaxDays:   houses.MaxDays,
		MinDays:   houses.MinDays,
		Price:     houses.Price,
		RoomCount: houses.RoomCount,
		Title:     houses.Title,
		Unit:      houses.Unit,
		UserName:  userName.(string),
	})

	if err != nil {
		fmt.Println("调用远程服务错误", err)
	}

	//返回数据
	ctx.JSON(http.StatusOK, resp)
}

//上传房屋图片
func PostHousesImage(ctx *gin.Context) {
	//获取数据
	houseId := ctx.Param("id")
	fileHeader, err := ctx.FormFile("house_image")
	//校验数据
	if houseId == "" || err != nil {
		fmt.Println("传入数据不完整", err)
		return
	}

	//三种校验 大小,类型,防止重名  fastdfs
	if fileHeader.Size > 50000000 {
		fmt.Println("文件过大,请重新选择")
		return
	}

	fileExt := path.Ext(fileHeader.Filename)
	if fileExt != ".png" && fileExt != ".jpg" {
		fmt.Println("文件类型错误,请重新选择")
		return
	}

	r := utils.GetConsul()
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("discovery:///house"),
		grpc.WithDiscovery(r),
	)
	if err != nil {
		fmt.Println("grpc.Dial err", err)
		return
	}
	defer conn.Close()
	gClient := house.NewHouseClient(conn)
	ctx1, cel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cel()
	resp, err := gClient.UploadHouseImg(ctx1, &house.ImgReq{
		HouseId:  houseId,
		FileName: fileHeader.Filename,
	})
	if err != nil {
		fmt.Println("远程服务获取失败", err)
	}
	ctx.JSON(http.StatusOK, resp)
}

// GetHouseInfo 展示房屋详情
func GetHouseInfo(ctx *gin.Context) {
	//获取数据
	houseId := ctx.Param("id")
	//校验数据
	if houseId == "" {
		fmt.Println("获取数据错误")
		return
	}
	userName := sessions.Default(ctx).Get("userName")
	//处理数据
	r := utils.GetConsul()
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("discovery:///house"),
		grpc.WithDiscovery(r),
	)
	if err != nil {
		fmt.Println("grpc.Dial err", err)
		return
	}
	defer conn.Close()
	gClient := house.NewHouseClient(conn)
	ctx1, cel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cel()

	resp, err := gClient.GetHouseDetail(ctx1, &house.DetailReq{
		HouseId:  houseId,
		UserName: userName.(string),
	})

	if err != nil {
		fmt.Println("远程服务获取失败", err)
	}

	//返回数据
	ctx.JSON(http.StatusOK, resp)
}

//展示首页轮播图
func GetIndex(ctx *gin.Context) {
	//处理数据
	r := utils.GetConsul()
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("discovery:///house"),
		grpc.WithDiscovery(r),
	)
	if err != nil {
		fmt.Println("grpc.Dial err", err)
		return
	}
	defer conn.Close()
	gClient := house.NewHouseClient(conn)
	ctx1, cel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cel()

	resp, err := gClient.GetIndexHouse(ctx1, &house.IndexReq{})
	if err != nil {
		fmt.Println("远程服务获取失败", err)
	}

	//返回数据
	ctx.JSON(http.StatusOK, resp)
}

// GetHouses 搜索房屋
func GetHouses(ctx *gin.Context) {
	//获取数据
	//areaId
	aid := ctx.Query("aid")
	//start day
	sd := ctx.Query("sd")
	//end day
	ed := ctx.Query("ed")
	//排序方式
	//sk := ctx.Query("sk")
	//page  第几页
	//ctx.Query("p")
	//校验数据
	//fmt.Println(aid, sd, ed, sk)
	//if aid == "" || sd == "" || ed == "" || sk == "" {
	//	fmt.Println("传入数据不完整")
	//	return
	//}
	//处理数据
	r := utils.GetConsul()
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("discovery:///house"),
		grpc.WithDiscovery(r),
	)
	if err != nil {
		fmt.Println("grpc.Dial err", err)
		return
	}
	defer conn.Close()
	gClient := house.NewHouseClient(conn)
	ctx1, cel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cel()
	resp, err := gClient.SearchHouse(ctx1, &house.SearchReq{
		Aid: aid,
		Sd:  sd,
		Ed:  ed,
		//Sk:  sk,
	})
	if err != nil {
		fmt.Println("远程服务获取失败", err)
	}
	//fmt.Println(resp.Data)
	//返回数据
	ctx.JSON(http.StatusOK, resp)
}

type Order struct {
	EndDate   string `json:"end_date"`
	HouseId   string `json:"house_id"`
	StartDate string `json:"start_date"`
}

func GetMsg(ctx *gin.Context) {
	//id := ctx.Query("hid")
	//fmt.Println(id)
	fmt.Println("1111111111111")
	resp := make(map[string]string)
	resp["errno"] = "200"
	resp["errmsg"] = "ok"
	ctx.JSON(http.StatusOK, resp)
}

type Comment struct {
	OrderId string `json:"order_id"`
	Comment string `json:"comment"`
}

//func PutComment(ctx *gin.Context) {
//	//获取数据
//	id := ctx.Param("id")
//	fmt.Println("*****id = ", id, "*********")
//	var comment Comment
//	err := ctx.Bind(&comment)
//	//校验数据
//	if err != nil || id == "" {
//		fmt.Println("获取数据错误", err)
//		return
//	}
//	fmt.Println(comment)
//	fmt.Println("1111111111111")
//	resp := make(map[string]string)
//	resp["errno"] = "200"
//	resp["errmsg"] = "ok"
//	ctx.JSON(http.StatusOK, resp)
//}
func PutComment(ctx *gin.Context) {
	//获取数据
	id := ctx.Param("id")
	//fmt.Println("*****id = ", id, "*********")
	var comment Comment
	err := ctx.Bind(&comment)
	//校验数据
	if err != nil || id == "" {
		fmt.Println("获取数据错误", err)
		return
	}
	fmt.Println(comment)
	//fmt.Println("1111111111111")
	var order model.OrderHouse
	//oid, _ := strconv.Atoi(id)
	err = model.GlobalDB.Table("order_house").Where("id = (?)", id).Find(&order).Error
	if err != nil {
		fmt.Println(err)
		return
	}
	var obj model.Report
	obj.Houseid = int(order.HouseId)
	obj.Cause = comment.Comment
	fmt.Println()
	fmt.Println(obj)
	err = model.GlobalDB.Table("report").Create(&obj).Error
	err = model.GlobalDB.Table("house").Where("id = (?)", obj.Houseid).Update("state", 1).Error
	if err != nil {
		fmt.Println(err)
	}
	resp := make(map[string]string)
	resp["errno"] = "0"
	resp["errmsg"] = "ok"
	ctx.JSON(http.StatusOK, resp)
}

//下订单
func PostOrders(ctx *gin.Context) {
	//获取数据
	var order Order
	err := ctx.Bind(&order)

	//校验数据
	if err != nil {
		fmt.Println("获取数据错误", err)
		return
	}
	//获取用户名
	userName := sessions.Default(ctx).Get("userName")

	r := utils.GetConsul()
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("discovery:///userOrder"),
		grpc.WithDiscovery(r),
	)
	if err != nil {
		fmt.Println("grpc.Dial err", err)
		return
	}
	defer conn.Close()
	gClient := userOrder.NewUserOrderClient(conn)
	ctx1, cel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cel()
	resp, err := gClient.CreateOrder(ctx1, &userOrder.Request{
		StartDate: order.StartDate,
		EndDate:   order.EndDate,
		HouseId:   order.HouseId,
		UserName:  userName.(string),
	})
	if err != nil {
		fmt.Println("远程服务获取失败", err)
	}
	//返回数据
	ctx.JSON(http.StatusOK, resp)
}

//获取订单信息
func GetUserOrder(ctx *gin.Context) {
	//获取get请求传参
	role := ctx.Query("role")
	//校验数据
	if role == "" {
		fmt.Println("获取数据失败")
		return
	}
	//调用远程服务
	r := utils.GetConsul()
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("discovery:///userOrder"),
		grpc.WithDiscovery(r),
	)
	if err != nil {
		fmt.Println("grpc.Dial err", err)
		return
	}
	defer conn.Close()
	gClient := userOrder.NewUserOrderClient(conn)
	ctx1, cel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cel()
	resp, err := gClient.GetOrderInfo(ctx1, &userOrder.GetReq{
		Role:     role,
		UserName: sessions.Default(ctx).Get("userName").(string),
	})
	if err != nil {
		fmt.Println("远程服务获取失败", err)
	}
	//返回数据
	ctx.JSON(http.StatusOK, resp)
}

type Status struct {
	Action string `json:"action"`
	Reason string `json:"reason"`
}

//更新订单状态
func PutOrders(ctx *gin.Context) {
	//获取数据
	id := ctx.Param("id")
	var statusStu Status
	err := ctx.Bind(&statusStu)

	//校验数据
	if err != nil || id == "" {
		fmt.Println("获取数据错误", err)
		return
	}

	//处理数据   更新订单状态
	r := utils.GetConsul()
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("discovery:///userOrder"),
		grpc.WithDiscovery(r),
	)
	if err != nil {
		fmt.Println("grpc.Dial err", err)
		return
	}
	defer conn.Close()
	gClient := userOrder.NewUserOrderClient(conn)
	ctx1, cel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cel()
	resp, err := gClient.UpdateStatus(ctx1, &userOrder.UpdateReq{
		Action: statusStu.Action,
		Reason: statusStu.Reason,
		Id:     id,
	})
	if err != nil {
		fmt.Println("远程服务获取失败", err)
	}
	//返回数据
	ctx.JSON(http.StatusOK, resp)
}
