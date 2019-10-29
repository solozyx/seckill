package rabbitmq

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

// 订阅模式
func NewRabbitMQPubSub(exchangeName string) *rabbitmq {
	// 不需要设置 queue
	// 设置 exchange
	// 不需要设置 routing key
	return newRabbitMQ("", exchangeName, "")
}

// 订阅模式 生产消息
func (r *rabbitmq) PublishPub(message string) {
	// 1.申请声明交换机 如果存在该交换机则不创建 如果不存在则创建
	err := r.c.ExchangeDeclare(
		// 交换机名称
		r.Exchange,
		// 交换机类型 发布订阅模式 使用fanout广播类型
		"fanout",
		// durable 是否持久化
		true,
		// autoDelete 是否自动删除
		false,
		// internal 设置true表示这个exchange不可以被client用来推送消息,仅用来进行exchange和exchange之间的绑定
		// 通常设置为false
		false,
		// noWait 是否阻塞 false表示非不阻塞就是阻塞
		false,
		// 其他参数
		nil,
	)
	r.failOnErr(err, "Failed to declare an exchange")

	// 2.发送消息
	err = r.c.Publish(
		// 发送消息到哪个交换机
		r.Exchange,
		// routing key 不设置
		"",
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
}

// 订阅模式 消费消息
func (r *rabbitmq) RecieveSub() {
	// 1.申请声明交换机
	err := r.c.ExchangeDeclare(
		r.Exchange,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	r.failOnErr(err, "Failed to declare an exchange")

	// 2.申请声明队列
	q, err := r.c.QueueDeclare(
		// queueName 这里注意队列名称不要写留空 表示随机生成队列
		"",
		false,
		false,
		// exclusive 排他性 true 表示排他
		true,
		false,
		nil,
	)
	r.failOnErr(err, "Failed to declare a queue")

	// 3.绑定队列到交换机
	err = r.c.QueueBind(
		// 队列名称 使用服务端随机生成的该队列名称
		q.Name,
		// 在pub/sub模式 这里的 routing key 必须留空 否则就不是订阅模式
		"",
		// 交换机
		r.Exchange,
		false,
		nil)

	// 4.消费消息
	msgs, err := r.c.Consume(
		// 消费的队列名称
		q.Name,
		"",
		// autoAck 是否自动应答 true自动应答 false需要自己实现应答
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
