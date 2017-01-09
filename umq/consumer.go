package umq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/ucloud/umq-sdk-go/umq/websocket"
)

type startConsumeReq struct {
	OrganizationId uint64
	QueueId        string
	ConsumerId     string
	ConsumerToken  string
}

type subscribeInfo struct {
	subscribe      bool
	lastConnTime   time.Time
	stop           chan *subscribeInfo
	conn           *websocket.Conn
	retryConnTimes uint8
	mutex          *sync.Mutex
}

// UmqConsumer consumer的实例
type UmqConsumer struct {
	client        *UmqClient
	consumerID    string
	consumerToken string
	subInfo       map[string]*subscribeInfo
	mutex         *sync.Mutex
	token         string
}

type GetMsgResponse struct {
	Count    int
	Messages []Message
}

type AckMsgRequest struct {
	MessageID []string `json:"MessageID"`
}

type AckMsgResponse struct {
	FailMessageID []string `json:"FailMessageID"`
}

func newConsumer(client *UmqClient, consumerID, consumerToken string) *UmqConsumer {
	return &UmqConsumer{
		client:        client,
		consumerID:    consumerID,
		consumerToken: consumerToken,
		subInfo:       make(map[string]*subscribeInfo),
		mutex:         &sync.Mutex{},
	}
}

// GetMsg 获取queueId对应的topic的num条消息
func (consumer *UmqConsumer) GetMsg(queueId string, num int) ([]Message, error) {
	url := *(consumer.client.baseURL)
	url.Path = fmt.Sprintf("/%s/%s/message", consumer.client.projectID, queueId)
	params := url.Query()
	params.Set("count", strconv.Itoa(num))
	url.RawQuery = params.Encode()
	resp := GetMsgResponse{}
	err := sendHTTPRequest(url.String(), "GET", nil, consumer.token, &resp, 0)
	if err != nil {
		return nil, err
	}
	return resp.Messages, nil
}

// AckMsg ack queueid对应topic的一条消息，msgId为该消息的message id
func (consumer *UmqConsumer) AckMsg(queueId string, msgId []string) (msgID []string, err error) {
	url := *(consumer.client.baseURL)
	url.Path = fmt.Sprintf("/%s/%s/message", consumer.client.projectID, queueId)
	req := &AckMsgRequest{
		MessageID: msgId,
	}
	data, err := json.Marshal(req)
	if err != nil {
		return make([]string, 0), err
	}
	resp := AckMsgResponse{}
	err = sendHTTPRequest(url.String(), "DELETE", bytes.NewReader(data), consumer.token, &resp, 0)
	if resp.FailMessageID == nil {
		return make([]string, 0), fmt.Errorf("Fail to ack message")
	}
	return resp.FailMessageID, err
}

// UnSubscribe 停止订阅queueId指向的topic
func (consumer *UmqConsumer) UnSubscribe(queueId string) error {
	var (
		info *subscribeInfo
		ok   bool
	)
	consumer.mutex.Lock()
	if info, ok = consumer.subInfo[queueId]; !ok {
		consumer.mutex.Unlock()
		return nil
	}
	delete(consumer.subInfo, queueId)
	info.subscribe = false
	consumer.mutex.Unlock()
	info.stop <- info
	return nil
}

// SubscribeQueue 订阅queueId指向的topic
// 成功订阅之后，消息会通过msgHandler回调
func (consumer *UmqConsumer) SubscribeQueue(queueId string, msgHandler MsgHandler) error {
	consumer.mutex.Lock()
	if _, ok := consumer.subInfo[queueId]; ok {
		consumer.mutex.Unlock()
		return fmt.Errorf("Already subscribed")
	}
	consumer.mutex.Unlock()

	conn, err := consumer.handshake(queueId)
	if err != nil {
		return err
	}
	subInfo := &subscribeInfo{
		subscribe:      true,
		lastConnTime:   time.Now(),
		stop:           make(chan *subscribeInfo),
		conn:           conn,
		retryConnTimes: 0,
		mutex:          &sync.Mutex{},
	}
	ackMsg := make(chan string)

	go func() {
		for {
			select {
			case MsgId := <-ackMsg:
				if MsgId == "" {
				} else {
					consumer.AckMsg(queueId, []string{MsgId})
				}
			case info := <-subInfo.stop:
				info.mutex.Lock()
				info.subscribe = false
				info.conn.Close()
				info.mutex.Unlock()
				return
			}
		}
	}()

	consumer.mutex.Lock()
	consumer.subInfo[queueId] = subInfo
	consumer.mutex.Unlock()

	for {
		err = consumer.loopReceive(subInfo.conn, ackMsg, msgHandler)
		connected, err := consumer.reconnect(queueId)
		if err != nil {
			return err
		}
		if !connected {
			return nil
		}
	}
}

func (consumer *UmqConsumer) loopReceive(conn *websocket.Conn, ackMsg chan string, msgHandler MsgHandler) error {
	for {
		var msgBuf []byte
		err := websocket.Message.Receive(conn, &msgBuf)
		if err != nil {
			return err
		}

		var data wsMessagePack
		err = json.Unmarshal(msgBuf, &data)
		if err != nil {
			return err
		}
		msgHandler(ackMsg, data.Data)
	}
}

func (consumer *UmqConsumer) reconnect(queueId string) (connected bool, err error) {
	var (
		target *subscribeInfo
		ok     bool
		conn   *websocket.Conn
	)

	connected = false
	consumer.mutex.Lock()
	if target, ok = consumer.subInfo[queueId]; !ok {
		consumer.mutex.Unlock()
		return false, nil
	}
	consumer.mutex.Unlock()
	target.conn.Close()
	for {
		target.mutex.Lock()
		sub := target.subscribe
		target.mutex.Unlock()
		if !sub {
			return
		}
		conn, err = consumer.handshake(queueId)
		if err != nil {
			target.retryConnTimes++
			sleepTime := 1000 * int(target.retryConnTimes)
			if sleepTime > 10000 {
				// maximum period is 10 seconds
				sleepTime = 10000
			}
			sleepTime += rand.Intn(1000)
			time.Sleep(time.Duration(sleepTime) * time.Millisecond)
			continue
		}
		target.lastConnTime = time.Now()
		target.retryConnTimes = 0
		break
	}
	target.mutex.Lock()
	target.conn = conn
	target.mutex.Unlock()
	connected = true
	return
}

func (consumer *UmqConsumer) handshake(queueId string) (*websocket.Conn, error) {
	path := fmt.Sprintf("/%s/%s/message/subscription?permits=1000", consumer.client.projectID, queueId)
	header := http.Header{}
	header.Add("content-type", "application/json")
	header.Add("Authorization", consumer.token)
	wsConn, err := websocket.Dial("ws:"+consumer.client.wsUrl+path, "", header)
	if err != nil {
		return nil, err
	}
	return wsConn, nil
}
