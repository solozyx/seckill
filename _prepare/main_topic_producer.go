package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/solozyx/seckill/_prepare/rabbitmq"
)

func main() {
	p1 := rabbitmq.NewRabbitMQTopic("exchange_topic", "test.topic.one")
	p2 := rabbitmq.NewRabbitMQTopic("exchange_topic", "test.topic.two")
	for i := 0; i <= 10; i++ {
		p1.PublishTopic("topic.one " + strconv.Itoa(i))
		p2.PublishTopic("topic.two " + strconv.Itoa(i))
		time.Sleep(1 * time.Second)
		fmt.Println(i)
	}
	p1.Destroy()
	p2.Destroy()
}
