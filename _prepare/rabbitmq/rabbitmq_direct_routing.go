package rabbitmq

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

// 路由模式
func NewRabbitMQRouting(exchangeName string, routingKey string) *rabbitmq {
	return newRabbitMQ("", exchangeName, routingKey)
}

// 路由模式 生产消息
func (r *rabbitmq) PublishRouting(message string) {
	err := r.c.ExchangeDeclare(
		r.Exchange,
		// 直连 direct 广播 fanout
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	r.failOnErr(err, "Failed to declare an exchange")

	// 2.发送消息
	err = r.c.Publish(
		r.Exchange,
		// routing key 要设置
		r.Key,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
}

// 路由模式 消费消息
func (r *rabbitmq) ConsumeRouting() {
	err := r.c.ExchangeDeclare(
		r.Exchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	r.failOnErr(err, "Failed to declare an exchange")

	q, err := r.c.QueueDeclare(
		// 这里注意队列名称不要写
		"",
		false,
		false,
		true,
		false,
		nil,
	)
	r.failOnErr(err, "Failed to declare a queue")

	// routing key 绑定队列到交换机
	err = r.c.QueueBind(
		q.Name,
		// 需要绑定 routing key
		r.Key,
		r.Exchange,
		false,
		nil)

	//消费消息
	msgs, err := r.c.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
		}
	}()

	fmt.Printf("退出请按 CTRL+C \n")

	<-forever
}
