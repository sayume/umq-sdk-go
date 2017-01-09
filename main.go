package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	umq "github.com/ucloud/umq-sdk-go/umq"
)

type Configuration struct {
	// Host 就是队列所在的URL
	Host string
	// 账户的公钥
	PublicKey string
	// 账户的私钥
	PrivateKey string
	// 地域，参考 https://docs.ucloud.cn/api/summary/regionlist
	Region string
	// 账户名
	Account string
	// 项目ID，类似 org-0ex02v
	ProjectID string
	// 生产者ID
	ProducerID string
	// 生产者Token
	ProducerToken string
	// 消费者ID
	ConsumerID string
	// 消费者Token
	ConsumerToken string
	// 主题名称
	TopicName string `json:"QueueID"`
}

func main() {
	file, err := os.Open("conf.json")
	if err != nil {
		panic(err)
	}
	decoder := json.NewDecoder(file)
	conf := Configuration{}
	err = decoder.Decode(&conf)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(conf)

	//配置umq http client，其中HttpAddr和WsAddr可从console的umq主页上查询，PublicKey和PrivateKey可从console的账户主页上获得。
	client, err := umq.NewClient(umq.UmqConfig{
		Host:       conf.Host,
		PublicKey:  conf.PublicKey,
		PrivateKey: conf.PrivateKey,
		Region:     conf.Region,
		Account:    conf.Account,
		ProjectID:  conf.ProjectID,
	})

	if err != nil {
		fmt.Println(err)
		return
	}

	// Create consumer and producer
	consumer := client.NewConsumer(conf.ConsumerID, conf.ConsumerToken)
	producer := client.NewProducer(conf.ProducerID, conf.ProducerToken)
	// Publish a message to a topic
	msgID, err := producer.PublishMsg(conf.TopicName, strings.NewReader("message body"))
	if err != nil {
		fmt.Println("publish error", err)
		return
	}
	fmt.Println("publish done", msgID)
	// Get messages from topic
	msgs, err := consumer.GetMsg(conf.TopicName, 10)
	if err != nil {
		fmt.Println("get message error ", err)
		return
	}
	fmt.Println("get message done", msgs)
	messageIDList := make([]string, 0)
	for _, m := range msgs {
		messageIDList = append(messageIDList, m.MessageID)
	}
	// Ack messages
	fails, err := consumer.AckMsg(conf.TopicName, messageIDList)
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(fails) != 0 {
		fmt.Println("We have unacked messages.")
		return
	}
	fmt.Println("ack message done")
	// Create a new topic
	isSuccess, err := client.CreateTopic("topic name", "pubsub or queue", "topic description", -1)
	if err != nil || !isSuccess {
		fmt.Println(err)
		return
	}

	// List topics
	queueInfoList, err := client.ListTopic(10, 0)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Queue info list")
	fmt.Println(queueInfoList)
	// Create publisher
	isSuccess, err = client.CreateRole("Pub")
	if err != nil || !isSuccess {
		fmt.Println(err)
		return
	}
	// Create consumer
	isSuccess, err = client.CreateRole("Sub")
	if err != nil || !isSuccess {
		fmt.Println(err)
		return
	}
	// List roles
	roleInfoList, err := client.ListRole(10, 0, "Pub")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("List role info")
	fmt.Println(roleInfoList)
	// Delete an existing topic
	isSuccess, err = client.DeleteTopic("topic name")
	if err != nil || !isSuccess {
		fmt.Println(err)
		return
	}

	return
}
