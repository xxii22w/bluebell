package logic

import (
	"bluebell/dao/mysql"
	"bluebell/models"
	"bluebell/pkg/jwt"
	"bluebell/pkg/snowflake"
)

func SignUp(p *models.ParamSignUp) (err error) {
	// 判断用户存不存在
	if err := mysql.CheckUserExist(p.Username);err != nil {
		return err
	}
	// 生成uid
	userID := snowflake.GenID()
	// 构造一个User示例
	user := &models.User{
		UserID: userID,
		Username: p.Username,
		Password: p.Password,
	}
	return mysql.InsertUser(user)
}

func Login(p *models.ParamLogin) (user *models.User,err error) {
	user = &models.User{
		Username: p.Username,
		Password: p.Password,
	}
	// 传递的是指针，就能拿到user.UserId
	if err := mysql.Login(user);err != nil {
		return nil,err
	}
	// 生成JWT
	token,err := jwt.GenToken(user.UserID,user.Username)
	if err != nil {
		return
	}
	user.Token = token
	return
}