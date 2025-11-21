package service

import (
	"fmt"

	"github.com/rezeropoint/nexlyn/service/iotquery/iotquery"

	"github.com/zeromicro/go-zero/zrpc"
)

// NewSensorDataService 创建传感器数据服务实例
// 返回 gRPC 生成的 IoTQuery 接口（不做任何封装）
//
// 设计理念：
//   - LynxGraph Core 不定义业务接口
//   - 积木直接使用 gRPC 接口（类似 Skylark 积木直接使用 go-skylark 接口）
//   - 本函数只负责创建 gRPC 客户端，不做业务逻辑封装
func NewSensorDataService(rpcClient zrpc.Client) (iotquery.IoTQuery, error) {
	if rpcClient == nil {
		return nil, fmt.Errorf("gRPC客户端不能为空")
	}

	// 直接返回 gRPC 生成的客户端
	return iotquery.NewIoTQuery(rpcClient), nil
}
