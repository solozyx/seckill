package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/solozyx/seckill/_prepare/rabbitmq"
)

func main() {
	mqPubSub := rabbitmq.NewRabbitMQPubSub("exchange_pubsub")
	for i := 0; i < 100; i++ {
		mqPubSub.PublishPub("订阅模式生产第" + strconv.Itoa(i) + "条" + "消息数据")
		fmt.Println("订阅模式生产第" + strconv.Itoa(i) + "条" + "消息数据")
		time.Sleep(1 * time.Second)
	}
	mqPubSub.Destroy()
}
