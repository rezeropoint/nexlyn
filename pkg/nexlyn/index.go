package plugin_nexlyn

import (
	"fmt"
	"time"

	"github.com/rezeropoint/casbinx/core"
	"github.com/rezeropoint/casbinx/engine"

	"m7s.live/v5"
	plugin_gb28181pro "m7s.live/v5/plugin/gb28181"
)

var _ = m7s.InstallPlugin[NexlynPlugin](m7s.PluginMeta{})

type NexlynPlugin struct {
	m7s.Plugin
	Auth           Auth                             `yaml:"Auth"`          // JWT认证配置
	WatcherConfig  core.WatcherConfig               `yaml:"WatcherConfig"` // CasbinX Watcher配置
	CasbinxDsn     string                           `yaml:"CasbinxDsn"`    // CasbinX数据库连接字符串
	Casbinx        engine.CasbinX                   // CasbinX实例
	gb28181Plugin  *plugin_gb28181pro.GB28181Plugin // GB28181插件实例
	paramExtractor *PathParamExtractor              // 路径参数提取器
}

// OnInit 可选的初始化
func (n *NexlynPlugin) Start() error {

	// 初始化CasbinX权限管理（必须启用）
	if n.CasbinxDsn == "" {
		n.Error("CasbinxDsn配置不能为空，权限管理系统必须启用")
		return fmt.Errorf("CasbinxDsn配置不能为空")
	}

	config := core.Config{
		Dsn: n.CasbinxDsn,
		PossiblePaths: []string{
			"etc/casbin_model.conf",
			"pkg/nexlyn/etc/casbin_model.conf",
		},
		Security: core.DefaultSecurityConfig(),
		Watcher:  n.WatcherConfig,
	}
	casbinx, err := engine.NewCasbinx(config)
	if err != nil {
		n.Error("CasbinX初始化失败", "error", err)
		return fmt.Errorf("CasbinX初始化失败: %v", err)
	}
	n.Casbinx = casbinx
	n.Info("✓ CasbinX权限管理系统初始化成功")

	// 阻塞等待 GB28181 插件，但有超时机制
	if !n.waitForGB28181Plugin() {
		n.Warn("GB28181 插件加载失败，部分功能将不可用，但路由仍会注册以供测试")
		// 暂时允许继续运行以便测试路由框架
	} else {
		n.Info("✓ GB28181 插件连接成功")
	}

	n.Info("✓ NexlynPlugin 初始化完成", "Auth", n.Auth)
	return nil
}

// waitForGB28181Plugin 等待 GB28181 插件加载完成，有超时机制
func (n *NexlynPlugin) waitForGB28181Plugin() bool {
	timeout := 30 * time.Second               // 30秒超时，等待插件加载完成
	ticker := time.NewTicker(2 * time.Second) // 每2秒检查一次
	defer ticker.Stop()

	timeoutChan := time.After(timeout)

	// 首先立即检查一次
	n.gb28181Plugin = n.findGB28181Plugin()
	if n.gb28181Plugin != nil {
		n.Info("✓ GB28181 插件实例初始化成功")
		return true
	}

	n.Info("正在等待 GB28181 插件加载...")

	for {
		select {
		case <-timeoutChan:
			n.Error("等待 GB28181 插件超时")
			return false
		case <-ticker.C:
			n.gb28181Plugin = n.findGB28181Plugin()
			if n.gb28181Plugin != nil {
				n.Info("✓ GB28181 插件实例初始化成功")
				return true
			}
			n.Debug("正在等待 GB28181 插件加载...")
		}
	}
}

// findGB28181Plugin 查找并返回 GB28181 插件实例
func (n *NexlynPlugin) findGB28181Plugin() *plugin_gb28181pro.GB28181Plugin {
	// 遍历所有插件查找 GB28181
	for plugin := range n.Server.Plugins.Range {
		n.Debug("正在等待 GB28181 插件加载...", "plugin", plugin.Meta.Name)
		if plugin.Meta.Name == "GB28181" {
			if gb, ok := plugin.GetHandler().(*plugin_gb28181pro.GB28181Plugin); ok {
				return gb
			}
		}
	}
	return nil
}
