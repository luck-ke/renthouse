package model

import (
	"encoding/json"
	"fmt"
	"github.com/gomodule/redigo/redis"
)

/* 区域信息 table_name = area */ //区域信息是需要我们手动添加到数据库中的
type Area struct {
	Id   int    `json:"aid"`                  //区域编号     1    2
	Name string `gorm:"size:32" json:"aname"` //区域名字     昌平 海淀
	//Houses []*House `json:"houses"` //区域所有的房屋   与房屋表进行关联
}

//获取所有地域信息
func GetArea() ([]Area, error) {
	//连接数据库
	var areas []Area

	//从缓存中获取数据  从redis中获取数据
	conn := GlobalRedis.Get()
	//关闭,释放资源
	areaByte, _ := redis.Bytes(conn.Do("get", "areaData"))
	if len(areaByte) == 0 {
		//从mysql中获取数据
		if err := GlobalDB.Find(&areas).Error; err != nil {
			return areas, err
		}
		//序列化数据,存入redis中
		//把数据存入redis中
		areaJson, err := json.Marshal(areas)
		if err != nil {
			return nil, err
		}
		_, err = conn.Do("set", "areaData", areaJson)
		fmt.Println(err)

		fmt.Println("从mysql中获取数据")
	} else {
		json.Unmarshal(areaByte, &areas)
		fmt.Println("从redis中获取数据")
	}
	return areas, nil
}
