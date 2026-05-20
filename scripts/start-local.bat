@echo off
echo ========================================
echo 微服务构建平台 - 本地启动脚本
echo ========================================
echo.

echo [1/8] 启动 authentication-service...
start "auth-service" cmd /k "cd apps/authentication-service && go run main.go"
timeout /t 3 /nobreak > nul

echo [2/8] 启动 user-service...
start "user-service" cmd /k "cd apps/user-service && go run main.go"
timeout /t 2 /nobreak > nul

echo [3/8] 启动 project-service...
start "project-service" cmd /k "cd apps/project-service && go run main.go"
timeout /t 2 /nobreak > nul

echo [4/8] 启动 generator-service...
start "generator-service" cmd /k "cd apps/generator-service && go run main.go"
timeout /t 2 /nobreak > nul

echo [5/8] 启动 operations-service...
start "operations-service" cmd /k "cd apps/operations-service && go run main.go"
timeout /t 2 /nobreak > nul

echo [6/8] 启动 cluster-service...
start "cluster-service" cmd /k "cd apps/cluster-service && go run main.go"
timeout /t 2 /nobreak > nul

echo [7/8] 启动 api-gateway...
start "api-gateway" cmd /k "cd apps/api-gateway && go run main.go"
timeout /t 2 /nobreak > nul

echo [8/8] 启动 web-admin...
start "web-admin" cmd /k "cd apps/web-admin && npm run dev"

echo.
echo ========================================
echo 所有服务启动完成！
echo ========================================
echo.
echo 访问地址：
echo - 前端界面: http://localhost:3000
echo - API网关: http://localhost:8080
echo.
echo 默认登录：
echo - 用户名: admin
echo - 密码: admin123
echo.
pause