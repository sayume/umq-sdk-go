package umq

import (
	"encoding/json"
	"fmt"
)

//PublishMsg 发布消息
func (publisher *UmqProducer) PublishMsg(queueID, content string) error {
	req := map[string]string{
		"Action":         "PublishMsg",
		"Region":         publisher.client.region,
		"QueueId":        queueID,
		"PublisherId":    publisher.producerID,
		"PublisherToken": publisher.token,
		"Content":        content,
		"OrganizationId": publisher.client.organizationID,
	}

	resp, err := sendHTTPRequest(publisher.client.httpAddr, req, 10)
	if err != nil {
		return err
	}
	var replyBody map[string]interface{}
	err = json.Unmarshal(resp, &replyBody)
	if err != nil {
		return err
	}

	if resultCode, ok := replyBody["RetCode"].(float64); !ok || int(resultCode) != 0 {
		errMsg := ""
		if msg, ok := replyBody["Message"].(string); ok {
			errMsg = msg
		}
		return fmt.Errorf("Fail to publish message: %s", errMsg)
	}
	return nil
}
