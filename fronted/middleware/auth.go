package middleware

import (
	"github.com/kataras/iris"

	"github.com/solozyx/seckill/conf"
)

func AuthUserLogin(ctx iris.Context) {
	uid := ctx.GetCookie(conf.CookieName)
	if uid == "" {
		ctx.Application().Logger().Debug("必须先登录!")
		ctx.Redirect("/user/login")
		return
	}
	ctx.Application().Logger().Debug("用户已经登录")
	ctx.Next()
}
