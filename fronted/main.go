package main

import (
	"context"
	"time"

	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
	"github.com/kataras/iris/sessions"

	"github.com/solozyx/seckill/dao"
	"github.com/solozyx/seckill/datasource"
	"github.com/solozyx/seckill/fronted/middleware"
	"github.com/solozyx/seckill/fronted/web/controllers"
	"github.com/solozyx/seckill/service"
)

func main() {
	// 1.创建iris 实例
	app := iris.New()
	// 2.设置错误模式，在mvc模式下提示错误
	app.Logger().SetLevel("debug")
	// 3.注册模板
	tmplate := iris.HTML("./fronted/web/views", ".html").
		Layout("shared/layout.html").
		Reload(true)
	app.RegisterView(tmplate)
	// 4.设置模板
	app.StaticWeb("/public", "./fronted/web/public")

	// 访问生成好的html静态文件
	// app.StaticWeb("/html", "./fronted/web/htmlProductShow")

	// 出现异常跳转到指定页面
	app.OnAnyErrorCode(func(ctx iris.Context) {
		ctx.ViewData("message", ctx.Values().GetStringDefault("message", "访问的页面出错！"))
		ctx.ViewLayout("")
		ctx.View("shared/error.html")
	})

	// 连接数据库
	db, err := datasource.NewMysqlConn()
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// session
	session := sessions.New(sessions.Config{
		// cookie名称
		Cookie: "seckill",
		// 过期时间
		Expires: 60 * time.Minute,
	})

	// 注册 user 控制器
	userDao := dao.NewUserManager(db)
	userService := service.NewUserService(userDao)
	userParty := app.Party("/user")
	user := mvc.New(userParty)
	user.Register(userService, ctx, session.Start)
	user.Handle(new(controllers.UserController))

	// 注册 product 控制器
	productDao := dao.NewProductManager(db)
	productService := service.NewProductService(productDao)
	productParty := app.Party("/product")
	// 秒杀用户登录验证中间件
	productParty.Use(middleware.AuthUserLogin)
	product := mvc.New(productParty)
	// 注册service和session
	product.Register(productService, ctx, session.Start)
	product.Handle(new(controllers.ProductController))

	app.Run(
		iris.Addr("0.0.0.0:8082"),
		// 不检测 iris 版本
		// iris.WithoutVersionChecker,
		// 忽略 server err 服务不关闭
		iris.WithoutServerError(iris.ErrServerClosed),
		// 服务优化
		iris.WithOptimizations,
	)
}
