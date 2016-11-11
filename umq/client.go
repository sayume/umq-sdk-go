package umq

import (
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
)

type httpResult struct {
	Action  string        `json:"Action"`
	RetCode int           `json:"RetCode"`
	Message string        `json:"Message"`
	DataSet []interface{} `json:"DataSet"`
}

type httpResult2 struct {
	Action  string      `json:"Action"`
	RetCode int         `json:"RetCode"`
	Message string      `json:"Message"`
	DataSet interface{} `json:"DataSet"`
}

type getMessagePack struct {
	Action  string      `json:"Action"`
	RetCode int         `json:"RetCode"`
	Message string      `json:"Message, omitempty"`
	Data    MessageInfo `json:"Data"`
}

type wsMessagePack struct {
	Action  string  `json:"Action"`
	RetCode int     `json:"RetCode"`
	Message string  `json:"Message, omitempty"`
	Data    Message `json:"Data"`
}

// CreateQueue 创建queue
func (client *UmqClient) CreateQueue(projectID, couponID, remark, queueName, pushType, qos string) (interface{}, error) {
	if pushType != "Direct" && pushType != "Fanout" {
		return nil, errors.New("Push type can only be 'Direct' or 'Fanout'.")
	}
	req := map[string]string{
		"Action":    "UmqCreateQueue",
		"Region":    client.region,
		"CouponId":  couponID,
		"Remark":    remark,
		"QueueName": queueName,
		"PushType":  pushType,
		"QoS":       qos,
		"ProjectId": client.projectID,
		"PublicKey": client.publicKey,
	}

	res, err := sendAPIHttpRequest(req, client.privateKey, 10)
	if err != nil {
		return nil, err
	}
	var result httpResult2
	err = json.Unmarshal(res, &result)
	if err != nil {
		return nil, err
	}
	if result.RetCode == 0 {
		data := result.DataSet
		queueResult := data.(map[string]interface{})
		return queueResult["QueueId"], nil
	} else {
		return nil, errors.New(result.Message)
	}
}

// DeleteQueue 删除queue
func (client *UmqClient) DeleteQueue(queueId string, projectId string) (interface{}, error) {
	req := map[string]string{
		"Action":    "UmqDeleteQueue",
		"Region":    client.region,
		"QueueId":   queueId,
		"PublicKey": client.publicKey,
	}
	if projectId != "" {
		req["ProjectId"] = projectId
	}

	res, err := sendAPIHttpRequest(req, client.privateKey, 10)
	if err != nil {
		return nil, err
	}
	var result httpResult2
	err = json.Unmarshal(res, &result)
	if err != nil {
		return nil, err
	}
	if result.RetCode == 0 {
		return queueId, nil
	} else {
		return nil, errors.New(result.Message)
	}
}

// ListQueue 获取队列列表
func (client *UmqClient) ListQueue(limit int, offset int, projectId string) (interface{}, error) {
	var resultList []QueueInfo
	req := map[string]string{
		"Action":    "UmqGetQueue",
		"Region":    client.region,
		"Limit":     strconv.Itoa(limit),
		"Offset":    strconv.Itoa(offset),
		"PublicKey": client.publicKey,
	}
	if projectId != "" {
		req["ProjectId"] = projectId
	}

	res, err := sendAPIHttpRequest(req, client.privateKey, 10)
	if err != nil {
		return nil, err
	}
	var result httpResult
	err = json.Unmarshal(res, &result)
	if err != nil {
		return nil, err
	}

	if result.RetCode == 0 {
		data := result.DataSet
		for _, d := range data {
			var queueInfo QueueInfo
			info := d.(map[string]interface{})
			queueInfo.QueueId = info["QueueId"].(string)
			queueInfo.QueueName = info["QueueName"].(string)
			queueInfo.PushType = info["PushType"].(string)
			queueInfo.MsgTTL = int(info["MsgTtl"].(float64))
			queueInfo.CreateTime = int64(info["CreateTime"].(float64))
			queueInfo.HttpAddr = info["HttpAddr"].(string)

			reqPublisher := map[string]string{
				"Action":    "UmqGetRole",
				"Region":    client.region,
				"QueueId":   queueInfo.QueueId,
				"Role":      "Pub",
				"Limit":     "100",
				"Offset":    "0",
				"PublicKey": client.publicKey,
			}
			resPublisher, err := sendAPIHttpRequest(reqPublisher, client.privateKey, 10)
			if err != nil {
				return nil, err
			}
			var publisherResult httpResult
			err = json.Unmarshal(resPublisher, &publisherResult)
			if err != nil {
				return nil, err
			}
			if publisherResult.RetCode != 0 {
				return nil, errors.New(publisherResult.Message)
			}
			pubData := publisherResult.DataSet
			publisherList := make([]Role, 0)
			for _, p := range pubData {
				pub := p.(map[string]interface{})
				publisherList = append(publisherList[:], Role{
					Id:         pub["Id"].(string),
					Token:      pub["Token"].(string),
					CreateTime: int64(pub["CreateTime"].(float64)),
				})
			}
			queueInfo.PublisherList = publisherList

			reqConsumer := map[string]string{
				"Action":    "UmqGetRole",
				"Region":    client.region,
				"QueueId":   queueInfo.QueueId,
				"Role":      "Sub",
				"Limit":     "100",
				"Offset":    "0",
				"PublicKey": client.publicKey,
			}
			resConsumer, err := sendAPIHttpRequest(reqConsumer, client.privateKey, 10)
			if err != nil {
				return nil, err
			}
			var consumerResult httpResult
			err = json.Unmarshal(resConsumer, &consumerResult)
			if err != nil {
				return nil, err
			}
			if consumerResult.RetCode != 0 {
				return nil, errors.New(consumerResult.Message)
			}
			subData := consumerResult.DataSet
			consumerList := make([]Role, 0)
			for _, s := range subData {
				sub := s.(map[string]interface{})
				consumerList = append(consumerList[:], Role{
					Id:         sub["Id"].(string),
					Token:      sub["Token"].(string),
					CreateTime: int64(sub["CreateTime"].(float64)),
				})
			}
			queueInfo.ConsumerList = consumerList

			resultList = append(resultList[:], queueInfo)
		}

		return resultList, nil
	} else {
		return nil, errors.New(result.Message)
	}
}

