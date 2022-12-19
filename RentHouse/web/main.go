package main

import (
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"net/http"
	"test05kratos/web/controller"
	"test05kratos/web/model"
	"test05kratos/web/utils"
)

func main() {
	//初始化路由
	r := gin.Default()

	err := model.InitDb()
	if err != nil {
		fmt.Println("数据库初始化失败 : ", err)
		return
	}

	//r.GET("/h", controller.GetHouses)

	// 路由匹配
	r.Static("/home", "./web/view")

	//session容器初始化
	store, _ := redis.NewStore(20, "tcp", "127.0.0.1:6379", "", []byte("session"))

	r1 := r.Group("/api/v1.0")
	{
		r1.GET("/imagecode/:uuid", controller.GetImageCd)
		r1.GET("/smscode/:mobile", controller.GetSmscd)
		r1.POST("/users", controller.PostRet)
		r1.GET("/areas", controller.GetArea)

		//路由过滤
		r1.Use(sessions.Sessions("mysession", store))
		r1.GET("/session", controller.GetSession)
		r1.POST("/sessions", controller.PostLogin)
		//路由过滤器   登录的情况下才能执行一下路由请求
		r1.Use(Filter)
		r1.DELETE("/session", controller.DeleteSession)
		r1.POST("/user/avatar", controller.PostAvatar)
		r1.GET("/user", controller.GetUserInfo)
		r1.PUT("/user/name", controller.PutUserInfo)
		r1.POST("/user/auth", controller.PutUserAuth)
		r1.GET("/user/auth", controller.GetUserInfo)

		//获取已发布房源信息
		r1.GET("/user/houses", controller.GetUserHouses)
		r1.POST("/houses", controller.PostHouses)
		//添加房源图片
		r1.POST("/houses/:id/images", controller.PostHousesImage)
		//展示房屋详情
		r1.GET("/houses/:id", controller.GetHouseInfo)
		//展示首页轮播图
		r1.GET("/house/index", controller.GetIndex)

		//搜索房屋
		r1.GET("/houses", controller.GetHouses)

		r1.GET("/msgorder", controller.GetMsg)

		//下订单
		r1.POST("/orders", controller.PostOrders)
		//获取订单
		r1.GET("/user/orders", controller.GetUserOrder)
		//同意/拒绝订单
		r1.PUT("/orders/:id/status", controller.PutOrders)
		//发表评论
		r1.PUT("/orders/:id/comment", controller.PutComment)

		//r.NoRoute(func(c *gin.Context) {
		//	// 实现内部重定向
		//	c.HTML(http.StatusOK, "search.html", gin.H{
		//		"title": "404",
		//	})
		//})
	}

	//启动运行
	r.Run(":8080")
}

//路由过滤器
func Filter(ctx *gin.Context) {
	//登录校验
	session := sessions.Default(ctx)
	userName := session.Get("userName")
	resp := make(map[string]interface{})
	if userName == nil {
		resp["errno"] = utils.RECODE_SESSIONERR
		resp["errmsg"] = utils.RecodeText(utils.RECODE_SESSIONERR)
		ctx.JSON(http.StatusOK, resp)
		ctx.Abort()
		return
	}
	//计算这个业务耗时
	//fmt.Println("next之前打印", time.Now())

	//执行函数
	ctx.Next()

	//fmt.Println("next之后打印....")
}
