package main

import (
	"context"
	"errors"
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
	"log"
	"os"
	"unicode"
	"user/api/user"
	"user/model"
	"user/utils"
)

type User struct {
	user.UnimplementedUserServer
}

//GetUser(context.Context, *Request) (*Response, error)
func (*User) GetUser(ctx context.Context, req *user.Request) (*user.Response, error) {
	//根据用户名获取用户信息 在mysql数据库中
	myUser, err := model.GetUserInfo(req.Name)
	if err != nil {
		return &user.Response{
			Errno:  utils.RECODE_DATAERR,
			Errmsg: utils.RecodeText(utils.RECODE_DATAERR),
		}, err
	}

	//获取一个结构体对象
	var userInfo user.UserInfo
	userInfo.UserId = int32(myUser.ID)
	userInfo.Name = myUser.Name
	userInfo.Mobile = myUser.Mobile
	userInfo.RealName = myUser.Real_name
	userInfo.IdCard = myUser.Id_card
	userInfo.AvatarUrl = "http://127.0.0.1:9000/image/" + myUser.Avatar_url

	return &user.Response{
		Errno:  utils.RECODE_OK,
		Errmsg: utils.RecodeText(utils.RECODE_OK),
		Data:   &userInfo,
	}, nil
}

//UpdateUserName(context.Context, *UpdateReq)
func (*User) UpdateUserName(ctx context.Context, req *user.UpdateReq) (*user.UpdateResp, error) {
	//根据传递过来的用户名更新数据中新的用户名
	err := model.UpdateUserName(req.OldName, req.NewName)
	if err != nil {
		fmt.Println("更新失败", err)
		return &user.UpdateResp{
			Errno:  utils.RECODE_DATAERR,
			Errmsg: utils.RecodeText(utils.RECODE_DATAERR),
		}, err
	}
	resp := &user.UpdateResp{}
	resp.Errno = utils.RECODE_OK
	resp.Errmsg = utils.RecodeText(utils.RECODE_OK)
	var nameData user.NameData
	nameData.Name = req.NewName
	resp.Data = &nameData

	return resp, nil
}

//UploadAvatar(context.Context, *UploadReq) (*UploadResp, error)
func (*User) UploadAvatar(ctx context.Context, req *user.UploadReq) (*user.UploadResp, error) {
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
	filePath := "D:\\goimage\\user\\" + req.FileName
	contentType := "image/png"

	// Upload the zip file with FPutObject
	info, err := minioClient.FPutObject(c, bucketName, objectName, filePath, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		fmt.Println("文件上传失败")
		log.Fatalln(err)
		return &user.UploadResp{
			Errno:  utils.RECODE_DATAERR,
			Errmsg: utils.RecodeText(utils.RECODE_DATAERR),
		}, err
	}
	//fmt.Println(info)
	log.Printf("Successfully uploaded %s of size %d\n", objectName, info.Size)
	//把存储凭证写入数据库
	err = model.SaveUserAvatar(req.UserName, req.FileName)
	if err != nil {
		fmt.Println("存储用户头像错误", err)
		return &user.UploadResp{
			Errno:  utils.RECODE_DBERR,
			Errmsg: utils.RecodeText(utils.RECODE_DBERR),
		}, err
	}
	resp := &user.UploadResp{}
	resp.Errno = utils.RECODE_OK
	resp.Errmsg = utils.RecodeText(utils.RECODE_OK)
	var uploadData user.UploadData
	uploadData.AvatarUrl = "http://127.0.0.1:9000/image/" + req.FileName
	resp.Data = &uploadData
	return resp, nil
}

//AuthUpdate(context.Context, *AuthReq) (*AuthResp, error)
func (*User) AuthUpdate(ctx context.Context, req *user.AuthReq) (*user.AuthResp, error) {
	//假判断
	for _, v := range req.RealName {
		if !unicode.Is(unicode.Han, v) || len(req.IdCard) != 18 {
			fmt.Println("输入不是中文")
			return &user.AuthResp{
				Errno:  utils.RECODE_DATAERR,
				Errmsg: utils.RecodeText(utils.RECODE_DATAERR),
			}, errors.New("姓名或身份证号输入错误")
		}
	}

	//存储真实姓名和真是身份证号  数据库
	err := model.SaveRealName(req.UserName, req.RealName, req.IdCard)
	if err != nil {
		return &user.AuthResp{
			Errno:  utils.RECODE_DBERR,
			Errmsg: utils.RecodeText(utils.RECODE_DBERR),
		}, err
	}
	return &user.AuthResp{
		Errno:  utils.RECODE_OK,
		Errmsg: utils.RecodeText(utils.RECODE_OK),
	}, nil
}

func main() {
	model.InitRedis()
	model.InitDb()

	logger := log2.NewStdLogger(os.Stdout)

	consulClient, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Fatal(err)
	}

	grpcSrv := grpc.NewServer(
		grpc.Address(":12344"),
		grpc.Middleware(
			recovery.Recovery(),
			logging.Server(logger),
		),
	)

	s := &User{}
	user.RegisterUserServer(grpcSrv, s)

	r := consul.New(consulClient)
	app := kratos.New(
		kratos.Name("user"),
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
