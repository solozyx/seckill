package datasource

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/streadway/amqp"

	"github.com/solozyx/seckill/model"
	"github.com/solozyx/seckill/service"
)

// 连接信息
//const MQURL = "amqp://seckilluser:seckilluser@172.31.96.59:5672/seckill"
const MQURL = "amqp://seckilluser:seckilluser@127.0.0.1:5672/seckill"

type RabbitMQ struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	// 队列名称
	QueueName string
	// 交换机名称
	Exchange string
	// bind Key 名称
	Key string
	// 连接信息
	MQUrl string
	sync.Mutex
}

func newRabbitMQ(queueName string, exchange string, key string) *RabbitMQ {
	return &RabbitMQ{QueueName: queueName, Exchange: exchange, Key: key, MQUrl: MQURL}
}

// 断开rabbitmq通信信道channel 和 连接connection
func (r *RabbitMQ) Destroy() {
	r.ch.Close()
	r.conn.Close()
}

// 错误处理
func (r *RabbitMQ) failOnErr(err error, message string) {
	if err != nil {
		panic(fmt.Sprintf("%s:%s", message, err))
	}
}

// 简单模式下RabbitMQ实例
func NewRabbitMQSimple(queueName string) *RabbitMQ {
	// 创建RabbitMQ实例 简单模型只需要队列名称
	rabbitmq := newRabbitMQ(queueName, "", "")
	var err error
	// 获取connection
	rabbitmq.conn, err = amqp.Dial(rabbitmq.MQUrl)
	rabbitmq.failOnErr(err, "failed to connect rabbitmq server")
	// 获取rabbitmq通信信道channel
	rabbitmq.ch, err = rabbitmq.conn.Channel()
	rabbitmq.failOnErr(err, "failed to open a rabbitmq channel")
	return rabbitmq
}

// 直接模式队列生产
func (r *RabbitMQ) PublishSimple(message string) error {
	r.Lock()
	defer r.Unlock()

	// 1.申请队列 如果队列不存在会自动创建 存在则跳过创建
	_, err := r.ch.QueueDeclare(
		r.QueueName,
		// 不持久化
		false,
		// 不自动删除
		false,
		// 不具有排他性
		false,
		// 不阻塞处理
		false,
		// 额外属性
		nil,
	)
	if err != nil {
		return err
	}
	// 调用rabbitmq信道 发送消息到队列中
	r.ch.Publish(
		r.Exchange,
		r.QueueName,
		// 如果为true 根据自身 exchange 类型和 routingKey 规则无法找到符合条件的队列会把消息返还给发送者
		false,
		// 如果为true 当 exchange 发送消息到队列后发现队列上没有消费者 则会把消息返还给发送者
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
	return nil
}

// simple 模式下消费者
func (r *RabbitMQ) ConsumeSimple(orderService service.IOrderService, productService service.IProductService) {
	// 1.申请队列 如果队列不存在会自动创建 存在则跳过创建
	q, err := r.ch.QueueDeclare(
		r.QueueName,
		// 不持久化
		false,
		// 不自动删除
		false,
		// 不具有排他性
		false,
		// 不阻塞处理
		false,
		// 额外属性
		nil,
	)
	if err != nil {
		fmt.Println(err)
	}

	// 消费者流控防止MySQL暴库 autoAck(设置为false) + Qos 保证消费端每次只消费1个消息
	// 当消费完成1个消息 手动Ack告诉server端发另外1个消息过来 不要一次性消费大量消息
	r.ch.Qos(
		// 当前消费者一次能接受的最大消息数量,当前数量的消息没消费完,则不接收新消息
		1,
		// 服务器传递的最大容量 以八位字节为单位
		0,
		// 如果设置为true 则整个通信信道channel全局可用 这里只对自己的消费队列进行设置
		false,
	)

	// 接收消息
	msgs, err := r.ch.Consume(
		// queue
		q.Name,
		// consumer 用来区分多个消费者
		"",
		// auto-ack 是否自动应答 这里用手动应答控制消费端流控
		false,
		// exclusive 是否独有
		false,
		// no-local 设置为true表示不能将同一个connection中生产者发送的消息传递给这个connection中的消费者
		false,
		// no-wait 是否阻塞
		false,
		// args
		nil,
	)
	if err != nil {
		fmt.Println(err)
	}

	forever := make(chan bool)
	// 启用协程处理消息
	go func() {
		for d := range msgs {
			fmt.Printf("rabbitmq consumer received a message = %v", d.Body)
			message := &model.Message{}
			err := json.Unmarshal(d.Body, message)
			if err != nil {
				fmt.Println(err)
			}
			// TODO:WARNING 执行2次MySQL插入 有可能导致数据不一致
			// 插入订单
			_, err = orderService.InsertOrderByMessage(message)
			if err != nil {
				fmt.Println(err)
			}
			// 扣除商品数量
			err = productService.SubNumberOne(message.ProductID)
			if err != nil {
				fmt.Println(err)
			}
			// 消费端设置为手动Ack 当消费完成1个消息Ack告诉server端该消息消费完成 server把该消息删除
			// true  表示确认所有未确认消息,在批量消费设置true
			// false 表示确认当前消息
			// TODO:NOTICE 如果不写 d.Ack() 消费完成不做应答给server端 消费完成的消息不会从server端队列删除
			//  这种情况下,如果消费端断开了,server端会把已经投递给消费端而未删除的消息重新回到队列
			//  造成已经消费的消息有可能被其他消费者消费 [消息重复消费] 消息不满足幂等性 后果严重
			d.Ack(false)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
