package main

import (
	"context"

	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"

	"github.com/solozyx/seckill/backend/web/controllers"
	"github.com/solozyx/seckill/dao"
	"github.com/solozyx/seckill/datasource"
	"github.com/solozyx/seckill/service"
)

func main() {
	//1.创建iris 实例
	app := iris.New()
	//2.设置错误模式，在mvc模式下提示错误
	app.Logger().SetLevel("debug")
	//3.注册模板 模板(相对)路径 模板文件后缀.html
	tmplate := iris.HTML("./backend/web/views", ".html").
		// 设置布局
		Layout("shared/layout.html").
		Reload(true)
	app.RegisterView(tmplate)
	//4.设置静态资源 通过 domain_name/assets 访问到静态资源
	app.StaticWeb("/assets", "./backend/web/assets")
	//5.出现异常跳转到指定页面
	app.OnAnyErrorCode(func(ctx iris.Context) {
		ctx.ViewData("message", ctx.Values().GetStringDefault("message", "访问的页面出错！"))
		// 这里不设置 layout
		ctx.ViewLayout("")
		ctx.View("shared/error.html")
	})

	//6.连接数据库
	db, err := datasource.NewMysqlConn()
	if err != nil {
		panic(err)
	}

	//7.注册控制器 实现路由
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 秒杀商品管理
	productDao := dao.NewProductManager(db)
	productService := service.NewProductService(productDao)
	productParty := app.Party("/product")
	product := mvc.New(productParty)
	product.Register(ctx, productService)
	product.Handle(new(controllers.ProductController))

	// 订单管理
	orderDao := dao.NewOrderManager(db)
	orderService := service.NewOrderService(orderDao)
	orderParty := app.Party("/order")
	order := mvc.New(orderParty)
	order.Register(ctx, orderService)
	order.Handle(new(controllers.OrderController))

	//8.启动服务
	app.Run(
		// 使用 80端口需要 sudo权限 debug模式启动没有 80端口权限
		iris.Addr(":8080"),
		// 是否检测 iris框架版本 设置为不检测
		// iris.WithoutVersionChecker,
		// 忽略 iris 框架错误
		iris.WithoutServerError(iris.ErrServerClosed),
		// 使用优化
		iris.WithOptimizations,
	)
}
