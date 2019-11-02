package main

import (
	"github.com/solozyx/seckill/_prepare/rabbitmq"
)

func main() {
	mqPubSub := rabbitmq.NewRabbitMQPubSub("exchange_pubsub")
	mqPubSub.ConsumeSub()
}
