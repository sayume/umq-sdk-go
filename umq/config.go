package umq

const (
	// RegionCnBj2 地域: 北京2
	RegionCnBj2 = "cn-bj2"
	// RegionCnGd 地域: 广东
	RegionCnGd = "cn-gd"
)

// UmqConfig 描述了Umq的配置信息
type UmqConfig struct {
	Host       string // URL，例如 air.bj2.umq.service.ucloud.cn
	Region     string // 地域, 例如 RegionCnBj2
	Account    string // 账户信息
	ProjectID  string // 项目ID
	PublicKey  string // 账户的公钥
	PrivateKey string // 账户的私钥
}
