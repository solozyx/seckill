package controllers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"text/template"

	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
	"github.com/kataras/iris/sessions"

	"github.com/solozyx/seckill/conf"
	"github.com/solozyx/seckill/datasource"
	"github.com/solozyx/seckill/model"
	"github.com/solozyx/seckill/service"
)

type ProductController struct {
	Ctx iris.Context
	// TODO:NOTICE 使用session是临时的,后续会把session优化掉,在高并发场景,session数据维护消耗大
	Session        *sessions.Session
	ProductService service.IProductService
	OrderService   service.IOrderService
	RabbitMQ       *datasource.RabbitMQ
}

var (
	// 静态文件模版目录
	templatePath = "./fronted/web/views/template/"
	// 生成静态资源 html 文件保存目录
	htmlOutPath = "./fronted/web/static_product/"
	// 生成静态资源 html 文件名
	htmlProduct = "product.html"
)

// 秒杀商品详情
func (p *ProductController) GetDetail() mvc.View {
	// TODO:这里直接硬编码 后续要改为接收商品id
	// id := p.Ctx.URLParam("ProductID")
	product, err := p.ProductService.GetProductById(1)
	if err != nil {
		p.Ctx.Application().Logger().Error(err)
	}

	return mvc.View{
		// 商品详情展示布局模板文件
		Layout: "shared/product_layout.html",
		// 商品详情展示模板文件
		Name: "product/view.html",
		Data: iris.Map{
			"product": product,
		},
	}
}

// GET domain_name/generate/html?productID=1
// 生成秒杀商品静态 html 文件
func (p *ProductController) GetGenerateHtml() {
	// 1.获取 html 模板
	tpl, err := template.ParseFiles(filepath.Join(templatePath, "product.html"))
	if err != nil {
		p.Ctx.Application().Logger().Error(err)
	}
	// 2.获取html生成路径
	filePath := filepath.Join(htmlOutPath, htmlProduct)
	// 3.获取模板渲染数据
	productString := p.Ctx.URLParam("productID")
	productID, err := strconv.Atoi(productString)
	if err != nil {
		p.Ctx.Application().Logger().Debug(err)
	}
	// TODO:NOTICE 这里生成商品详情页查询1次MySQL数据库 [对性能没有影响]
	//  因为用户直接访问的是生成后的静态资源 用户不走这里
	product, err := p.ProductService.GetProductById(int64(productID))
	if err != nil {
		p.Ctx.Application().Logger().Debug(err)
	}
	//4.生成静态文件
	generateStaticHtml(p.Ctx, tpl, filePath, product)
}

// 生成 html 静态文件
// Template 会把HTML模板的注释过滤掉 直接以text文本输出
// filePath 静态资源生成存放目录
func generateStaticHtml(ctx iris.Context, template *template.Template, filePath string, product *model.Product) {
	// 1.判断静态文件是否存在
	if exist(filePath) {
		err := os.Remove(filePath)
		if err != nil {
			ctx.Application().Logger().Error(err)
		}
	}
	// 2.生成静态文件
	// TODO:WARNING 使用 0777
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		ctx.Application().Logger().Error(err)
	}
	defer file.Close()
	// 把product数据渲染到HTML模板 输出静态商品详情页
	template.Execute(file, &product)
}

// 判断文件是否存在
func exist(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil || os.IsExist(err)
}

// 基于MySQL的秒杀抢购流程
func (p *ProductController) GetOrder() []byte {
	productString := p.Ctx.URLParam("productID")
	userString := p.Ctx.GetCookie(conf.CookieName)
	productID, err := strconv.ParseInt(productString, 10, 64)
	if err != nil {
		p.Ctx.Application().Logger().Debug(err)
	}
	userID, err := strconv.ParseInt(userString, 10, 64)
	if err != nil {
		p.Ctx.Application().Logger().Debug(err)
	}

	// 把消息投递到 rabbitmq
	message := model.NewMessage(userID, productID)
	byteMessage, err := json.Marshal(message)
	if err != nil {
		p.Ctx.Application().Logger().Debug(err)
	}
	err = p.RabbitMQ.PublishSimple(string(byteMessage))
	if err != nil {
		p.Ctx.Application().Logger().Debug(err)
	}
	return []byte("true")
}

// 基于MySQL的秒杀抢购流程
func (p *ProductController) deprecatedGetOrder() mvc.View {
	productString := p.Ctx.URLParam("productID")
	userString := p.Ctx.GetCookie(conf.CookieName)
	productID, err := strconv.ParseInt(productString, 10, 64)
	if err != nil {
		p.Ctx.Application().Logger().Debug(err)
	}
	// TODO:WARNING 直接查询MySQL数据库获取秒杀商品信息 性能低下
	product, err := p.ProductService.GetProductById(productID)
	if err != nil {
		p.Ctx.Application().Logger().Debug(err)
	}
	var orderID int64
	showMessage := "抢购失败！"
	// 判断商品数量是否满足需求
	if product.ProductNum > 0 {
		// 扣除商品数量
		product.ProductNum -= 1
		// TODO:ERROR 直接更新MySQL数据库获取秒杀商品信息 性能低下 在高并发场景会超卖错误
		err := p.ProductService.UpdateProduct(product)
		if err != nil {
			p.Ctx.Application().Logger().Debug(err)
		}
		// 创建订单
		userID, err := strconv.Atoi(userString)
		if err != nil {
			p.Ctx.Application().Logger().Debug(err)
		}
		order := &model.Order{
			UserId:      int64(userID),
			ProductId:   productID,
			OrderStatus: model.OrderSuccess,
		}
		// TODO:WARNING 新建订单 直接更新MySQL数据库获取秒杀商品信息 性能低下
		orderID, err = p.OrderService.InsertOrder(order)
		if err != nil {
			p.Ctx.Application().Logger().Debug(err)
		} else {
			showMessage = "抢购成功！"
		}
	}
	return mvc.View{
		Layout: "shared/product_layout.html",
		Name:   "product/result.html",
		Data: iris.Map{
			"orderID":     orderID,
			"showMessage": showMessage,
		},
	}
}
