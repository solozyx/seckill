package service

import (
	"errors"

	"golang.org/x/crypto/bcrypt"

	"github.com/solozyx/seckill/dao"
	"github.com/solozyx/seckill/model"
)

type IUserService interface {
	IsPwdSuccess(userName string, pwd string) (user *model.User, isOk bool)
	AddUser(user *model.User) (userId int64, err error)
}

type UserService struct {
	userDao dao.IUser
}

func NewUserService(dao dao.IUser) IUserService {
	return &UserService{dao}
}

func (u *UserService) IsPwdSuccess(userName string, pwd string) (user *model.User, isOk bool) {
	user, err := u.userDao.Select(userName)
	if err != nil {
		return
	}
	isOk, _ = validatePassword(pwd, user.HashPassword)
	if !isOk {
		return &model.User{}, false
	}
	return
}

func (u *UserService) AddUser(user *model.User) (userId int64, err error) {
	pwdByte, err := generatePassword(user.HashPassword)
	if err != nil {
		return userId, err
	}
	user.HashPassword = string(pwdByte)
	return u.userDao.Insert(user)
}

func generatePassword(userPassword string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(userPassword), bcrypt.DefaultCost)
}

func validatePassword(userPassword string, hashed string) (isOK bool, err error) {
	if err = bcrypt.CompareHashAndPassword([]byte(hashed), []byte(userPassword)); err != nil {
		return false, errors.New("密码比对错误！")
	}
	return true, nil
}
