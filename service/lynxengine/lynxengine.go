package main

import (
	"flag"
	"fmt"

	"github.com/rezeropoint/nexlyn/service/lynxengine/internal/config"
	"github.com/rezeropoint/nexlyn/service/lynxengine/internal/server"
	"github.com/rezeropoint/nexlyn/service/lynxengine/internal/svc"
	"github.com/rezeropoint/nexlyn/service/lynxengine/pb"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/config.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 创建服务上下文
	ctx := svc.NewServiceContext(c)

	// 创建gRPC服务器
	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		// 注册服务
		pb.RegisterLynxEngineServer(grpcServer, server.NewLynxEngineServer(ctx))

		// 启用反射（用于grpcurl等工具调试）
		if c.Mode == service.DevMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting LynxGraph Engine gRPC server at %s...\n", c.ListenOn)

	// 启动服务
	group := service.NewServiceGroup()
	group.Add(s)
	group.Start()
}
