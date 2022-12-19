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
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	house2 "house/api/house"
	"house/model"
	"house/utils"
	"log"
	"os"
	"strconv"
)

type House struct {
	house2.UnimplementedHouseServer
}

func (*House) PubHouse(ctx context.Context, req *house2.Request) (*house2.Response, error) {
	//上传房屋业务  把获取到的房屋数据插入数据库
	houseId, err := model.AddHouse(req)
	if err != nil {
		return &house2.Response{
			Errno:  utils.RECODE_DBERR,
			Errmsg: utils.RecodeText(utils.RECODE_DBERR),
			Data:   nil,
		}, nil
	}

	var h house2.HouseData
	h.HouseId = strconv.Itoa(houseId)
	resp := &house2.Response{}
	resp.Errno = utils.RECODE_OK
	resp.Errmsg = utils.RecodeText(utils.RECODE_OK)
	resp.Data = &h

	return resp, nil
}

func (*House) UploadHouseImg(ctx context.Context, req *house2.ImgReq) (*house2.ImgResp, error) {
	c := context.Background()
	endpoint := "127.0.0.1:9000"
	accessKeyID := "minioadmin"
	secretAccessKey := "minioadmin"
	useSSL := false

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalln(err)
	}

	bucketName := "image"
	location := "us-east-1"

	err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			log.Printf("We already own %s\n", bucketName)
		} else {
			log.Fatalln(err)
		}
	} else {
		log.Printf("Successfully created %s\n", bucketName)
	}

	// Upload the zip file
	objectName := req.FileName
	filePath := "D:\\goimage\\house\\" + req.FileName
	contentType := "image/png"

	// Upload the zip file with FPutObject
	info, err := minioClient.FPutObject(c, bucketName, objectName, filePath, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		fmt.Println("文件上传失败")
		log.Fatalln(err)
		return &house2.ImgResp{
			Errno:  utils.RECODE_DATAERR,
			Errmsg: utils.RecodeText(utils.RECODE_DATAERR),
		}, err
	}
	log.Printf("Successfully uploaded %s of size %d\n", objectName, info.Size)

	//把存储凭证写入数据库
	err = model.SaveHouseImg(req.HouseId, req.FileName)
	if err != nil {
		fmt.Println("存储用户头像错误", err)
		return &house2.ImgResp{
			Errno:  utils.RECODE_DBERR,
			Errmsg: utils.RecodeText(utils.RECODE_DBERR),
		}, err
	}
	resp := &house2.ImgResp{}
	resp.Errno = utils.RECODE_OK
	resp.Errmsg = utils.RecodeText(utils.RECODE_OK)

	var img house2.ImgData
	img.Url = "http://127.0.0.1:9000/image/" + req.FileName

	resp.Data = &img

	return resp, nil
}

func (*House) GetHouseInfo(ctx context.Context, req *house2.GetReq) (*house2.GetResp, error) {
	//根据用户名获取所有的房屋数据
	houseInfos, err := model.GetUserHouse(req.UserName)
	if err != nil {
		return &house2.GetResp{
			Errno:  utils.RECODE_DBERR,
			Errmsg: utils.RecodeText(utils.RECODE_DBERR),
		}, err
	}

	var getData house2.GetData
	getData.Houses = houseInfos
	resp := &house2.GetResp{}
	resp.Errno = utils.RECODE_OK
	resp.Errmsg = utils.RecodeText(utils.RECODE_OK)
	resp.Data = &getData
	//fmt.Println(*resp.Data)

	return resp, nil
}

func (*House) GetHouseDetail(ctx context.Context, req *house2.DetailReq) (*house2.DetailResp, error) {
	//根据houseId获取所有的返回数据
	respData, err := model.GetHouseDetail(req.HouseId, req.UserName)
	if err != nil {
		return &house2.DetailResp{
			Errno:  utils.RECODE_DBERR,
			Errmsg: utils.RecodeText(utils.RECODE_DBERR),
		}, err
	}
	resp := &house2.DetailResp{}
	resp.Errno = utils.RECODE_OK
	resp.Errmsg = utils.RecodeText(utils.RECODE_OK)
	resp.Data = &respData

	return resp, nil
}

func (*House) GetIndexHouse(ctx context.Context, req *house2.IndexReq) (*house2.GetResp, error) {
	//获取房屋信息
	houseResp, err := model.GetIndexHouse()
	if err != nil {
		return &house2.GetResp{
			Errno:  utils.RECODE_DBERR,
			Errmsg: utils.RecodeText(utils.RECODE_DBERR),
		}, err
	}
	resp := &house2.GetResp{}
	resp.Errno = utils.RECODE_OK
	resp.Errmsg = utils.RecodeText(utils.RECODE_OK)

	resp.Data = &house2.GetData{Houses: houseResp}

	return resp, nil
}

func (*House) SearchHouse(ctx context.Context, req *house2.SearchReq) (*house2.GetResp, error) {
	//根据传入的参数,查询符合条件的房屋信息
	houseResp, err := model.SearchHouse(req.Aid, req.Sd, req.Ed, req.Sk)
	if err != nil {
		return &house2.GetResp{
			Errno:  utils.RECODE_DBERR,
			Errmsg: utils.RecodeText(utils.RECODE_DBERR),
		}, err
	}
	fmt.Println(houseResp)
	resp := &house2.GetResp{}
	resp.Errno = utils.RECODE_OK
	resp.Errmsg = utils.RecodeText(utils.RECODE_OK)
	resp.Data = &house2.GetData{Houses: houseResp}

	return resp, nil
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
		grpc.Address(":12345"),
		grpc.Middleware(
			recovery.Recovery(),
			logging.Server(logger),
		),
	)

	s := &House{}
	house2.RegisterHouseServer(grpcSrv, s)

	r := consul.New(consulClient)
	app := kratos.New(
		kratos.Name("house"),
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
