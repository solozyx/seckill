package rabbitmq

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

// 连接信息 amqp://用户名:密码@192.168.174.134:5672/虚拟主机
const MQ_URL = "amqp://root:root@192.168.174.134:5672/seckill"

//rabbitMQ结构体
type rabbitmq struct {
	// 连接
	conn *amqp.Connection
	// 通信信道
	c *amqp.Channel
	// url
	url string
	// 队列
	QueueName string
	// 交换机
	Exchange string
	// bind Key routing key
	Key string
}

func newRabbitMQ(queueName string, exchange string, key string) *rabbitmq {
	mq := &rabbitmq{
		QueueName: queueName,
		Exchange:  exchange,
		Key:       key,
		url:       MQ_URL,
	}

	var err error
	// 获取通信连接 Connection
	mq.conn, err = amqp.Dial(mq.url)
	mq.failOnErr(err, "failed to connect rabbitmq server")
	// 获取通信信道 Channel
	mq.c, err = mq.conn.Channel()
	mq.failOnErr(err, "failed to open a 通信信道")
	return mq
}

// 断开通信信道 和 连接
func (r *rabbitmq) Destroy() {
	r.c.Close()
	r.conn.Close()
}

// 错误处理
func (r *rabbitmq) failOnErr(err error, message string) {
	if err != nil {
		log.Fatalf("%s:%s", message, err)
		panic(fmt.Sprintf("%s:%s", message, err))
	}
}
