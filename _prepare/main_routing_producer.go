package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/solozyx/seckill/_prepare/rabbitmq"
)

func main() {
	p1 := rabbitmq.NewRabbitMQRouting("exchange_routing", "routing1")
	p2 := rabbitmq.NewRabbitMQRouting("exchange_routing", "routing2")
	for i := 0; i <= 10; i++ {
		p1.PublishRouting("routing1 " + strconv.Itoa(i))
		p2.PublishRouting("routing2 " + strconv.Itoa(i))
		time.Sleep(1 * time.Second)
		fmt.Println(i)
	}
}
