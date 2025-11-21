package dispatcher

type Config struct {
	ServiceName       string // 服务名称
	PodName           string // Pod 名称
	LynxGraphGRPCURL  string // LynxGraph gRPC 服务地址（例如: lynxengine:9999）
}
