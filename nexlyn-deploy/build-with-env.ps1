# 环境变量构建脚本 - 使用 SSH 密钥环境变量
# 使用方法: .\build-with-env.ps1 [service_name]

param(
    [string]$Service = "all"
)

# Write-Host "=== Nexlyn Docker 构建脚本 (环境变量版) ===" -ForegroundColor Green

# 检查 Docker 是否运行
# try {
#     docker version | Out-Null
#     Write-Host "✓ Docker 正在运行" -ForegroundColor Green
# } catch {
#     Write-Host "✗ Docker 未运行，请先启动 Docker Desktop" -ForegroundColor Red
#     exit 1
# }

# 检查 SSH 密钥是否存在
$sshKeyPath = "$env:USERPROFILE\.ssh\id_ed25519"
if (-not (Test-Path $sshKeyPath)) {
    Write-Host "✗ 未找到 SSH 密钥: $sshKeyPath" -ForegroundColor Red
    Write-Host "请确保你有 GitHub SSH 密钥" -ForegroundColor Yellow
    exit 1
}

# Write-Host "✓ 找到 SSH 密钥: $sshKeyPath" -ForegroundColor Green

# 读取 SSH 私钥内容
$sshPrivateKey = Get-Content $sshKeyPath -Raw
if (-not $sshPrivateKey) {
    Write-Host "✗ 无法读取 SSH 私钥" -ForegroundColor Red
    exit 1
}

# 设置环境变量
$env:SSH_PRIVATE_KEY = $sshPrivateKey
# Write-Host "✓ 设置 SSH_PRIVATE_KEY 环境变量" -ForegroundColor Green

# 测试 GitHub SSH 连接
# Write-Host "测试 GitHub SSH 连接..." -ForegroundColor Cyan
# $sshTest = ssh -T git@github.com 2>&1
# if ($LASTEXITCODE -eq 1 -and $sshTest -match "successfully authenticated") {
#     Write-Host "✓ GitHub SSH 连接正常" -ForegroundColor Green
# } else {
#     Write-Host "✗ GitHub SSH 连接失败" -ForegroundColor Red
#     Write-Host "请检查你的 SSH 密钥是否正确添加到 GitHub" -ForegroundColor Yellow
#     exit 1
# }

# 启用 BuildKit
$env:DOCKER_BUILDKIT = 1
# Write-Host "✓ 启用 Docker BuildKit" -ForegroundColor Green

# 根据参数选择构建的服务
switch ($Service.ToLower()) {
    "lynxmanager" {
        # Write-Host "构建 lynxmanager 服务..." -ForegroundColor Cyan
        docker-compose build lynxmanager
    }
    "lynxengine" {
        # Write-Host "构建 lynxengine 服务..." -ForegroundColor Cyan
        docker-compose build lynxengine
    }
    "iotquery" {
        # Write-Host "构建 iotquery 服务..." -ForegroundColor Cyan
        docker-compose build iotquery
    }
    "iotmanager" {
        # Write-Host "构建 iotmanager 服务..." -ForegroundColor Cyan
        docker-compose build iotmanager
    }
    "eventhandler" {
        # Write-Host "构建 eventhandler 服务..." -ForegroundColor Cyan
        docker-compose build eventhandler
    }
    "eventsync" {
        # Write-Host "构建 eventsync 服务..." -ForegroundColor Cyan
        docker-compose build eventsync
    }
    "mediahandler" {
        # Write-Host "构建 mediahandler 服务..." -ForegroundColor Cyan
        docker-compose build mediahandler
    }
    "backend" {
        # Write-Host "构建 backend 服务..." -ForegroundColor Cyan
        docker-compose build backend
    }
    "frontend" {
        # Write-Host "构建 frontend 服务..." -ForegroundColor Cyan
        docker-compose build frontend
    }
    "all" {
        # Write-Host "构建所有服务..." -ForegroundColor Cyan
        docker-compose build
    }
    default {
        Write-Host "未知服务: $Service" -ForegroundColor Red
        Write-Host "可用选项: backend, iotmanager, eventhandler, eventsync, mediahandler, lynxmanager, lynxengine, iotquery, frontend, all" -ForegroundColor Yellow
        exit 1
    }
}

# 清理环境变量
Remove-Item Env:SSH_PRIVATE_KEY -ErrorAction SilentlyContinue

# if ($LASTEXITCODE -eq 0) {
#     Write-Host "✓ 构建完成！" -ForegroundColor Green
#     Write-Host "运行 'docker-compose up -d' 启动服务" -ForegroundColor Cyan
# } else {
#     Write-Host "✗ 构建失败！" -ForegroundColor Red
#     exit 1
# }
