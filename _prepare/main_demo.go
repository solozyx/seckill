package main

import (
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

// 声明变量
// rabbitmq 连接
var conn *amqp.Connection

// rabbitmq channel
var channel *amqp.Channel
var count = 0

const (
	// 队列名称
	queueName = "test_queue"
	exchange  = ""
	mqurl     = "amqp://guest:guest@127.0.0.1:5673"
)

//错误处理函数
func failOnErr(err error, msg string) {
	if err != nil {
		log.Fatalf("%s:%s", msg, err)
		panic(fmt.Sprintf("%s:%s", msg, err))
	}
}

func connect() {
	var err error
	//连接 rabbitmq
	conn, err = amqp.Dial(mqurl)
	failOnErr(err, "failed to connect")
	//获取 channel
	channel, err = conn.Channel()
	failOnErr(err, "failed to open a channel")
}

func close() {
	//1.关闭channel
	channel.Close()
	//2.关闭连接
	conn.Close()
}

//消息生产
func push() {
	// 1.判断是否存在channel
	if channel == nil {
		connect()
	}
	// 2.消息
	message := "Hello simple!"

	// 3.声明队列
	q, err := channel.QueueDeclare(
		queueName,
		false,
		false,
		false,
		false,
		nil,
	)

	// 4.判断错误
	if err != nil {
		fmt.Println(err)
	}

	// 5.生产消息
	channel.Publish(exchange, q.Name, false,
		false, amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
}

// 消费端
func receive() {
	// 1.判断 channel 是否存在
	if channel == nil {
		connect()
	}

	// 2.声明队列
	q, err := channel.QueueDeclare(
		queueName,
		false,
		false,
		false,
		false,
		nil,
	)
	// 3.消费消息
	msg, err := channel.Consume(q.Name, "",
		true,
		false,
		false,
		false,
		nil)
	failOnErr(err, "获取消费信息异常")

	msgForver := make(chan bool)

	go func() {
		for d := range msg {
			// 相同效果 把[]byte类型转化为字符串类型
			// s := queue.BytesToString(&d.Body)
			s := string(d.Body)
			count++
			fmt.Printf("接收信息是 %s -- %d\n", s, count)
		}
	}()

	fmt.Printf("退出请按 CTRL+C \n")

	<-msgForver
}

func main() {
	go func() {
		for {
			push()
			time.Sleep(1 * time.Second)
		}
	}()
	receive()
	fmt.Println("生产消费完成")
	close()
}
