package main

import (
	"github.com/solozyx/seckill/_prepare/rabbitmq"
)

func main() {
	c2 := rabbitmq.NewRabbitMQRouting("exchange_routing", "routing2")
	c2.ConsumeRouting()
}
