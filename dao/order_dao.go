package dao

import (
	"database/sql"
	"strconv"

	"github.com/solozyx/seckill/datasource"
	"github.com/solozyx/seckill/model"
)

type IOrder interface {
	Conn() error
	// 返回插入订单id
	Insert(*model.Order) (int64, error)
	// 入参订单id
	Delete(int64) bool
	Update(*model.Order) error
	SelectById(int64) (*model.Order, error)
	SelectAll() ([]*model.Order, error)
	// 订单关联消息
	SelectAllWithInfo() (map[int]map[string]string, error)
}

type OrderManager struct {
	mysqlConn *sql.DB
}

func NewOrderManager(db *sql.DB) IOrder {
	return &OrderManager{mysqlConn: db}
}

func (o *OrderManager) Conn() error {
	// mysql断开需要重新连接
	if o.mysqlConn == nil {
		mysql, err := datasource.NewMysqlConn()
		if err != nil {
			return err
		}
		o.mysqlConn = mysql
	}
	return nil
}

func (o *OrderManager) Insert(order *model.Order) (productID int64, err error) {
	if err = o.Conn(); err != nil {
		return
	}
	sql := `insert order set user_id = ?,product_id = ?,order_status = ?`
	stmt, err := o.mysqlConn.Prepare(sql)
	if err != nil {
		return
	}
	defer stmt.Close()

	result, err := stmt.Exec(order.UserId, order.ProductId, order.OrderStatus)
	if err != nil {
		return
	}
	return result.LastInsertId()
}

func (o *OrderManager) Delete(orderID int64) bool {
	if err := o.Conn(); err != nil {
		return false
	}
	sql := `delete from order where id = ?`
	stmt, err := o.mysqlConn.Prepare(sql)
	if err != nil {
		return false
	}
	defer stmt.Close()

	_, err = stmt.Exec(orderID)
	if err != nil {
		return false
	}
	return true
}

func (o *OrderManager) Update(order *model.Order) (err error) {
	if err = o.Conn(); err != nil {
		return
	}

	sql := `update order set user_id=?,product_id=?,order_status=? where id=?`
	stmt, err := o.mysqlConn.Prepare(sql)
	if err != nil {
		return
	}
	defer stmt.Close()

	// int64 --> string
	orderId := strconv.FormatInt(order.ID, 10)
	_, err = stmt.Exec(order.UserId, order.ProductId, order.OrderStatus, orderId)
	return
}

func (o *OrderManager) SelectById(orderID int64) (order *model.Order, err error) {
	if err = o.Conn(); err != nil {
		return &model.Order{}, err
	}

	sql := "select * from order where id=" + strconv.FormatInt(orderID, 10)
	row, err := o.mysqlConn.Query(sql)
	if err != nil {
		return &model.Order{}, err
	}
	defer row.Close()

	result := datasource.GetResultRow(row)
	if len(result) == 0 {
		return &model.Order{}, err
	}

	order = &model.Order{}
	datasource.DataToStructByTagSql(result, order)
	return
}

func (o *OrderManager) SelectAll() (orderList []*model.Order, err error) {
	if err = o.Conn(); err != nil {
		return nil, err
	}
	sql := `Select * from order`
	rows, err := o.mysqlConn.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := datasource.GetResultRows(rows)
	if len(result) == 0 {
		return nil, err
	}

	for _, v := range result {
		order := &model.Order{}
		datasource.DataToStructByTagSql(v, order)
		orderList = append(orderList, order)
	}
	return
}

func (o *OrderManager) SelectAllWithInfo() (OrderMap map[int]map[string]string, err error) {
	if err = o.Conn(); err != nil {
		return nil, err
	}
	sql := `select o.id,p.product_name,o.order_status 
				from seckill.order as o 
					left join seckill.product as p 
						on o.product_id=p.id`
	rows, err := o.mysqlConn.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return datasource.GetResultRows(rows), err
}
