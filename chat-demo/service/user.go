package service

import (
	"chat-demo/model"
	"chat-demo/serializer"
)

// 用户注册服务结构体
type UserRegisterService struct {
	UserName string `json:"user_name" form:"user_name"`
	Password string `json:"password" form:"password"`
}

// 注册服务
func (service *UserRegisterService) Register() serializer.Response {
	var user model.User
	count := 0
	// 判断是否重名
	model.DB.Model(&model.User{}).Where("user_name=?", service.UserName).First(&user).Count(&count)
	if count != 0 {
		return serializer.Response{
			Status: 400,
			Msg:    "用户名已经存在",
		}
	}
	user = model.User{
		UserName: service.UserName,
	}
	if err := user.SetPassword(service.Password); err != nil {
		return serializer.Response{
			Status: 500,
			Msg:    "加密出错",
		}
	}
	model.DB.Create(&user)
	return serializer.Response{
		Status: 200,
		Msg:    "创建成功",
	}
}
