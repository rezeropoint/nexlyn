package main

import (
	"flag"
	"fmt"

	"github.com/rezeropoint/nexlyn/service/iotquery/internal/config"
	"github.com/rezeropoint/nexlyn/service/iotquery/internal/server"
	"github.com/rezeropoint/nexlyn/service/iotquery/internal/svc"
	"github.com/rezeropoint/nexlyn/service/iotquery/pb"

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
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterIoTQueryServer(grpcServer, server.NewIoTQueryServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
