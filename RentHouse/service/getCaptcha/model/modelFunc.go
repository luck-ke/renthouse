package model

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
)

func SaveImgRnd(uuid, rnd string) error {
	//连接数据库
	conn, err := redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		fmt.Println("redis Dial err: ", err)
		return err
	}
	defer conn.Close()

	//操作数据库
	_, err = conn.Do("setex", uuid, 60*5, rnd)
	return err
	//链接redis
	//conn := RedisPool.Get()
	//	//存储验证码
	//	_,err := conn.Do("setex",uuid,60 * 5,rnd)
	//	return err
}
