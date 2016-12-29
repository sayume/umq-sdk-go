package umq

import (
	"fmt"
	"io"
)

type PublishResponse struct {
	MessageID string
}

//PublishMsg 发布消息
func (publisher *UmqProducer) PublishMsg(queueID string, message io.Reader) (msgID string, err error) {
	url := *(publisher.client.baseURL)
	url.Path = fmt.Sprintf("/%s/%s/message", publisher.client.projectID, queueID)
	return sendHTTPRequestWithStringOutput(url.String(), "POST", message, publisher.token, 0)
}
