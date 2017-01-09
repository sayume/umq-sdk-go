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
	MessageID string `json:"messageID"`
	Content   string `json:"content"`
}

// Role 角色的结构体
type Role struct {
	RoleId     string `json:"RoleId"`
	RoleToken      string `json:"RoleToken`
	CreateTime int64 `json:"CreateTime`
	Status string `json:"Status"`
}

// QueueInfo 队列的信息
type QueueInfo struct {
	QueueId     string
	QueueName   string
	MessageTTL  int
	CreateTime  int64
	Description string
}

// UmqProducer UMQ生产者的实例
type UmqProducer struct {
	client     *UmqClient
	producerID string
	token      string
}
