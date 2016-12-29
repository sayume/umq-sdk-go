package umq

import (
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
	"sync"
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

type getTopicResult struct {
	Action  string        `json:"Action`
	RetCode int           `json:"RetCode`
	Message string        `json:"Message"`
	Queues  []interface{} `json:"Queues"`
}

type getRoleResult struct {
	Action  string `json:"Action`
	RetCode int    `json:"RetCode`
	Message string `json:"Message"`
	Roles   []Role `json:"Roles"`
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

type UmqClient struct {
	email          string
	region         string
	httpAddr       string
	wsUrl          string
	wsAddr         string
	publicKey      string
	privateKey     string
	projectID      string // 这个是缓存的project id
	organizationID string // 这个是转换出来的数字的org id
	baseURL        *url.URL
	apiUrl         string
}

const defaultTimeout = 60 //second

// CreateQueue 创建queue
func (client *UmqClient) CreateTopic(queueName, queueType, description string, messageTTL int) (bool, error) {
	req := map[string]string{
		"Action":    "CreateUMQTopic",
		"Region":    client.region,
		"QueueName": queueName,
		"QueueType": queueType,
		"ProjectId": client.projectID,
		"PublicKey": client.publicKey,
	}
	if messageTTL > 0 {
		req["MessageTTL"] = strconv.Itoa(messageTTL)
	}
	if description != "" {
		req["Description"] = description
	}

	res, err := sendAPIHttpRequest(req, client.privateKey, defaultTimeout)
	if err != nil {
		return false, err
	}
	var result httpResult
	err = json.Unmarshal(res, &result)
	if err != nil {
		return false, err
	}
	if result.RetCode == 0 {
		return true, nil
	} else {
		return false, errors.New(result.Message)
	}
}

// DeleteQueue 删除queue
func (client *UmqClient) DeleteTopic(queueName string) (bool, error) {
	req := map[string]string{
		"Action":    "DeleteUMQTopic",
		"Region":    client.region,
		"QueueName": queueName,
		"PublicKey": client.publicKey,
		"ProjectId": client.projectID,
	}

	res, err := sendAPIHttpRequest(req, client.privateKey, defaultTimeout)
	if err != nil {
		return false, err
	}
	var result httpResult
	err = json.Unmarshal(res, &result)
	if err != nil {
		return false, err
	}
	if result.RetCode == 0 {
		return true, nil
	} else {
		return false, errors.New(result.Message)
	}
}

// ListQueue 获取队列列表
func (client *UmqClient) ListTopic(limit int, offset int) ([]QueueInfo, error) {
	var resultList []QueueInfo
	if limit < 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	req := map[string]string{
		"Action":    "GetUMQTopics",
		"Region":    client.region,
		"Limit":     strconv.Itoa(limit),
		"Offset":    strconv.Itoa(offset),
		"ProjectId": client.projectID,
		"PublicKey": client.publicKey,
	}

	res, err := sendAPIHttpRequest(req, client.privateKey, defaultTimeout)
	if err != nil {
		return resultList, err
	}
	var result getTopicResult
	err = json.Unmarshal(res, &result)
	if err != nil {
		return resultList, err
	}

	if result.RetCode == 0 {
		data := result.Queues
		for _, d := range data {
			var queueInfo QueueInfo
			info := d.(map[string]interface{})
			queueInfo.QueueId = info["QueueId"].(string)
			queueInfo.QueueName = info["QueueName"].(string)
			queueInfo.MessageTTL = int(info["MessageTTL"].(float64))
			queueInfo.CreateTime = int64(info["CreateTime"].(float64))
			queueInfo.Description = info["Description"].(string)

			resultList = append(resultList[:], queueInfo)
		}

		return resultList, nil
	} else {
		return resultList, errors.New(result.Message)
	}
}

// CreateRole 创建角色
func (client *UmqClient) CreateRole(roleType string) (bool, error) {
	req := map[string]string{
		"Action":    "CreateUMQRole",
		"Region":    client.region,
		"ProjectId": client.projectID,
		"RoleType":  roleType,
		"PublicKey": client.publicKey,
	}

	res, err := sendAPIHttpRequest(req, client.privateKey, defaultTimeout)
	if err != nil {
		return false, err
	}
	var result httpResult
	err = json.Unmarshal(res, &result)
	if err != nil {
		return false, err
	}

	if result.RetCode == 0 {
		return true, nil
	} else {
		return false, errors.New(result.Message)
	}
}

// DeleteRole 删除角色
func (client *UmqClient) DeleteRole(roleId string, roleType string) (interface{}, error) {
	req := map[string]string{
		"Action":    "DeleteUMQRole",
		"Region":    client.region,
		"ProjectId": client.projectID,
		"RoleType":  roleType,
		"RoleId":    roleId,
		"PublicKey": client.publicKey,
	}

	res, err := sendAPIHttpRequest(req, client.privateKey, defaultTimeout)
	if err != nil {
		return nil, err
	}
	var result httpResult
	err = json.Unmarshal(res, &result)
	if err != nil {
		return nil, err
	}

	if result.RetCode == 0 {
		return result, nil
	} else {
		return nil, errors.New(result.Message)
	}
}

// ListRole 展示角色
func (client *UmqClient) ListRole(limit, offset int, roleType string) ([]Role, error) {
	if limit < 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	req := map[string]string{
		"Action":    "GetUMQRoles",
		"Region":    client.region,
		"ProjectId": client.projectID,
		"RoleType":  roleType,
		"Limit":     strconv.Itoa(limit),
		"Offset":    strconv.Itoa(offset),
		"PublicKey": client.publicKey,
	}

	res, err := sendAPIHttpRequest(req, client.privateKey, defaultTimeout)
	if err != nil {
		return nil, err
	}
	var result getRoleResult
	err = json.Unmarshal(res, &result)
	if err != nil {
		return nil, err
	}

	if result.RetCode == 0 {
		return result.Roles, nil
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
	return &UmqConsumer{
		client:  client,
		token:   consumerID + ":" + consumerToken,
		subInfo: make(map[string]*subscribeInfo),
		mutex:   &sync.Mutex{},
	}
}

// CreateClient 创建client
func NewClient(config UmqConfig) (*UmqClient, error) {
	baseURL, err := url.Parse(config.Host)
	// baseURL.Scheme = "http"
	// baseURL.Host = config.Host
	if err != nil {
		return nil, err
	}
	return &UmqClient{
		email:      config.Account,
		region:     config.Region,
		baseURL:    baseURL,
		wsUrl:      baseURL.Host,
		publicKey:  config.PublicKey,
		privateKey: config.PrivateKey,
		projectID:  config.ProjectID,
		apiUrl:     "http://api.ucloud.cn",
	}, nil
}
