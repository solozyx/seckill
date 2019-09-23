package service

import (
	"github.com/solozyx/seckill/dao"
	"github.com/solozyx/seckill/model"
)

type IProductService interface {
	// 入参商品id
	GetProductById(int64) (*model.Product, error)
	GetAllProduct() ([]*model.Product, error)
	DeleteProductById(int64) bool
	// 返回插入商品的id
	InsertProduct(product *model.Product) (int64, error)
	UpdateProduct(product *model.Product) error
	SubNumberOne(productID int64) error
}

type ProductService struct {
	productDao dao.IProduct
}

func NewProductService(dao dao.IProduct) IProductService {
	return &ProductService{dao}
}

func (p *ProductService) GetProductById(productID int64) (*model.Product, error) {
	return p.productDao.SelectById(productID)
}

func (p *ProductService) GetAllProduct() ([]*model.Product, error) {
	return p.productDao.SelectAll()
}

func (p *ProductService) DeleteProductById(productID int64) bool {
	return p.productDao.Delete(productID)
}

func (p *ProductService) InsertProduct(product *model.Product) (int64, error) {
	return p.productDao.Insert(product)
}

func (p *ProductService) UpdateProduct(product *model.Product) error {
	return p.productDao.Update(product)
}

func (p *ProductService) SubNumberOne(productID int64) error {
	return p.productDao.SubProductNum(productID)
}
