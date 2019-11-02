package main

import (
	"github.com/solozyx/seckill/_prepare/rabbitmq"
)

func main() {
	c1 := rabbitmq.NewRabbitMQRouting("exchange_routing", "routing1")
	c1.ConsumeRouting()
}
