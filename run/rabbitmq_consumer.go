package main

import (
	"fmt"

	"github.com/solozyx/seckill/conf"
	"github.com/solozyx/seckill/dao"
	"github.com/solozyx/seckill/datasource"
	"github.com/solozyx/seckill/service"
)

func main() {
	db, err := datasource.NewMysqlConn()
	if err != nil {
		fmt.Println(err)
	}
	productDao := dao.NewProductManager(db)
	productService := service.NewProductService(productDao)
	orderDao := dao.NewOrderManager(db)
	orderService := service.NewOrderService(orderDao)

	rabbitmqConsumerSimple := datasource.NewRabbitMQSimple(conf.SeckillQueueName)
	rabbitmqConsumerSimple.ConsumeSimple(orderService, productService)
}
