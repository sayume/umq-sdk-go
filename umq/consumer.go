package umq

import (
	"encoding/json"
	"fmt"
	"math/rand"
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
func (consumer *UmqConsumer) GetMsg(queueId string, num int) (*MessageInfo, error) {
	req := map[string]string{
		"Action":         "GetMsg",
		"QueueId":        queueId,
		"Region":         consumer.client.region,
		"ConsumerId":     consumer.consumerID,
		"ConsumerToken":  consumer.consumerToken,
		"OrganizationId": consumer.client.organizationID,
		"Num":            strconv.Itoa(num),
	}

	res, err := sendHTTPRequest(consumer.client.httpAddr, req, 10)
	if err != nil {
		return nil, err
	}
	var resBody getMessagePack
	err = json.Unmarshal(res, &resBody)
	if err != nil {
		return nil, err
	}

	if resBody.RetCode != 0 {
		return nil, fmt.Errorf("Fail to get message: %s", resBody.Message)
	}
	return &resBody.Data, nil
}

// AckMsg ack queueid对应topic的一条消息，msgId为该消息的message id
func (consumer *UmqConsumer) AckMsg(queueId, msgId string) error {
	req := map[string]string{
		"Action":        "AckMsg",
		"Region":        consumer.client.region,
		"QueueId":       queueId,
		"ConsumerId":    consumer.consumerID,
		"ConsumerToken": consumer.consumerToken,
		"MsgId":         msgId,
	}

	resp, err := sendHTTPRequest(consumer.client.httpAddr, req, 10)
	if err != nil {
		return err
	}

	var resBody map[string]interface{}
	err = json.Unmarshal(resp, &resBody)
	if err != nil {
		return err
	}

	if resultCode, ok := resBody["RetCode"].(float64); !ok || int(resultCode) != 0 {
		errMsg := ""
		if msg, ok := resBody["Message"].(string); ok {
			errMsg = msg
		}
		return fmt.Errorf("Fail to ack message: %s", errMsg)
	}
	return nil
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
					consumer.AckMsg(queueId, MsgId)
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
	wsConn, err := websocket.Dial(consumer.client.wsUrl, "", consumer.client.wsAddr)
	if err != nil {
		return nil, err
	}
	orgId, _ := strconv.ParseUint(consumer.client.organizationID, 10, 64)
	wsData := startConsumeReq{
		OrganizationId: orgId,
		QueueId:        queueId,
		ConsumerId:     consumer.consumerID,
		ConsumerToken:  consumer.consumerToken,
	}

	subReq := map[string]interface{}{
		"Action": "ConsumeMsg",
		"Data":   wsData,
	}

	buffer, err := json.Marshal(subReq)
	if err != nil {
		wsConn.Close()
		return nil, err
	}

	_, err = wsConn.Write(buffer)
	if err != nil {
		wsConn.Close()
		return nil, err
	}

	//订阅回包
	var subRes []byte
	err = websocket.Message.Receive(wsConn, &subRes)
	if err != nil {
		wsConn.Close()
		return nil, err
	}
	return wsConn, nil
}
