package main

import (
	"github.com/solozyx/seckill/_prepare/rabbitmq"
)

func main() {
	c1 := rabbitmq.NewRabbitMQTopic("exchange_topic", "#")
	c1.ConsumeTopic()
}
