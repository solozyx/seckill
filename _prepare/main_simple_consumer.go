package main

import (
	"github.com/solozyx/seckill/_prepare/rabbitmq"
)

func main() {
	mqSimple := rabbitmq.NewRabbitMQSimple("test_simple")
	mqSimple.ConsumeSimple()
}
