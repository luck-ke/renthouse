package model

type User struct {
	ID            int    //用户编号
	Name          string `gorm:"size:32;unique;not null"` //用户名
	Password_hash string `gorm:"size:128" `               //用户密码加密的  hash
	Mobile        string `gorm:"size:11;unique" `         //手机号
	Real_name     string `gorm:"size:32" `                //真实姓名  实名认证
	Id_card       string `gorm:"size:20" `                //身份证号  实名认证
	Avatar_url    string `gorm:"size:256" `               //用户头像路径
	//Houses        []*House      //用户发布的房屋信息  一个人多套房
	//Orders        []*OrderHouse //用户下的订单       一个人多次订单
}

//获取用户信息
func GetUserInfo(userName string) (User, error) {
	//连接数据库
	var user User
	err := GlobalDB.Where("name = ?", userName).Find(&user).Error
	return user, err
}

//存储用户头像   更新
func SaveUserAvatar(userName, avatarUrl string) error {
	return GlobalDB.Model(new(User)).Where("name = ?", userName).Update("avatar_url", avatarUrl).Error
}

//更新用户名
func UpdateUserName(oldName, newName string) error {
	//更新  链式调用
	return GlobalDB.Model(new(User)).
		Where("name = ?", oldName).
		Update("name", newName).Error
}

//存储用户真实姓名
func SaveRealName(userName, realName, idCard string) error {
	return GlobalDB.Model(new(User)).Where("name = ?", userName).
		Updates(map[string]interface{}{"real_name": realName, "id_card": idCard}).Error
}
