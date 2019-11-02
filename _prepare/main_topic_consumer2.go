package main

import (
	"github.com/solozyx/seckill/_prepare/rabbitmq"
)

func main() {
	c2 := rabbitmq.NewRabbitMQTopic("exchange_topic", "test.*.two")
	c2.ConsumeTopic()
}
