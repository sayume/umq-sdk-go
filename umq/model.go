package umq

type HttpResult struct {
	Action  string        `json:"Action"`
	RetCode int           `json:"RetCode"`
	Message string        `json:"Message"`
	DataSet []interface{} `json:"DataSet"`
}

type HttpResult2 struct {
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

type MessageInfo struct {
	Msgs      []Message `json:"Msgs"`
	IsStacked uint32    `json: "IsStacked"`
}

type Message struct {
	MsgId   string `json:"MsgId"`
	MsgBody string `json:"MsgBody"`
}

type Role struct {
	Id         string
	Token      string
	CreateTime int64
}

type QueueInfo struct {
	QueueId       string
	QueueName     string
	PushType      string
	MsgTTL        int
	CreateTime    int64
	HttpAddr      string
	QoS           string
	PublisherList []Role
	ConsumerList  []Role
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
}

type UmqProducer struct {
	client     *UmqClient
	producerID string
	token      string
}

type startConsumeReq struct {
	OrganizationId uint64
	QueueId        string
	ConsumerId     string
	ConsumerToken  string
}