// CreateRole 创建角色
func (client *UmqClient) CreateRole(queueId string, num int, role string, projectId string) (interface{}, error) {
	if role != "Pub" && role != "Sub" {
		return nil, errors.New("Role can only be Pub or Sub.")
	}

	req := map[string]string{
		"Action":    "UmqCreateRole",
		"Region":    client.region,
		"QueueId":   queueId,
		"Num":       strconv.Itoa(num),
		"Role":      role,
		"PublicKey": client.publicKey,
	}
	if projectId != "" {
		req["ProjectId"] = projectId
	}

	res, err := sendAPIHttpRequest(req, client.privateKey, 10)
	if err != nil {
		return nil, err
	}
	var result httpResult
	err = json.Unmarshal(res, &result)
	if err != nil {
		return nil, err
	}

	resultList := make([]Role, 0)
	if result.RetCode == 0 {
		data := result.DataSet
		for _, d := range data {
			info := d.(map[string]interface{})
			roleInfo := Role{
				Id:         info["Id"].(string),
				Token:      info["Token"].(string),
				CreateTime: int64(info["CreateTime"].(float64)),
			}
			resultList = append(resultList[:], roleInfo)
		}
		return resultList, nil
	} else {
		return nil, errors.New(result.Message)
	}
}

// DeleteRole 删除角色
func (client *UmqClient) DeleteRole(queueId string, roleId string, role string) (interface{}, error) {
	if role != "Pub" && role != "Sub" {
		return nil, errors.New("Role can only be Pub or Sub.")
	}

	req := map[string]string{
		"Action":    "UmqDeleteRole",
		"Region":    client.region,
		"QueueId":   queueId,
		"Role":      role,
		"RoleId":    roleId,
		"PublicKey": client.publicKey,
	}

	res, err := sendAPIHttpRequest(req, client.privateKey, 10)
	if err != nil {
		return nil, err
	}
	var result httpResult
	err = json.Unmarshal(res, &result)
	if err != nil {
		return nil, err
	}

	if result.RetCode == 0 {
		return roleId, nil
	} else {
		return nil, errors.New(result.Message)
	}
}

// NewProducer 创建一个生产者实例
func (client *UmqClient) NewProducer(producerID, producerToken string) *UmqProducer {
	return &UmqProducer{
		client: client,
		token:  producerID + ":" + producerToken,
	}
}

// NewConsumer 创建一个consumer实例
func (client *UmqClient) NewConsumer(consumerID, consumerToken string) *UmqConsumer {
	return newConsumer(client, consumerID, consumerToken)
}

// CreateClient 创建client
func NewClient(config UmqConfig) (*UmqClient, error) {
	baseURL, err := url.Parse(config.Host)
	if err != nil {
		return nil, err
	}
	return &UmqClient{
		email:          config.Account,
		region:         config.Region,
		baseURL:        baseURL,
		publicKey:      config.PublicKey,
		privateKey:     config.PrivateKey,
		organizationID: "1234",
		projectID:      config.ProjectID,
	}, nil
}
