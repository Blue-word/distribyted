package server

import (
	"errors"
	"github.com/distribyted/distribyted/model"
	"github.com/distribyted/distribyted/module"
)

func UserLogin(username, password string) (*model.UserToken, error) {
	var user model.UserToken
	if username == "" || password == "" {
		return nil, errors.New("账号或密码不能为空")
	}
	// 从数据库中判断账号密码
	//module.Badger.AddUser("admin", "admin")
	pass, err := module.Badger.GetUserPassword(username)
	if err != nil {
		return nil, err
	}
	// todo 密码加密
	if pass != password {
		return nil, errors.New("账号或密码错误")
	}
	// todo token生成
	user.Token = "admin-token"
	return &user, nil
}

func UserInfo(token string) (*model.UserInfo, error) {
	var userInfo model.UserInfo
	if token == "" {
		return nil, errors.New("token为空")
	}
	// todo 从token解析用户
	userInfo.Roles = []string{"admin"}
	userInfo.Introduction = "I am a super administrator"
	userInfo.Avatar = "https://wpimg.wallstcn.com/f778738c-e4f8-4870-b634-56703b4acafe.gif"
	userInfo.Name = "Super Admin"
	return &userInfo, nil
}
