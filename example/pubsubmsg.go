package example

import (
	"fmt"
	"sync/atomic"

	umq "github.com/ucloud/umq-sdk-go/umq"
)

var t = `
UCLOUD目前拥有7大地域（Region），分别是北京一、北京二、广东、浙江、上海、香港及美国加州。每个地域下开设有多个可用区（Zone），每个可用区所提供的网络带宽有所不同，在操作与公网相关的API时，需要主机选定可用区所支持的网络线路。
具体地域和可用区名称，及对应网络线路如下：
`

func RunPubSub(client *umq.UmqClient,
	producerId string, producerToken string,
	consumerId string, consumerToken string,
	queueId string) {
	producer := client.NewProducer(producerId, producerToken)
	consumer1 := client.NewConsumer(consumerId, consumerToken)

	go func() {
		for i := 0; i < 10; i++ {
			go producer.PublishMsg(queueId, t)
		}
	}()

	var count int32 = 0
	consumer1.SubscribeQueue(queueId, func(c chan string, msg umq.Message) {
		// 如果需要新起goroutine，需要手动处理
		go func() {
			fmt.Println("receive message", msg)
			atomic.AddInt32(&count, 1)
			nc := atomic.LoadInt32(&count)
			c <- msg.MsgId
			fmt.Println(nc)
			if nc == 10 {
				consumer1.UnSubscribe(queueId)
			}
		}()
	})

}
