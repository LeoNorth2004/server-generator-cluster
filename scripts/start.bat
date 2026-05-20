@echo off
chcp 65001 >nul 2>&1
setlocal enabledelayedexpansion

echo ============================================
echo   Generator Platform - 统一启动工具
echo ============================================
echo.
echo   请选择启动方式:
echo.
echo   [1] 本地开发环境 (Go + Node.js)
echo   [2] Docker 容器化部署
echo   [3] Kubernetes 集群部署
echo   [4] 查看帮助信息
echo   [0] 退出
echo.

set /p choice="请输入选项 (0-4): "

if "%choice%"=="1" goto local
if "%choice%"=="2" goto docker
if "%choice%"=="3" goto k8s
if "%choice%"=="4" goto help
if "%choice%"=="0" goto end

echo 无效选项，请重新运行脚本
goto end

:local
echo.
echo ============================================
echo   启动本地开发环境...
echo ============================================
echo.
call start-local.bat
goto end

:docker
echo.
echo ============================================
echo   Docker 容器化部署...
echo ============================================
powershell -ExecutionPolicy Bypass -File "%~dp0start-docker.ps1"
goto end

:k8s
echo.
echo ============================================
echo   Kubernetes 集群部署...
echo ============================================
powershell -ExecutionPolicy Bypass -File "%~dp0start-k8s.ps1"
goto end

:help
echo.
echo ============================================
echo   帮助信息
echo ============================================
echo.
echo   本地开发:
echo     需要: Go 1.21+, Node.js 18+, PostgreSQL, Redis
echo     启动: 运行选项1 或直接执行 start-local.bat
echo.
echo   Docker部署:
echo     需要: Docker Desktop
echo     启动: 运行选项2 或执行 docker-compose up -d --build
echo.
echo   K8S部署:
echo     需要: kubectl, Docker, Kubernetes集群
echo     启动: 运行选项3 或执行 .\start-k8s.ps1
echo.
echo   常用命令:
echo     make build        - 构建所有Docker镜像
echo     make deploy       - 部署到Kubernetes
echo     make status       - 查看Pod状态
echo     docker-compose up -d  - Docker快速启动
echo.
echo   默认登录:
echo     用户名: admin
echo     密码:   admin123
echo.
goto end

:end
pause
