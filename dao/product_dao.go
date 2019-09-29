package dao

import (
	"database/sql"
	"strconv"

	"github.com/solozyx/seckill/datasource"
	"github.com/solozyx/seckill/model"
)

type IProduct interface {
	// 连接数据库
	Conn() error
	// 返回插入记录id值
	Insert(*model.Product) (int64, error)
	// 入参商品id 返回删除是否成功
	Delete(int64) bool
	Update(*model.Product) error
	// 入参商品id
	SelectById(int64) (*model.Product, error)
	SelectAll() ([]*model.Product, error)
	// 扣除商品数量
	SubProductNum(productID int64) error
}

// Go语言实现接口,非显式,struct实现了interface定义的全部方法,方法名称入参出参匹配,就是struct实现了接口
type ProductManager struct {
	mysqlConn *sql.DB
}

// Go的struct没有构造函数
// 返回类型 IProduct 接口 保证在代码层面struct必须实现接口中定义的全部方法,才能成功创建 ProductManager 的实例
// 如果没有实现接口 代码在编译时报错
func NewProductManager(db *sql.DB) IProduct {
	return &ProductManager{mysqlConn: db}
}

// 数据连接库
func (p *ProductManager) Conn() error {
	if p.mysqlConn == nil {
		mysql, err := datasource.NewMysqlConn()
		if err != nil {
			return err
		}
		p.mysqlConn = mysql
	}
	return nil
}

func (p *ProductManager) Insert(product *model.Product) (productId int64, err error) {
	//1.判断连接是否存在
	if err := p.Conn(); err != nil {
		return
	}
	//2.准备sql
	sql := `insert product set product_name=?,product_num=?,product_image=?,product_url=?`
	stmt, err := p.mysqlConn.Prepare(sql)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	//3.传入参数
	result, err := stmt.Exec(product.ProductName, product.ProductNum, product.ProductImage, product.ProductUrl)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (p *ProductManager) Delete(productID int64) bool {
	if err := p.Conn(); err != nil {
		return false
	}
	sql := `delete from product where id=?`
	stmt, err := p.mysqlConn.Prepare(sql)
	if err != nil {
		return false
	}
	defer stmt.Close()

	_, err = stmt.Exec(strconv.FormatInt(productID, 10))
	if err != nil {
		return false
	}
	return true
}

func (p *ProductManager) Update(product *model.Product) error {
	if err := p.Conn(); err != nil {
		return err
	}
	sql := `update product set product_name=?,product_num=?,product_image=?,product_url=? where id=?`
	stmt, err := p.mysqlConn.Prepare(sql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	productId := strconv.FormatInt(product.ID, 10)
	_, err = stmt.Exec(product.ProductName, product.ProductNum, product.ProductImage, product.ProductUrl, productId)
	if err != nil {
		return err
	}
	return nil
}

func (p *ProductManager) SelectById(productID int64) (product *model.Product, err error) {
	if err := p.Conn(); err != nil {
		return &model.Product{}, err
	}
	sql := "select * from product where id=" + strconv.FormatInt(productID, 10)
	row, err := p.mysqlConn.Query(sql)
	if err != nil {
		return &model.Product{}, err
	}
	defer row.Close()

	result := datasource.GetResultRow(row)
	if len(result) == 0 {
		return &model.Product{}, nil
	}
	product = &model.Product{}
	datasource.DataToStructByTagSql(result, product)
	return
}

func (p *ProductManager) SelectAll() (productList []*model.Product, err error) {
	if err := p.Conn(); err != nil {
		return nil, err
	}
	sql := `select * from product`
	rows, err := p.mysqlConn.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := datasource.GetResultRows(rows)
	if len(result) == 0 {
		return nil, nil
	}

	for _, v := range result {
		product := &model.Product{}
		datasource.DataToStructByTagSql(v, product)
		productList = append(productList, product)
	}
	return
}

// 让rabbitmq消费端调用 扣除商品数量
func (p *ProductManager) SubProductNum(productID int64) error {
	if err := p.Conn(); err != nil {
		return err
	}
	sql := `update product set product_num = product_num-1 where id = ?`
	stmt, err := p.mysqlConn.Prepare(sql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	productId := strconv.FormatInt(productID, 10)
	_, err = stmt.Exec(productId)
	return err
}
