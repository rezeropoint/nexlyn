/**
 * @name 代理的配置
 * @see 在生产环境 代理是无法生效的，所以这里没有生产环境的配置
 * -------------------------------
 * The agent cannot take effect in the production environment
 * so there is no configuration of the production environment
 * For details, please see
 * https://pro.ant.design/docs/deploy
 *
 * @doc https://umijs.org/docs/guides/proxy
 */
export default {
  // 如果需要自定义本地开发服务器  请取消注释按需调整
  dev: {
    // 优先将流媒体相关接口转发到 mediahandler (8080端口)
    // 按照 nginx.conf 中的优先级顺序配置

    // SSE 汇总接口 - 优先级最高
    '/api/summary/': {
      target: 'http://localhost:8080',
      changeOrigin: true,
      pathRewrite: { '^': '' },
      timeout: 0, // SSE需要无限超时
      // 禁用webpack-dev-server的压缩和缓冲
      compress: false,
      // SSE流式传输配置
      onProxyReq: (proxyReq: any, req: any, _res: any) => {
        // 设置SSE请求头
        if (req.url.includes('/sse')) {
          proxyReq.setHeader('Accept', 'text/event-stream');
          proxyReq.setHeader('Cache-Control', 'no-cache');
          proxyReq.setHeader('Connection', 'keep-alive');
        }
      },
      onProxyRes: (proxyRes: any, req: any, _res: any) => {
        // SSE 必须设置这些响应头
        if (req.url.includes('/sse')) {
          // 删除可能导致缓冲的头
          delete proxyRes.headers['content-encoding'];
          delete proxyRes.headers['content-length'];

          // 设置正确的响应头
          proxyRes.headers['content-type'] = 'text/event-stream';
          proxyRes.headers['cache-control'] = 'no-cache, no-transform';
          proxyRes.headers.connection = 'keep-alive';
          proxyRes.headers['x-accel-buffering'] = 'no';
        }
      },
    },

    // 流媒体接口
    '/api/stream/': {
      target: 'http://localhost:8080',
      changeOrigin: true,
      pathRewrite: { '^': '' },
      timeout: 30000, // 30秒超时
    },

    // 订阅者接口
    '/api/subscribers/': {
      target: 'http://localhost:8080',
      changeOrigin: true,
      pathRewrite: { '^': '' },
      timeout: 30000, // 30秒超时
    },

    // GB28181 接口代理到 mediahandler
    '/gb28181/': {
      target: 'http://localhost:8080',
      changeOrigin: true,
      pathRewrite: { '^': '' },
      timeout: 30000,
    },

    // Nexlyn 接口代理到 mediahandler
    '/nexlyn/': {
      target: 'http://localhost:8080',
      changeOrigin: true,
      pathRewrite: { '^': '' },
      timeout: 30000,
    },

    // WebSocket FLV 播放
    '/flv/': {
      target: 'http://localhost:8080',
      changeOrigin: true,
      ws: true, // WebSocket 支持
      pathRewrite: { '^': '' },
      timeout: 60000,
    },

    // IoT 服务接口转发到 iotmanager 服务
    // 本地开发: 在docker-compose.yml中取消注释iotmanager服务的端口映射 "8889:8888"
    // 或者本地运行: 修改 restful/iotmanager/etc/config.yaml 的 Port 为 8889，然后运行服务

    // 设备类别接口
    '/api/v1/device-categories': {
      target: 'http://localhost:8889',
      changeOrigin: true,
      pathRewrite: { '^': '' },
      timeout: 10000,
    },

    // 设备模板接口
    '/api/v1/sensor-template': {
      target: 'http://localhost:8889',
      changeOrigin: true,
      pathRewrite: { '^': '' },
      timeout: 10000,
    },

    // 设备管理接口
    '/api/v1/device-bindings': {
      target: 'http://localhost:8889',
      changeOrigin: true,
      pathRewrite: { '^': '' },
      timeout: 10000,
    },

    // 设备标签接口
    '/api/v1/device-tags': {
      target: 'http://localhost:8889',
      changeOrigin: true,
      pathRewrite: { '^': '' },
      timeout: 10000,
    },

    // 平台管理接口
    '/api/v1/platforms': {
      target: 'http://localhost:8889',
      changeOrigin: true,
      pathRewrite: { '^': '' },
      timeout: 10000,
    },

    // 时序数据查询接口
    '/api/v1/timeseries': {
      target: 'http://localhost:8889',
      changeOrigin: true,
      pathRewrite: { '^': '' },
      timeout: 30000, // 30秒超时，查询可能较慢
    },

    // AI Box设备控制接口
    '/api/v1/device-control': {
      target: 'http://localhost:8889',
      changeOrigin: true,
      pathRewrite: { '^': '' },
      timeout: 15000, // 15秒超时，MQTT控制可能需要时间
    },

    // HTTP 接收配置管理接口
    '/api/v1/http-receive': {
      target: 'http://localhost:8889',
      changeOrigin: true,
      pathRewrite: { '^': '' },
      timeout: 10000,
    },

    // EventHandler 服务接口转发到 eventhandler 服务
    // 本地开发: eventhandler 服务的端口映射为 "8890:8888"
    '/api/v1/skylark-platform': {
      target: 'http://localhost:8890',
      changeOrigin: true,
      pathRewrite: { '^': '' },
      timeout: 10000,
    },
    '/api/v1/event-configs': {
      target: 'http://localhost:8890',
      changeOrigin: true,
      pathRewrite: { '^': '' },
      timeout: 10000,
    },
    '/api/v1/org-mappings': {
      target: 'http://localhost:8890',
      changeOrigin: true,
      pathRewrite: { '^': '' },
      timeout: 10000,
    },
    '/api/v1/events': {
      target: 'http://localhost:8890',
      changeOrigin: true,
      pathRewrite: { '^': '' },
      timeout: 30000, // 查询可能较慢
    },
    // 用户任务接口（my-assignments, my-proposed-journeys）
    '/api/v1/my-': {
      target: 'http://localhost:8890',
      changeOrigin: true,
      pathRewrite: { '^': '' },
      timeout: 10000,
    },
    // 流程操作接口（flows）
    '/api/v1/flows': {
      target: 'http://localhost:8890',
      changeOrigin: true,
      pathRewrite: { '^': '' },
      timeout: 10000,
    },

    // LynxGraph Manager 服务接口转发到 lynxmanager 服务
    // 本地开发: lynxmanager 服务的端口映射为 "8891:8888"
    // 包括: 信息原子、逻辑图配置等接口
    '/api/lynxmanager/': {
      target: 'http://localhost:8891',
      changeOrigin: true,
      pathRewrite: { '^': '' },
      timeout: 10000,
    },

    // 其他 API 接口转发到 backend (8888端口)
    // 包括: 认证、用户、权限、租户、组织等接口
    '/api/': {
      target: 'http://localhost:8888',
      changeOrigin: true,
      pathRewrite: { '^/api': '/api' },
      timeout: 10000, // 10秒超时，对应 nginx 配置
      headers: {
        Connection: 'close', // 禁用 Keep-Alive，避免频繁请求时 ECONNRESET
      },
    },
  },
  /**
   * @name 详细的代理配置
   * @doc https://github.com/chimurai/http-proxy-middleware
   */
  test: {
    // localhost:8000/api/** -> https://preview.pro.ant.design/api/**
    '/api/': {
      target: 'https://proapi.azurewebsites.net',
      changeOrigin: true,
      pathRewrite: { '^': '' },
    },
  },
  pre: {
    '/api/': {
      target: 'your pre url',
      changeOrigin: true,
      pathRewrite: { '^': '' },
    },
  },
};
