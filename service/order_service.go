package service

import (
	"github.com/solozyx/seckill/dao"
	"github.com/solozyx/seckill/model"
)

type IOrderService interface {
	// 入参订单id
	GetOrderById(int64) (*model.Order, error)
	DeleteOrderById(int64) bool
	UpdateOrder(*model.Order) error
	// 出参订单id
	InsertOrder(*model.Order) (int64, error)
	GetAllOrder() ([]*model.Order, error)
	GetAllOrderInfo() (map[int]map[string]string, error)
	// 让rabbitmq消费端调用 创建订单 返回订单id
	InsertOrderByMessage(*model.Message) (int64, error)
}

type OrderService struct {
	orderDao dao.IOrder
}

func NewOrderService(dao dao.IOrder) IOrderService {
	return &OrderService{dao}
}

func (o *OrderService) GetOrderById(orderID int64) (order *model.Order, err error) {
	return o.orderDao.SelectById(orderID)
}

func (o *OrderService) DeleteOrderById(orderID int64) bool {
	return o.orderDao.Delete(orderID)
}

func (o *OrderService) UpdateOrder(order *model.Order) error {
	return o.orderDao.Update(order)
}

func (o *OrderService) InsertOrder(order *model.Order) (orderID int64, err error) {
	return o.orderDao.Insert(order)
}

func (o *OrderService) GetAllOrder() ([]*model.Order, error) {
	return o.orderDao.SelectAll()
}

func (o *OrderService) GetAllOrderInfo() (map[int]map[string]string, error) {
	return o.orderDao.SelectAllWithInfo()
}

// 让rabbitmq消费端调用 创建订单 返回订单id
func (o *OrderService) InsertOrderByMessage(message *model.Message) (orderID int64, err error) {
	order := &model.Order{
		UserId:      message.UserID,
		ProductId:   message.ProductID,
		OrderStatus: model.OrderSuccess,
	}
	return o.InsertOrder(order)
}
