package umq

import (
	"encoding/json"
	"fmt"
	"strconv"
)

//获取项目ID
func getOrganizationId(url, email, projectId, publicKey, privateKey string) (string, error) {
	req := map[string]string{
		"Action":            "GetOrganizationId",
		"UserEmail":         email,
		"OrganizationAlias": projectId,
		"PublicKey":         publicKey,
	}

	res, err := sendUMQAPIHttpRequest(url, req, privateKey, 10)
	if err != nil {
		return "", err
	}
	var result map[string]interface{}
	err = json.Unmarshal(res, &result)
	if err != nil {
		return "", fmt.Errorf("Fail while identify the project id %s", err.Error())
	}
	if retCode, ok := result["RetCode"].(float64); !ok || int(retCode) != 0 {
		ss := ""
		if msg, ok := result["Message"].(string); ok {
			ss = msg
		}
		return "", fmt.Errorf("Fail to identify the project id: %s", ss)
	}
	if data, ok := result["Data"].(float64); !ok {
		return "", fmt.Errorf("Fail to identify the project id: Server Error")
	} else {
		return strconv.Itoa(int(data)), nil
	}
}
