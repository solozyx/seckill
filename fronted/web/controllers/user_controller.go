package controllers

import (
	"strconv"

	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
	"github.com/kataras/iris/sessions"

	"github.com/solozyx/seckill/comm"
	"github.com/solozyx/seckill/conf"
	"github.com/solozyx/seckill/model"
	"github.com/solozyx/seckill/service"
)

type UserController struct {
	Ctx     iris.Context
	Service service.IUserService
	Session *sessions.Session
}

// 秒杀用户注册页面
// GET domain_name/register
func (c *UserController) GetRegister() mvc.View {
	return mvc.View{
		Name: "user/register.html",
	}
}

// POST domain_name/register
func (c *UserController) PostRegister() {
	// 表单字段比较少 适用 字段多参照backend form映射struct
	var (
		nickName = c.Ctx.FormValue("NickName")
		userName = c.Ctx.FormValue("UserName")
		password = c.Ctx.FormValue("Password")
	)
	// 表单校验 ... github.com/ozzo-validation
	user := &model.User{
		UserName:     userName,
		NickName:     nickName,
		HashPassword: password,
	}
	_, err := c.Service.AddUser(user)
	if err != nil {
		c.Ctx.Application().Logger().Debug(err)
		c.Ctx.Redirect("/user/error")
		return
	}
	// 重定向 登录页面
	c.Ctx.Redirect("/user/login")
	return
}

// GET domain_name/login
func (c *UserController) GetLogin() mvc.View {
	return mvc.View{
		Name: "user/login.html",
	}
}

// POST domain_name/login
func (c *UserController) PostLogin() mvc.Response {
	var (
		userName = c.Ctx.FormValue("UserName")
		password = c.Ctx.FormValue("Password")
	)
	// 验证账号密码是否正确
	user, isOk := c.Service.IsPwdSuccess(userName, password)
	if !isOk {
		// 通过 mvc.Response 跳转页面
		return mvc.Response{
			// 跳转登录页面
			Path: "/user/login",
		}
	}

	// 写入用户ID到cookie中
	comm.GlobalCookie(c.Ctx, conf.CookieName, strconv.FormatInt(user.ID, 10))
	// 设置服务端session 一般的web登录会使用服务端session
	c.Session.Set("userId", strconv.FormatInt(user.ID, 10))

	return mvc.Response{
		// 跳转秒杀页面
		Path: "/product/",
	}
}
