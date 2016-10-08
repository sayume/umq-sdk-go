package umq

// MsgHandler 订阅函数使用的回调函数
// 其中 channel c 用来ack这条消息，一个常见的MsgHandler的实现如下
//   func handleMessage(c chan string, msg Message) {
//   	 //处理消息
//		 c <- msg.MsgId //  处理完成之后在这里ack消息
//   }
type MsgHandler func(c chan string, Msg Message)

// MessageInfo the structure of messages
type MessageInfo struct {
	Msgs      []Message `json:"Msgs"`
	IsStacked uint32    `json:"IsStacked"`
}

// Message 消息的结构体
type Message struct {
	MsgId   string `json:"MsgId"`
	MsgBody string `json:"MsgBody"`
}

// Role 角色的结构体
type Role struct {
	Id         string
	Token      string
	CreateTime int64
}

// QueueInfo 队列的信息
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

// UmqClient UMQ的客户端实例
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

// UmqProducer UMQ生产者的实例
type UmqProducer struct {
	client     *UmqClient
	producerID string
	token      string
}
