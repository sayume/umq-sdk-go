package main

import (
	"encoding/json"
	"os"

	. "github.com/ucloud/umq-sdk-go/example"
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
	// queue ID
	QueueID string
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
		panic(err)
	}

	//配置umq http client，其中HttpAddr和WsAddr可从console的umq主页上查询，PublicKey和PrivateKey可从console的账户主页上获得。
	client, err := umq.CreateClient(umq.UmqConfig{
		Host:       conf.Host,
		PublicKey:  conf.PublicKey,
		PrivateKey: conf.PrivateKey,
		Region:     conf.Region,
		Account:    conf.Account,
		ProjectID:  conf.ProjectID,
	})

	if err != nil {
		panic(err)
	}

	RunPubSub(client, conf.ProducerID, conf.ProducerToken,
		conf.ConsumerID, conf.ConsumerToken,
		conf.QueueID)

	return
}
