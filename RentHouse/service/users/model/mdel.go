package model

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var GlobalDB *gorm.DB
var GlobalRedis redis.Pool

func InitDb() error {
	//打开数据库
	//拼接链接字符串
	dsn := "root:010927@tcp(127.0.0.1:3306)/renthouse?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		panic("failed to connect database")
	}
	if err != nil {
		fmt.Println("链接数据库失败", err)
		return err
	}
	GlobalDB = db
	//连接池设置
	sqlDB, err := db.DB()
	sqlDB.SetMaxIdleConns(50)
	sqlDB.SetMaxOpenConns(70)
	sqlDB.SetConnMaxLifetime(60 * 5)

	//表的创建
	//err = db.AutoMigrate(new(User), new(House), new(Area), new(Facility), new(HouseImage), new(OrderHouse))
	return err
}

//初始化redis链接
func InitRedis() {
	GlobalRedis = redis.Pool{
		MaxIdle:     20,
		MaxActive:   50,
		IdleTimeout: 60 * 5,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", "127.0.0.1:6379")
		},
	}
}
