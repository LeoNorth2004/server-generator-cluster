# ============================================
# Generator Platform - 本地开发一键启动脚本 (PowerShell版)
# ============================================
# 用法: .\start-local.ps1
# 前置要求: Go 1.21+, Node.js 18+, PostgreSQL, Redis

$ErrorActionPreference = "Stop"

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  微服务代码生成平台 - 本地开发环境启动" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

# 检查前置依赖
Write-Host "[检查] 验证前置依赖..." -ForegroundColor Yellow

$goVersion = go version 2>$null
if (-not $goVersion) {
    Write-Host "[错误] Go 未安装! 请安装 Go 1.21+" -ForegroundColor Red
    exit 1
}
Write-Host "[OK] Go: $goVersion" -ForegroundColor Green

$nodeVersion = node --version 2>$null
if (-not $nodeVersion) {
    Write-Host "[错误] Node.js 未安装! 请安装 Node.js 18+" -ForegroundColor Red
    exit 1
}
Write-Host "[OK] Node.js: $nodeVersion" -ForegroundColor Green

Write-Host ""

# 启动顺序: 基础设施 -> 核心服务 -> 网关 -> 前端

Write-Host "[1/8] 启动 authentication-service (认证服务) :8082..." -ForegroundColor Yellow
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd apps/authentication-service; go run main.go"
Start-Sleep -Seconds 3

Write-Host "[2/8] 启动 user-service (用户服务) :8081..." -ForegroundColor Yellow
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd apps/user-service; go run main.go"
Start-Sleep -Seconds 2

Write-Host "[3/8] 启动 project-service (项目管理服务) :8083..." -ForegroundColor Yellow
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd apps/project-service; go run main.go"
Start-Sleep -Seconds 2

Write-Host "[4/8] 启动 generator-service (代码生成服务) :8084..." -ForegroundColor Yellow
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd apps/generator-service; go run main.go"
Start-Sleep -Seconds 2

Write-Host "[5/8] 启动 operations-service (运维服务) :8085..." -ForegroundColor Yellow
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd apps/operations-service; go run main.go"
Start-Sleep -Seconds 2

Write-Host "[6/8] 启动 cluster-service (集群管理服务) :8086..." -ForegroundColor Yellow
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd apps/cluster-service; go run main.go"
Start-Sleep -Seconds 2

Write-Host "[7/8] 启动 api-gateway (API网关) :8080..." -ForegroundColor Yellow
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd apps/api-gateway; go run main.go"
Start-Sleep -Seconds 2

Write-Host "[8/8] 启动 web-admin (前端界面) :3000..." -ForegroundColor Yellow
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd apps/web-admin; npm run dev"

Write-Host ""
Write-Host "============================================" -ForegroundColor Green
Write-Host "  所有服务启动完成!" -ForegroundColor Green
Write-Host "============================================" -ForegroundColor Green
Write-Host ""
Write-Host "访问地址:" -ForegroundColor White
Write-Host "  前端界面: http://localhost:3000" -ForegroundColor Cyan
Write-Host "  API网关:   http://localhost:8080" -ForegroundColor Cyan
Write-Host "  健康检查: http://localhost:8080/health" -ForegroundColor Cyan
Write-Host ""
Write-Host "默认登录凭据:" -ForegroundColor White
Write-Host "  用户名: admin" -ForegroundColor Yellow
Write-Host "  密码:   admin123" -ForegroundColor Yellow
Write-Host ""
Write-Host "服务端口列表:" -ForegroundColor White
Write-Host "  authentication-service : 8082" -ForegroundColor Gray
Write-Host "  user-service           : 8081" -ForegroundColor Gray
Write-Host "  project-service        : 8083" -ForegroundColor Gray
Write-Host "  generator-service      : 8084" -ForegroundColor Gray
Write-Host "  operations-service     : 8085" -ForegroundColor Gray
Write-Host "  cluster-service        : 8086" -ForegroundColor Gray
Write-Host "  api-gateway            : 8080" -ForegroundColor Gray
Write-Host "  web-admin              : 3000" -ForegroundColor Gray
Write-Host ""

Read-Host "按 Enter 键退出..."
