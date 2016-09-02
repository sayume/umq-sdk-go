package main

import (
	"fmt"
	httpclient "umq-sdk/go/http_v2"
)

func DoMsg(c chan string, Msg interface{}) {
	fmt.Println(Msg)
	c <- ""
	return
}

func main() {
	httpclient.OrganizationId = 25543
	httpclient.QueueId = "umq-wmwhlk"
	httpclient.HttpAddr = "http://192.168.153.41:6318"
	httpclient.PublisherToken = "e349dcbb67521fe082cf88d0892ef3f6"
	httpclient.ConsumerToken = "fafd13a67063ddd00d12a4c85664616d"
	httpclient.WsAddr = "http://192.168.153.41:6138/"
	httpclient.WsUrl = "ws://192.168.153.41:6318/ws"

	//发送消息
	//res, err := httpclient.PublishMsg("hello from xiaoding")
	//主动拉取消息
	//res, err := httpclient.GetMsg("3")
	//回执消息
	//res, err := httpclient.AckMsg("umq-wmwhlk-9")
	//err := httpclient.SubscribeQueue(DoMsg)
	//获取项目Id
	//res, err := httpclient.GetOrganizationId("uhosttest@ucloud.cn", "org-23313")
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//fmt.Println(res)
	return
}
