package rabbitmq

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

// 话题模式
func NewRabbitMQTopic(exchangeName string, routingKey string) *rabbitmq {
	return newRabbitMQ("", exchangeName, routingKey)
}

// 话题模式 生产消息
func (r *rabbitmq) PublishTopic(message string) {
	err := r.c.ExchangeDeclare(
		r.Exchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	r.failOnErr(err, "Failed to declare an exchange")

	err = r.c.Publish(
		r.Exchange,
		r.Key,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
}

// 话题模式接受消息
// 要注意 key 规则
// 匹配符 * 用于匹配一个单词
// 匹配符 # 用于匹配多个单词可以是零个
// 匹配 topic.* 表示匹配 topic.hello
// 但是 topic.hello.one 需要用 topic.# 才能匹配到
func (r *rabbitmq) ConsumeTopic() {
	err := r.c.ExchangeDeclare(
		r.Exchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	r.failOnErr(err, "Failed to declare an exchange")

	q, err := r.c.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
	)
	r.failOnErr(err, "Failed to declare a queue")

	err = r.c.QueueBind(
		q.Name,
		// 在 PubSub模式下这里的key要为空
		r.Key,
		r.Exchange,
		false,
		nil)

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
