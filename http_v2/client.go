package httpclient

import (
	"encoding/json"
	"fmt"
	"strconv"
	"umq-sdk/go/http_v2/websocket"
)

var (
	OrganizationId = 10000                                  //组织ID
	QueueId        = "qid_xxxx|1"                           //队列ID
	HttpAddr       = "http://xxxxx:6318"                    //服务地址
	PublisherId    = "PID_e349dcbb67521fe082cf88d0892ef3f6" //发布者ID
	PublisherToken = "xxxxx"                                //发布者鉴权Token
	ConsumerId     = "CID_fafd13a67063ddd00d12a4c85664616d" //消费者ID
	ConsumerToken  = "xxxxxx"                               //消费者鉴权Token
	WsAddr         = "http://xxxx:6318/"                    //Websocket服务IP
	WsUrl          = "ws://xxxx:6318/ws"                    //Websocket服务IP
)

type MsgHandler func(c chan string, Msg interface{})

//回执消息
func AckMsg(MsgId string) (interface{}, error) {
	ack_req := map[string]interface{}{
		"Action":     "AckMsg",
		"QueueId":    QueueId,
		"ConsumerId": ConsumerId,
		"MsgId":      MsgId,
	}

	ack_res, err := sendHttpRequest(HttpAddr, ack_req, 100)
	if err != nil {
		return nil, err
	}

	var ack_res_body map[string]interface{}
	err = json.Unmarshal(ack_res, &ack_res_body)
	if err != nil {
		return nil, err
	}

	return ack_res_body, nil
}

//发布消息
func PublishMsg(Content string) (interface{}, error) {
	pub_req := map[string]interface{}{
		"Action":         "PublishMsg",
		"QueueId":        QueueId,
		"OrganizationId": strconv.FormatInt(int64(OrganizationId), 10),
		"PublisherId":    PublisherId,
		"PublisherToken": PublisherToken,
		"Content":        Content,
	}

	pub_res, err := sendHttpRequest(HttpAddr, pub_req, 100)
	if err != nil {
		return nil, err
	}
	var pub_res_body map[string]interface{}
	err = json.Unmarshal(pub_res, &pub_res_body)
	if err != nil {
		return nil, err
	}
	return pub_res_body, nil
}

//主动拉取消息
func GetMsg(Num string) (interface{}, error) {
	get_req := map[string]interface{}{
		"Action":         "GetMsg",
		"QueueId":        QueueId,
		"OrganizationId": strconv.FormatInt(int64(OrganizationId), 10),
		"ConsumerId":     ConsumerId,
		"ConsumerToken":  ConsumerToken,
		"Num":            Num,
	}

	get_res, err := sendHttpRequest(HttpAddr, get_req, 100)
	if err != nil {
		return nil, err
	}
	var get_res_body map[string]interface{}
	err = json.Unmarshal(get_res, &get_res_body)
	if err != nil {
		return nil, err
	}
	return get_res_body, nil
}

//订阅消息
func SubscribeQueue(msgHandler MsgHandler) error {
	type StartConsumeReq struct {
		OrganizationId uint64
		QueueId        string
		ConsumerId     string
		ConsumerToken  string
	}

	wsConn, err := websocket.Dial(WsUrl, "", WsAddr)
	if err != nil {
		fmt.Println("Dial error")
		return err
	}

	fmt.Println("connect websocket successfully")

	ws_data := new(StartConsumeReq)
	ws_data.OrganizationId = uint64(OrganizationId)
	ws_data.QueueId = QueueId
	ws_data.ConsumerId = ConsumerId
	ws_data.ConsumerToken = ConsumerToken

	sub_req := map[string]interface{}{
		"Action": "ConsumeMsg",
		"Data":   ws_data,
	}
	buffer, err := json.Marshal(sub_req)
	if err != nil {
		fmt.Println("Encoding json request error")
		return err
	}

	_, err = wsConn.Write(buffer)
	if err != nil {
		fmt.Println("Send request error")
		return err
	}
	//订阅回包
	var sub_res []byte
	err = websocket.Message.Receive(wsConn, &sub_res)
	if err != nil {
		fmt.Println("Receive sub response error")
		return err
	}

	ackMsg := make(chan string)
	go func() {
		for {
			select {
			case MsgId := <-ackMsg:
				if MsgId == "" {
					fmt.Println("MsgHandler error")
				} else {
					AckMsg(MsgId)
				}
			}
		}
	}()

	for {
		var sub_msg []byte
		//TODO: 重试逻辑
		err = websocket.Message.Receive(wsConn, &sub_msg)
		if err != nil {
			fmt.Println("Receive push msg error")
			break
		}

		var data map[string]interface{}
		err = json.Unmarshal(sub_msg, &data)
		if err != nil {
			fmt.Println("Decoding msg error")
		}

		go msgHandler(ackMsg, data)
	}
	return nil
}

//获取项目ID
func GetOrganizationId(email, projectId string) (interface{}, error) {
	org_req := map[string]interface{}{
		"Action":            "GetOrganizationId",
		"UserEmail":         email,
		"OrganizationAlias": projectId,
	}

	org_res, err := sendHttpRequest(HttpAddr, org_req, 100)
	if err != nil {
		return nil, err
	}
	var org_res_body map[string]interface{}
	err = json.Unmarshal(org_res, &org_res_body)
	if err != nil {
		return nil, err
	}
	return org_res_body, nil
}
