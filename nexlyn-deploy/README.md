# Nexlyn — 快速启动

简洁说明，适合本地开发与构建私有依赖（通过 SSH 私钥）。

前提
- 已安装 Docker Desktop
- 已在 GitHub 添加你的 SSH 公钥，并在本机存在私钥 `~/.ssh/id_ed25519`
- 在项目根目录运行命令（此文件位于 `nexlyn-deploy/`）

使用推荐脚本（最简单）
```powershell
# 构建并生成镜像（全部服务）
.\build-with-env.ps1

# 或只构建单个服务，例如 mediahandler：
.\build-with-env.ps1 mediahandler
```

手动（不使用脚本）
```powershell
# 启用 BuildKit
$env:DOCKER_BUILDKIT = 1

# 将本地 SSH 私钥内容临时传给构建（安全：会在构建后删除）
$env:SSH_PRIVATE_KEY = Get-Content ~/.ssh/id_ed25519 -Raw

# 构建并启动 mediahandler（或其它服务）
docker-compose up -d --build mediahandler

# 可选：删除临时环境变量
Remove-Item Env:SSH_PRIVATE_KEY -ErrorAction SilentlyContinue
```

常用命令
- 查看服务状态：`docker-compose ps`
- 查看日志：`docker-compose logs -f`
- 停止服务：`docker-compose down`

备注
- 本仓库保留 `build-with-env.ps1` 作为推荐方式，便于一次性完成密钥注入与构建。
- 如果需要更详细步骤或故障排查，可在仓库中查看脚本内容或向维护者咨询。
