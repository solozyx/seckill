package rabbitmq

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

// 创建简单Simple模式下RabbitMQ实例
func NewRabbitMQSimple(queueName string) *rabbitmq {
	// Simple模式 默认使用 default exchange
	return newRabbitMQ(queueName, "", "")
}

// Simple模式 生产消息
func (r *rabbitmq) PublishSimple(message string) {
	// 1.向 rabbitmq server 声明申请队列
	// 如果队列不存在会自动创建 存在则跳过创建
	// 保证消息一定能发到队列中
	_, err := r.c.QueueDeclare(
		// 队列名称
		r.QueueName,
		// durable 消息数据是否持久化 通常设置false非持久化
		// 消息存储到队列中,没被消费,服务重启,消息数据就删除了
		false,
		// autoDelete 是否自动删除,消费者断开连接,是否把消息从队列中删除
		// 通常设置为false 不做自动删除
		false,
		// exclusive 是否具有排他性 true表示创建只对自己可见的队列 其他用户不能访问
		false,
		// noWait 是否阻塞处理 生产者发送消息后是否等待服务端响应 false不等待服务端响应
		false,
		// args 额外属性
		nil,
	)
	if err != nil {
		fmt.Println(err)
	}

	// 调用通信信道Channel 发送消息到队列中
	r.c.Publish(
		// Simple模式 默认使用 default exchange 是 direct类型
		r.Exchange,
		r.QueueName,
		// mandatory 强制性的 推荐设置为 false
		// 如果为 true 根据自身exchange类型和routing key规则
		// 如果无法找到符合条件的队列 会把消息返还给生产者
		false,
		// immediate 立即的 推荐设置为 false
		// 如果为 true 当exchange发送消息到队列后 发现队列上没有消费者
		// 消息不会存储在队列中 会把消息返还给生产者
		// 为 false 队列上没有消费者 消息也会暂时存储在队列中
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
}

// Simple模式 消费消息
func (r *rabbitmq) ConsumeSimple() {
	q, err := r.c.QueueDeclare(
		r.QueueName,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		fmt.Println(err)
	}

	// 接收消息
	msgs, err := r.c.Consume(
		// 队列
		q.Name,
		// consumer 用来区分多个消费者 这里不区分消费者
		"",
		// autoAck 是否自动应答
		// 默认true 表示消费者接收到1个消息消费完毕,是否自动告诉rabbitmq服务端该消息消费完成
		// 服务端可以把消息从队列中删除了
		// 设置false 需要手动实现Ack回调函数 手动通知服务端消费完成,服务端才把消息从队列删除
		true,
		// exclusive 是否具有排他性 该用户创建的队列只有自己可见 其他用户不可见 false没有排他性
		false,
		// noLocal 设置为true 表示不能将同一个Connection中生产的消息传递给这个Connection中的消费者
		false,
		// noWait 消费时队列是否阻塞 消费完这个消息 下个消息才能给到消费者 false表示阻塞
		false,
		// args 其他参数
		nil,
	)
	if err != nil {
		fmt.Println(err)
	}

	forever := make(chan bool)

	// 启用协程处理消息
	go func() {
		for d := range msgs {
			// 消息逻辑处理 可以自行设计逻辑
			log.Printf("Received a message: %s", d.Body)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")

	<-forever
}
