@echo off
chcp 65001 >nul 2>&1
title Generator Platform - 一键启动工具

echo.
echo ============================================================
echo   Generator Platform - 一键启动工具 (小白专用版)
echo ============================================================
echo.
echo [INFO] 正在检查环境...
echo.

where docker >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Docker 未安装！请先安装 Docker Desktop。
    echo         下载地址: https://www.docker.com/products/docker-desktop/
    pause
    exit /b 1
)

docker info >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Docker 未运行！请先启动 Docker Desktop。
    pause
    exit /b 1
)

echo [OK] Docker 环境正常
echo.

echo ============================================================
echo   选择启动模式:
echo ============================================================
echo.
echo   [1] Docker Compose 模式 (推荐，最简单)
echo   [2] 停止所有服务
echo   [3] 查看服务状态
echo   [4] 打开浏览器访问
echo   [0] 退出
echo.
set /p choice=请输入选项 (0-4):

if "%choice%"=="1" goto deploy
if "%choice%"=="2" goto stop
if "%choice%"=="3" goto status
if "%choice%"=="4" goto open
if "%choice%"=="0" goto end

echo [ERROR] 无效选项
pause
goto end

:deploy
echo.
echo ============================================================
echo   正在构建和启动服务...
echo ============================================================
echo.
echo [1/2] 构建 Docker 镜像...
echo.

docker-compose build --no-cache
if %errorlevel% neq 0 (
    echo [ERROR] 镜像构建失败！
    pause
    exit /b 1
)

echo.
echo [2/2] 启动所有服务...
echo.

docker-compose up -d
if %errorlevel% neq 0 (
    echo [ERROR] 服务启动失败！
    pause
    exit /b 1
)

echo.
echo ============================================================
echo   ✅ 启动成功！
echo ============================================================
echo.
echo   等待服务初始化中（约10秒）...
timeout /t 10 /nobreak >nul

echo.
echo   访问地址:
echo     前端界面: http://localhost:3000
echo     API接口: http://localhost:8080/api/v1
echo.
echo   登录凭据:
echo     用户名: admin
echo     密码:   admin123
echo.
echo   提示: 首次启动会自动创建数据库和管理员账户
echo.
set /p openbrowser=是否立即打开浏览器? (y/n):
if /i "%openbrowser%"=="y" (
    start http://localhost:3000
)
goto end

:stop
echo.
echo ============================================================
echo   正在停止所有服务...
echo ============================================================
echo.

docker-compose down

echo.
echo ✅ 所有服务已停止
goto end

:status
echo.
echo ============================================================
echo   服务运行状态:
echo ============================================================
echo.

docker-compose ps

echo.
echo   访问地址:
echo     前端: http://localhost:3000
echo     API:  http://localhost:8080/api/v1
echo.
goto end

:open
echo.
echo 正在打开浏览器...

start http://localhost:3000

echo ✅ 已打开浏览器
goto end

:end
echo.
pause
