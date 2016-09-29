# UMQ Go SDK 使用指南

## 1.配置http client

配置client

Method: CreateClient

params (HttpClient)

|      Name      |  Type   | Description | Required |
| :------------: | :-----: | :--------- | :------: |
|   OrganizationId     | Integer  |    组织ID     |   Yes    |
|  PublicKey   | String  |    用户公钥    |   Yes    |
| PrivateKey | String  |  用户私钥  |   Yes    |
|    HttpAddr     | String  |   httpAPI地址   |   Yes    |
|    WsAddr     | String  |   ws地址   |   Yes    |
|    WsUrl     | String  |   ws url  |   Yes    |

## 2.操作queue

创建队列

Method: CreateQueue

Params

|   Name    |  Type   |               Description                | Required |
| :-------: | :-----: | :-------------------------------------- | :------: |
|  Region   | String  |                   地区Id                   |   Yes    |
| CouponId  | String  |                  优惠券Id                   |    Yes    |
|  Remark   | String  |                  域名组描述                   |    Yes    |
| QueueName | String  |                业务组信息/队列名                 |   Yes    |
| PushType  | Integer |      发送方式, e.g "Direct" or "Fanout"      |   Yes    |
|    QoS    | String  | 是否需要对消费进行服务质量管控。枚举值为: "Yes",表示消费消息时客户端需要确认消息已收到(Ack模式)；"No",表示消费消息时不需要确认(NoAck模式)。默认为"Yes"。 |    Yes    |

Return

|  Name   |  Type   |      Description       |
| :-----: | :-----: | :-------------------- |
| Err  |  error |       错误消息            |
| QueueID|  String  | 生成的queue id |

删除队列

Method: DeleteQueue

Parameters

|   Name    |  Type   | Description | Required |
| :-------: | :-----: | :--------- | :------: |
|  ProjectId   | String  |    项目Id     |   Yes    |
| QueueId |  String  |    队列Id     |   Yes    |

Return

|  Name   |  Type   |      Description       |
| :-----: | :-----: | :-------------------- |
| err  |  error |       错误消息            |
| queueId |  String  | 删除的queue的Id |

展示队列

Method: ListQueue

Params

|   Name    |  Type   | Description | Required |
| :-------: | :-----: | :--------- | :------: |
|  ProjectId   | String  |    项目Id     |   Yes    |
| Limit |  int  |   数量     |   Yes    |
| Offset |  int  |   偏移量     |   Yes    |

Return

|  Name   |  Type   |      Description       |
| :-----: | :-----: | :-------------------- |
| queueInfo |  Array |       队列信息            |
| err  |  error |       错误消息            |

QueueInfo

|  Name   |  Type   |      Description       |
| :-----: | :-----: | :-------------------- |
| queueId |  String |       队列ID            |
| queueName |  String |       队列名称           |
| PushType |  String |       推送类型           |
| MsgTTL |  Int |       消息失效时间            |
| CreateTime |  Int |       队列建立时间            |
| HttpAddr |  String |       队列http地址           |
| PublisherList |  Array(Role) |       生产者列表           |
| ConsumerList|  Array(Role) |       消费者列表            |

Role

|  Name   |  Type   |      Description       |
| :-----: | :-----: | :-------------------- |
| Id |  String |       角色ID            |
| Token |  String |       角色Token           |
| CreateTime |  Int |       创建时间           |

## 3.操作角色

创建生产者(Pub)，消费者(Sub)

Method: CreateRole

Params

|  Name   |  Type   |      Description       |
| :-----: | :-----: | :-------------------- |
| QueueId |  String |       队列ID            |
| Num |  Int|       数量           |
| Role |  String |       角色           |
| ProjectId |  String |       项目ID           |

Return

|  Name   |  Type   |      Description       |
| :-----: | :-----: | :-------------------- |
| Role |  Role |       角色信息           |
| err  |  error |       错误消息            |

删除角色

Method: DeleteRole

Params

|  Name   |  Type   |      Description       |
| :-----: | :-----: | :-------------------- |
| QueueId |  String |       队列ID            |
| RoleId |  String|       角色ID           |
| Role |  String |       角色           |

Return

|  Name   |  Type   |      Description       |
| :-----: | :-----: | :-------------------- |
| RoleId |  String |       角色ID           |
| err  |  error |       错误消息            |

## 消息操作

推送消息

Method: PublishMsg

Params

|  Name   |  Type   |      Description       |
| :-----: | :-----: | :-------------------- |
| ProjectId |  String |       项目ID            |
| QueueId |  String|       队列ID           |
| PublisherId |  String |       生产者ID          |
| PublisherToken |  String |       生产者Token          |
| Content |  String |       消息内容          |

Return

|  Name   |  Type   |      Description       |
| :-----: | :-----: | :-------------------- |
| ResultCode |  Bool|       操作是否成功           |
| err  |  error |       错误消息            |

拉取消息

Method: GetMsg

Params

|  Name   |  Type   |      Description       |
| :-----: | :-----: | :-------------------- |
| ProjectId |  String |       项目ID            |
| QueueId |  String|       队列ID           |
| ConsumerId |  String |       生产者ID          |
| ConsumerToken |  String |       生产者Token          |
| Num |  Int |       拉取消息条数          |
| Content |  String |       消息内容          |

Return

|  Name   |  Type   |      Description       |
| :-----: | :-----: | :-------------------- |
| Msgs |  Array(Msg)|       消息           |
| err  |  error |       错误消息            |

Msg

|  Name   |  Type   |      Description       |
| :-----: | :-----: | :-------------------- |
| MsgId |  String|       消息ID           |
| MsgBody  |  String |       消息内容            |
| MsgTime |  Int64 |       消息传递时间           |

回执消息

Method: AckMsg

Params

|  Name   |  Type   |      Description       |
| :-----: | :-----: | :-------------------- |
| QueueId |  String|       队列ID           |
| ConsumerId |  String |       生产者ID          |
| MsgId |  String |       消息ID         |

Return

|  Name   |  Type   |      Description       |
| :-----: | :-----: | :-------------------- |
| ResultCode |  Bool|       操作是否成功           |
| err  |  error |       错误消息            |

订阅消息

Method: SubscribeQueue

Params

|  Name   |  Type   |      Description       |
| :-----: | :-----: | :-------------------- |
| OrganizationId |  String |       组织ID            |
| QueueId |  String|       队列ID           |
| ConsumerId |  String |       生产者ID          |
| ConsumerToken |  String |       生产者Token          |
| MsgHandler |  func |       ack函数          |



