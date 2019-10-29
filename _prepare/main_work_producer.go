package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/solozyx/seckill/_prepare/rabbitmq"
)

func main() {
	mqWork := rabbitmq.NewRabbitMQSimple("test_work")
	for i := 0; i <= 100; i++ {
		mqWork.PublishSimple("test_work_" + strconv.Itoa(i))
		time.Sleep(1 * time.Second)
		fmt.Println(i)
	}
	mqWork.Destroy()
}
