package umq

const (
	// RegionCnBj2 地域: 北京2
	RegionCnBj2 = "cn-bj2"
	// RegionCnGd 地域: 广东
	RegionCnGd = "cn-gd"
)

// UmqConfig 描述了Umq的配置信息
type UmqConfig struct {
	// 主机地址 例如 air.bj2.umq.service.ucloud.cn
	Host string
	// 地域, 例如 RegionCnBj2
	Region string
	// 账户信息
	Account string
	// 项目ID
	ProjectID string
	// 账户的公钥
	PublicKey string
	// 账户的私钥
	PrivateKey string
}
