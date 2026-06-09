# Server Generator Cluster - 详细安装说明

> **微服务代码生成平台** | Go + React + Kubernetes + Docker

---

## 目录

- [一、系统要求](#一系统要求)
- [二、环境准备](#二环境准备)
  - [2.1 安装 Docker](#21-安装-docker)
  - [2.2 安装 Node.js](#22-安装-nodejs)
  - [2.3 安装 Go](#23-安装-go)
  - [2.4 安装 K3d（可选，用于 K8s 开发）](#24-安装-k3d可选用于-k8s-开发)
  - [2.5 安装 kubectl](#25-安装-kubectl)
  - [2.6 安装 PostgreSQL 和 Redis（本地开发）](#26-安装-postgresql-和-redis本地开发)
- [三、项目结构概览](#三项目结构概览)
- [四、安装方式选择](#四安装方式选择)
  - [4.1 方式一：Docker Compose 部署（推荐新手）](#41-方式一docker-compose-部署推荐新手)
  - [4.2 方式二：K3d (Kubernetes) 部署（推荐开发）](#42-方式二k3d-kubernetes-部署推荐开发)
  - [4.3 方式三：本地源码运行（适合二次开发）](#43-方式三本地源码运行适合二次开发)
- [五、详细安装步骤](#五详细安装步骤)
  - [5.1 获取项目代码](#51-获取项目代码)
  - [5.2 环境变量配置](#52-环境变量配置)
  - [5.3 构建镜像](#53-构建镜像)
  - [5.4 启动服务](#54-启动服务)
  - [5.5 验证安装](#55-验证安装)
- [六、各部署方式的完整操作指南](#六各部署方式的完整操作指南)
  - [6.1 Docker Compose 完整流程](#61-docker-compose-完整流程)
  - [6.2 K3d (K8s) 完整流程](#62-k3d-k8s-完整流程)
  - [6.3 本地源码运行完整流程](#63-本地源码运行完整流程)
- [七、服务端口说明](#七服务端口说明)
- [八、默认账号信息](#八默认账号信息)
- [九、常用运维命令](#九常用运维命令)
- [十、故障排查](#十故障排查)
- [十一、安全建议](#十一安全建议)

---

## 一、系统要求

### 最低配置

| 项目 | 最低要求 | 推荐配置 |
|------|----------|----------|
| **操作系统** | Windows 10/11, macOS 12+, Ubuntu 20.04+ | Windows 11, macOS 14+, Ubuntu 22.04+ |
| **CPU** | 4 核心以上 | 8 核心以上 |
| **内存** | 8 GB RAM | 16 GB RAM |
| **磁盘空间** | 20 GB 可用空间 | 50 GB 可用空间（含 Docker 镜像） |
| **网络** | 可访问互联网（下载依赖和镜像） | 稳定的互联网连接 |

### 软件依赖版本

| 软件 | 最低版本 | 推荐版本 | 用途 |
|------|----------|----------|------|
| **Docker Desktop** | >= 24.0 | >= 27.0 | 容器化运行 |
| **Node.js** | >= 18.x | >= 20 LTS | 前端构建与开发 |
| **npm / pnpm** | >= 9.x / 8.x | 最新版 | 包管理器 |
| **Go** | >= 1.21 | >= 1.22 | 后端微服务编译 |
| **K3d** (可选) | >= 5.7 | >= 5.7 | 本地 K8s 集群 |
| **kubectl** (可选) | >= 1.28 | >= 1.30 | K8s 命令行工具 |
| **Git** | >= 2.30 | 最新版 | 版本控制 |

---

## 二、环境准备

### 2.1 安装 Docker

#### Windows

1. 访问 [Docker Desktop 下载页](https://www.docker.com/products/docker-desktop/)
2. 下载并运行 `Docker Desktop Installer.exe`
3. 安装过程中确保勾选 **"Use WSL 2 instead of Hyper-V"**
4. 安装完成后重启电脑
5. 启动 Docker Desktop，等待状态变为绿色 "Running"
6. 验证安装：

```powershell
docker --version
docker compose version
```

#### macOS

```bash
brew install --cask docker
```

或从官网下载 `.dmg` 文件安装。

#### Linux (Ubuntu/Debian)

```bash
# 安装 Docker
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER

# 安装 Docker Compose 插件
sudo apt-get update
sudo apt-get install docker-compose-plugin

# 重新登录使组权限生效
newgrp docker
```

### 2.2 安装 Node.js

#### 方式 A：使用 nvm（推荐）

**Windows:**

1. 从 [nvm-windows releases](https://github.com/coreybutler/nvm-windows/releases) 下载安装包
2. 安装后打开新的终端：

```powershell
nvm install 20
nvm use 20
node --version   # 应显示 v20.x.x
npm --version    # 应显示 10.x.x
```

**macOS/Linux:**

```bash
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash
source ~/.bashrc
nvm install 20
nvm use 20
```

#### 方式 B：直接安装

从 [Node.js 官网](https://nodejs.org/) 下载 LTS 版本安装包。

### 2.3 安装 Go

#### Windows

1. 访问 [Go 下载页](https://go.dev/dl/)
2. 下载 `go1.21.x.windows-amd64.msi`（或更高版本）
3. 双击安装，按默认路径完成
4. 打开新终端验证：

```powershell
go version
# 输出: go version go1.21.x windows/amd64
```

#### macOS/Linux

```bash
# macOS
brew install go

# Linux
wget https://go.dev/dl/go1.21.13.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.13.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

### 2.4 安装 K3d（可选，用于 K8s 开发）

K3d 是在 Docker 中运行轻量级 Kubernetes 集群的工具。

#### Windows

```powershell
# 使用 winget 安装
winget install k3d

# 或手动下载
# 从 https://github.com/k3d-io/k3d/releases 下载最新版本
# 将 k3d.exe 放入 PATH 目录（如 C:\Windows\System32）
```

#### macOS/Linux

```bash
curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
```

验证安装：

```powershell
k3d version
```

### 2.5 安装 kubectl

#### Windows

```powershell
winget install Kubernetes.kubectl
```

或从 [Kubernetes 发布页](https://kubernetes.io/docs/tasks/tools/install-kubectl-windows/) 手动下载。

#### macOS/Linux

```bash
# macOS
brew install kubectl

# Linux
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl
sudo mv kubectl /usr/local/bin/
```

验证：

```powershell
kubectl version --client
```

### 2.6 安装 PostgreSQL 和 Redis（本地开发）

如果选择**本地源码运行**方式，需要在本机安装 PostgreSQL 和 Redis。

#### PostgreSQL (Windows)

1. 从 [PostgreSQL 官网](https://www.postgresql.org/download/windows/) 下载安装器
2. 安装时设置密码为 `123456`（与项目默认配置一致）
3. 创建数据库 `generator_platform`：

```sql
CREATE DATABASE generator_platform;
```

#### Redis (Windows)

Redis 在 Windows 上推荐通过 Docker 运行：

```powershell
docker run -d --name redis-dev -p 6379:6379 redis:7-alpine
```

或使用 Memurai（Windows 原生 Redis 替代品）。

---

## 三、项目结构概览

```
server-generator-cluster/
│
├── apps/                              # 微服务应用目录
│   ├── api-gateway/                   # API 网关 (Go, :8080)
│   │   ├── main.go                    # 入口文件
│   │   └── Dockerfile                 # Docker 构建文件
│   ├── authentication-service/        # 认证服务 (Go, :8082)
│   │   ├── main.go
│   │   └── Dockerfile
│   ├── user-service/                  # 用户服务 (Go, :8081)
│   │   ├── main.go
│   │   └── Dockerfile
│   ├── project-service/               # 项目管理服务 (Go, :8083)
│   │   ├── main.go
│   │   └── Dockerfile
│   ├── generator-service/             # 代码生成服务 (Go, :8084)
│   │   ├── main.go
│   │   └── Dockerfile
│   ├── operations-service/            # 运维服务 (Go, :8085)
│   │   ├── main.go
│   │   └── Dockerfile
│   ├── cluster-service/               # 集群管理服务 (Go, :8086)
│   │   ├── main.go
│   │   └── Dockerfile
│   └── web-admin/                     # React 前端管理界面 (:3000)
│       ├── src/                       # 源代码
│       ├── public/                    # 静态资源
│       ├── package.json               # Node.js 依赖
│       ├── vite.config.js             # Vite 配置
│       ├── Dockerfile                 # Docker 构建文件
│       ├── nginx.k8s.conf             # Nginx K8s 配置
│       ├── nginx.docker.conf          # Nginx Docker Compose 配置
│       └── docker-entrypoint.sh       # 入口脚本
│
├── libs/                              # 公共库
│   └── go-common/                     # Go 公共模块（被所有 Go 服务共享）
│
├── infra/                             # 基础设施配置
│   └── k8s/                           # Kubernetes 部署清单
│       ├── namespace.yaml             # 命名空间定义
│       ├── rbac.yaml                  # RBAC 权限控制
│       ├── postgres.yaml              # PostgreSQL 部署
│       ├── redis.yaml                 # Redis 部署
│       ├── api-gateway.yaml           # API 网关部署
│       ├── auth-service.yaml          # 认证服务部署
│       ├── user-service.yaml          # 用户服务部署
│       ├── project-service.yaml       # 项目服务部署
│       ├── generator-service.yaml     # 生成器服务部署
│       ├── operations-service.yaml    # 运维服务部署
│       ├── cluster-service.yaml       # 集群服务部署
│       ├── web-admin.yaml             # 前端部署
│       └── ingress.yaml               # Ingress 路由规则
│
├── scripts/                           # 自动化脚本
│   ├── start-local.ps1                # PowerShell 一键启动脚本
│   ├── start-local.bat                # CMD 一键启动脚本
│   └── start.bat                      # 通用启动脚本
│
├── test/                              # 测试代码
├── docs/                              # 项目文档
│
├── .env                               # 环境变量（本地，不提交 Git）
├── .env.example                       # 环境变量模板
├── .gitignore                         # Git 忽略规则
├── docker-compose.yaml                # Docker Compose 编排文件
├── Makefile                           # Make 构建命令
├── LICENSE                            # MIT 许可证
└── README.md                          # 项目说明文档
```

---

## 四、安装方式选择

本项目支持三种安装方式，根据你的需求选择：

| 特性 | Docker Compose | K3d (K8s) | 本地源码运行 |
|------|---------------|-----------|-------------|
| **难度** | ⭐ 简单 | ⭐⭐ 中等 | ⭐⭐⭐ 进阶 |
| **适用场景** | 快速体验/演示 | 开发/测试/学习 | 二次开发/调试 |
| **隔离性** | 容器级别 | Pod 级别 | 进程级别 |
| **资源占用** | 中等 | 较高 | 较低 |
| **启动速度** | 快 (~30秒) | 中等 (~1分钟) | 最快 (~10秒) |
| **热重载** | 不支持 | 不支持 | 支持 |
| **生产就绪** | 部分 | 是 | 否 |

---

## 五、详细安装步骤

### 5.1 获取项目代码

```bash
# 克隆仓库（替换为你的实际地址）
git clone <your-repository-url>
cd server-generator-cluster
```

### 5.2 环境变量配置

**重要**: 复制环境变量模板并根据实际情况修改：

```bash
# 复制根目录环境变量模板
cp .env.example .env

# 如果是前端项目，也复制前端的
cp apps/web-admin/.env.example apps/web-admin/.env 2>/dev/null || true
```

编辑根目录的 `.env` 文件：

```env
# ============================================
# 数据库配置
# ============================================
DB_HOST=localhost            # Docker Compose 模式改为 postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=123456           # 生产环境请修改！
DB_NAME=generator_platform

# ============================================
# Redis 配置
# ============================================
REDIS_HOST=localhost         # Docker Compose 模式改为 redis
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# ============================================
# JWT 配置
# ============================================
JWT_SECRET=your-super-secret-key-change-in-production  # 生产环境必须修改！
JWT_EXPIRE=24h

# ============================================
# 服务端口配置
# ============================================
API_GATEWAY_PORT=8080
USER_SERVICE_PORT=8081
AUTH_SERVICE_PORT=8082
PROJECT_SERVICE_PORT=8083
GENERATOR_SERVICE_PORT=8084
OPERATIONS_SERVICE_PORT=8085
CLUSTER_SERVICE_PORT=8086
WEB_ADMIN_PORT=3000

# ============================================
# 服务间通信 URL（Docker 内部通信用服务名）
# ============================================
AUTH_SERVICE_URL=http://localhost:8082
USER_SERVICE_URL=http://localhost:8081
PROJECT_SERVICE_URL=http://localhost:8083
GENERATOR_SERVICE_URL=http://localhost:8084
OPERATIONS_SERVICE_URL=http://localhost:8085
CLUSTER_SERVICE_URL=http://localhost:8086

# ============================================
# Docker 配置
# ============================================
IMAGE_PREFIX=generator-platform
```

前端环境变量 (`apps/web-admin/.env`)：

```env
VITE_API_URL=/api/v1        # Vite 开发代理模式（转发到 localhost:8080）
# 生产环境部署到 K8s/Docker 时保持此值不变
```

### 5.3 构建镜像

#### 构建 Go 后端服务镜像（全部 7 个）

```bash
# 方式一：使用 Makefile（推荐）
make build

# 方式二：手动逐个构建
docker build -t generator-platform/api-gateway:latest -f apps/api-gateway/Dockerfile .
docker build -t generator-platform/auth-service:latest -f apps/authentication-service/Dockerfile .
docker build -t generator-platform/user-service:latest -f apps/user-service/Dockerfile .
docker build -t generator-platform/project-service:latest -f apps/project-service/Dockerfile .
docker build -t generator-platform/generator-service:latest -f apps/generator-service/Dockerfile .
docker build -t generator-platform/operations-service:latest -f apps/operations-service/Dockerfile .
docker build -t generator-platform/cluster-service:latest -f apps/cluster-service/Dockerfile .

# 构建前端镜像
docker build -t generator-platform/web-admin:latest \
  --build-arg DEPLOY_MODE=k8s \
  -f apps/web-admin/Dockerfile .
```

#### 构建前端（独立开发时需要）

```bash
cd apps/web-admin
npm install
npm run build
```

### 5.4 启动服务

根据选择的部署方式，跳转到对应的章节：
- [Docker Compose 部署 → 6.1](#61-docker-compose-完整流程)
- [K3d (K8s) 部署 → 6.2](#62-k3d-k8s-完整流程)
- [本地源码运行 → 6.3](#63-本地源码运行完整流程)

### 5.5 验证安装

无论哪种部署方式，都可以通过以下方式验证：

```bash
# 1. 检查 Web Admin 前端是否可访问
# 浏览器打开 http://localhost:3000（或对应端口）

# 2. 检查 API Gateway 是否正常
curl http://localhost:8080/health

# 3. 尝试登录接口
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

---

## 六、各部署方式的完整操作指南

### 6.1 Docker Compose 完整流程

#### 步骤 1：构建所有镜像

```bash
make build
```

或者只构建你需要的部分：

```bash
# 仅构建后端服务
docker build -t generator-platform/api-gateway:latest -f apps/api-gateway/Dockerfile .
docker build -t generator-platform/auth-service:latest -f apps/authentication-service/Dockerfile .
# ... 其他服务同理
```

#### 步骤 2：启动所有服务

```bash
# 一键启动所有服务（后台运行）
docker compose up -d

# 查看日志（前台运行，可实时查看输出）
docker compose up

# 只启动特定服务
docker compose up -d postgres redis      # 先启动基础设施
docker compose up -d auth-service user-service  # 再启动核心服务
docker compose up -d api-gateway web-admin      # 最后启动网关和前端
```

#### 步骤 3：验证服务状态

```bash
# 查看所有容器状态
docker compose ps

# 查看某个服务的日志
docker compose logs -f api-gateway
docker compose logs -f web-admin

# 查看所有服务日志
docker compose logs --tail=50
```

#### 步骤 4：访问服务

| 服务 | 地址 | 说明 |
|------|------|------|
| **Web Admin 前端** | http://localhost:3000 | 主界面 |
| **API Gateway** | http://localhost:8080 | 后端统一入口 |
| **认证服务** | http://localhost:8082 | 直接访问（调试用） |

#### 步骤 5：停止和清理

```bash
# 停止所有服务（保留数据）
docker compose down

# 停止并删除 Volume 数据（完全清理）
docker compose down -v

# 重启单个服务
docker compose restart api-gateway

# 重新构建并启动
docker compose up -d --build
```

#### Docker Compose 端口映射说明

`docker-compose.yaml` 中定义了以下端口映射：

| 外部端口 | 内部端口 | 服务 | 用途 |
|----------|----------|------|------|
| 3000 | 3000 | web-admin | 前端界面 |
| 8080 | 8080 | api-gateway | API 网关 |
| 5432 | 5432 | postgres | 数据库（仅开发环境暴露） |
| 6379 | 6379 | redis | 缓存（仅开发环境暴露） |
| 8081 | 8081 | user-service | 用户服务 |
| 8082 | 8082 | auth-service | 认证服务 |
| 8083 | 8083 | project-service | 项目服务 |
| 8084 | 8084 | generator-service | 生成器服务 |
| 8085 | 8085 | operations-service | 运维服务 |
| 8086 | 8086 | cluster-service | 集群服务 |

可通过环境变量自定义端口：

```bash
# 自定义前端端口
FRONTEND_PORT=8080 docker compose up -d web-admin

# 自定义 API 端口
API_PORT=9090 docker compose up -d api-gateway
```

---

### 6.2 K3d (K8s) 完整流程

#### 步骤 1：创建 K3d 集群

```bash
# 创建集群（包含 1 个 server 节点 + 2 个 agent 节点）
k3d cluster create gen-platform-test \
  --api-port 6443 \
  --port "13000:80@loadbalancer" \       # Web Admin 前端入口
  --port "18080:8080@loadbalancer" \     # API Gateway 入口
  --agents 2

# 验证集群状态
k3d cluster list
kubectl get nodes
```

#### 步骤 2：获取 Kubeconfig

```bash
# 合并 kubeconfig 到本地
k3d kubeconfig merge gen-platform-test --kubeconfig-switch-context

# 或者导出到指定文件
k3d kubeconfig merge gen-platform-test -o ./k3d-kubeconfig.yaml
export KUBECONFIG=./k3d-kubeconfig.yaml  # Linux/macOS
$env:KUBECONFIG="./k3d-kubeconfig.yaml"  # PowerShell
```

#### 步骤 3：导入镜像到 K3d 集群

K3d 集群内部无法直接拉取本地 Docker 镜像，需要先导入：

```bash
# 导入所有项目镜像
k3d image import generator-platform/api-gateway:latest -c gen-platform-test
k3d image import generator-platform/auth-service:latest -c gen-platform-test
k3d image import generator-platform/user-service:latest -c gen-platform-test
k3d image import generator-platform/project-service:latest -c gen-platform-test
k3d image import generator-platform/generator-service:latest -c gen-platform-test
k3d image import generator-platform/operations-service:latest -c gen-platform-test
k3d image import generator-platform/cluster-service:latest -c gen-platform-test
k3d image import generator-platform/web-admin:latest -c gen-platform-test

# 导入基础设施数像
k3d image import postgres:15-alpine -c gen-platform-test
k3d image import redis:7-alpine -c gen-platform-test
k3d image import nginx:alpine -c gen-platform-test
```

#### 步骤 4：部署应用到 K8s

```bash
# 方式一：使用 Makefile 一键部署
make apply

# 方式二：手动逐步部署（推荐顺序）
kubectl apply -f infra/k8s/namespace.yaml
kubectl apply -f infra/k8s/rbac.yaml
kubectl apply -f infra/k8s/postgres.yaml
kubectl apply -f infra/k8s/redis.yaml
kubectl apply -f infra/k8s/auth-service.yaml
kubectl apply -f infra/k8s/user-service.yaml
kubectl apply -f infra/k8s/project-service.yaml
kubectl apply -f infra/k8s/generator-service.yaml
kubectl apply -f infra/k8s/operations-service.yaml
kubectl apply -f infra/k8s/cluster-service.yaml
kubectl apply -f infra/k8s/api-gateway.yaml
kubectl apply -f infra/k8s/web-admin.yaml
kubectl apply -f infra/k8s/ingress.yaml
```

#### 步骤 5：等待 Pod 就绪

```bash
# 观察所有 Pod 启动过程
watch -n 3 'kubectl get pods -n generator-platform'

# PowerShell 替代方案
while ($true) { Clear-Host; kubectl get pods -n generator-platform; Start-Sleep 3 }

# 等待所有 Pod 就绪
kubectl wait --for=condition=ready pod --all -n generator-platform --timeout=180s
```

#### 步骤 6：端口转发访问服务

由于 K3d 的 Traefik Ingress 与当前 Ingress 配置可能不兼容，推荐使用 port-forward：

```bash
# 终端 1：转发 Web Admin
kubectl port-forward svc/web-admin 3000:3000 -n generator-platform

# 终端 2：转发 API Gateway
kubectl port-forward svc/api-gateway 8080:8080 -n generator-platform
```

然后访问：
- **Web Admin**: http://localhost:3000
- **API Gateway**: http://localhost:8080

#### 步骤 7：常用管理命令

```bash
# 查看所有资源
make get
# 或
kubectl get all -n generator-platform

# 查看服务日志
make logs SERVICE=web-admin
# 或
kubectl logs -f -n generator-platform -l app=web-admin --tail=100

# 重启某个服务
make restart SERVICE=api-gateway
# 或
kubectl rollout restart deployment/api-gateway -n generator-platform

# 扩缩容
make scale SERVICE=user-service REPLICAS=3

# 进入容器调试
make exec SERVICE=api-gateway
# 或
kubectl exec -it -n generator-platform -l app=api-gateway -- /bin/sh

# 删除所有资源
make delete
# 或
kubectl delete namespace generator-platform

# 删除整个集群
k3d cluster delete gen-platform-test
```

---

### 6.3 本地源码运行完整流程

此方式适合对项目进行二次开发和深度调试。每个服务都在本地进程中运行，支持热重载。

#### 步骤 1：准备基础设施

确保 PostgreSQL 和 Redis 正在运行：

```bash
# 使用 Docker 快速启动（推荐）
docker run -d --name dev-postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=123456 \
  -e POSTGRES_DB=generator_platform \
  -p 5432:5432 \
  postgres:15-alpine

docker run -d --name dev-redis \
  -p 6379:6379 \
  redis:7-alpine
```

#### 步骤 2：配置环境变量

确保根目录 `.env` 文件存在且配置正确（参考 [5.2 环境变量配置](#52-环境变量配置)）。

#### 步骤 3：一键启动（推荐）

**PowerShell (Windows):**

```powershell
.\scripts\start-local.ps1
```

**CMD (Windows):**

```cmd
scripts\start-local.bat
```

该脚本会自动按顺序启动所有 8 个服务，并在新窗口中运行。

#### 步骤 4：手动逐个启动（进阶用户）

如果你想更精细地控制每个服务，可以手动在单独的终端中启动：

```bash
# 终端 1: 认证服务
cd apps/authentication-service
go run main.go

# 终端 2: 用户服务
cd apps/user-service
go run main.go

# 终端 3: 项目管理服务
cd apps/project-service
go run main.go

# 终端 4: 代码生成服务
cd apps/generator-service
go run main.go

# 终端 5: 运维服务
cd apps/operations-service
go run main.go

# 终端 6: 集群管理服务
cd apps/cluster-service
go run main.go

# 终端 7: API 网关（必须在其他服务之后启动）
cd apps/api-gateway
go run main.go

# 终端 8: React 前端
cd apps/web-admin
npm install        # 首次运行需要安装依赖
npm run dev
```

#### 步骤 5：前端开发模式说明

前端使用 Vite 开发服务器，具有以下特性：

- **HMR (热模块替换)**: 修改代码后浏览器自动刷新
- **API 代理**: `/api` 请求自动转发到 `http://localhost:8080`
- **端口**: 默认运行在 `http://localhost:3000`

Vite 代理配置位于 `apps/web-admin/vite.config.js`：

```javascript
server: {
  port: 3000,
  proxy: {
    '/api': {
      target: 'http://localhost:8080',  // 转发到 API 网关
      changeOrigin: true,
    },
  },
},
```

#### 步骤 6：停止所有服务

- 按 `Ctrl+C` 停止当前终端的服务
- 关闭所有已打开的服务终端窗口
- 清理 Docker 基础设施（如需要）：

```bash
docker stop dev-postgres dev-redis
docker rm dev-postgres dev-redis
```

---

## 七、服务端口说明

### 服务架构图

```
                    ┌─────────────────┐
                    │   Web Admin     │  :3000
                    │  (React 前端)    │
                    └────────┬────────┘
                             │ /api/*
                             ▼
                    ┌─────────────────┐
                    │  API Gateway    │  :8080
                    │  (路由+鉴权)     │
                    └────────┬────────┘
            ┌───────────────┼───────────────┐
            ▼               ▼               ▼
    ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
    │ Auth Service │ │ User Service │ │Project Svc   │
    │    :8082     │ │    :8081     │ │    :8083     │
    └──────────────┘ └──────────────┘ └──────────────┘
            │               │               │
            ▼               ▼               ▼
    ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
    │Generator Svc │ │Operations Svc│ │Cluster Svc   │
    │    :8084     │ │    :8085     │ │    :8086     │
    └──────────────┘ └──────────────┘ └──────────────┘
            │               │
            ▼               ▼
    ┌──────────────┐ ┌──────────────┐
    │  PostgreSQL  │ │    Redis     │
    │    :5432     │ │    :6379     │
    └──────────────┘ └──────────────┘
```

### 端口速查表

| 端口 | 服务 | 协议 | 说明 |
|------|------|------|------|
| **3000** | Web Admin | HTTP | React 前端管理界面 |
| **8080** | API Gateway | HTTP | 统一 API 入口，路由分发 |
| **8081** | User Service | HTTP | 用户 CRUD 管理 |
| **8082** | Auth Service | HTTP | 登录注册、JWT 签发 |
| **8083** | Project Service | HTTP | 项目管理与代码生成触发 |
| **8084** | Generator Service | HTTP | 代码模板生成引擎 |
| **8085** | Operations Service | HTTP | 系统监控、健康检查、日志 |
| **8086** | Cluster Service | HTTP | K8s/Docker 集群管理 |
| **5432** | PostgreSQL | TCP | 关系型数据库 |
| **6379** | Redis | TCP | 缓存 / Session 存储 |

---

## 八、默认账号信息

| 角色 | 用户名 | 密码 | 说明 |
|------|--------|------|------|
| **管理员** | admin | admin123 | 拥有所有权限 |

> ⚠️ **首次登录后请立即修改密码！**

---

## 九、常用运维命令

### Docker Compose 相关

```bash
# 查看服务状态
docker compose ps

# 查看实时日志
docker compose logs -f <service-name>

# 重启服务
docker compose restart <service-name>

# 重新构建并重启
docker compose up -d --build <service-name>

# 进入容器
docker compose exec <service-name> /bin/sh

# 停止所有服务
docker compose down

# 停止并清除数据
docker compose down -v
```

### Kubernetes (K3d) 相关

```bash
# 设置命名空间上下文（简化后续命令）
alias k='kubectl -n generator-platform'
# PowerShell: function k { kubectl -n generator-platform @args }

# 查看 Pod 状态
k get pods -o wide

# 查看 Service
k get svc

# 查看日志
k logs -f <pod-name>

# 进入容器
k exec -it <pod-name> -- /bin/sh

# 重启部署
k rollout restart deployment/<name>

# 扩缩容
k scale deployment/<name> --replicas=3

# 查看事件（排查问题）
k get events --sort-by='.lastTimestamp'

# 查看资源描述
k describe pod <pod-name>
```

### 本地开发相关

```bash
# Go 服务：修改代码后自动重新编译运行（go run 天然支持）

# 前端：Vite HMR 自动刷新，无需手动操作

# 格式化 Go 代码
gofmt -w apps/*/main.go

# 检查前端代码规范
cd apps/web-admin && npm run lint
```

---

## 十、故障排查

### 常见问题及解决方案

#### Q1: Docker Compose 启动报错 "port already in use"

**原因**: 端口被其他进程占用

**解决**:
```bash
# 查找占用端口的进程
netstat -ano | findstr :3000   # Windows
lsof -i :3000                   # macOS/Linux

# 结束进程（将 PID 替换为实际的进程 ID）
taskkill /PID <PID> /F          # Windows
kill -9 <PID>                   # macOS/Linux

# 或修改 docker-compose.yaml 中的端口映射
```

#### Q2: K3d Pod 状态为 ImagePullBackOff

**原因**: K3d 集群内无法拉取镜像

**解决**:
```bash
# 确保镜像已导入 K3d
k3d image import <image-name>:<tag> -c gen-platform-test

# 重启失败的 Pod
kubectl delete pod <pod-name> -n generator-platform
# 或
kubectl rollout restart deployment/<service> -n generator-platform
```

#### Q3: 前端页面空白 / API 请求 404

**原因**: 前端无法连接到后端 API

**解决**:
1. 确认 API Gateway 已启动：`curl http://localhost:8080/health`
2. 检查 Vite 代理配置（`apps/web-admin/vite.config.js`）
3. 检查前端 `.env` 中的 `VITE_API_URL` 设置
4. 如果是 K8s/Docker 部署，确认 Ingress 或 port-forward 已正确配置

#### Q4: PostgreSQL 连接失败

**错误信息**: `connection refused`, `password authentication failed`

**解决**:
1. 确认 PostgreSQL 容器正在运行：`docker ps | grep postgres`
2. 检查 `.env` 中的 `DB_HOST`, `DB_PASSWORD` 是否匹配
3. 确认数据库已创建：进入 PostgreSQL 容器执行 `psql -U postgres -c "\l"`
4. Docker Compose 模式下 `DB_HOST` 应为 `postgres`（服务名），而非 `localhost`

#### Q5: Redis 连接失败

**解决**:
```bash
# 确认 Redis 容器运行中
docker ps | grep redis

# 测试 Redis 连接
docker exec -it <redis-container> redis-cli ping
# 应返回 PONG
```

#### Q6: Go 编译报错 "module not found"

**解决**:
```bash
# 进入对应服务目录
cd apps/<service-name>

# 整理依赖
go mod tidy

# 重新编译
go build -o <service-name> .
```

#### Q7: npm install 报错网络超时

**解决**（国内用户）:

```bash
# 使用淘宝镜像源
npm config set registry https://registry.npmmirror.com

# 或使用 pnpm
npm install -g pnpm
pnpm install
```

#### Q8: K3d 集群创建失败

**常见原因及解决**:

```bash
# 端口冲突
k3d cluster create my-cluster --api-port 6444  # 更改 API 端口

# 内存不足
# 关闭不必要的应用程序，或减少 agent 节点数
k3d cluster create my-cluster --agents 0  # 单节点模式

# WSL2 问题（Windows）
wsl --update
wsl --shutdown
# 然后重启 Docker Desktop
```

#### Q9: 前端 npm run build 报错内存不足

**解决**:
```bash
# 增加 Node.js 内存限制
set NODE_OPTIONS=--max-old-space-size=4096   # Windows CMD
$env:NODE_OPTIONS="--max-old-space-size=4096" # PowerShell
export NODE_OPTIONS=--max-old-space-size=4096 # Linux/macOS

npm run build
```

#### Q10: JWT Token 验证失败

**原因**: 各服务的 `JWT_SECRET` 不一致

**解决**:
- 确保 `.env` 文件中所有服务使用相同的 `JWT_SECRET`
- Docker Compose 模式下检查 `docker-compose.yaml` 中各服务的 environment 配置
- K8s 模式下考虑使用 Secret 统一管理

---

## 十一、安全建议

> ⚠️ 以下安全措施在生产环境中**必须执行**！

### 必须修改的安全配置

| 配置项 | 当前值 | 操作 |
|--------|--------|------|
| `DB_PASSWORD` | `123456` | 改为强密码（至少 16 位，含大小写字母+数字+特殊字符） |
| `JWT_SECRET` | `generator-platform-secret-key-2024` | 改为随机生成的长字符串（至少 32 字符） |
| `REDIS_PASSWORD` | 空 | 设置 Redis 密码 |
| 默认管理员密码 | `admin123` | 登录后立即修改 |

### 生成安全的 JWT_SECRET

```bash
# 方法一：使用 OpenSSL
openssl rand -base64 32

# 方法二：使用 Python
python -c "import secrets; print(secrets.token_urlsafe(32))"

# 方法三：使用 Node.js
node -e "console.log(require('crypto').randomBytes(24).toString('base64'))"
```

### 生产环境额外建议

1. **HTTPS**: 使用 Let's Encrypt 或自有证书配置 TLS
2. **防火墙**: 只开放必要端口（80/443），其余服务不对外暴露
3. **数据库**: 不要将 PostgreSQL 和 Redis 的端口暴露到公网
4. **CORS**: API Gateway 严格限制允许的 Origin
5. **Rate Limiting**: API Gateway 添加限流策略
6. **日志审计**: 敏感操作记录审计日志
7. **定期备份**: PostgreSQL 定期自动备份
8. **密钥轮换**: JWT Secret 定期更换
9. **镜像安全**: 使用私有镜像仓库，扫描漏洞
10. **RBAC**: K8s 集群启用 RBAC，遵循最小权限原则

---

## 附录：快速开始命令汇总

### 第一次安装（复制粘贴即可）

```bash
# ===== 1. 克隆项目 =====
git clone <your-repo-url>
cd server-generator-cluster

# ===== 2. 配置环境变量 =====
cp .env.example .env
# 编辑 .env 修改密码等敏感信息

# ===== 3. Docker Compose 一键启动 =====
make build              # 构建镜像
docker compose up -d    # 启动服务
sleep 30                # 等待初始化
docker compose ps       # 检查状态

# ===== 4. 访问 =====
# 浏览器打开 http://localhost:3000
# 用户名: admin  密码: admin123
```

### K3d 快速启动

```bash
# ===== 1. 创建集群 =====
k3d cluster create gen-platform-test \
  --api-port 6443 \
  --port "13000:80@loadbalancer" \
  --port "18080:8080@loadbalancer"

# ===== 2. 获取 Kubeconfig =====
k3d kubeconfig merge gen-platform-test --kubeconfig-switch-context

# ===== 3. 构建并导入镜像 =====
make build
for img in $(docker images --format "{{.Repository}}:{{.Tag}}" | grep generator-platform); do
  k3d image import $img -c gen-platform-test
done

# ===== 4. 部署 =====
make apply

# ===== 5. 端口转发 =====
kubectl port-forward svc/web-admin 3000:3000 -n generator-platform &
kubectl port-forward svc/api-gateway 8080:8080 -n generator-platform &

# ===== 6. 访问 =====
# http://localhost:3000
```

### 本地开发快速启动

```bash
# ===== 1. 启动基础设施数据库 =====
docker run -d --name dev-postgres -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=123456 -e POSTGRES_DB=generator_platform -p 5432:5432 postgres:15-alpine
docker run -d --name dev-redis -p 6379:6379 redis:7-alpine

# ===== 2. 一键启动所有服务 =====
./scripts/start-local.ps1    # PowerShell
# scripts\start-local.bat     # CMD

# ===== 3. 访问 =====
# http://localhost:3000
```

---

> **文档版本**: v1.0
> **最后更新**: 2026-05-20
> **适用版本**: Server Generator Cluster 全版本
> 如有问题，请提交 Issue 或联系维护者。
