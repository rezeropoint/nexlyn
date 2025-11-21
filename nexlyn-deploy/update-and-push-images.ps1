# 确保在执行前登录到Docker Hub或私有仓库
# docker login

param(
    [string]$Service = "all",
    [string]$Registry = "registry.cn-hangzhou.aliyuncs.com/rezeropoint/"
)

# 脚本路径
$ScriptPath = Split-Path -Parent $MyInvocation.MyCommand.Definition
$ProjectRoot = Join-Path $ScriptPath ".."

# 获取新的镜像标签
$NewTag = Read-Host "请输入新的镜像标签（例如：v1.0.0）："
if ([string]::IsNullOrWhiteSpace($NewTag)) {
    Write-Host "错误：镜像标签不能为空。" -ForegroundColor Red
    exit 1
}

Write-Host "将使用标签：$NewTag 来推送镜像。" -ForegroundColor Green

# 规范化并校验 Registry 前缀
$DefaultRegistryPrefix = $Registry
if ([string]::IsNullOrWhiteSpace($DefaultRegistryPrefix)) {
    Write-Host "错误：必须提供 Registry 前缀（例如：registry.cn-hangzhou.aliyuncs.com/rezeropoint/）。" -ForegroundColor Red
    exit 1
}
if (-not $DefaultRegistryPrefix.EndsWith("/")) { $DefaultRegistryPrefix = "$DefaultRegistryPrefix/" }

# 所有服务列表
$AllServices = @("backend", "mediahandler", "iotmanager", "eventhandler", "eventsync", "frontend","lynxmanager","lynxengine", "iotquery")

$ServicesToProcess = @()

if ($Service.ToLower() -eq "all") {
    $ServicesToProcess = $AllServices
} elseif ($AllServices -contains $Service.ToLower()) {
    $ServicesToProcess = @($Service.ToLower())
} else {
    Write-Host "错误：未知服务 '$Service'。可用选项：all, $($AllServices -join ', ')" -ForegroundColor Red
    exit 1
}

# 并发推送
$jobs = @()
foreach ($CurrentService in $ServicesToProcess) {
    $jobs += Start-Job -Name "push-$CurrentService" -ArgumentList $CurrentService, $DefaultRegistryPrefix, $NewTag -ScriptBlock {
        param($ServiceName, $RegistryPrefix, $Tag)
        $LocalImageName = "nexlyn-${ServiceName}:latest"
        if (-not $RegistryPrefix.EndsWith("/")) { $RegistryPrefix = "$RegistryPrefix/" }

        # 推送指定版本标签
        $TargetImageName = "nexlyn-${ServiceName}:${Tag}"
        $FullImageName = "$RegistryPrefix$TargetImageName"

        & docker tag $LocalImageName $FullImageName | Out-Null
        if ($LASTEXITCODE -ne 0) {
            [PSCustomObject]@{ Service=$ServiceName; Image=$FullImageName; Status="tag_failed" }
            return
        }

        & docker push $FullImageName | Out-Null
        if ($LASTEXITCODE -ne 0) {
            [PSCustomObject]@{ Service=$ServiceName; Image=$FullImageName; Status="push_failed" }
            return
        }

        # 同时推送 latest 标签
        $LatestImageName = "nexlyn-${ServiceName}:latest"
        $FullLatestImageName = "$RegistryPrefix$LatestImageName"

        & docker tag $LocalImageName $FullLatestImageName | Out-Null
        if ($LASTEXITCODE -ne 0) {
            [PSCustomObject]@{ Service=$ServiceName; Image=$FullImageName; Status="ok" }
            [PSCustomObject]@{ Service=$ServiceName; Image=$FullLatestImageName; Status="tag_failed (latest)" }
            return
        }

        & docker push $FullLatestImageName | Out-Null
        if ($LASTEXITCODE -ne 0) {
            [PSCustomObject]@{ Service=$ServiceName; Image=$FullImageName; Status="ok" }
            [PSCustomObject]@{ Service=$ServiceName; Image=$FullLatestImageName; Status="push_failed (latest)" }
            return
        }

        # 返回两个成功结果
        [PSCustomObject]@{ Service=$ServiceName; Image=$FullImageName; Status="ok" }
        [PSCustomObject]@{ Service=$ServiceName; Image=$FullLatestImageName; Status="ok" }
    }
}

Wait-Job -Job $jobs | Out-Null
$results = Receive-Job -Job $jobs
Remove-Job -Job $jobs | Out-Null

# 仅保留我们返回的结构化结果，过滤掉字符串等杂项输出
$results = $results | Where-Object { $_ -is [pscustomobject] -and $_.PSObject.Properties['Status'] }

foreach ($r in $results) {
    if ($r.Status -eq "ok") {
        Write-Host ("[{0}] 推送成功：{1}" -f $r.Service, $r.Image) -ForegroundColor Green
    } else {
        Write-Host ("[{0}] {1}：{2}" -f $r.Service, $r.Status, $r.Image) -ForegroundColor Red
    }
}

Write-Host "所有指定服务的镜像推送已完成。" -ForegroundColor Green
