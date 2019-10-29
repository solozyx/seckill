package main

import (
	"github.com/solozyx/seckill/_prepare/rabbitmq"
)

func main() {
	mqWork := rabbitmq.NewRabbitMQSimple("test_work")
	mqWork.ConsumeSimple()
}
