package controllers

import (
	"strconv"

	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"

	"github.com/solozyx/seckill/comm"
	"github.com/solozyx/seckill/conf"
	"github.com/solozyx/seckill/model"
	"github.com/solozyx/seckill/service"
)

type ProductController struct {
	Ctx            iris.Context
	ProductService service.IProductService
}

func (p *ProductController) GetAll() mvc.View {
	productArray, _ := p.ProductService.GetAllProduct()
	return mvc.View{
		// 指定渲染模板
		Name: "product/view.html",
		Data: iris.Map{
			"productArray": productArray,
		},
	}
}

// 修改商品
func (p *ProductController) PostUpdate() {
	product := &model.Product{}
	p.Ctx.Request().ParseForm()
	dec := comm.NewDecoder(&comm.DecoderOptions{TagName: conf.FormTagName})
	// 把form表单数据映射到product模型
	if err := dec.Decode(p.Ctx.Request().Form, product); err != nil {
		p.Ctx.Application().Logger().Debug(err)
	}
	err := p.ProductService.UpdateProduct(product)
	if err != nil {
		p.Ctx.Application().Logger().Debug(err)
	}
	// 重定向
	p.Ctx.Redirect("/product/all")
}

// 显示添加商品页面
func (p *ProductController) GetAdd() mvc.View {
	return mvc.View{
		Name: "product/add.html",
	}
}

// 处理添加商品逻辑
func (p *ProductController) PostAdd() {
	product := &model.Product{}
	p.Ctx.Request().ParseForm()
	dec := comm.NewDecoder(&comm.DecoderOptions{TagName: conf.FormTagName})
	if err := dec.Decode(p.Ctx.Request().Form, product); err != nil {
		p.Ctx.Application().Logger().Debug(err)
	}
	_, err := p.ProductService.InsertProduct(product)
	if err != nil {
		p.Ctx.Application().Logger().Debug(err)
	}
	p.Ctx.Redirect("/product/all")
}

// 展示商品修改页面
func (p *ProductController) GetManager() mvc.View {
	idString := p.Ctx.URLParam("id")
	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		p.Ctx.Application().Logger().Debug(err)
	}
	product, err := p.ProductService.GetProductById(id)
	if err != nil {
		p.Ctx.Application().Logger().Debug(err)
	}

	return mvc.View{
		Name: "product/manager.html",
		Data: iris.Map{
			"product": product,
		},
	}
}

func (p *ProductController) GetDelete() {
	idString := p.Ctx.URLParam("id")
	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		p.Ctx.Application().Logger().Debug(err)
	}
	isOk := p.ProductService.DeleteProductById(id)
	if isOk {
		p.Ctx.Application().Logger().Debug("删除商品成功，ID为：" + idString)
	} else {
		p.Ctx.Application().Logger().Debug("删除商品失败，ID为：" + idString)
	}
	p.Ctx.Redirect("/product/all")
}
