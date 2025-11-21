package graph

import (
	"github.com/rezeropoint/etcdtrigger"

	"github.com/rezeropoint/nexlyn/pkg/lynxgraph/core"
)

// Config 图注册表配置
type Config struct {
	RunMode    core.Mode
	EtcdConfig etcdtrigger.Config // Etcd配置（包内建立连接）
}

// 常量定义
const (
	Collection = "graph"
	// etcd 键格式
	EtcdKeyFormat = "%s/%s/%s" // prefix/action/id
)
