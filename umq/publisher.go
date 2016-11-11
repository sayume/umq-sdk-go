package umq

import (
	"fmt"
	"io"
)

// UmqProducer UMQ生产者的实例
type UmqProducer struct {
	client *UmqClient
	token  string
}

type PublishResponse struct {
}

//PublishMsg 发布消息
func (publisher *UmqProducer) PublishMsg(queueID string, message io.Reader) error {
	url := *(publisher.client.baseURL)
	url.Path = fmt.Sprintf("/%s/%s/message", publisher.client.projectID, queueID)
	resp := &PublishResponse{}
	err := sendHTTPRequest(url.String(), "POST", message, publisher.token, resp)
	return err
}
