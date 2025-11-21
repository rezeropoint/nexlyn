package svc

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/rezeropoint/nexlyn/restful/mediahandler/internal/config"

	_ "github.com/lib/pq" // postgres 支持
	"m7s.live/v5"

	_ "github.com/rezeropoint/nexlyn/pkg/nexlyn"

	_ "m7s.live/v5/plugin/cascade"

	_ "m7s.live/v5/plugin/debug"
	_ "m7s.live/v5/plugin/flv"
	_ "m7s.live/v5/plugin/gb28181"
	_ "m7s.live/v5/plugin/hls"
	_ "m7s.live/v5/plugin/logrotate"
	_ "m7s.live/v5/plugin/mp4"
	_ "m7s.live/v5/plugin/onvif"
	_ "m7s.live/v5/plugin/preview"
	_ "m7s.live/v5/plugin/rtmp"
	_ "m7s.live/v5/plugin/rtp"
	_ "m7s.live/v5/plugin/rtsp"
	_ "m7s.live/v5/plugin/sei"
	_ "m7s.live/v5/plugin/snap"
	_ "m7s.live/v5/plugin/srt"
	_ "m7s.live/v5/plugin/test"
	_ "m7s.live/v5/plugin/transcode"
	_ "m7s.live/v5/plugin/webrtc"
	_ "m7s.live/v5/plugin/webtransport"
)

type ServiceContext struct {
	PodName   string
	Config    config.Config
	M7sServer *m7s.Server
}

func NewServiceContext(c config.Config) *ServiceContext {

	podName := os.Getenv("POD_NAME")
	if podName == "" {
		podName = "none"
	}

	conf := flag.String("c", "config.yaml", "config file")
	flag.Parse()

	// 先构造 ServiceContext，返回地址；m7s 实例异步就绪后再注入字段
	sc := &ServiceContext{
		Config:  c,
		PodName: podName,
	}

	// 启动 m7s 引擎
	go m7s.Run(context.Background(), *conf)

	// 后台等待 m7s Server 实例创建后注入到 ServiceContext
	go func() {
		for {
			if srv, ok := m7s.Servers.Find(func(s *m7s.Server) bool { return true }); ok {
				sc.M7sServer = srv
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	return sc
}
