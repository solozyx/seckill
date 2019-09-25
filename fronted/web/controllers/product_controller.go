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

	"github.com/solozyx/seckill/model"
	"github.com/solozyx/seckill/service"
)

type ProductController struct {
	Ctx iris.Context
	// TODO:NOTICE 使用session是临时的,后续会把session优化掉,在高并发场景,session数据维护消耗大
	Session        *sessions.Session
	ProductService service.IProductService
}

var (
	//生成的Html保存目录
	htmlOutPath = "./fronted/web/htmlProductShow/"
	//静态文件模版目录
	templatePath = "./fronted/web/views/template/"
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

func (p *ProductController) GetGenerateHtml() {
	productString := p.Ctx.URLParam("productID")
	productID, err := strconv.Atoi(productString)
	if err != nil {
		p.Ctx.Application().Logger().Debug(err)
	}

	//1.获取模版
	contenstTmp, err := template.ParseFiles(filepath.Join(templatePath, "product.html"))
	if err != nil {
		p.Ctx.Application().Logger().Debug(err)
	}
	//2.获取html生成路径
	fileName := filepath.Join(htmlOutPath, "htmlProduct.html")

	//3.获取模版渲染数据
	product, err := p.ProductService.GetProductByID(int64(productID))
	if err != nil {
		p.Ctx.Application().Logger().Debug(err)
	}
	//4.生成静态文件
	generateStaticHtml(p.Ctx, contenstTmp, fileName, product)
}

//生成html静态文件
func generateStaticHtml(ctx iris.Context, template *template.Template, fileName string, product *model.Product) {
	//1.判断静态文件是否存在
	if exist(fileName) {
		err := os.Remove(fileName)
		if err != nil {
			ctx.Application().Logger().Error(err)
		}
	}
	//2.生成静态文件
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		ctx.Application().Logger().Error(err)
	}
	defer file.Close()
	template.Execute(file, &product)
}

//判断文件是否存在
func exist(fileName string) bool {
	_, err := os.Stat(fileName)
	return err == nil || os.IsExist(err)
}

func (p *ProductController) GetOrder() []byte {
	productString := p.Ctx.URLParam("productID")
	userString := p.Ctx.GetCookie("uid")
	productID, err := strconv.ParseInt(productString, 10, 64)
	if err != nil {
		p.Ctx.Application().Logger().Debug(err)
	}
	userID, err := strconv.ParseInt(userString, 10, 64)
	if err != nil {
		p.Ctx.Application().Logger().Debug(err)
	}

	//创建消息体
	message := model.NewMessage(userID, productID)
	//类型转化
	byteMessage, err := json.Marshal(message)
	if err != nil {
		p.Ctx.Application().Logger().Debug(err)
	}

	err = p.RabbitMQ.PublishSimple(string(byteMessage))
	if err != nil {
		p.Ctx.Application().Logger().Debug(err)
	}

	return []byte("true")

	//product, err := p.ProductService.GetProductByID(int64(productID))
	//if err != nil {
	//	p.Ctx.Application().Logger().Debug(err)
	//}
	//var orderID int64
	//showMessage := "抢购失败！"
	////判断商品数量是否满足需求
	//if product.ProductNum > 0 {
	//	//扣除商品数量
	//	product.ProductNum -= 1
	//	err := p.ProductService.UpdateProduct(product)
	//	if err != nil {
	//		p.Ctx.Application().Logger().Debug(err)
	//	}
	//	//创建订单
	//	userID, err := strconv.Atoi(userString)
	//	if err != nil {
	//		p.Ctx.Application().Logger().Debug(err)
	//	}
	//
	//	order := &datamodels.Order{
	//		UserId:      int64(userID),
	//		ProductId:   int64(productID),
	//		OrderStatus: datamodels.OrderSuccess,
	//	}
	//	//新建订单
	//	orderID, err = p.OrderService.InsertOrder(order)
	//	if err != nil {
	//		p.Ctx.Application().Logger().Debug(err)
	//	} else {
	//		showMessage = "抢购成功！"
	//	}
	//}

}
