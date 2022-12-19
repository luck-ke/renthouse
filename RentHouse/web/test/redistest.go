package main

//func main() {
//	//连接数据库
//	conn, err := redis.Dial("tcp", "127.0.0.1:6379")
//	if err != nil {
//		fmt.Println("redis Dial err: ", err)
//		return
//	}
//	defer conn.Close()
//
//	//操作数据库
//	reply, err := conn.Do("set", "testK1", "testV1")
//
//	//使用回复助手
//	str, err := redis.String(reply, err)
//	fmt.Println(str, err)
//}
